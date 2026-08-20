package chunking

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/prismaid/review/config"
)

// Binding associates an extraction prompt with its source document and chunk.
type Binding struct {
	DocumentIndex int
	Filename      string
	ChunkIndex    int
	ChunkCount    int
}

// ExpectedGroup describes one provider/model response artifact expected for a
// document. It lets Merge report a document for which the model returned no
// responses at all.
type ExpectedGroup struct {
	DocumentIndex  int
	Filename       string
	Provider       string
	Model          string
	SequenceNumber int
	ChunkCount     int
}

// Conflict records a non-fatal disagreement between chunk values and the
// deterministic rule used to resolve it.
type Conflict struct {
	DocumentIndex int    `json:"document_index"`
	Filename      string `json:"filename"`
	Field         string `json:"field"`
	Resolution    string `json:"resolution"`
}

// Coercion records a non-fatal normalization of an LLM-produced value.
type Coercion struct {
	DocumentIndex int    `json:"document_index"`
	Filename      string `json:"filename"`
	ChunkIndex    int    `json:"chunk_index"`
	Field         string `json:"field"`
	SourceType    string `json:"source_type"`
	Action        string `json:"action"`
}

// Failure records one document/provider/model artifact that could not be
// merged. Other documents continue and remain available in the result file.
type Failure struct {
	DocumentIndex  int    `json:"document_index"`
	Filename       string `json:"filename"`
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	SequenceNumber int    `json:"sequence_number"`
	Message        string `json:"message"`
}

// MergeReport contains the complete non-fatal merge audit trail.
type MergeReport struct {
	Conflicts []Conflict `json:"conflicts"`
	Coercions []Coercion `json:"coercions"`
	Failures  []Failure  `json:"failures"`
}

type groupedResponseKey struct {
	DocumentIndex  int
	Provider       string
	Model          string
	SequenceNumber int
}

type groupedResponse struct {
	binding  Binding
	response definitions.Response
}

type fieldValue struct {
	value   interface{}
	missing bool
}

type normalizationNote struct {
	chunkIndex int
	sourceType string
	action     string
}

// Merge combines chunk-level responses into document-level responses. Invalid
// values from an LLM are normalized or dropped according to the configured
// rule. A failed document group is reported and skipped; only malformed global
// extraction output can make Merge itself fail.
func Merge(results string, bindings map[string]Binding, cfg config.ChunkingConfig) (string, MergeReport, error) {
	return MergeWithExpected(results, bindings, nil, cfg)
}

// MergeWithExpected additionally records artifacts that were expected but for
// which the LLM returned no response.
func MergeWithExpected(results string, bindings map[string]Binding, expected []ExpectedGroup, cfg config.ChunkingConfig) (string, MergeReport, error) {
	var output definitions.Output
	if err := json.Unmarshal([]byte(results), &output); err != nil {
		return "", MergeReport{}, fmt.Errorf("parse chunk extraction output: %w", err)
	}

	report := MergeReport{}
	groups := make(map[groupedResponseKey][]groupedResponse)
	for _, response := range output.Responses {
		binding, ok := bindings[response.SequenceID]
		if !ok {
			report.Failures = append(report.Failures, Failure{DocumentIndex: -1, Message: fmt.Sprintf("missing source-document binding for response sequence %q", response.SequenceID)})
			continue
		}
		key := groupedResponseKey{binding.DocumentIndex, response.Provider, response.Model, response.SequenceNumber}
		groups[key] = append(groups[key], groupedResponse{binding: binding, response: response})
	}

	expectedByKey := make(map[groupedResponseKey]ExpectedGroup, len(expected))
	keys := make([]groupedResponseKey, 0, len(expected))
	for _, item := range expected {
		key := groupedResponseKey{item.DocumentIndex, item.Provider, item.Model, item.SequenceNumber}
		if _, exists := expectedByKey[key]; !exists {
			expectedByKey[key] = item
			keys = append(keys, key)
		}
	}
	// Keep Merge useful to direct callers that do not provide an expected plan.
	if len(keys) == 0 {
		for key := range groups {
			keys = append(keys, key)
		}
	}
	sortGroupKeys(keys)

	merged := definitions.Output{Metadata: output.Metadata}
	for _, key := range keys {
		group := groups[key]
		expectedGroup, hasExpected := expectedByKey[key]
		if !hasExpected && len(group) > 0 {
			expectedGroup = ExpectedGroup{DocumentIndex: key.DocumentIndex, Filename: group[0].binding.Filename, Provider: key.Provider, Model: key.Model, SequenceNumber: key.SequenceNumber, ChunkCount: group[0].binding.ChunkCount}
		}
		sort.Slice(group, func(i, j int) bool { return group[i].binding.ChunkIndex < group[j].binding.ChunkIndex })
		if err := validateGroup(group, expectedGroup); err != nil {
			report.Failures = append(report.Failures, failureFor(expectedGroup, err))
			continue
		}
		modelResponse, conflicts, coercions, err := mergeGroup(group, cfg)
		if err != nil {
			report.Failures = append(report.Failures, failureFor(expectedGroup, err))
			continue
		}
		report.Conflicts = append(report.Conflicts, conflicts...)
		report.Coercions = append(report.Coercions, coercions...)
		merged.Responses = append(merged.Responses, definitions.Response{Provider: key.Provider, Model: key.Model, SequenceID: strconv.Itoa(key.DocumentIndex + 1), SequenceNumber: key.SequenceNumber, ModelResponses: []string{modelResponse}})
	}

	encoded, err := json.Marshal(merged)
	if err != nil {
		return "", report, fmt.Errorf("encode merged chunk output: %w", err)
	}
	return string(encoded), report, nil
}

