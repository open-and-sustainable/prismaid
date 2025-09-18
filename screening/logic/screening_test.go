package logic

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestScreeningWithBasicConfig tests the screening function with a basic configuration
func TestScreeningWithBasicConfig(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "screening_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test input CSV file
	inputFile := filepath.Join(tempDir, "test_input.csv")
	inputContent := `id,title,abstract
1,"Study on Climate Change","This is a research article about climate change and its effects on global temperatures."
2,"Review of Climate Studies","This is a systematic review of climate change literature published in the last decade."
3,"Editorial: Climate Action","This editorial discusses the urgent need for climate action."
4,"Study on Climate Change","This is a research article about climate change and its effects on global temperatures."
5,"Estudio sobre el cambio climático","Este es un artículo de investigación sobre el cambio climático."
`
	if err := ioutil.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Create test configuration
	outputFile := filepath.Join(tempDir, "test_output")
	config := `
[project]
name = "Test Screening"
author = "Test"
version = "1.0"
input_file = "` + inputFile + `"
output_file = "` + outputFile + `"
text_column = "abstract"
identifier_column = "id"
output_format = "csv"
log_level = "low"

[filters]

[filters.deduplication]
enabled = true
use_ai = false
compare_fields = ["title", "abstract"]

[filters.language]
enabled = true
accepted_languages = ["en"]
use_ai = false

[filters.article_type]
enabled = true
exclude_reviews = true
exclude_editorials = true
exclude_letters = false
`

	// Run screening
	err = Screen(config)
	if err != nil {
		t.Fatalf("Screening failed: %v", err)
	}

	// Check if output file was created
	outputCSV := outputFile + ".csv"
	if _, err := os.Stat(outputCSV); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", outputCSV)
	}

	// Read and verify output
	outputData, err := ioutil.ReadFile(outputCSV)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(outputData)

	// Debug: Print the output for inspection
	t.Logf("Output CSV content:\n%s", outputStr)

	// Check that required columns exist
	if !strings.Contains(outputStr, "tag_detected_language") {
		t.Error("Output should contain language tag column")
	}

	// Check that article type was classified
	if !strings.Contains(outputStr, "tag_article_type") {
		t.Error("Output should contain article type tag column")
	}

	// Check that include/exclusion columns exist
	if !strings.Contains(outputStr, "include") {
		t.Error("Output should contain include column")
	}

	if !strings.Contains(outputStr, "exclusion_reason") {
		t.Error("Output should contain exclusion_reason column")
	}

	// Count the number of lines to verify records were processed
	lines := strings.Split(outputStr, "\n")
	if len(lines) < 3 { // Header + at least 2 data rows
		t.Errorf("Expected at least 3 lines in output, got %d", len(lines))
	}

	// Check for exclusions (we expect some papers to be excluded)
	if !strings.Contains(outputStr, "false") {
		t.Error("Expected some papers to be excluded (include=false)")
	}
}

// TestDeduplicationFilter tests the deduplication functionality
func TestDeduplicationFilter(t *testing.T) {
	// Create test manuscript records
	records := []ManuscriptRecord{
		{
			ID: "1",
			OriginalData: map[string]string{
				"title":    "Climate Change Study",
				"abstract": "Research on global warming",
			},
			Text:    "Research on global warming",
			Tags:    make(map[string]interface{}),
			Include: true,
		},
		{
			ID: "2",
			OriginalData: map[string]string{
				"title":    "Climate Change Study",
				"abstract": "Research on global warming",
			},
			Text:    "Research on global warming",
			Tags:    make(map[string]interface{}),
			Include: true,
		},
		{
			ID: "3",
			OriginalData: map[string]string{
				"title":    "Different Study",
				"abstract": "Research on ocean acidification",
			},
			Text:    "Research on ocean acidification",
			Tags:    make(map[string]interface{}),
			Include: true,
		},
	}

	result := &ScreeningResult{
		TotalRecords: len(records),
		Records:      records,
		Statistics:   make(map[string]int),
	}

	config := DeduplicationConfig{
		Enabled:       true,
		UseAI:         false,
		CompareFields: []string{"title", "abstract"},
	}

	err := applyDeduplicationFilter(result, config)
	if err != nil {
		t.Fatalf("Deduplication filter failed: %v", err)
	}

	// Check that record 2 is marked as duplicate
	if result.Records[1].Include != false {
		t.Error("Duplicate record should be excluded")
	}

	if result.Records[1].Tags["is_duplicate"] != true {
		t.Error("Duplicate record should have is_duplicate tag")
	}

	// Check that records 1 and 3 are not duplicates
	if result.Records[0].Include != true {
		t.Error("Original record should be included")
	}

	if result.Records[2].Include != true {
		t.Error("Unique record should be included")
	}
}

