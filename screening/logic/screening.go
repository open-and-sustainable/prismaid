package logic

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/open-and-sustainable/alembica/utils/logger"
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
	Deduplication  DeduplicationConfig  `toml:"deduplication"`
	Language       LanguageConfig       `toml:"language"`
	ArticleType    ArticleTypeConfig    `toml:"article_type"`
	TopicRelevance TopicRelevanceConfig `toml:"topic_relevance"`
	LLM            []LLMConfig          `toml:"llm"`
}

// DeduplicationConfig for duplicate detection
type DeduplicationConfig struct {
	Enabled       bool     `toml:"enabled"`
	UseAI         bool     `toml:"use_ai"`         // Use AI for similarity detection
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
	Enabled           bool `toml:"enabled"`
	UseAI             bool `toml:"use_ai"` // Use AI for classification vs rule-based
	ExcludeReviews    bool `toml:"exclude_reviews"`
	ExcludeEditorials bool `toml:"exclude_editorials"`
	ExcludeLetters    bool `toml:"exclude_letters"`

	// Empirical vs Theoretical filtering
	ExcludeTheoretical bool `toml:"exclude_theoretical"`
	ExcludeEmpirical   bool `toml:"exclude_empirical"`
	ExcludeMethods     bool `toml:"exclude_methods"`

	// Study scope filtering
	ExcludeSingleCase bool `toml:"exclude_single_case"`
	ExcludeSample     bool `toml:"exclude_sample"`

	IncludeTypes []string `toml:"include_types"` // Specific types to include
}

// TopicRelevanceConfig for topic-based filtering
type TopicRelevanceConfig struct {
	Enabled      bool                       `toml:"enabled"`
	UseAI        bool                       `toml:"use_ai"`        // Use AI for relevance scoring
	Topics       []string                   `toml:"topics"`        // List of topic descriptions
	MinScore     float64                    `toml:"min_score"`     // Minimum score (0-1) to include
	ScoreWeights TopicRelevanceScoreWeights `toml:"score_weights"` // Weights for different scoring components
}

