package logic

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/open-and-sustainable/prismaid/screening/filters"
)

// ScreeningConfig represents the TOML configuration for screening
type ScreeningConfig struct {
	Project ProjectConfig `toml:"project"`
	Filters FiltersConfig `toml:"filters"`
}

// ProjectConfig contains basic project information
type ProjectConfig struct {
	Name             string `toml:"name"`
	Author           string `toml:"author"`
	Version          string `toml:"version"`
	InputFile        string `toml:"input_file"`
	OutputFile       string `toml:"output_file"`
	TextColumn       string `toml:"text_column"`       // Column containing text/path to text files
	IdentifierColumn string `toml:"identifier_column"` // Column for unique identifiers
	OutputFormat     string `toml:"output_format"`     // csv or json
	LogLevel         string `toml:"log_level"`         // low, medium, high
}

// FiltersConfig contains settings for each screening filter
type FiltersConfig struct {
	Deduplication DeduplicationConfig `toml:"deduplication"`
	Language      LanguageConfig      `toml:"language"`
	ArticleType   ArticleTypeConfig   `toml:"article_type"`
	LLM           []LLMConfig         `toml:"llm"`
}

// DeduplicationConfig for duplicate detection
type DeduplicationConfig struct {
	Enabled       bool     `toml:"enabled"`
	Method        string   `toml:"method"`         // "exact", "fuzzy", "semantic"
	Threshold     float64  `toml:"threshold"`      // Similarity threshold for fuzzy/semantic
	CompareFields []string `toml:"compare_fields"` // Fields to compare for duplication
}

// LanguageConfig for language detection
type LanguageConfig struct {
	Enabled           bool     `toml:"enabled"`
	AcceptedLanguages []string `toml:"accepted_languages"` // e.g., ["en", "es", "fr"]
	UseAI             bool     `toml:"use_ai"`             // Use AI for detection vs rule-based
}

// ArticleTypeConfig for article classification
type ArticleTypeConfig struct {
	Enabled           bool     `toml:"enabled"`
	ExcludeReviews    bool     `toml:"exclude_reviews"`
	ExcludeEditorials bool     `toml:"exclude_editorials"`
	ExcludeLetters    bool     `toml:"exclude_letters"`
	IncludeTypes      []string `toml:"include_types"` // Specific types to include
}

// LLMConfig for AI model configuration (reused from main project)
type LLMConfig struct {
	Provider    string  `toml:"provider"`
	APIKey      string  `toml:"api_key"`
	Model       string  `toml:"model"`
	Temperature float64 `toml:"temperature"`
	TPMLimit    int     `toml:"tpm_limit"`
	RPMLimit    int     `toml:"rpm_limit"`
}

// ManuscriptRecord represents a single manuscript with tags
type ManuscriptRecord struct {
	ID              string                 `json:"id"`
	OriginalData    map[string]string      `json:"original_data"`
	Text            string                 `json:"-"` // Text content (not exported to JSON)
	Tags            map[string]interface{} `json:"tags"`
	ExclusionReason string                 `json:"exclusion_reason,omitempty"`
	Include         bool                   `json:"include"`
}

// ScreeningResult contains the complete screening results
type ScreeningResult struct {
	TotalRecords    int                `json:"total_records"`
	IncludedRecords int                `json:"included_records"`
	ExcludedRecords int                `json:"excluded_records"`
	Records         []ManuscriptRecord `json:"records"`
	Statistics      map[string]int     `json:"statistics"`
}

