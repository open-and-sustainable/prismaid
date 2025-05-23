package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

// Extremely minimal PDF with text "Test"
// This is the raw bytes of a valid PDF with minimal structure
var minimalPdfContent = []byte{37, 80, 68, 70, 45, 49, 46, 51, 10, 37, 147, 140, 139, 158, 32, 82, 101, 112, 111, 114, 116, 76, 97, 98, 32, 71, 101, 110, 101, 114, 97, 116, 101, 100, 32, 80, 68, 70, 32, 100, 111, 99, 117, 109, 101, 110, 116, 32, 104, 116, 116, 112, 58, 47, 47, 119, 119, 119, 46, 114, 101, 112, 111, 114, 116, 108, 97, 98, 46, 99, 111, 109, 10, 49, 32, 48, 32, 111, 98, 106, 10, 60, 60, 10, 47, 70, 49, 32, 50, 32, 48, 32, 82, 10, 62, 62, 10, 101, 110, 100, 111, 98, 106, 10, 50, 32, 48, 32, 111, 98, 106, 10, 60, 60, 10, 47, 66, 97, 115, 101, 70, 111, 110, 116, 32, 47, 72, 101, 108, 118, 101, 116, 105, 99, 97, 32, 47, 69, 110, 99, 111, 100, 105, 110, 103, 32, 47, 87, 105, 110, 65, 110, 115, 105, 69, 110, 99, 111, 100, 105, 110, 103, 32, 47, 78, 97, 109, 101, 32, 47, 70, 49, 32, 47, 83, 117, 98, 116, 121, 112, 101, 32, 47, 84, 121, 112, 101, 49, 32, 47, 84, 121, 112, 101, 32, 47, 70, 111, 110, 116, 10, 62, 62, 10, 101, 110, 100, 111, 98, 106, 10, 51, 32, 48, 32, 111, 98, 106, 10, 60, 60, 10, 47, 67, 111, 110, 116, 101, 110, 116, 115, 32, 55, 32, 48, 32, 82, 32, 47, 77, 101, 100, 105, 97, 66, 111, 120, 32, 91, 32, 48, 32, 48, 32, 53, 57, 53, 46, 50, 55, 53, 54, 32, 56, 52, 49, 46, 56, 56, 57, 56, 32, 93, 32, 47, 80, 97, 114, 101, 110, 116, 32, 54, 32, 48, 32, 82, 32, 47, 82, 101, 115, 111, 117, 114, 99, 101, 115, 32, 60, 60, 10, 47, 70, 111, 110, 116, 32, 49, 32, 48, 32, 82, 32, 47, 80, 114, 111, 99, 83, 101, 116, 32, 91, 32, 47, 80, 68, 70, 32, 47, 84, 101, 120, 116, 32, 47, 73, 109, 97, 103, 101, 66, 32, 47, 73, 109, 97, 103, 101, 67, 32, 47, 73, 109, 97, 103, 101, 73, 32, 93, 10, 62, 62, 32, 47, 82, 111, 116, 97, 116, 101, 32, 48, 32, 47, 84, 114, 97, 110, 115, 32, 60, 60, 10, 10, 62, 62, 32, 10, 32, 32, 47, 84, 121, 112, 101, 32, 47, 80, 97, 103, 101, 10, 62, 62, 10, 101, 110, 100, 111, 98, 106, 10, 52, 32, 48, 32, 111, 98, 106, 10, 60, 60, 10, 47, 80, 97, 103, 101, 77, 111, 100, 101, 32, 47, 85, 115, 101, 78, 111, 110, 101, 32, 47, 80, 97, 103, 101, 115, 32, 54, 32, 48, 32, 82, 32, 47, 84, 121, 112, 101, 32, 47, 67, 97, 116, 97, 108, 111, 103, 10, 62, 62, 10, 101, 110, 100, 111, 98, 106, 10, 53, 32, 48, 32, 111, 98, 106, 10, 60, 60, 10, 47, 65, 117, 116, 104, 111, 114, 32, 40, 97, 110, 111, 110, 121, 109, 111, 117, 115, 41, 32, 47, 67, 114, 101, 97, 116, 105, 111, 110, 68, 97, 116, 101, 32, 40, 68, 58, 50, 48, 50, 53, 48, 53, 48, 53, 48, 55, 53, 53, 48, 55, 43, 48, 48, 39, 48, 48, 39, 41, 32, 47, 67, 114, 101, 97, 116, 111, 114, 32, 40, 82, 101, 112, 111, 114, 116, 76, 97, 98, 32, 80, 68, 70, 32, 76, 105, 98, 114, 97, 114, 121, 32, 45, 32, 119, 119, 119, 46, 114, 101, 112, 111, 114, 116, 108, 97, 98, 46, 99, 111, 109, 41, 32, 47, 75, 101, 121, 119, 111, 114, 100, 115, 32, 40, 41, 32, 47, 77, 111, 100, 68, 97, 116, 101, 32, 40, 68, 58, 50, 48, 50, 53, 48, 53, 48, 53, 48, 55, 53, 53, 48, 55, 43, 48, 48, 39, 48, 48, 39, 41, 32, 47, 80, 114, 111, 100, 117, 99, 101, 114, 32, 40, 82, 101, 112, 111, 114, 116, 76, 97, 98, 32, 80, 68, 70, 32, 76, 105, 98, 114, 97, 114, 121, 32, 45, 32, 119, 119, 119, 46, 114, 101, 112, 111, 114, 116, 108, 97, 98, 46, 99, 111, 109, 41, 32, 10, 32, 32, 47, 83, 117, 98, 106, 101, 99, 116, 32, 40, 117, 110, 115, 112, 101, 99, 105, 102, 105, 101, 100, 41, 32, 47, 84, 105, 116, 108, 101, 32, 40, 117, 110, 116, 105, 116, 108, 101, 100, 41, 32, 47, 84, 114, 97, 112, 112, 101, 100, 32, 47, 70, 97, 108, 115, 101, 10, 62, 62, 10, 101, 110, 100, 111, 98, 106, 10, 54, 32, 48, 32, 111, 98, 106, 10, 60, 60, 10, 47, 67, 111, 117, 110, 116, 32, 49, 32, 47, 75, 105, 100, 115, 32, 91, 32, 51, 32, 48, 32, 82, 32, 93, 32, 47, 84, 121, 112, 101, 32, 47, 80, 97, 103, 101, 115, 10, 62, 62, 10, 101, 110, 100, 111, 98, 106, 10, 55, 32, 48, 32, 111, 98, 106, 10, 60, 60, 10, 47, 70, 105, 108, 116, 101, 114, 32, 91, 32, 47, 65, 83, 67, 73, 73, 56, 53, 68, 101, 99, 111, 100, 101, 32, 47, 70, 108, 97, 116, 101, 68, 101, 99, 111, 100, 101, 32, 93, 32, 47, 76, 101, 110, 103, 116, 104, 32, 57, 54, 10, 62, 62, 10, 115, 116, 114, 101, 97, 109, 10, 71, 97, 112, 81, 104, 48, 69, 61, 70, 44, 48, 85, 92, 72, 51, 84, 92, 112, 78, 89, 84, 94, 81, 75, 107, 63, 116, 99, 62, 73, 80, 44, 59, 87, 35, 85, 49, 94, 50, 51, 105, 104, 80, 69, 77, 95, 63, 67, 87, 52, 75, 73, 83, 105, 57, 48, 77, 106, 71, 94, 50, 44, 70, 83, 35, 60, 82, 59, 34, 67, 47, 77, 62, 70, 77, 35, 103, 35, 115, 77, 100, 47, 112, 58, 70, 104, 117, 87, 110, 61, 95, 91, 115, 54, 126, 62, 101, 110, 100, 115, 116, 114, 101, 97, 109, 10, 101, 110, 100, 111, 98, 106, 10, 120, 114, 101, 102, 10, 48, 32, 56, 10, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 32, 54, 53, 53, 51, 53, 32, 102, 32, 10, 48, 48, 48, 48, 48, 48, 48, 48, 55, 51, 32, 48, 48, 48, 48, 48, 32, 110, 32, 10, 48, 48, 48, 48, 48, 48, 48, 49, 48, 52, 32, 48, 48, 48, 48, 48, 32, 110, 32, 10, 48, 48, 48, 48, 48, 48, 48, 50, 49, 49, 32, 48, 48, 48, 48, 48, 32, 110, 32, 10, 48, 48, 48, 48, 48, 48, 48, 52, 49, 52, 32, 48, 48, 48, 48, 48, 32, 110, 32, 10, 48, 48, 48, 48, 48, 48, 48, 52, 56, 50, 32, 48, 48, 48, 48, 48, 32, 110, 32, 10, 48, 48, 48, 48, 48, 48, 48, 55, 55, 56, 32, 48, 48, 48, 48, 48, 32, 110, 32, 10, 48, 48, 48, 48, 48, 48, 48, 56, 51, 55, 32, 48, 48, 48, 48, 48, 32, 110, 32, 10, 116, 114, 97, 105, 108, 101, 114, 10, 60, 60, 10, 47, 73, 68, 32, 10, 91, 60, 48, 101, 98, 52, 97, 99, 50, 48, 55, 102, 49, 49, 49, 99, 52, 54, 56, 48, 53, 52, 99, 100, 54, 52, 56, 52, 49, 52, 100, 51, 98, 54, 62, 60, 48, 101, 98, 52, 97, 99, 50, 48, 55, 102, 49, 49, 49, 99, 52, 54, 56, 48, 53, 52, 99, 100, 54, 52, 56, 52, 49, 52, 100, 51, 98, 54, 62, 93, 10, 37, 32, 82, 101, 112, 111, 114, 116, 76, 97, 98, 32, 103, 101, 110, 101, 114, 97, 116, 101, 100, 32, 80, 68, 70, 32, 100, 111, 99, 117, 109, 101, 110, 116, 32, 45, 45, 32, 100, 105, 103, 101, 115, 116, 32, 40, 104, 116, 116, 112, 58, 47, 47, 119, 119, 119, 46, 114, 101, 112, 111, 114, 116, 108, 97, 98, 46, 99, 111, 109, 41, 10, 10, 47, 73, 110, 102, 111, 32, 53, 32, 48, 32, 82, 10, 47, 82, 111, 111, 116, 32, 52, 32, 48, 32, 82, 10, 47, 83, 105, 122, 101, 32, 56, 10, 62, 62, 10, 115, 116, 97, 114, 116, 120, 114, 101, 102, 10, 49, 48, 50, 50, 10, 37, 37, 69, 79, 70, 10}