func sortGroupKeys(keys []groupedResponseKey) {
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].DocumentIndex != keys[j].DocumentIndex {
			return keys[i].DocumentIndex < keys[j].DocumentIndex
		}
		if keys[i].Provider != keys[j].Provider {
			return keys[i].Provider < keys[j].Provider
		}
		if keys[i].Model != keys[j].Model {
			return keys[i].Model < keys[j].Model
		}
		return keys[i].SequenceNumber < keys[j].SequenceNumber
	})
}

func failureFor(group ExpectedGroup, err error) Failure {
	return Failure{DocumentIndex: group.DocumentIndex, Filename: group.Filename, Provider: group.Provider, Model: group.Model, SequenceNumber: group.SequenceNumber, Message: err.Error()}
}

func validateGroup(responses []groupedResponse, expected ExpectedGroup) error {
	if len(responses) != expected.ChunkCount {
		return fmt.Errorf("returned %d of %d expected chunk responses", len(responses), expected.ChunkCount)
	}
	for index, item := range responses {
		if item.binding.ChunkIndex != index {
			return fmt.Errorf("missing chunk response %d", index+1)
		}
		if item.response.Error != nil {
			return fmt.Errorf("chunk %d failed: %s", item.binding.ChunkIndex+1, item.response.Error.Message)
		}
		if len(item.response.ModelResponses) == 0 {
			return fmt.Errorf("chunk %d returned no model response", item.binding.ChunkIndex+1)
		}
	}
	return nil
}

func mergeGroup(responses []groupedResponse, cfg config.ChunkingConfig) (string, []Conflict, []Coercion, error) {
	if responses[0].response.SequenceNumber != 1 {
		return mergeAuxiliary(responses), nil, nil, nil
	}
	objects := make([]map[string]interface{}, 0, len(responses))
	for _, item := range responses {
		var object map[string]interface{}
		if err := json.Unmarshal([]byte(item.response.ModelResponses[0]), &object); err != nil {
			return "", nil, nil, fmt.Errorf("chunk %d returned invalid JSON: %w", item.binding.ChunkIndex+1, err)
		}
		objects = append(objects, object)
	}
	fields := make([]string, 0, len(cfg.Merge))
	for field := range cfg.Merge {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	merged := make(map[string]interface{}, len(fields))
	conflicts := make([]Conflict, 0)
	coercions := make([]Coercion, 0)
	for _, field := range fields {
		values := make([]fieldValue, len(objects))
		for index, object := range objects {
			value, ok := object[field]
			values[index] = fieldValue{value: value, missing: !ok}
		}
		value, conflict, resolution, notes, err := mergeFieldValues(values, cfg.Merge[field])
		if err != nil {
			return "", nil, nil, fmt.Errorf("field %q: %w", field, err)
		}
		merged[field] = value
		if conflict {
			conflicts = append(conflicts, Conflict{DocumentIndex: responses[0].binding.DocumentIndex, Filename: responses[0].binding.Filename, Field: field, Resolution: resolution})
		}
		for _, note := range notes {
			coercions = append(coercions, Coercion{DocumentIndex: responses[0].binding.DocumentIndex, Filename: responses[0].binding.Filename, ChunkIndex: note.chunkIndex, Field: field, SourceType: note.sourceType, Action: note.action})
		}
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return "", nil, nil, fmt.Errorf("encode merged response: %w", err)
	}
	return string(encoded), conflicts, coercions, nil
}

func mergeAuxiliary(responses []groupedResponse) string {
	unique, seen := make([]string, 0, len(responses)), make(map[string]struct{}, len(responses))
	for _, item := range responses {
		value, key := item.response.ModelResponses[0], normalizedText(item.response.ModelResponses[0])
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, value)
	}
	return strings.Join(unique, "\n\n")
}