// TopicRelevanceScoreWeights defines the weights for different scoring components
type TopicRelevanceScoreWeights struct {
	KeywordMatch   float64 `toml:"keyword_match"`   // Weight for keyword matching
	ConceptMatch   float64 `toml:"concept_match"`   // Weight for concept matching
	FieldRelevance float64 `toml:"field_relevance"` // Weight for field/domain relevance
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
	LowerFieldMap   map[string]string      `json:"-"` // Lowercase to original field name mapping
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
	LLMConfigs      []LLMConfig        `json:"-"` // Pass LLM configs through for filters
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

	// Setup logger based on configuration
	if config.Project.LogLevel == "high" {
		logger.SetupLogging(logger.File, config.Project.OutputFile)
	} else if config.Project.LogLevel == "medium" {
		logger.SetupLogging(logger.Stdout, config.Project.OutputFile)
	} else {
		logger.SetupLogging(logger.Silent, config.Project.OutputFile) // default value
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
		LLMConfigs:   config.Filters.LLM,
	}

	// Apply filters
	previousFilterUsedAI := false

	if config.Filters.Deduplication.Enabled {
		// Check if we need to wait before this AI filter
		if previousFilterUsedAI && config.Filters.Deduplication.UseAI && len(config.Filters.LLM) > 0 {
			logger.Info("Waiting 30 seconds before next AI-assisted filter...")
			time.Sleep(30 * time.Second)
		}

		if err := applyDeduplicationFilter(result, config.Filters.Deduplication); err != nil {
			return fmt.Errorf("deduplication filter error: %v", err)
		}

		// Update flag for next filter
		previousFilterUsedAI = config.Filters.Deduplication.UseAI && len(config.Filters.LLM) > 0
	}

	if config.Filters.Language.Enabled {
		// Check if we need to wait before this AI filter
		if previousFilterUsedAI && config.Filters.Language.UseAI && len(config.Filters.LLM) > 0 {
			logger.Info("Waiting 30 seconds before next AI-assisted filter...")
			time.Sleep(30 * time.Second)
		}

		if err := applyLanguageFilter(result, config.Filters.Language, config.Filters.LLM); err != nil {
			return fmt.Errorf("language filter error: %v", err)
		}

		// Update flag for next filter
		previousFilterUsedAI = config.Filters.Language.UseAI && len(config.Filters.LLM) > 0
	}

	if config.Filters.ArticleType.Enabled {
		// Check if we need to wait before this AI filter
		if previousFilterUsedAI && config.Filters.ArticleType.UseAI && len(config.Filters.LLM) > 0 {
			logger.Info("Waiting 30 seconds before next AI-assisted filter...")
			time.Sleep(30 * time.Second)
		}

		if err := applyArticleTypeFilter(result, config.Filters.ArticleType, config.Filters.LLM); err != nil {
			return fmt.Errorf("article type filter error: %v", err)
		}

		// Update flag for next filter
		previousFilterUsedAI = config.Filters.ArticleType.UseAI && len(config.Filters.LLM) > 0
	}

	if config.Filters.TopicRelevance.Enabled {
		// Check if we need to wait before this AI filter
		if previousFilterUsedAI && config.Filters.TopicRelevance.UseAI && len(config.Filters.LLM) > 0 {
			logger.Info("Waiting 30 seconds before next AI-assisted filter...")
			time.Sleep(30 * time.Second)
		}

		if err := applyTopicRelevanceFilter(result, config.Filters.TopicRelevance, config.Filters.LLM); err != nil {
			return fmt.Errorf("topic relevance filter error: %v", err)
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

	// Create lowercase map for field lookups while preserving original names
	headerMap := make(map[string]string) // lowercase -> original
	for _, col := range header {
		headerMap[strings.ToLower(col)] = col
	}

	// Find column indices (case-insensitive)
	textIdx := -1
	textColumnLower := strings.ToLower(textColumn)
	for i, col := range header {
		if strings.ToLower(col) == textColumnLower {
			textIdx = i
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
			OriginalData:  make(map[string]string),
			LowerFieldMap: make(map[string]string),
			Tags:          make(map[string]interface{}),
			Include:       true, // Initially include all records
		}

		// Use row number as unique internal ID
		manuscript.ID = fmt.Sprintf("%d", recordNum)

		// Store all original data
		for i, value := range record {
			if i < len(header) {
				manuscript.OriginalData[header[i]] = value
				// Store lowercase mapping for case-insensitive lookups
				manuscript.LowerFieldMap[strings.ToLower(header[i])] = header[i]
			}
		}

		// Get text content (could be a path to file or actual text)
		if textIdx < len(record) {
			textContent := record[textIdx]
			// Check if it's a file path
			if fileExists(textContent) {
				content, err := os.ReadFile(textContent)
				if err != nil {
					logger.Error("Could not read file %s: %v", textContent, err)
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
	scanner := bufio.NewScanner(file)

	// Read header
	if !scanner.Scan() {
		return nil, fmt.Errorf("error reading TSV header")
	}
	header := strings.Split(scanner.Text(), "\t")

	// Create lowercase map for field lookups while preserving original names
	headerMap := make(map[string]string) // lowercase -> original
	for _, col := range header {
		headerMap[strings.ToLower(col)] = col
	}

	// Find column indices (case-insensitive)
	textIdx := -1
	textColumnLower := strings.ToLower(textColumn)
	for i, col := range header {
		colLower := strings.ToLower(col)
		if colLower == textColumnLower {
			textIdx = i
		}
	}

	if textIdx == -1 {
		return nil, fmt.Errorf("text column '%s' not found in TSV", textColumn)
	}

	// Read records
	var manuscripts []ManuscriptRecord
	recordNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Split(line, "\t")

		recordNum++

		// Create manuscript record
		manuscript := ManuscriptRecord{
			OriginalData:  make(map[string]string),
			LowerFieldMap: make(map[string]string),
			Tags:          make(map[string]interface{}),
			Include:       true, // Initially include all records
		}

		// Use row number as unique internal ID
		manuscript.ID = fmt.Sprintf("%d", recordNum)

		// Store all original data
		for i, value := range values {
			if i < len(header) {
				manuscript.OriginalData[header[i]] = value
				// Store lowercase mapping for case-insensitive lookups
				manuscript.LowerFieldMap[strings.ToLower(header[i])] = header[i]
			}
		}

		// Set text (abstract)
		if textIdx >= 0 && textIdx < len(values) {
			manuscript.Text = values[textIdx]
		}

		// Try to read the text from file if it's a path
		if manuscript.Text != "" && fileExists(manuscript.Text) {
			content, err := os.ReadFile(manuscript.Text)
			if err == nil {
				manuscript.Text = string(content)
			}
		}

		manuscripts = append(manuscripts, manuscript)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading TSV file: %v", err)
	}

	return manuscripts, nil
}

// applyDeduplicationFilter applies deduplication logic
func applyDeduplicationFilter(result *ScreeningResult, config DeduplicationConfig) error {
	// Convert ManuscriptRecord to filters.ManuscriptData
	manuscripts := make([]filters.ManuscriptData, len(result.Records))
	for i, record := range result.Records {
		manuscripts[i] = filters.ManuscriptData{
			ID:            record.ID,
			OriginalData:  record.OriginalData,
			LowerFieldMap: record.LowerFieldMap,
			Text:          record.Text,
		}
	}

	// Convert config to filters.DeduplicationConfig
	filterConfig := filters.DeduplicationConfig{
		UseAI:         config.UseAI,
		CompareFields: config.CompareFields,
		LLMConfigs:    convertLLMConfigs(result.LLMConfigs),
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

	if config.UseAI && len(llmConfigs) > 0 {
		// Batch AI processing for all included records
		var manuscriptsToProcess []map[string]string
		var recordIndices []int

		// Collect all included records for batch processing
		for i := range result.Records {
			if result.Records[i].Include {
				manuscriptsToProcess = append(manuscriptsToProcess, result.Records[i].OriginalData)
				recordIndices = append(recordIndices, i)
			}
		}

		if len(manuscriptsToProcess) > 0 {
			// Convert LLMConfig to interface{} for filter function
			llmInterfaces := convertLLMConfigs(llmConfigs)

			// Batch process all manuscripts
			logger.Info("Processing %d manuscripts for language detection with AI", len(manuscriptsToProcess))
			languages := filters.BatchDetectLanguagesWithAI(manuscriptsToProcess, llmInterfaces)

			// Apply results
			for idx, recordIdx := range recordIndices {
				language := languages[fmt.Sprintf("%d", idx)]

				// Store AI detection result
				if language != "" && language != "unknown" {
					result.Records[recordIdx].Tags["ai_detected_language"] = language
				}
				result.Records[recordIdx].Tags["detected_language"] = language

				// Check if language is accepted
				isAccepted := false
				for _, acceptedLang := range config.AcceptedLanguages {
					if strings.EqualFold(language, acceptedLang) {
						isAccepted = true
						break
					}
				}

				if !isAccepted && language != "unknown" {
					result.Records[recordIdx].Include = false
					result.Records[recordIdx].ExclusionReason = fmt.Sprintf("Language not accepted: %s", language)
					excludedCount++
				}
			}
		}
	} else {
		// Non-AI: Process each record individually
		for i := range result.Records {
			if !result.Records[i].Include {
				continue // Skip already excluded records
			}

			var language string

			// Try to get title from original data
			titleText := ""
			for field, value := range result.Records[i].OriginalData {
				if strings.ToLower(field) == "title" {
					titleText = value
					break
				}
			}

			// Get abstract text (from Text field which is populated from text_column)
			abstractText := result.Records[i].Text

			var titleLang, abstractLang string

			// Detect title language if available
			if titleText != "" {
				titleLang, _ = filters.DetectLanguage(titleText)
			}

			// Detect abstract language if available
			if abstractText != "" {
				abstractLang, _ = filters.DetectLanguage(abstractText)
			}

			// Prioritize title language (as many journals translate abstracts to English)
			if titleLang != "" && titleLang != "unknown" {
				language = titleLang
			} else if abstractLang != "" && abstractLang != "unknown" {
				language = abstractLang
			} else {
				language = "unknown"
			}

			// Store both detected languages for transparency
			if titleLang != "" {
				result.Records[i].Tags["title_language"] = titleLang
			}
			if abstractLang != "" {
				result.Records[i].Tags["abstract_language"] = abstractLang
			}

			// Store detected language
			result.Records[i].Tags["detected_language"] = language

			// Check if language is accepted
			isAccepted := false
			for _, acceptedLang := range config.AcceptedLanguages {
				if strings.EqualFold(language, acceptedLang) {
					isAccepted = true
					break
				}
			}

			if !isAccepted && language != "unknown" {
				result.Records[i].Include = false
				result.Records[i].ExclusionReason = fmt.Sprintf("Language not accepted: %s", language)
				excludedCount++
			}
		}
	}

	result.Statistics["language_excluded"] = excludedCount
	return nil
}

// applyArticleTypeFilter applies article type classification
func applyArticleTypeFilter(result *ScreeningResult, config ArticleTypeConfig, llmConfigs []LLMConfig) error {
	excludedCount := 0

	if config.UseAI && len(llmConfigs) > 0 {
		// Batch AI processing for all included records
		var manuscriptsToProcess []map[string]string
		var recordIndices []int

		// Collect all included records for batch processing
		for i := range result.Records {
			if result.Records[i].Include {
				manuscriptsToProcess = append(manuscriptsToProcess, result.Records[i].OriginalData)
				recordIndices = append(recordIndices, i)
			}
		}

		if len(manuscriptsToProcess) > 0 {
			// Convert LLMConfig to interface{} for filter function
			llmInterfaces := convertLLMConfigs(llmConfigs)

			// Batch process all manuscripts
			logger.Info("Processing %d manuscripts for article type classification with AI", len(manuscriptsToProcess))
			classifications := filters.BatchClassifyArticleTypesWithAI(manuscriptsToProcess, llmInterfaces)

			// Apply results
			for idx, recordIdx := range recordIndices {
				classification := classifications[fmt.Sprintf("%d", idx)]

				// Store complete classification in tags
				result.Records[recordIdx].Tags["article_type"] = classification.PrimaryType
				result.Records[recordIdx].Tags["all_article_types"] = classification.AllTypes
				result.Records[recordIdx].Tags["methodological_types"] = classification.MethodologicalTypes
				result.Records[recordIdx].Tags["scope_types"] = classification.ScopeTypes
				result.Records[recordIdx].Tags["type_scores"] = classification.TypeScores

				// Check exclusion rules against all classified types
				shouldExclude, excludedTypes := checkArticleTypeExclusion(classification, config)

				if shouldExclude {
					result.Records[recordIdx].Include = false
					result.Records[recordIdx].ExclusionReason = fmt.Sprintf("Article type excluded: %s", strings.Join(excludedTypes, ", "))
					excludedCount++
				}
			}
		}
	} else {
		// Non-AI: Process each record individually
		for i := range result.Records {
			if !result.Records[i].Include {
				continue // Skip already excluded records
			}

			// Use rule-based classification with text
			classification, err := filters.ClassifyArticleTypes(result.Records[i].Text, nil)
			if err != nil {
				logger.Error("Article type classification failed for %s: %v", result.Records[i].ID, err)
				continue
			}

			// Store complete classification in tags
			result.Records[i].Tags["article_type"] = classification.PrimaryType
			result.Records[i].Tags["all_article_types"] = classification.AllTypes
			result.Records[i].Tags["methodological_types"] = classification.MethodologicalTypes
			result.Records[i].Tags["scope_types"] = classification.ScopeTypes
			result.Records[i].Tags["type_scores"] = classification.TypeScores

			// Check exclusion rules against all classified types
			shouldExclude, excludedTypes := checkArticleTypeExclusion(classification, config)

			if shouldExclude {
				result.Records[i].Include = false
				result.Records[i].ExclusionReason = fmt.Sprintf("Article type excluded: %s", strings.Join(excludedTypes, ", "))
				excludedCount++
			}
		}
	}

	result.Statistics["article_type_excluded"] = excludedCount
	return nil
}

// checkArticleTypeExclusion checks if an article should be excluded based on its types
func checkArticleTypeExclusion(classification *filters.ArticleClassification, config ArticleTypeConfig) (bool, []string) {
	shouldExclude := false
	excludedTypes := []string{}

	// Check traditional publication type exclusions
	if config.ExcludeReviews {
		if filters.HasAnyArticleType(classification, filters.ReviewArticle, filters.SystematicReview, filters.MetaAnalysis) {
			shouldExclude = true
			excludedTypes = append(excludedTypes, "review")
		}
	}

	if config.ExcludeEditorials && filters.HasArticleType(classification, filters.Editorial) {
		shouldExclude = true
		excludedTypes = append(excludedTypes, "editorial")
	}

	if config.ExcludeLetters && filters.HasArticleType(classification, filters.Letter) {
		shouldExclude = true
		excludedTypes = append(excludedTypes, "letter")
	}

	// Check methodological type exclusions
	if config.ExcludeTheoretical && filters.HasArticleType(classification, filters.TheoreticalPaper) {
		shouldExclude = true
		excludedTypes = append(excludedTypes, "theoretical")
	}

	if config.ExcludeEmpirical && filters.HasArticleType(classification, filters.EmpiricalStudy) {
		shouldExclude = true
		excludedTypes = append(excludedTypes, "empirical")
	}

	if config.ExcludeMethods && filters.HasArticleType(classification, filters.MethodsPaper) {
		shouldExclude = true
		excludedTypes = append(excludedTypes, "methods")
	}

	// Check study scope exclusions
	if config.ExcludeSingleCase && filters.HasArticleType(classification, filters.SingleCaseStudy) {
		shouldExclude = true
		excludedTypes = append(excludedTypes, "single_case")
	}

	if config.ExcludeSample && filters.HasArticleType(classification, filters.SampleStudy) {
		shouldExclude = true
		excludedTypes = append(excludedTypes, "sample_study")
	}

	return shouldExclude, excludedTypes
}

// applyTopicRelevanceFilter applies topic relevance scoring to filter off-topic manuscripts
func applyTopicRelevanceFilter(result *ScreeningResult, config TopicRelevanceConfig, llmConfigs []LLMConfig) error {
	logger.Info("Applying topic relevance filter...")
	excludedCount := 0

	if config.UseAI && len(llmConfigs) > 0 {
		// Batch AI processing for all included records
		var manuscriptsToProcess []map[string]string
		var recordIndices []int

		// Collect all included records for batch processing
		for i := range result.Records {
			if result.Records[i].Include {
				manuscriptsToProcess = append(manuscriptsToProcess, result.Records[i].OriginalData)
				recordIndices = append(recordIndices, i)
			}
		}

		if len(manuscriptsToProcess) > 0 {
			// Convert LLMConfig to interface{} for filter function
			llmInterfaces := convertLLMConfigs(llmConfigs)

			// Batch process all manuscripts
			logger.Info("Processing %d manuscripts for topic relevance with AI", len(manuscriptsToProcess))
			relevanceScores := filters.BatchCalculateTopicRelevanceWithAI(manuscriptsToProcess, config.Topics, config.MinScore, llmInterfaces)

			// Apply results
			for idx, recordIdx := range recordIndices {
				relevanceScore := relevanceScores[fmt.Sprintf("%d", idx)]

				// Store relevance score in tags
				if result.Records[recordIdx].Tags == nil {
					result.Records[recordIdx].Tags = make(map[string]any)
				}
				result.Records[recordIdx].Tags["topic_relevance_score"] = relevanceScore.OverallScore
				result.Records[recordIdx].Tags["topic_relevance_confidence"] = relevanceScore.Confidence
				result.Records[recordIdx].Tags["matched_keywords"] = relevanceScore.MatchedKeywords
				result.Records[recordIdx].Tags["matched_concepts"] = relevanceScore.MatchedConcepts

				// Apply minimum score threshold
				if relevanceScore.OverallScore < config.MinScore {
					result.Records[recordIdx].Include = false
					result.Records[recordIdx].ExclusionReason = fmt.Sprintf(
						"Topic relevance score (%.2f) below minimum threshold (%.2f)",
						relevanceScore.OverallScore,
						config.MinScore,
					)
					excludedCount++

					logger.Info("Excluded manuscript %s - relevance score: %.2f < %.2f",
						result.Records[recordIdx].ID,
						relevanceScore.OverallScore,
						config.MinScore,
					)
				}
			}
		}
	} else {
		// Non-AI: Process each record individually
		for i := range result.Records {
			if !result.Records[i].Include {
				continue // Skip already excluded records
			}

			// Convert score weights
			weights := filters.ScoreWeights{
				KeywordMatch:   config.ScoreWeights.KeywordMatch,
				ConceptMatch:   config.ScoreWeights.ConceptMatch,
				FieldRelevance: config.ScoreWeights.FieldRelevance,
			}

			relevanceScore, err := filters.CalculateTopicRelevance(
				result.Records[i].OriginalData,
				config.Topics,
				weights,
			)

			if err != nil {
				logger.Error("Failed to calculate topic relevance for record %s: %v", result.Records[i].ID, err)
				// Don't exclude on error, just log and continue
				continue
			}

			// Store relevance score in tags
			if result.Records[i].Tags == nil {
				result.Records[i].Tags = make(map[string]any)
			}
			result.Records[i].Tags["topic_relevance_score"] = relevanceScore.OverallScore
			result.Records[i].Tags["topic_relevance_confidence"] = relevanceScore.Confidence
			result.Records[i].Tags["matched_keywords"] = relevanceScore.MatchedKeywords
			result.Records[i].Tags["matched_concepts"] = relevanceScore.MatchedConcepts

			// Apply minimum score threshold
			if relevanceScore.OverallScore < config.MinScore {
				result.Records[i].Include = false
				result.Records[i].ExclusionReason = fmt.Sprintf(
					"Topic relevance score (%.2f) below minimum threshold (%.2f)",
					relevanceScore.OverallScore,
					config.MinScore,
				)
				excludedCount++

				logger.Info("Excluded manuscript %s - relevance score: %.2f < %.2f",
					result.Records[i].ID,
					relevanceScore.OverallScore,
					config.MinScore,
				)
			}
		}
	}

	result.Statistics["topic_relevance_excluded"] = excludedCount
	logger.Info("Topic relevance filter: excluded %d manuscripts", excludedCount)
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
	allTags := make(map[string]bool)
	if len(result.Records) > 0 {
		// Original columns
		for key := range result.Records[0].OriginalData {
			header = append(header, key)
		}
		// Collect all unique tags from all records
		for _, record := range result.Records {
			for key := range record.Tags {
				allTags[key] = true
			}
		}
		// Add tag columns for all unique tags
		for key := range allTags {
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
	logger.Info("\n=== Screening Summary ===")
	logger.Info("Total Records: %d", result.TotalRecords)
	logger.Info("Included: %d", result.IncludedRecords)
	logger.Info("Excluded: %d", result.ExcludedRecords)

	if logLevel == "medium" || logLevel == "high" {
		logger.Info("\n--- Exclusion Statistics ---")
		for key, value := range result.Statistics {
			logger.Info("%s: %d", key, value)
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
			logger.Info("\nDetailed log saved to: %s", logFile)
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

// convertLLMConfigs converts LLMConfig to interface{} for filter usage
func convertLLMConfigs(configs []LLMConfig) []any {
	result := make([]any, len(configs))
	for i, config := range configs {
		result[i] = map[string]any{
			"provider":    config.Provider,
			"api_key":     config.APIKey,
			"model":       config.Model,
			"temperature": config.Temperature,
			"tpm_limit":   config.TPMLimit,
			"rpm_limit":   config.RPMLimit,
		}
	}
	return result
}
