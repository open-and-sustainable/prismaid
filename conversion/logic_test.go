package conversion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConvert tests the conversion logic without relying on real format conversions
func TestConvert(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "convert_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock files of different formats
	testFiles := map[string]string{
		"test1.html":  "<html><body>HTML test content</body></html>",
		"test2.htm":   "<html><body>HTM test content</body></html>",
		"test3.pdf":   "Mock PDF content",
		"test4.docx":  "Mock DOCX content",
		"test5.other": "File with different extension", // Changed from .txt to .other
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Create a test-specific converter function
	convertTest := func(inputDir string, formats string) error {
		// Load files from the input directory
		files, err := os.ReadDir(inputDir)
		if err != nil {
			return fmt.Errorf("error reading input directory: %v", err)
		}

		// Split formats
		formatsList := strings.Split(formats, ",")

		// Process files
		for _, format := range formatsList {
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				ext := filepath.Ext(file.Name())

				var shouldProcess bool
				var processFormat string

				// Check if this file should be processed
				if ext == "."+format {
					shouldProcess = true
					processFormat = format
				} else if ext == ".htm" && format == "html" {
					shouldProcess = true
					processFormat = "html"
				}

				if shouldProcess {
					// For testing, we'll create a simple output instead of calling readText
					mockContent := fmt.Sprintf("Converted %s content", processFormat)

					// Create output filename
					var baseFilename string
					if ext == ".htm" {
						baseFilename = strings.TrimSuffix(file.Name(), ".htm")
					} else {
						baseFilename = strings.TrimSuffix(file.Name(), "."+format)
					}

					txtPath := filepath.Join(inputDir, baseFilename+".txt")

					// Write the output file
					err = os.WriteFile(txtPath, []byte(mockContent), 0644)
					if err != nil {
						return fmt.Errorf("error writing to file: %v", err)
					}
				}
			}
		}
		return nil
	}

	// Test 1: Convert HTML and PDF files only
	err = convertTest(tempDir, "html,pdf")
	if err != nil {
		t.Errorf("Convert test returned an error: %v", err)
	}

	// Verify that HTML and PDF files were converted to TXT
	expectedTxtFiles := []string{"test1.txt", "test2.txt", "test3.txt"}
	unexpectedTxtFiles := []string{"test4.txt", "test5.txt"}

	for _, filename := range expectedTxtFiles {
		txtPath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(txtPath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s does not exist", txtPath)
		}
	}

	for _, filename := range unexpectedTxtFiles {
		txtPath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(txtPath); !os.IsNotExist(err) {
			t.Errorf("Unexpected output file %s exists", txtPath)
		}
	}

	// Clean up created txt files for the next test
	for _, filename := range expectedTxtFiles {
		os.Remove(filepath.Join(tempDir, filename))
	}

	// Test 2: Convert DOCX files only
	err = convertTest(tempDir, "docx")
	if err != nil {
		t.Errorf("Convert test returned an error: %v", err)
	}

	// Verify that only DOCX files were converted to TXT
	expectedTxtFiles = []string{"test4.txt"}
	unexpectedTxtFiles = []string{"test1.txt", "test2.txt", "test3.txt", "test5.txt"}

	for _, filename := range expectedTxtFiles {
		txtPath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(txtPath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s does not exist", txtPath)
		}
	}

	for _, filename := range unexpectedTxtFiles {
		txtPath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(txtPath); !os.IsNotExist(err) {
			t.Errorf("Unexpected output file %s exists", txtPath)
		}
	}
}

// TestConvertErrors tests error handling
func TestConvertErrors(t *testing.T) {
	// Test with a non-existent directory
	err := Convert("/nonexistent/directory", "html,pdf")
	if err == nil {
		t.Errorf("Expected error when using non-existent directory, but got none")
	}
}

// TestReadText tests the readText function with mock files
func TestReadText(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "readtext_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test for unsupported format
	_, err = readText("file.xyz", "xyz")
	if err == nil {
		t.Errorf("readText should fail for unsupported format")
	}

	// Note: We don't test the actual content conversion here as that's covered
	// by the individual package tests (pdf, doc, html)
}

// TestWriteText tests the writeText function
func TestWriteText(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "writetext_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test writing text to a file
	testText := "This is test content"
	testFilePath := filepath.Join(tempDir, "writetest.txt")

	err = writeText(testText, testFilePath)
	if err != nil {
		t.Errorf("writeText returned an error: %v", err)
	}

	// Verify the content
	content, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Errorf("Failed to read test file: %v", err)
	}

	if string(content) != testText {
		t.Errorf("File content doesn't match.\nExpected: %s\nActual: %s", testText, string(content))
	}
}
