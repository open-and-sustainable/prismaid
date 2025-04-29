package doc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadDocx(t *testing.T) {
	// Look for a test DOCX file in the testdata directory
	testFilePath := filepath.Join("testdata", "sample.docx")

	// Skip this test if the test file doesn't exist
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Skip("Test DOCX file not found at testdata/sample.docx. Skipping test.")
	}

	// Call the ReadDocx function
	result, err := ReadDocx(testFilePath)
	if err != nil {
		t.Errorf("ReadDocx returned an error: %v", err)
		return
	}

	// Verify the output contains expected text
	// Replace this with text that exists in your sample file
	expectedText := "sample text"
	if !strings.Contains(result, expectedText) {
		t.Errorf("Converted text does not contain expected content.\nExpected to find: %s\nActual content: %s",
			expectedText, result)
	}
}

// For integration testing with an external file
func TestReadDocxFromPath(t *testing.T) {
	// This test can be used with a command-line specified file path
	// Example: go test -v -run TestReadDocxFromPath -testdocx=/path/to/your/test.docx

	testFilePath := os.Getenv("TEST_DOCX_PATH")
	if testFilePath == "" {
		t.Skip("No test file specified. Set TEST_DOCX_PATH environment variable to run this test.")
	}

	// Call the ReadDocx function
	result, err := ReadDocx(testFilePath)
	if err != nil {
		t.Errorf("ReadDocx returned an error: %v", err)
		return
	}

	// Just verify we got some content back
	if len(strings.TrimSpace(result)) == 0 {
		t.Errorf("ReadDocx returned empty content")
	} else {
		t.Logf("Successfully extracted %d characters from %s", len(result), testFilePath)
	}
}