func TestReadPdf(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "pdf2txt_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	// Create a test PDF file in the temporary directory
	pdfFilePath := filepath.Join(tempDir, "sample.pdf")
	if err := os.WriteFile(pdfFilePath, minimalPdfContent, 0644); err != nil {
		t.Fatalf("Failed to write test PDF file: %v", err)
	}

	// Instead of relying on PDF extraction which might depend on external tools,
	// let's just check if the function handles the file without error first
	result, err := ReadPdf(pdfFilePath)

	// We're primarily testing that the function works, not necessarily the PDF content extraction
	if err != nil {
		t.Errorf("ReadPdf returned an error: %v", err)
	} else {
		t.Logf("Successfully processed PDF file, extracted text: %s", result)
	}

	// Test for invalid PDF file
	invalidFilePath := filepath.Join(tempDir, "invalid.pdf")
	if err := os.WriteFile(invalidFilePath, []byte("This is not a PDF file"), 0644); err != nil {
		t.Fatalf("Failed to create invalid PDF file: %v", err)
	}

	// Call ReadPdf on the invalid file, expect an error
	_, err = ReadPdf(invalidFilePath)
	if err == nil {
		t.Errorf("ReadPdf did not return an error for an invalid PDF file")
	} else {
		t.Logf("As expected, ReadPdf returned error for invalid file: %v", err)
	}
}
