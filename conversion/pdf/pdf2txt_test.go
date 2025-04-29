package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPdf(t *testing.T) {
	// Look for a test PDF file in the testdata directory
	testFilePath := filepath.Join("testdata", "sample.pdf")

	// Skip this test if the test file doesn't exist
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Skip("Test PDF file not found at testdata/sample.pdf. Skipping test.")
	}

	// Call the ReadPdf function
	result, err := ReadPdf(testFilePath)
	if err != nil {
		t.Errorf("ReadPdf returned an error: %v", err)
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

// Test for empty or corrupted PDF
func TestReadPdfWithInvalidFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "readpdf_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an empty file that's not a valid PDF
	invalidFilePath := filepath.Join(tempDir, "invalid.pdf")
	err = os.WriteFile(invalidFilePath, []byte("This is not a PDF file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid PDF file: %v", err)
	}

	// Call ReadPdf on the invalid file, expect an error
	_, err = ReadPdf(invalidFilePath)
	if err == nil {
		t.Errorf("ReadPdf did not return an error for an invalid PDF file")
	}
}

// Test with environment variable for dynamic testing
func TestReadPdfFromPath(t *testing.T) {
	// This test can be used with a command-line specified file path
	// Example: go test -v -run TestReadPdfFromPath -testpdf=/path/to/your/test.pdf

	testFilePath := os.Getenv("TEST_PDF_PATH")
	if testFilePath == "" {
		t.Skip("No test file specified. Set TEST_PDF_PATH environment variable to run this test.")
	}

	// Call the ReadPdf function
	result, err := ReadPdf(testFilePath)
	if err != nil {
		t.Errorf("ReadPdf returned an error: %v", err)
		return
	}

	// Just verify we got some content back
	if len(strings.TrimSpace(result)) == 0 {
		t.Errorf("ReadPdf returned empty content")
	} else {
		t.Logf("Successfully extracted %d characters from %s", len(result), testFilePath)
		// Print first 100 chars of result for debugging
		if len(result) > 100 {
			t.Logf("First 100 chars: %s", result[:100])
		} else {
			t.Logf("Content: %s", result)
		}
	}
}

// Test fallback mechanism
func TestReadPdfFallback(t *testing.T) {
	// This test would ideally use a PDF that triggers the fallback mechanism
	// Such a PDF might have content that isn't extractable by the primary method

	testFilePath := filepath.Join("testdata", "requires_fallback.pdf")

	// Skip this test if the test file doesn't exist
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Skip("Test PDF file not found at testdata/requires_fallback.pdf. Skipping test.")
	}

	// Call the ReadPdf function
	result, err := ReadPdf(testFilePath)
	if err != nil {
		t.Errorf("ReadPdf returned an error: %v", err)
		return
	}

	// Verify we got some text back
	if len(strings.TrimSpace(result)) == 0 {
		t.Errorf("ReadPdf fallback mechanism returned empty content")
	}
}
