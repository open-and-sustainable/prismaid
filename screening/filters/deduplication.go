package filters

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/alembica/extraction"
	"github.com/open-and-sustainable/alembica/utils/logger"
)

// ManuscriptData represents the data structure for a manuscript
type ManuscriptData struct {
	ID            string
	OriginalData  map[string]string
	LowerFieldMap map[string]string // Lowercase to original field name mapping
	Text          string
}

// DeduplicationConfig represents the configuration for deduplication
type DeduplicationConfig struct {
	UseAI         bool
	CompareFields []string
	LLMConfigs    []any // LLM configurations for AI-based deduplication
}

// FindDuplicates identifies duplicate manuscripts based on the configuration
// Returns a map where key is manuscript ID and value is a tuple (isDuplicate, originalID)
func FindDuplicates(records []ManuscriptData, config DeduplicationConfig) map[string][2]interface{} {
	if config.UseAI && len(config.LLMConfigs) > 0 {
		logger.Info("Using AI-based deduplication with %d models", len(config.LLMConfigs))
		return findAIMatches(records, config)
	}
	logger.Info("Using simple matching for deduplication")
	return findSimpleMatches(records, config.CompareFields)
}

// findSimpleMatches finds matches allowing for single character differences
func findSimpleMatches(manuscripts []ManuscriptData, compareFields []string) map[string][2]interface{} {
	duplicates := make(map[string][2]interface{})

	// Initialize all as non-duplicates
	for _, manuscript := range manuscripts {
		duplicates[manuscript.ID] = [2]interface{}{false, ""}
	}

	// Compare each manuscript with others
	for i := 0; i < len(manuscripts); i++ {
		if duplicates[manuscripts[i].ID][0].(bool) {
			continue // Skip if already marked as duplicate
		}

		for j := i + 1; j < len(manuscripts); j++ {
			if duplicates[manuscripts[j].ID][0].(bool) {
				continue // Skip if already marked as duplicate
			}

			if areSimpleMatches(manuscripts[i], manuscripts[j], compareFields) {
				// Mark the second one as duplicate of the first
				duplicates[manuscripts[j].ID] = [2]interface{}{true, manuscripts[i].ID}
			}
		}
	}

	return duplicates
}