// TestLanguageDetection tests the language detection functionality
func TestLanguageDetection(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{
			text:     "This is an English text about scientific research and methodology.",
			expected: "en",
		},
		{
			text:     "Este es un texto en español sobre investigación científica.",
			expected: "es",
		},
		{
			text:     "Ceci est un texte en français sur la recherche scientifique.",
			expected: "fr",
		},
		{
			text:     "Dies ist ein deutscher Text über wissenschaftliche Forschung.",
			expected: "de",
		},
	}

	for _, test := range tests {
		records := []ManuscriptRecord{
			{
				ID:           "1",
				OriginalData: make(map[string]string),
				Text:         test.text,
				Tags:         make(map[string]interface{}),
				Include:      true,
			},
		}

		result := &ScreeningResult{
			TotalRecords: 1,
			Records:      records,
			Statistics:   make(map[string]int),
		}

		config := LanguageConfig{
			Enabled:           true,
			AcceptedLanguages: []string{"en"},
			UseAI:             false,
		}

		err := applyLanguageFilter(result, config, nil)
		if err != nil {
			t.Fatalf("Language filter failed: %v", err)
		}

		detectedLang, ok := result.Records[0].Tags["detected_language"].(string)
		if !ok {
			t.Error("Language tag should be a string")
			continue
		}

		if detectedLang != test.expected {
			t.Errorf("Expected language %s, got %s for text: %s", test.expected, detectedLang, test.text[:30])
		}
	}
}

