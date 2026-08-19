package chunking

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/prismaid/review/config"
)

// Binding associates an extraction prompt with its source document and chunk.
// DocumentIndex is zero-based internally; Merge restores one-based sequence
// IDs for existing result writers.
type Binding struct {
	DocumentIndex int
	Filename      string
	ChunkIndex    int
	ChunkCount    int
}

// Conflict records values from more than one chunk that required a merge
// decision. Conflicts are reported in the review log without changing the
// configured review schema.
type Conflict struct {
	DocumentIndex int
	Filename      string
	Field         string
}

// MergeReport contains non-fatal merge conflicts for review logging. Conflicts
// preserve schema conformance by remaining in the report rather than adding
// unconfigured fields to model output.
type MergeReport struct {
	Conflicts []Conflict
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

// Merge combines chunk-level responses into one response for every source
// document, provider, model, and response artifact. Primary extraction
// responses use the explicit field rules in cfg. Auxiliary justification and
// summary responses are deduplicated and concatenated per document. Missing,
// failed, malformed, or unbound chunk responses stop the review rather than
// silently producing a partial document-level extraction.
func Merge(results string, bindings map[string]Binding, cfg config.ChunkingConfig) (string, MergeReport, error) {
	var output definitions.Output
	if err := json.Unmarshal([]byte(results), &output); err != nil {
		return "", MergeReport{}, fmt.Errorf("parse chunk extraction output: %w", err)
	}

	groups := make(map[groupedResponseKey][]groupedResponse)
	for _, response := range output.Responses {
		binding, ok := bindings[response.SequenceID]
		if !ok {
			return "", MergeReport{}, fmt.Errorf("missing source-document binding for response sequence %q", response.SequenceID)
		}
		key := groupedResponseKey{
			DocumentIndex:  binding.DocumentIndex,
			Provider:       response.Provider,
			Model:          response.Model,
			SequenceNumber: response.SequenceNumber,
		}
		groups[key] = append(groups[key], groupedResponse{binding: binding, response: response})
	}

	keys := make([]groupedResponseKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
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

	merged := definitions.Output{Metadata: output.Metadata}
	report := MergeReport{}
	for _, key := range keys {
		responses := groups[key]
		sort.Slice(responses, func(i, j int) bool {
			return responses[i].binding.ChunkIndex < responses[j].binding.ChunkIndex
		})
		if err := validateGroup(responses); err != nil {
			return "", report, err
		}

		modelResponse, conflicts, err := mergeGroup(responses, cfg)
		if err != nil {
			return "", report, err
		}
		report.Conflicts = append(report.Conflicts, conflicts...)
		merged.Responses = append(merged.Responses, definitions.Response{
			Provider:       key.Provider,
			Model:          key.Model,
			SequenceID:     strconv.Itoa(key.DocumentIndex + 1),
			SequenceNumber: key.SequenceNumber,
			ModelResponses: []string{modelResponse},
		})
	}

	encoded, err := json.Marshal(merged)
	if err != nil {
		return "", report, fmt.Errorf("encode merged chunk output: %w", err)
	}
	return string(encoded), report, nil
}

func validateGroup(responses []groupedResponse) error {
	if len(responses) == 0 {
		return fmt.Errorf("cannot merge an empty response group")
	}
	expected := responses[0].binding.ChunkCount
	if len(responses) != expected {
		return fmt.Errorf("document %q returned %d of %d expected chunk responses", responses[0].binding.Filename, len(responses), expected)
	}
	for index, item := range responses {
		if item.binding.ChunkIndex != index {
			return fmt.Errorf("document %q is missing chunk response %d", item.binding.Filename, index+1)
		}
		if item.response.Error != nil {
			return fmt.Errorf("document %q chunk %d failed: %s", item.binding.Filename, index+1, item.response.Error.Message)
		}
		if len(item.response.ModelResponses) == 0 {
			return fmt.Errorf("document %q chunk %d returned no model response", item.binding.Filename, index+1)
		}
	}
	return nil
}

func mergeGroup(responses []groupedResponse, cfg config.ChunkingConfig) (string, []Conflict, error) {
	if responses[0].response.SequenceNumber != 1 {
		return mergeAuxiliary(responses), nil, nil
	}

	objects := make([]map[string]interface{}, 0, len(responses))
	for _, item := range responses {
		var object map[string]interface{}
		if err := json.Unmarshal([]byte(item.response.ModelResponses[0]), &object); err != nil {
			return "", nil, fmt.Errorf("document %q chunk %d returned invalid JSON: %w", item.binding.Filename, item.binding.ChunkIndex+1, err)
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
	for _, field := range fields {
		values := make([]interface{}, 0, len(objects))
		for index, object := range objects {
			value, ok := object[field]
			if !ok {
				return "", nil, fmt.Errorf("document %q chunk %d is missing configured review field %q", responses[index].binding.Filename, index+1, field)
			}
			values = append(values, value)
		}

		value, conflict, err := mergeField(field, values, cfg.Merge[field])
		if err != nil {
			return "", nil, fmt.Errorf("merge document %q field %q: %w", responses[0].binding.Filename, field, err)
		}
		merged[field] = value
		if conflict {
			conflicts = append(conflicts, Conflict{
				DocumentIndex: responses[0].binding.DocumentIndex,
				Filename:      responses[0].binding.Filename,
				Field:         field,
			})
		}
	}

	encoded, err := json.Marshal(merged)
	if err != nil {
		return "", nil, fmt.Errorf("encode merged response: %w", err)
	}
	return string(encoded), conflicts, nil
}

func mergeAuxiliary(responses []groupedResponse) string {
	unique := make([]string, 0, len(responses))
	seen := make(map[string]struct{}, len(responses))
	for _, item := range responses {
		value := item.response.ModelResponses[0]
		key := normalizedText(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, value)
	}
	return strings.Join(unique, "\n\n")
}

func mergeField(field string, values []interface{}, rule config.MergeRule) (interface{}, bool, error) {
	conflict := distinctValueCount(values) > 1
	switch rule.Rule {
	case "union":
		value, _, err := mergeUnion(values, rule.Sentinels)
		return value, conflict, err
	case "ordinal":
		value, err := mergeOrdinal(values, rule.Order)
		return value, conflict, err
	case "categorical":
		value, err := mergeCategorical(values, rule.Defaults, rule.TieBreak)
		return value, conflict, err
	case "unique_text":
		value, err := mergeUniqueText(values, rule.Separator, rule.MaxLength)
		return value, conflict, err
	case "numeric":
		value, err := mergeNumeric(values, rule.Operation)
		return value, conflict, err
	case "metadata":
		value, err := mergeMetadata(field, values, rule.OnMismatch)
		return value, conflict, err
	default:
		return nil, false, fmt.Errorf("unsupported merge rule %q", rule.Rule)
	}
}

func mergeUnion(values []interface{}, sentinels []string) (interface{}, bool, error) {
	items := make([]string, 0)
	seen := make(map[string]struct{})
	for _, value := range values {
		array, ok := value.([]interface{})
		if !ok {
			return nil, false, fmt.Errorf("union rule requires arrays")
		}
		for _, entry := range array {
			item, ok := entry.(string)
			if !ok {
				return nil, false, fmt.Errorf("union rule requires string array values")
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}

	hasRealValue := false
	for _, item := range items {
		if !containsFold(sentinels, item) {
			hasRealValue = true
			break
		}
	}
	if hasRealValue {
		filtered := items[:0]
		for _, item := range items {
			if !containsFold(sentinels, item) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	return items, len(items) > 1, nil
}

func mergeOrdinal(values []interface{}, order []string) (interface{}, error) {
	ranks := make(map[string]int, len(order))
	for index, value := range order {
		ranks[value] = index
	}
	best := ""
	bestRank := -1
	valueKind := 0
	for _, value := range values {
		candidate := ""
		switch typed := value.(type) {
		case string:
			if valueKind == 0 {
				valueKind = 1
			} else if valueKind != 1 {
				return nil, fmt.Errorf("ordinal values must all have the same type")
			}
			candidate = typed
		case bool:
			if valueKind == 0 {
				valueKind = 2
			} else if valueKind != 2 {
				return nil, fmt.Errorf("ordinal values must all have the same type")
			}
			candidate = strconv.FormatBool(typed)
		default:
			return nil, fmt.Errorf("ordinal rule requires strings or booleans")
		}
		rank, ok := ranks[candidate]
		if !ok {
			return nil, fmt.Errorf("ordinal value %q is not included in the configured order", candidate)
		}
		if rank > bestRank {
			best, bestRank = candidate, rank
		}
	}
	if valueKind == 2 {
		return best == "true", nil
	}
	return best, nil
}

func mergeCategorical(values []interface{}, defaults []string, tieBreak string) (string, error) {
	if tieBreak != "first" {
		return "", fmt.Errorf("unsupported categorical tie break %q", tieBreak)
	}
	candidates := make([]string, 0, len(values))
	for _, value := range values {
		candidate, err := scalarString(value)
		if err != nil {
			return "", err
		}
		if !containsFold(defaults, candidate) {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		for _, value := range values {
			candidate, err := scalarString(value)
			if err != nil {
				return "", err
			}
			candidates = append(candidates, candidate)
		}
	}
	return majorityFirst(candidates), nil
}

func mergeUniqueText(values []interface{}, separator string, maxLength int) (string, error) {
	unique := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("unique_text rule requires strings")
		}
		if text == "" || isNearDuplicate(text, unique) {
			continue
		}
		unique = append(unique, text)
	}
	return truncateRunes(strings.Join(unique, separator), maxLength), nil
}

func mergeNumeric(values []interface{}, operation string) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("numeric rule requires at least one value")
	}
	result := 0.0
	for index, value := range values {
		number, ok := value.(float64)
		if !ok {
			return 0, fmt.Errorf("numeric rule requires numbers")
		}
		if index == 0 || (operation == "max" && number > result) || (operation == "min" && number < result) {
			result = number
		}
		if operation == "mean" && index > 0 {
			result += number
		}
	}
	if operation == "mean" {
		return result / float64(len(values)), nil
	}
	if operation != "max" && operation != "min" {
		return 0, fmt.Errorf("unsupported numeric operation %q", operation)
	}
	return result, nil
}

func mergeMetadata(field string, values []interface{}, onMismatch string) (string, error) {
	first, err := scalarString(values[0])
	if err != nil {
		return "", err
	}
	for _, value := range values[1:] {
		candidate, err := scalarString(value)
		if err != nil {
			return "", err
		}
		if candidate != first && onMismatch == "error" {
			return "", fmt.Errorf("metadata values differ and on_mismatch is error")
		}
	}
	return first, nil
}

func scalarString(value interface{}) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case bool:
		return strconv.FormatBool(typed), nil
	default:
		return "", fmt.Errorf("expected a string or boolean value")
	}
}

func majorityFirst(values []string) string {
	counts := make(map[string]int, len(values))
	best, bestCount := values[0], 0
	for _, value := range values {
		counts[value]++
		if counts[value] > bestCount {
			best, bestCount = value, counts[value]
		}
	}
	return best
}

func distinctValueCount(values []interface{}) int {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		encoded, _ := json.Marshal(value)
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