// Screen performs the main screening process
func Screen(tomlConfiguration string) error {
	// Parse TOML configuration
	var config ScreeningConfig
	if _, err := toml.Decode(tomlConfiguration, &config); err != nil {
		return fmt.Errorf("error parsing TOML configuration: %v", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return fmt.Errorf("configuration validation error: %v", err)
	}

	// Load input data
	manuscripts, err := loadInputData(config.Project.InputFile, config.Project.TextColumn, config.Project.IdentifierColumn)
	if err != nil {
		return fmt.Errorf("error loading input data: %v", err)
	}

	// Initialize screening result
	result := &ScreeningResult{
		TotalRecords: len(manuscripts),
		Records:      manuscripts,
		Statistics:   make(map[string]int),
	}

	// Apply filters
	if config.Filters.Deduplication.Enabled {
		if err := applyDeduplicationFilter(result, config.Filters.Deduplication); err != nil {
			return fmt.Errorf("deduplication filter error: %v", err)
		}
	}

	if config.Filters.Language.Enabled {
		if err := applyLanguageFilter(result, config.Filters.Language, config.Filters.LLM); err != nil {
			return fmt.Errorf("language filter error: %v", err)
		}
	}

	if config.Filters.ArticleType.Enabled {
		if err := applyArticleTypeFilter(result, config.Filters.ArticleType, config.Filters.LLM); err != nil {
			return fmt.Errorf("article type filter error: %v", err)
		}
	}

	// Calculate final statistics
	calculateStatistics(result)

	// Save results
	if err := saveResults(result, config.Project.OutputFile, config.Project.OutputFormat); err != nil {
		return fmt.Errorf("error saving results: %v", err)
	}

	// Log summary
	logSummary(result, config.Project.LogLevel)

	return nil
}

// loadInputData loads manuscripts from CSV or TXT file
func loadInputData(inputFile, textColumn, idColumn string) ([]ManuscriptRecord, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("cannot open input file: %v", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(inputFile))

	switch ext {
	case ".csv":
		return loadCSVData(file, textColumn, idColumn)
	case ".txt", ".tsv":
		return loadTSVData(file, textColumn, idColumn)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// loadCSVData loads data from CSV file
func loadCSVData(file io.Reader, textColumn, idColumn string) ([]ManuscriptRecord, error) {
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV header: %v", err)
	}

	// Find column indices
	textIdx := -1
	idIdx := -1
	for i, col := range header {
		if col == textColumn {
			textIdx = i
		}
		if col == idColumn {
			idIdx = i
		}
	}

	if textIdx == -1 {
		return nil, fmt.Errorf("text column '%s' not found in CSV", textColumn)
	}

	// Read records
	var manuscripts []ManuscriptRecord
	recordNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV record: %v", err)
		}

		recordNum++

		// Create manuscript record
		manuscript := ManuscriptRecord{
			OriginalData: make(map[string]string),
			Tags:         make(map[string]interface{}),
			Include:      true, // Initially include all records
		}

		// Set ID
		if idIdx >= 0 && idIdx < len(record) {
			manuscript.ID = record[idIdx]
		} else {
			manuscript.ID = fmt.Sprintf("record_%d", recordNum)
		}

		// Store all original data
		for i, value := range record {
			if i < len(header) {
				manuscript.OriginalData[header[i]] = value
			}
		}

		// Get text content (could be a path to file or actual text)
		if textIdx < len(record) {
			textContent := record[textIdx]
			// Check if it's a file path
			if fileExists(textContent) {
				content, err := os.ReadFile(textContent)
				if err != nil {
					fmt.Printf("Warning: Could not read file %s: %v\n", textContent, err)
					manuscript.Text = textContent // Use as is if file can't be read
				} else {
					manuscript.Text = string(content)
				}
			} else {
				manuscript.Text = textContent
			}
		}

		manuscripts = append(manuscripts, manuscript)
	}

	return manuscripts, nil
}

// loadTSVData loads data from TSV/TXT file
func loadTSVData(file io.Reader, textColumn, idColumn string) ([]ManuscriptRecord, error) {
	// Create a CSV reader configured for tab-separated values
	reader := csv.NewReader(file)
	reader.Comma = '\t'

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading TSV header: %v", err)
	}

	// Find column indices
	textIdx := -1
	idIdx := -1
	for i, col := range header {
		if col == textColumn {
			textIdx = i
		}
		if col == idColumn {
			idIdx = i
		}
	}

	if textIdx == -1 {
		return nil, fmt.Errorf("text column '%s' not found in TSV", textColumn)
	}

	// Read records
	var manuscripts []ManuscriptRecord
	recordNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading TSV record: %v", err)
		}

		recordNum++

		// Create manuscript record
		manuscript := ManuscriptRecord{
			OriginalData: make(map[string]string),
			Tags:         make(map[string]interface{}),
			Include:      true, // Initially include all records
		}

		// Set ID
		if idIdx >= 0 && idIdx < len(record) {
			manuscript.ID = record[idIdx]
		} else {
			manuscript.ID = fmt.Sprintf("record_%d", recordNum)
		}

		// Store all original data
		for i, value := range record {
			if i < len(header) {
				manuscript.OriginalData[header[i]] = value
			}
		}

		// Get text content (could be a path to file or actual text)
		if textIdx < len(record) {
			textContent := record[textIdx]
			// Check if it's a file path
			if fileExists(textContent) {
				content, err := os.ReadFile(textContent)
				if err != nil {
					fmt.Printf("Warning: Could not read file %s: %v\n", textContent, err)
					manuscript.Text = textContent // Use as is if file can't be read
				} else {
					manuscript.Text = string(content)
				}
			} else {
				manuscript.Text = textContent
			}
		}

		manuscripts = append(manuscripts, manuscript)
	}

	return manuscripts, nil
}