func mergeField(field string, values []interface{}, rule config.MergeRule) (interface{}, bool, error) {
	_ = field
	fieldValues := make([]fieldValue, len(values))
	for index, value := range values {
		fieldValues[index] = fieldValue{value: value}
	}
	value, conflict, _, _, err := mergeFieldValues(fieldValues, rule)
	return value, conflict, err
}

func mergeFieldValues(values []fieldValue, rule config.MergeRule) (interface{}, bool, string, []normalizationNote, error) {
	conflict := distinctFieldValueCount(values) > 1
	switch rule.Rule {
	case "union":
		value, notes := mergeUnion(values, rule.Sentinels)
		return value, conflict, "union and sentinel filtering", notes, nil
	case "ordinal":
		value, notes := mergeOrdinal(values, rule.Order)
		return value, conflict, "strongest configured ordinal", notes, nil
	case "categorical":
		value, notes := mergeCategorical(values, rule.Defaults, rule.TieBreak)
		return value, conflict, "non-default majority; first on tie", notes, nil
	case "unique_text":
		value, notes := mergeUniqueText(values, rule.Separator, rule.MaxLength)
		return value, conflict, "unique fragments concatenated", notes, nil
	case "numeric":
		value, notes := mergeNumeric(values, rule.Operation)
		return value, conflict, rule.Operation + " numeric value", notes, nil
	case "metadata":
		value, notes, err := mergeMetadata(values, rule.OnMismatch)
		return value, conflict, "first non-null metadata value", notes, err
	default:
		return nil, false, "", nil, fmt.Errorf("unsupported configured merge rule %q", rule.Rule)
	}
}

