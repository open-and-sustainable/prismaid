package prompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/open-and-sustainable/prismaid/review/config"
)

func TestParsePrompts(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Prompt: config.PromptConfig{
			Persona:        "Sample Persona",
			Task:           "Sample Task",
			ExpectedResult: "Sample Expected Result",
			Failsafe:       "Sample Failsafe",
			Definitions:    "Sample Definitions",
			Example:        "Sample Example",
		},
		Project: config.ProjectConfig{
			Configuration: config.ProjectConfiguration{
				InputDirectory: "test_data",
			},
		},
		Review: map[string]config.ReviewItem{
			"1": {Key: "test", Values: []string{"yes", "no"}},
		},
	}

	// Mock data
	os.Mkdir("test_data", 0755)
	defer os.RemoveAll("test_data")
	file, _ := os.Create(filepath.Join("test_data", "file.txt"))
	file.WriteString("Test file content")
	file.Close()

	// Execute
	prompts, filenames := parsePrompts(cfg)

	// Verify
	if len(prompts) == 0 || len(filenames) == 0 {
		t.Errorf("Expected non-empty results, got prompts: %d, filenames: %d", len(prompts), len(filenames))
	}
}

func TestGetReviewKeysByEntryOrder(t *testing.T) {
	cfg := &config.Config{
		Review: map[string]config.ReviewItem{
			"2": {Key: "beta"},
			"1": {Key: "alpha"},
		},
	}

	expected := []string{"1", "2"}
	result := GetReviewKeysByEntryOrder(cfg)
	if len(result) != len(expected) || result[0] != expected[0] || result[1] != expected[1] {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestSortReviewKeysAlphabetically(t *testing.T) {
	cfg := &config.Config{
		Review: map[string]config.ReviewItem{
			"1": {Key: "value2"},
			"2": {Key: "value1"},
		},
	}

	expected := []string{"value1", "value2"} // These should be the alphabetical order of the values, not the keys
	result := SortReviewKeysAlphabetically(cfg)
	if len(result) != len(expected) || result[0] != expected[0] || result[1] != expected[1] {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