// applyDeduplicationFilter applies deduplication logic
func applyDeduplicationFilter(result *ScreeningResult, config DeduplicationConfig) error {
	// Convert ManuscriptRecord to filters.ManuscriptData
	manuscripts := make([]filters.ManuscriptData, len(result.Records))
	for i, record := range result.Records {
		manuscripts[i] = filters.ManuscriptData{
			ID:           record.ID,
			OriginalData: record.OriginalData,
			Text:         record.Text,
		}
	}

	// Convert config to filters.DeduplicationConfig
	filterConfig := filters.DeduplicationConfig{
		Method:        config.Method,
		Threshold:     config.Threshold,
		CompareFields: config.CompareFields,
	}

	duplicates := filters.FindDuplicates(manuscripts, filterConfig)

	duplicateCount := 0
	for i := range result.Records {
		if dupInfo, exists := duplicates[result.Records[i].ID]; exists {
			isDuplicate := dupInfo[0].(bool)
			originalID := dupInfo[1].(string)

			if isDuplicate {
				result.Records[i].Tags["is_duplicate"] = true
				result.Records[i].Tags["duplicate_of"] = originalID
				result.Records[i].Include = false
				result.Records[i].ExclusionReason = fmt.Sprintf("Duplicate of %s", originalID)
				duplicateCount++
			}
		}
	}

	result.Statistics["duplicates_found"] = duplicateCount
	return nil
}

// applyLanguageFilter applies language detection
func applyLanguageFilter(result *ScreeningResult, config LanguageConfig, llmConfigs []LLMConfig) error {
	excludedCount := 0

	for i := range result.Records {
		if !result.Records[i].Include {
			continue // Skip already excluded records
		}

		var language string
		var err error

		if config.UseAI && len(llmConfigs) > 0 {
			// Convert LLMConfig to interface{} for filter function
			language, err = filters.DetectLanguageWithAI(result.Records[i].Text, llmConfigs[0])
		} else {
			language, err = filters.DetectLanguage(result.Records[i].Text)
		}

		if err != nil {
			fmt.Printf("Warning: Language detection failed for %s: %v\n", result.Records[i].ID, err)
			continue
		}

		result.Records[i].Tags["detected_language"] = language

		// Check if language is accepted
		languageAccepted := false
		for _, acceptedLang := range config.AcceptedLanguages {
			if strings.EqualFold(language, acceptedLang) {
				languageAccepted = true
				break
			}
		}

		if !languageAccepted {
			result.Records[i].Include = false
			result.Records[i].ExclusionReason = fmt.Sprintf("Language not accepted: %s", language)
			excludedCount++
		}
	}

	result.Statistics["language_excluded"] = excludedCount
	return nil
}

// applyArticleTypeFilter applies article type classification
func applyArticleTypeFilter(result *ScreeningResult, config ArticleTypeConfig, llmConfigs []LLMConfig) error {
	excludedCount := 0

	for i := range result.Records {
		if !result.Records[i].Include {
			continue // Skip already excluded records
		}

		// Convert LLMConfig slice to []interface{} for filter function
		var llmInterfaces []interface{}
		for _, llm := range llmConfigs {
			llmInterfaces = append(llmInterfaces, llm)
		}

		articleType, err := filters.ClassifyArticleType(result.Records[i].Text, llmInterfaces)
		if err != nil {
			fmt.Printf("Warning: Article type classification failed for %s: %v\n", result.Records[i].ID, err)
			continue
		}

		result.Records[i].Tags["article_type"] = articleType

		// Check exclusion rules
		shouldExclude := false
		exclusionReason := ""

		if config.ExcludeReviews && strings.Contains(strings.ToLower(articleType), "review") {
			shouldExclude = true
			exclusionReason = "Review article"
		} else if config.ExcludeEditorials && strings.Contains(strings.ToLower(articleType), "editorial") {
			shouldExclude = true
			exclusionReason = "Editorial"
		} else if config.ExcludeLetters && strings.Contains(strings.ToLower(articleType), "letter") {
			shouldExclude = true
			exclusionReason = "Letter"
		}

		// Check include types if specified
		if len(config.IncludeTypes) > 0 && !shouldExclude {
			typeIncluded := false
			for _, includeType := range config.IncludeTypes {
				if strings.EqualFold(articleType, includeType) {
					typeIncluded = true
					break
				}
			}
			if !typeIncluded {
				shouldExclude = true
				exclusionReason = fmt.Sprintf("Article type not in include list: %s", articleType)
			}
		}

		if shouldExclude {
			result.Records[i].Include = false
			result.Records[i].ExclusionReason = exclusionReason
			excludedCount++
		}
	}

	result.Statistics["article_type_excluded"] = excludedCount
	return nil
}