func mergeUnion(values []fieldValue, sentinels []string) ([]interface{}, []normalizationNote) {
	items, notes, seen := make([]interface{}, 0), make([]normalizationNote, 0), make(map[string]struct{})
	for index, value := range values {
		strings, current := stringsFor(value, index, true)
		notes = append(notes, current...)
		for _, item := range strings {
			if _, ok := seen[item]; !ok {
				seen[item] = struct{}{}
				items = append(items, item)
			}
		}
	}
	hasReal := false
	for _, item := range items {
		if !containsFold(sentinels, item.(string)) {
			hasReal = true
			break
		}
	}
	if hasReal {
		filtered := items[:0]
		for _, item := range items {
			if !containsFold(sentinels, item.(string)) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	return items, notes
}

func mergeOrdinal(values []fieldValue, order []string) (interface{}, []normalizationNote) {
	ranks := make(map[string]int, len(order))
	for index, value := range order {
		ranks[value] = index
	}
	best := ""
	if len(order) > 0 {
		best = order[0]
	}
	bestRank, allBoolean, seen := -1, true, false
	notes := make([]normalizationNote, 0)
	for index, value := range values {
		candidates, candidateNotes, booleans := ordinalCandidates(value, index)
		notes = append(notes, candidateNotes...)
		for candidateIndex, candidate := range candidates {
			rank, ok := ranks[candidate]
			if !ok {
				notes = append(notes, normalizationNote{index, "string", fmt.Sprintf("ignored value %q not present in order", candidate)})
				continue
			}
			seen = true
			allBoolean = allBoolean && booleans[candidateIndex]
			if rank > bestRank {
				best, bestRank = candidate, rank
			}
		}
	}
	if seen && allBoolean {
		return best == "true", notes
	}
	return best, notes
}

func mergeCategorical(values []fieldValue, defaults []string, tieBreak string) (string, []normalizationNote) {
	fallback := ""
	if len(defaults) > 0 {
		fallback = defaults[0]
	}
	candidates, notes := make([]string, 0, len(values)), make([]normalizationNote, 0)
	for index, value := range values {
		candidate, current := categoricalCandidate(value, index, defaults, fallback)
		candidates, notes = append(candidates, candidate), append(notes, current...)
	}
	nonDefaults := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if !containsFold(defaults, candidate) {
			nonDefaults = append(nonDefaults, candidate)
		}
	}
	if len(nonDefaults) > 0 {
		return majorityFirst(nonDefaults), notes
	}
	if len(candidates) == 0 {
		return fallback, notes
	}
	return majorityFirst(candidates), notes
}

func mergeUniqueText(values []fieldValue, separator string, maxLength int) (string, []normalizationNote) {
	unique, notes := make([]string, 0, len(values)), make([]normalizationNote, 0)
	for index, value := range values {
		fragments, current := textFragments(value, index)
		notes = append(notes, current...)
		text := strings.Join(fragments, separator)
		if text != "" && !isNearDuplicate(text, unique) {
			unique = append(unique, text)
		}
	}
	return truncateRunes(strings.Join(unique, separator), maxLength), notes
}

func mergeNumeric(values []fieldValue, operation string) (interface{}, []normalizationNote) {
	numbers, notes := make([]float64, 0), make([]normalizationNote, 0)
	for index, value := range values {
		current, currentNotes := numericValues(value, index)
		numbers, notes = append(numbers, current...), append(notes, currentNotes...)
	}
	if len(numbers) == 0 {
		return nil, notes
	}
	result := numbers[0]
	for _, number := range numbers[1:] {
		if operation == "max" && number > result {
			result = number
		}
		if operation == "min" && number < result {
			result = number
		}
		if operation == "mean" {
			result += number
		}
	}
	if operation == "mean" {
		result /= float64(len(numbers))
	}
	return result, notes
}

func mergeMetadata(values []fieldValue, onMismatch string) (interface{}, []normalizationNote, error) {
	var first string
	found := false
	notes := make([]normalizationNote, 0)
	for index, value := range values {
		candidate, ok, current := metadataValue(value, index)
		notes = append(notes, current...)
		if !ok {
			continue
		}
		if !found {
			first, found = candidate, true
			continue
		}
		if candidate != first {
			if onMismatch == "error" {
				return nil, notes, fmt.Errorf("metadata values differ and on_mismatch is error")
			}
			notes = append(notes, normalizationNote{index, "string", "metadata mismatch retained first non-null value"})
		}
	}
	if !found {
		return "", notes, nil
	}
	return first, notes, nil
}

func stringsFor(value fieldValue, index int, stringifyScalars bool) ([]string, []normalizationNote) {
	if value.missing {
		return nil, []normalizationNote{{index, "missing", "treated as empty"}}
	}
	if value.value == nil {
		return nil, []normalizationNote{{index, "null", "treated as empty"}}
	}
	convert := func(item interface{}) (string, bool, string) {
		switch typed := item.(type) {
		case string:
			return typed, true, ""
		case bool:
			if stringifyScalars {
				return strconv.FormatBool(typed), true, "stringified boolean"
			}
		case float64:
			if stringifyScalars {
				return strconv.FormatFloat(typed, 'f', -1, 64), true, "stringified number"
			}
		case json.Number:
			if stringifyScalars {
				return string(typed), true, "stringified number"
			}
		}
		return "", false, "dropped unsupported value"
	}
	if array, ok := value.value.([]interface{}); ok {
		out, notes := make([]string, 0, len(array)), make([]normalizationNote, 0)
		for _, item := range array {
			text, valid, action := convert(item)
			if valid {
				out = append(out, text)
				if action != "" {
					notes = append(notes, normalizationNote{index, sourceType(item), action})
				}
			} else {
				notes = append(notes, normalizationNote{index, sourceType(item), action})
			}
		}
		return out, notes
	}
	text, ok, action := convert(value.value)
	if !ok {
		return nil, []normalizationNote{{index, sourceType(value.value), action}}
	}
	return []string{text}, []normalizationNote{{index, sourceType(value.value), "coerced scalar to array" + suffixAction(action)}}
}

func ordinalCandidates(value fieldValue, index int) ([]string, []normalizationNote, []bool) {
	if value.missing || value.value == nil {
		return nil, []normalizationNote{{index, fieldType(value), "treated as absent"}}, nil
	}
	items := []interface{}{value.value}
	if array, ok := value.value.([]interface{}); ok {
		items = array
	}
	result, notes, booleans := make([]string, 0, len(items)), make([]normalizationNote, 0), make([]bool, 0, len(items))
	for _, item := range items {
		switch typed := item.(type) {
		case string:
			result, booleans = append(result, typed), append(booleans, false)
		case bool:
			result, booleans, notes = append(result, strconv.FormatBool(typed)), append(booleans, true), append(notes, normalizationNote{index, "boolean", "coerced boolean to ordinal"})
		default:
			notes = append(notes, normalizationNote{index, sourceType(item), "dropped unsupported ordinal value"})
		}
	}
	return result, notes, booleans
}

func categoricalCandidate(value fieldValue, index int, defaults []string, fallback string) (string, []normalizationNote) {
	if value.missing || value.value == nil {
		return fallback, []normalizationNote{{index, fieldType(value), "treated as default"}}
	}
	items := []interface{}{value.value}
	if array, ok := value.value.([]interface{}); ok {
		items = array
	}
	first := ""
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			continue
		}
		if first == "" {
			first = text
		}
		if !containsFold(defaults, text) {
			return text, []normalizationNote{{index, "array", "selected first non-default array value"}}
		}
	}
	if first != "" {
		return first, nil
	}
	return fallback, []normalizationNote{{index, sourceType(value.value), "treated as default after dropping unsupported categorical value"}}
}

