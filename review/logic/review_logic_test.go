package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
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
	err = Review(mockConfig)
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

	// Expect only the CSV header ("File Name")
	expectedContent := "Provider,Model,File Name\n"
	if string(content) != expectedContent {
		t.Errorf("Expected output file to contain header only, got: %s", string(content))
	}

	// Clean up the output file if it was created
	if err := os.Remove(outputFilePath); err != nil {
		t.Fatalf("Failed to clean up the output file: %v", err)
	}
}