// TestArticleTypeClassification tests the article type classification
func TestArticleTypeClassification(t *testing.T) {
	tests := []struct {
		text         string
		expectedType string
	}{
		{
			text:         "Methods: We conducted a randomized controlled trial with 100 participants. Results: The treatment group showed significant improvement (p<0.05).",
			expectedType: "research_article",
		},
		{
			text:         "This systematic review follows PRISMA guidelines. We searched PubMed, Embase, and Cochrane databases. Inclusion criteria were defined as...",
			expectedType: "systematic_review",
		},
		{
			text:         "Editorial: The recent developments in climate science require urgent attention from policymakers.",
			expectedType: "editorial",
		},
		{
			text:         "Dear Editor, We read with interest the recent article by Smith et al. and would like to comment on their methodology.",
			expectedType: "letter",
		},
	}

	for _, test := range tests {
		records := []ManuscriptRecord{
			{
				ID:           "1",
				OriginalData: make(map[string]string),
				Text:         test.text,
				Tags:         make(map[string]interface{}),
				Include:      true,
			},
		}

		result := &ScreeningResult{
			TotalRecords: 1,
			Records:      records,
			Statistics:   make(map[string]int),
		}

		config := ArticleTypeConfig{
			Enabled:           true,
			ExcludeReviews:    false,
			ExcludeEditorials: false,
			ExcludeLetters:    false,
		}

		err := applyArticleTypeFilter(result, config, nil)
		if err != nil {
			t.Fatalf("Article type filter failed: %v", err)
		}

		articleType, ok := result.Records[0].Tags["article_type"].(string)
		if !ok {
			t.Error("Article type tag should be a string")
			continue
		}

		if !strings.Contains(articleType, test.expectedType) {
			t.Errorf("Expected article type containing %s, got %s for text: %s", test.expectedType, articleType, test.text[:50])
		}
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	// Test missing input file
	config := &ScreeningConfig{
		Project: ProjectConfig{
			OutputFile: "output.csv",
			TextColumn: "abstract",
		},
	}

	err := validateConfig(config)
	if err == nil || !strings.Contains(err.Error(), "input_file") {
		t.Error("Should error on missing input_file")
	}

	// Test missing output file
	config = &ScreeningConfig{
		Project: ProjectConfig{
			InputFile:  "input.csv",
			TextColumn: "abstract",
		},
	}

	err = validateConfig(config)
	if err == nil || !strings.Contains(err.Error(), "output_file") {
		t.Error("Should error on missing output_file")
	}

	// Test missing text column
	config = &ScreeningConfig{
		Project: ProjectConfig{
			InputFile:  "input.csv",
			OutputFile: "output.csv",
		},
	}

	err = validateConfig(config)
	if err == nil || !strings.Contains(err.Error(), "text_column") {
		t.Error("Should error on missing text_column")
	}

	// Test no filters enabled
	config = &ScreeningConfig{
		Project: ProjectConfig{
			InputFile:  "input.csv",
			OutputFile: "output.csv",
			TextColumn: "abstract",
		},
		Filters: FiltersConfig{
			Deduplication: DeduplicationConfig{Enabled: false},
			Language:      LanguageConfig{Enabled: false},
			ArticleType:   ArticleTypeConfig{Enabled: false},
		},
	}

	err = validateConfig(config)
	if err == nil || !strings.Contains(err.Error(), "at least one filter") {
		t.Error("Should error when no filters are enabled")
	}

	// Test valid configuration
	config = &ScreeningConfig{
		Project: ProjectConfig{
			InputFile:  "input.csv",
			OutputFile: "output.csv",
			TextColumn: "abstract",
		},
		Filters: FiltersConfig{
			Deduplication: DeduplicationConfig{Enabled: true},
		},
	}

	err = validateConfig(config)
	if err != nil {
		t.Errorf("Valid configuration should not error: %v", err)
	}
}

// TestFuzzyMatching tests simple duplicate detection with single character differences
func TestFuzzyMatching(t *testing.T) {
	records := []ManuscriptRecord{
		{
			ID: "1",
			OriginalData: map[string]string{
				"title": "Climate Change and Global Warming",
			},
			Text:    "Study on climate effects",
			Tags:    make(map[string]interface{}),
			Include: true,
		},
		{
			ID: "2",
			OriginalData: map[string]string{
				"title": "Climate Change and Global Warmings", // Single char difference
			},
			Text:    "Study on climate effects",
			Tags:    make(map[string]interface{}),
			Include: true,
		},
		{
			ID: "3",
			OriginalData: map[string]string{
				"title": "Ocean Acidification Study",
			},
			Text:    "Research on ocean pH levels",
			Tags:    make(map[string]interface{}),
			Include: true,
		},
	}

	result := &ScreeningResult{
		TotalRecords: len(records),
		Records:      records,
		Statistics:   make(map[string]int),
	}

	config := DeduplicationConfig{
		Enabled:       true,
		UseAI:         false,
		CompareFields: []string{"title"},
	}

	err := applyDeduplicationFilter(result, config)
	if err != nil {
		t.Fatalf("Fuzzy deduplication failed: %v", err)
	}

	// Check that titles with single character difference are detected as duplicates
	if result.Records[1].Include != false {
		t.Error("Single character difference duplicate should be excluded")
	}

	// Check that dissimilar record is not marked as duplicate
	if result.Records[2].Include != true {
		t.Error("Dissimilar record should not be marked as duplicate")
	}
}

// TestOutputFormats tests CSV and JSON output generation
func TestOutputFormats(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "output_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	result := &ScreeningResult{
		TotalRecords:    3,
		IncludedRecords: 2,
		ExcludedRecords: 1,
		Records: []ManuscriptRecord{
			{
				ID: "1",
				OriginalData: map[string]string{
					"title": "Study 1",
				},
				Tags: map[string]interface{}{
					"detected_language": "en",
					"article_type":      "research_article",
				},
				Include: true,
			},
			{
				ID: "2",
				OriginalData: map[string]string{
					"title": "Study 2",
				},
				Tags: map[string]interface{}{
					"is_duplicate": true,
					"duplicate_of": "1",
				},
				Include:         false,
				ExclusionReason: "Duplicate of 1",
			},
		},
		Statistics: map[string]int{
			"duplicates_found": 1,
		},
	}

	// Test CSV output
	csvPath := filepath.Join(tempDir, "test_csv")
	err = saveResults(result, csvPath, "csv")
	if err != nil {
		t.Fatalf("Failed to save CSV results: %v", err)
	}

	if _, err := os.Stat(csvPath + ".csv"); os.IsNotExist(err) {
		t.Error("CSV file was not created")
	}

	// Test JSON output
	jsonPath := filepath.Join(tempDir, "test_json")
	err = saveResults(result, jsonPath, "json")
	if err != nil {
		t.Fatalf("Failed to save JSON results: %v", err)
	}

	if _, err := os.Stat(jsonPath + ".json"); os.IsNotExist(err) {
		t.Error("JSON file was not created")
	}
}