func textFragments(value fieldValue, index int) ([]string, []normalizationNote) {
	if value.missing || value.value == nil {
		return nil, []normalizationNote{{index, fieldType(value), "treated as empty text"}}
	}
	if text, ok := value.value.(string); ok {
		return []string{text}, nil
	}
	if array, ok := value.value.([]interface{}); ok {
		fragments, notes := make([]string, 0, len(array)), make([]normalizationNote, 0)
		for _, item := range array {
			if text, ok := item.(string); ok {
				fragments = append(fragments, text)
			} else {
				notes = append(notes, normalizationNote{index, sourceType(item), "dropped non-string text fragment"})
			}
		}
		return fragments, notes
	}
	return nil, []normalizationNote{{index, sourceType(value.value), "dropped unsupported text value"}}
}

func numericValues(value fieldValue, index int) ([]float64, []normalizationNote) {
	if value.missing || value.value == nil {
		return nil, []normalizationNote{{index, fieldType(value), "skipped absent numeric value"}}
	}
	items := []interface{}{value.value}
	if array, ok := value.value.([]interface{}); ok {
		items = array
	}
	numbers, notes := make([]float64, 0, len(items)), make([]normalizationNote, 0)
	for _, item := range items {
		number, ok := parseNumber(item)
		if !ok {
			notes = append(notes, normalizationNote{index, sourceType(item), "skipped non-numeric value"})
			continue
		}
		if sourceType(item) == "string" {
			notes = append(notes, normalizationNote{index, "string", "coerced numeric string"})
		}
		numbers = append(numbers, number)
	}
	return numbers, notes
}

func metadataValue(value fieldValue, index int) (string, bool, []normalizationNote) {
	if value.missing || value.value == nil {
		return "", false, []normalizationNote{{index, fieldType(value), "skipped absent metadata value"}}
	}
	text, ok := value.value.(string)
	if !ok {
		return "", false, []normalizationNote{{index, sourceType(value.value), "dropped non-string metadata value"}}
	}
	return text, true, nil
}

func parseNumber(value interface{}) (float64, bool) {
	var number float64
	switch typed := value.(type) {
	case float64:
		number = typed
	case float32:
		number = float64(typed)
	case int:
		number = float64(typed)
	case int64:
		number = float64(typed)
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		number = parsed
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, false
		}
		number = parsed
	default:
		return 0, false
	}
	return number, !math.IsNaN(number) && !math.IsInf(number, 0)
}

func sourceType(value interface{}) string {
	switch value.(type) {
	case nil:
		return "null"
	case []interface{}, []string:
		return "array"
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64, float32, int, int64, json.Number:
		return "number"
	default:
		return "object"
	}
}
func fieldType(value fieldValue) string {
	if value.missing {
		return "missing"
	}
	return sourceType(value.value)
}
func suffixAction(action string) string {
	if action == "" {
		return ""
	}
	return "; " + action
}

func majorityFirst(values []string) string {
	if len(values) == 0 {
		return ""
	}
	counts, best, bestCount := make(map[string]int, len(values)), values[0], 0
	for _, value := range values {
		counts[value]++
		if counts[value] > bestCount {
			best, bestCount = value, counts[value]
		}
	}
	return best
}
func distinctFieldValueCount(values []fieldValue) int {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value.missing {
			unique["<missing>"] = struct{}{}
			continue
		}
		encoded, _ := json.Marshal(value.value)
		unique[string(encoded)] = struct{}{}
	}
	return len(unique)
}
func containsFold(values []string, candidate string) bool {
	for _, value := range values {
		if strings.EqualFold(value, candidate) {
			return true
		}
	}
	return false
}
func normalizedText(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(text)), " ")
}
func isNearDuplicate(text string, existing []string) bool {
	normalized := normalizedText(text)
	for _, candidate := range existing {
		other := normalizedText(candidate)
		if normalized == other || strings.Contains(normalized, other) || strings.Contains(other, normalized) {
			return true
		}
	}
	return false
}
func truncateRunes(text string, maxLength int) string {
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}
	return string(runes[:maxLength])
}
