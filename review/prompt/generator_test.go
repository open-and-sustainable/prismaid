package prompt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/open-and-sustainable/alembica/definitions"
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

func TestPreparePlanSplitsOnlyOverContextDocuments(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "long.txt"), []byte("First paragraph is deliberately long enough to exceed the small configured limit.\n\nSecond paragraph is also deliberately long enough to require splitting."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "short.txt"), []byte("short"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Project: config.ProjectConfig{
			Configuration: config.ProjectConfiguration{
				InputDirectory: directory,
				Chunking: config.ChunkingConfig{
					Enabled:            true,
					InputContextTokens: 120,
					Merge: map[string]config.MergeRule{
						"status": {Rule: "ordinal", Order: []string{"no", "yes"}},
					},
				},
			},
			LLM: map[string]config.LLMItem{"1": {Provider: "SelfHosted", Model: "local"}},
		},
		Prompt: config.PromptConfig{Task: "Extract", ExpectedResult: "JSON"},
		Review: map[string]config.ReviewItem{"1": {Key: "status", Values: []string{"no", "yes"}}},
	}

	prepared, err := PreparePlan(cfg)
	if err != nil {
		t.Fatalf("PreparePlan returned an error: %v", err)
	}
	if !prepared.HasChunks {
		t.Fatal("expected the long document to be chunked")
	}
	if len(prepared.Filenames) != 2 {
		t.Fatalf("expected two source filenames, got %d", len(prepared.Filenames))
	}
	var input definitions.Input
	if err := json.Unmarshal([]byte(prepared.JSON), &input); err != nil {
		t.Fatalf("unmarshal prepared input: %v", err)
	}
	if len(input.Prompts) <= 2 {
		t.Fatalf("expected more prompts than documents, got %d", len(input.Prompts))
	}
	if len(prepared.Bindings) != len(input.Prompts) {
		t.Fatalf("expected a binding for every prompt, got %d bindings for %d prompts", len(prepared.Bindings), len(input.Prompts))
	}
	if input.Metadata.SchemaVersion != "v1" {
		t.Fatalf("expected Alembica schema version v1, got %q", input.Metadata.SchemaVersion)
	}
	if len(prepared.ExpectedGroups) != len(input.Models)*len(prepared.Filenames) {
		t.Fatalf("expected one primary response group per model and document, got %d", len(prepared.ExpectedGroups))
	}
	chunkCounts := make(map[string]int)
	for sequenceID, binding := range prepared.Bindings {
		if binding.ChunkCount < 1 || binding.ChunkIndex < 0 || binding.ChunkIndex >= binding.ChunkCount {
			t.Fatalf("invalid binding for sequence %s: %+v", sequenceID, binding)
		}
		chunkCounts[binding.Filename]++
	}
	if chunkCounts["long"] < 2 || chunkCounts["short"] != 1 {
		t.Fatalf("unexpected prompt-to-document chunk mapping: %+v", chunkCounts)
	}
	foundLongReport := false
	for _, report := range prepared.ChunkReports {
		if report.Filename == "long" {
			foundLongReport = true
			if report.FullPromptTokens <= 120 || report.ChunkCount < 2 {
				t.Fatalf("expected an over-limit long document report, got %+v", report)
			}
		}
	}
	if !foundLongReport {
		t.Fatal("missing chunk report for long document")
	}
	for _, prompt := range input.Prompts {
		if len([]byte(prompt.PromptContent)) > 120 {
			t.Fatalf("prompt %s exceeds configured byte fallback limit", prompt.SequenceID)
		}
	}
}
