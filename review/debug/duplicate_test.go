package debug

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-and-sustainable/prismaid/review/config"
)

func TestDuplicateInput(t *testing.T) {
	// Step 1: Create a temporary directory
	tempDir, err := os.MkdirTemp("", "duplicate_input_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	// Ensure the temporary directory is removed after the test
	defer os.RemoveAll(tempDir)

	// Step 2: Create test .txt files in the temporary directory
	fileContents := map[string]string{
		"file1.txt": "Content of file 1",
		"file2.txt": "Content of file 2",
	}
	for fileName, content := range fileContents {
		filePath := filepath.Join(tempDir, fileName)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
	}

	// Step 3: Set up the configuration
	cfg := &config.Config{
		Project: config.ProjectConfig{
			Configuration: config.ProjectConfiguration{
				InputDirectory: tempDir,
			},
		},
	}

	// Step 4: Call DuplicateInput
	err = DuplicateInput(cfg)
	if err != nil {
		t.Fatalf("DuplicateInput returned an error: %v", err)
	}

	// Step 5: Verify that duplicated files are created
	expectedFiles := []string{"file1.txt", "file2.txt", "file1_duplicate.txt", "file2_duplicate.txt"}
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	actualFileNames := make(map[string]bool)
	for _, file := range files {
		actualFileNames[file.Name()] = true
	}

	for _, expectedFile := range expectedFiles {
		if !actualFileNames[expectedFile] {
			t.Errorf("Expected file %s was not found in the directory", expectedFile)
		}
	}

	// Step 6: Verify contents of duplicated files
	for originalFile, content := range fileContents {
		duplicatedFile := strings.TrimSuffix(originalFile, ".txt") + "_duplicate.txt"
		duplicatedFilePath := filepath.Join(tempDir, duplicatedFile)

		duplicatedContent, err := os.ReadFile(duplicatedFilePath)
		if err != nil {
			t.Errorf("Failed to read duplicated file %s: %v", duplicatedFile, err)
			continue
		}

		if string(duplicatedContent) != content {
			t.Errorf("Content mismatch in duplicated file %s. Expected: %s, Got: %s", duplicatedFile, content, string(duplicatedContent))
		}
	}
}
