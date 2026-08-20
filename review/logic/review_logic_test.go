package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/prismaid/review/results"
)

const mockConfigDataTemplate = `
[project]
name = "Test Project"
author = "Test Author"
version = "1.0"

[project.configuration]
input_directory = "%s"
results_file_name = "%s/test_results"
output_format = "csv"
log_level = "low"
duplication = "no"
cot_justification = "no"
summary = "no"

[project.llm]
[project.llm.1]
provider = "OpenAI"
api_key = "test-api-key"
model = "gpt-4o-mini"
temperature = 0.5
tpm_limit = 0
rpm_limit = 0

[prompt]
task = "Map the concepts discussed in the paper."
expected_result = "A JSON object with the requested keys."

[review]
[review.1]
key = "concept"
values = [""]
`

func TestRunReviewWithTempFiles(t *testing.T) {
	// Create a temporary directory for output files
	tmpDir := t.TempDir()

	// Create a mock config string (TOML configuration)
	mockConfig := fmt.Sprintf(mockConfigDataTemplate, tmpDir, tmpDir)

	// Create a temporary file to simulate stdin user input
	inputFile, err := os.CreateTemp("", "input_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())     // Clean up
	_, err = inputFile.WriteString("n\n") // Simulate 'n' response
	if err != nil {
		t.Fatalf("Failed to write to temp input file: %v", err)
	}
	if _, err := inputFile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek input file: %v", err)
	}

	// Backup the original stdin and defer restoring it
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }() // Restore os.Stdin after the test

	// Redirect stdin to our input file
	os.Stdin = inputFile

	// Mock the exit function
	exitCode := 0
	exitFunc = func(code int) {
		exitCode = code
	}

	// Run the workflow by passing the TOML configuration string directly
	_, err = Review(mockConfig)
	if err != nil {
		t.Fatalf("RunReview failed: %v", err)
	}

	// Ensure the process was terminated with exit code 0
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", exitCode)
	}

	// Check that the output file was created
	outputFilePath := filepath.Join(tmpDir, "test_results.csv")
	if _, err := os.Stat(outputFilePath); err != nil {
		t.Fatalf("Expected output file to be created, but it was not found: %v", err)
	}

	// Read the content of the output file to ensure it's just the header
	content, err := os.ReadFile(outputFilePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Expect only the CSV header (base columns plus the configured review key)
	expectedContent := "Provider,Model,File Name,concept\n"
	if string(content) != expectedContent {
		t.Errorf("Expected output file to contain header only, got: %s", string(content))
	}

	// Clean up the output file if it was created
	if err := os.Remove(outputFilePath); err != nil {
		t.Fatalf("Failed to clean up the output file: %v", err)
	}
}

func TestReviewWritesPartialOutputAndChunkingReport(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	if err := os.Mkdir(inputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "good.txt"), []byte("GOOD_DOCUMENT"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "bad.txt"), []byte(strings.Repeat("BROKEN_DOCUMENT ", 200)), 0o600); err != nil {
		t.Fatal(err)
	}
	configuration := fmt.Sprintf(`
[project]
name = "Partial output test"
author = "Test"
version = "1.0"

[project.configuration]
input_directory = %q
results_file_name = %q
output_format = "csv"
log_level = "low"
duplication = "no"
cot_justification = "no"
summary = "no"

[project.configuration.chunking]
enabled = true
input_context_tokens = 500
overlap_tokens = 0

[project.configuration.chunking.merge.status]
rule = "ordinal"
order = ["no", "yes"]

[project.llm.1]
provider = "SelfHosted"
model = "test-model"
base_url = "http://example.invalid"

[prompt]
task = "Extract status"
expected_result = "Return JSON"

[review.1]
key = "status"
values = ["no", "yes"]
`, inputDir, filepath.Join(tmpDir, "results"))

	originalExtract := extractReview
	defer func() { extractReview = originalExtract }()
	extractReview = func(input string) (string, error) {
		var parsed definitions.Input
		if err := json.Unmarshal([]byte(input), &parsed); err != nil {
			return "", err
		}
		output := definitions.Output{Metadata: definitions.OutputMetadata{SchemaVersion: "v1"}}
		for _, prompt := range parsed.Prompts {
			if prompt.SequenceNumber != 1 {
				continue
			}
			response := `{"status":"yes"}`
			if strings.Contains(prompt.PromptContent, "BROKEN_DOCUMENT") {
				response = "not json"
			}
			output.Responses = append(output.Responses, definitions.Response{
				Provider: "SelfHosted", Model: "test-model", SequenceID: prompt.SequenceID,
				SequenceNumber: 1, ModelResponses: []string{response},
			})
		}
		encoded, err := json.Marshal(output)
		return string(encoded), err
	}

	result, err := Review(configuration)
	var partial *PartialReviewError
	if !errors.As(err, &partial) {
		t.Fatalf("expected partial review error, got result=%#v err=%v", result, err)
	}
	if result == nil || result.ManuscriptsProcessed != 2 || result.ManuscriptsSucceeded != 1 || result.ManuscriptsFailed != 1 {
		t.Fatalf("unexpected partial result: %#v", result)
	}
	csvData, err := os.ReadFile(result.OutputFile)
	if err != nil {
		t.Fatalf("expected preserved results file: %v", err)
	}
	if !strings.Contains(string(csvData), "good") || strings.Contains(string(csvData), "bad") {
		t.Fatalf("expected only good document in partial CSV:\n%s", csvData)
	}
	reportData, err := os.ReadFile(result.ChunkingReportFile)
	if err != nil {
		t.Fatalf("expected chunking report: %v", err)
	}
	var report results.ChunkingReport
	if err := json.Unmarshal(reportData, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Documents) != 2 || len(report.Documents[0].Failures) == 0 {
		t.Fatalf("expected failure recorded for first (bad) document, got %#v", report)
	}
}

// TestValidateConfig verifies that ValidateConfig accepts a complete review
// configuration and rejects configurations missing required fields.
func TestValidateConfig(t *testing.T) {
	valid := `
[project]
name = "Test"
[project.configuration]
input_directory = "/tmp/in"
results_file_name = "/tmp/out/results"
[project.llm]
[project.llm.1]
provider = "OpenAI"
model = "gpt-4o-mini"
[prompt]
task = "Map the concepts discussed in the paper."
expected_result = "A JSON object with the requested keys."
[review]
[review.1]
key = "interest rate"
values = [""]
`
	if err := ValidateConfig(valid); err != nil {
		t.Fatalf("expected valid review config, got error: %v", err)
	}

	invalid := []struct {
		name string
		toml string
	}{
		{"malformed toml", "[project]\nname"},
		{"missing prompt.task", `
[project.configuration]
input_directory = "/tmp/in"
results_file_name = "/tmp/out/results"
[project.llm.1]
provider = "OpenAI"
[prompt]
expected_result = "json"
[review.1]
key = "k"
`},
		{"missing review items", `
[project.configuration]
input_directory = "/tmp/in"
results_file_name = "/tmp/out/results"
[project.llm.1]
provider = "OpenAI"
[prompt]
task = "do"
expected_result = "json"
`},
		{"missing input_directory", `
[project.configuration]
results_file_name = "/tmp/out/results"
[project.llm.1]
provider = "OpenAI"
[prompt]
task = "do"
expected_result = "json"
[review.1]
key = "k"
`},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateConfig(tc.toml); err == nil {
				t.Fatalf("expected validation error, got nil")
			}
		})
	}
}