// areSimpleMatches checks if two manuscripts are matches allowing for single character differences
func areSimpleMatches(m1, m2 ManuscriptData, compareFields []string) bool {
	// First, check if DOI field exists and matches exactly - if so, it's a duplicate
	for _, field := range compareFields {
		fieldLower := strings.ToLower(field)
		if fieldLower == "doi" {
			// Try to get DOI with different case variations
			doi1 := getFieldValueWithMapping(m1.OriginalData, m1.LowerFieldMap, "doi")
			doi2 := getFieldValueWithMapping(m2.OriginalData, m2.LowerFieldMap, "doi")

			// If both have DOIs and they match exactly, it's a duplicate
			if doi1 != "" && doi2 != "" && doi1 == doi2 {
				return true
			}
		}
	}

	// Check for author+year+(title OR abstract) combination
	hasAuthors := false
	hasYear := false
	hasTitle := false
	hasAbstract := false

	for _, field := range compareFields {
		fieldLower := strings.ToLower(field)
		if fieldLower == "authors" || fieldLower == "author" {
			hasAuthors = true
		}
		if fieldLower == "year" {
			hasYear = true
		}
		if fieldLower == "title" {
			hasTitle = true
		}
		if fieldLower == "abstract" {
			hasAbstract = true
		}
	}

	// If we have the required fields for combination check
	if hasAuthors && hasYear && (hasTitle || hasAbstract) {
		authorsMatch := false
		yearMatch := false
		titleMatch := false
		abstractMatch := false

		// Check authors
		val1 := getFieldValueWithMapping(m1.OriginalData, m1.LowerFieldMap, "authors")
		if val1 == "" {
			val1 = getFieldValueWithMapping(m1.OriginalData, m1.LowerFieldMap, "author")
		}
		val2 := getFieldValueWithMapping(m2.OriginalData, m2.LowerFieldMap, "authors")
		if val2 == "" {
			val2 = getFieldValueWithMapping(m2.OriginalData, m2.LowerFieldMap, "author")
		}
		if val1 != "" && val2 != "" {
			authorsMatch = isSingleCharDifference(val1, val2)
		}

		// Check year
		val1 = getFieldValueWithMapping(m1.OriginalData, m1.LowerFieldMap, "year")
		val2 = getFieldValueWithMapping(m2.OriginalData, m2.LowerFieldMap, "year")
		if val1 != "" && val2 != "" {
			yearMatch = isSingleCharDifference(val1, val2)
		}

		// Check title
		if hasTitle {
			val1 = getFieldValueWithMapping(m1.OriginalData, m1.LowerFieldMap, "title")
			val2 = getFieldValueWithMapping(m2.OriginalData, m2.LowerFieldMap, "title")
			if val1 != "" && val2 != "" {
				titleMatch = isSingleCharDifference(val1, val2)
			}
		}

		// Check abstract
		if hasAbstract {
			val1 = getFieldValueWithMapping(m1.OriginalData, m1.LowerFieldMap, "abstract")
			val2 = getFieldValueWithMapping(m2.OriginalData, m2.LowerFieldMap, "abstract")
			if val1 != "" && val2 != "" {
				abstractMatch = isSingleCharDifference(val1, val2)
			}
		}

		// If authors and year match, and either title or abstract matches, it's a duplicate
		if authorsMatch && yearMatch && (titleMatch || abstractMatch) {
			return true
		}
	}

	// If no special logic matched, check all fields with AND logic
	for _, field := range compareFields {
		fieldLower := strings.ToLower(field)

		// Skip DOI as it was already checked
		if fieldLower == "doi" {
			continue
		}

		val1 := ""
		val2 := ""

		if fieldLower == "text" {
			val1 = normalizeText(m1.Text)
			val2 = normalizeText(m2.Text)
		} else {
			val1 = getFieldValueWithMapping(m1.OriginalData, m1.LowerFieldMap, field)
			val2 = getFieldValueWithMapping(m2.OriginalData, m2.LowerFieldMap, field)
		}

		// If both values are empty, skip this field
		if val1 == "" && val2 == "" {
			continue
		}

		// Check for exact match or single character difference
		if !isSingleCharDifference(val1, val2) {
			return false
		}
	}

	return true
}

// isSingleCharDifference checks if two strings are identical or differ by at most one character
func isSingleCharDifference(s1, s2 string) bool {
	// Exact match
	if s1 == s2 {
		return true
	}

	// Check if length difference is at most 1
	lenDiff := len(s1) - len(s2)
	if lenDiff < -1 || lenDiff > 1 {
		return false
	}

	// Same length - check for single substitution
	if lenDiff == 0 {
		differences := 0
		for i := 0; i < len(s1); i++ {
			if s1[i] != s2[i] {
				differences++
				if differences > 1 {
					return false
				}
			}
		}
		return differences <= 1
	}

	// Different length by 1 - check for single insertion/deletion
	shorter, longer := s1, s2
	if len(s1) > len(s2) {
		shorter, longer = s2, s1
	}

	i, j := 0, 0
	differences := 0
	for i < len(shorter) && j < len(longer) {
		if shorter[i] != longer[j] {
			differences++
			if differences > 1 {
				return false
			}
			j++ // Skip character in longer string
		} else {
			i++
			j++
		}
	}

	return true
}

// normalizeText normalizes text for comparison
func normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")
	return text
}