// calculateStatistics calculates final statistics
func calculateStatistics(result *ScreeningResult) {
	included := 0
	excluded := 0

	for _, record := range result.Records {
		if record.Include {
			included++
		} else {
			excluded++
		}
	}

	result.IncludedRecords = included
	result.ExcludedRecords = excluded
}

// saveResults saves screening results to file
func saveResults(result *ScreeningResult, outputFile, format string) error {
	switch strings.ToLower(format) {
	case "json":
		return saveJSONResults(result, outputFile)
	case "csv":
		return saveCSVResults(result, outputFile)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// saveJSONResults saves results as JSON
func saveJSONResults(result *ScreeningResult, outputFile string) error {
	file, err := os.Create(outputFile + ".json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// saveCSVResults saves results as CSV
func saveCSVResults(result *ScreeningResult, outputFile string) error {
	file, err := os.Create(outputFile + ".csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Build header
	var header []string
	if len(result.Records) > 0 {
		// Original columns
		for key := range result.Records[0].OriginalData {
			header = append(header, key)
		}
		// Tag columns
		for key := range result.Records[0].Tags {
			header = append(header, "tag_"+key)
		}
		// Status columns
		header = append(header, "include", "exclusion_reason")
	}

	if err := writer.Write(header); err != nil {
		return err
	}

	// Write records
	for _, record := range result.Records {
		row := make([]string, len(header))
		for i, col := range header {
			if val, ok := record.OriginalData[col]; ok {
				row[i] = val
			} else if strings.HasPrefix(col, "tag_") {
				tagName := strings.TrimPrefix(col, "tag_")
				if tagVal, ok := record.Tags[tagName]; ok {
					row[i] = fmt.Sprintf("%v", tagVal)
				}
			} else if col == "include" {
				row[i] = fmt.Sprintf("%v", record.Include)
			} else if col == "exclusion_reason" {
				row[i] = record.ExclusionReason
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// logSummary logs screening summary
func logSummary(result *ScreeningResult, logLevel string) {
	fmt.Printf("\n=== Screening Summary ===\n")
	fmt.Printf("Total Records: %d\n", result.TotalRecords)
	fmt.Printf("Included: %d\n", result.IncludedRecords)
	fmt.Printf("Excluded: %d\n", result.ExcludedRecords)

	if logLevel == "medium" || logLevel == "high" {
		fmt.Printf("\n--- Exclusion Statistics ---\n")
		for key, value := range result.Statistics {
			fmt.Printf("%s: %d\n", key, value)
		}
	}

	if logLevel == "high" {
		// Could save detailed log to file
		logFile := "screening_log.txt"
		file, err := os.Create(logFile)
		if err == nil {
			defer file.Close()
			fmt.Fprintf(file, "Detailed Screening Log\n")
			fmt.Fprintf(file, "======================\n\n")
			for _, record := range result.Records {
				fmt.Fprintf(file, "ID: %s\n", record.ID)
				fmt.Fprintf(file, "Include: %v\n", record.Include)
				if !record.Include {
					fmt.Fprintf(file, "Exclusion Reason: %s\n", record.ExclusionReason)
				}
				fmt.Fprintf(file, "Tags: %v\n\n", record.Tags)
			}
			fmt.Printf("\nDetailed log saved to: %s\n", logFile)
		}
	}
}

// validateConfig validates the configuration
func validateConfig(config *ScreeningConfig) error {
	if config.Project.InputFile == "" {
		return fmt.Errorf("input_file is required")
	}

	if config.Project.OutputFile == "" {
		return fmt.Errorf("output_file is required")
	}

	if config.Project.TextColumn == "" {
		return fmt.Errorf("text_column is required")
	}

	// Check if at least one filter is enabled
	if !config.Filters.Deduplication.Enabled &&
		!config.Filters.Language.Enabled &&
		!config.Filters.ArticleType.Enabled {
		return fmt.Errorf("at least one filter must be enabled")
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