// getFieldValueCaseInsensitive tries to get a field value with case-insensitive matching
func getFieldValueCaseInsensitive(data map[string]string, fieldName string) string {
	// Direct match first
	if val, exists := data[fieldName]; exists {
		return normalizeText(val)
	}

	// Try case variations
	fieldLower := strings.ToLower(fieldName)
	if val, exists := data[fieldLower]; exists {
		return normalizeText(val)
	}
	if val, exists := data[strings.ToUpper(fieldName)]; exists {
		return normalizeText(val)
	}
	if val, exists := data[strings.Title(fieldName)]; exists {
		return normalizeText(val)
	}

	return ""
}

// getFieldValueWithMapping tries to get a field value using the lowercase field mapping
func getFieldValueWithMapping(data map[string]string, lowerMap map[string]string, fieldName string) string {
	// Try direct match first
	if val, exists := data[fieldName]; exists {
		return normalizeText(val)
	}

	// Try using lowercase mapping
	fieldLower := strings.ToLower(fieldName)
	if originalField, exists := lowerMap[fieldLower]; exists {
		if val, exists := data[originalField]; exists {
			return normalizeText(val)
		}
	}

	return ""
}

// findAIMatches uses AI models to detect duplicates
func findAIMatches(manuscripts []ManuscriptData, config DeduplicationConfig) map[string][2]interface{} {
	duplicates := make(map[string][2]interface{})

	logger.Info("Starting AI-based duplicate detection for %d manuscripts", len(manuscripts))

	// Initialize all as non-duplicates
	for _, manuscript := range manuscripts {
		duplicates[manuscript.ID] = [2]interface{}{false, ""}
	}

	// Prepare AI model configurations
	var models []definitions.Model
	for _, llmConfig := range config.LLMConfigs {
		if llm, ok := llmConfig.(map[string]interface{}); ok {
			model := definitions.Model{
				Provider:     getStringValue(llm, "provider"),
				APIKey:       getStringValue(llm, "api_key"),
				Model:        getStringValue(llm, "model"),
				Temperature:  getFloatValue(llm, "temperature"),
				TPMLimit:     getIntValue(llm, "tpm_limit"),
				RPMLimit:     getIntValue(llm, "rpm_limit"),
				BaseURL:      getStringValue(llm, "base_url"),
				EndpointType: getStringValue(llm, "endpoint_type"),
				Region:       getStringValue(llm, "region"),
				ProjectID:    getStringValue(llm, "project_id"),
				Location:     getStringValue(llm, "location"),
				APIVersion:   getStringValue(llm, "api_version"),
			}
			models = append(models, model)
		}
	}

	if len(models) == 0 {
		logger.Info("No valid AI models configured, falling back to simple matching")
		// Fall back to simple matching if no valid models
		return findSimpleMatches(manuscripts, config.CompareFields)
	}

	logger.Info("Configured %d AI models for deduplication", len(models))

	// Build all comparison prompts
	type ComparisonPair struct {
		Index1 int
		Index2 int
		Prompt definitions.Prompt
	}
	var comparisons []ComparisonPair
	promptID := 1

	// Build field names list for context
	fieldsList := strings.Join(config.CompareFields, ", ")

	for i := 0; i < len(manuscripts); i++ {
		if duplicates[manuscripts[i].ID][0].(bool) {
			continue // Skip if already marked as duplicate
		}

		for j := i + 1; j < len(manuscripts); j++ {
			if duplicates[manuscripts[j].ID][0].(bool) {
				continue // Skip if already marked as duplicate
			}

			// Build the comparison data
			comparison1 := buildComparisonData(manuscripts[i], config.CompareFields)
			comparison2 := buildComparisonData(manuscripts[j], config.CompareFields)

			// Create the prompt
			prompt := fmt.Sprintf(`You are a scientific reviewer tasked with identifying duplicate manuscripts in a research database. You are provided with specific fields from two different records to compare.

CONTEXT:
- You are comparing the following fields: %s
- Records may have variations due to:
  * Author name formats (initials vs full names, middle names, order variations)
  * Character encoding issues (é→e, ü→u, ñ→n, ø→o, incorrect UTF-8 representation)
  * Non-standard character replacements (Müller→Mueller, Gómez→Gomez, Søren→Soren)
  * Technical simplifications in database entries
  * Minor transcription differences
  * Abbreviated vs full journal names
  * Different citation styles or formats
  * Minor typos or punctuation differences

IMPORTANT CONSIDERATIONS:
- If DOI is provided and identical, they are definitely duplicates
- For author names: "Smith, J." and "Smith, John" likely refer to the same person
- Character variations: "Müller" and "Mueller" or "André" and "Andre" are likely the same
- For titles: ignore minor differences in capitalization, punctuation, or small words
- For years: same year is a strong indicator if other fields match
- For abstracts: similar content with different phrasing may still be duplicates

MANUSCRIPT 1:
%s

MANUSCRIPT 2:
%s

TASK: Determine if these represent the same publication.
Respond with ONLY a JSON object: {"duplicate": true} or {"duplicate": false}`, fieldsList, comparison1, comparison2)

			comparisons = append(comparisons, ComparisonPair{
				Index1: i,
				Index2: j,
				Prompt: definitions.Prompt{
					PromptContent:  prompt,
					SequenceID:     fmt.Sprintf("%d", promptID),
					SequenceNumber: promptID,
				},
			})
			promptID++
		}
	}

	if len(comparisons) == 0 {
		logger.Info("No comparisons needed")
		return duplicates
	}

	logger.Info("Prepared %d comparison prompts for batch processing", len(comparisons))

	// Create all prompts for batch processing
	var prompts []definitions.Prompt
	for _, comp := range comparisons {
		prompts = append(prompts, comp.Prompt)
	}

	// Prepare the input for alembica with all prompts
	input := definitions.Input{
		Metadata: definitions.InputMetadata{
			Version:       "1.0",
			SchemaVersion: "1.0",
		},
		Models:  models,
		Prompts: prompts,
	}

	// Convert to JSON
	jsonInput, err := json.Marshal(input)
	if err != nil {
		logger.Error("Failed to marshal input for AI: %v", err)
		return duplicates
	}

	// Call alembica once with all prompts
	logger.Info("Calling AI model with batch of %d comparisons", len(comparisons))
	result, err := extraction.Extract(string(jsonInput))
	if err != nil {
		logger.Error("AI extraction failed: %v", err)
		return duplicates
	}

	// Parse the response
	var output definitions.Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		logger.Error("Failed to parse AI response: %v", err)
		return duplicates
	}

	// Process responses and update duplicates map
	for idx, comp := range comparisons {
		if idx < len(output.Responses) && len(output.Responses[idx].ModelResponses) > 0 {
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(output.Responses[idx].ModelResponses[0]), &response); err == nil {
				if duplicate, ok := response["duplicate"].(bool); ok && duplicate {
					logger.Info("AI detected duplicate: manuscript %s is duplicate of %s",
						manuscripts[comp.Index2].ID, manuscripts[comp.Index1].ID)
					duplicates[manuscripts[comp.Index2].ID] = [2]interface{}{true, manuscripts[comp.Index1].ID}
				}
			}
		}
	}

	return duplicates
}

// buildComparisonData builds a string representation of manuscript data for comparison
func buildComparisonData(m ManuscriptData, compareFields []string) string {
	var parts []string
	for _, field := range compareFields {
		var value string
		fieldLower := strings.ToLower(field)
		if fieldLower == "text" {
			value = m.Text
		} else {
			value = getFieldValueWithMapping(m.OriginalData, m.LowerFieldMap, field)
		}

		if value != "" {
			// Format field name for better readability
			formattedField := strings.ToUpper(strings.ReplaceAll(field, "_", " "))
			parts = append(parts, fmt.Sprintf("%s: %s", formattedField, value))
		}
	}

	if len(parts) == 0 {
		return "[No data available for comparison fields]"
	}

	return strings.Join(parts, "\n")
}

// Helper functions to extract values from interface maps
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0.0
}

func getIntValue(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}
