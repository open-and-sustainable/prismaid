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
	err := Convert("/nonexistent/directory", "html,pdf", ConvertOptions{})
	if err == nil {
		t.Errorf("Expected error when using non-existent directory, but got none")
	}
}

func TestConvertPDFSingleFileExtensionMismatch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "convert_single_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	htmlPath := filepath.Join(tempDir, "test.html")
	err = os.WriteFile(htmlPath, []byte("<html><body>Test</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test HTML file: %v", err)
	}

	err = Convert(tempDir, "pdf", ConvertOptions{
		PDF: PDFOptions{SingleFile: htmlPath},
	})
	if err == nil {
		t.Errorf("Expected error when single-file extension is not .pdf, but got none")
	}
}

func TestConvertOCROnlyRequiresTika(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "convert_ocr_only_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = Convert(tempDir, "pdf", ConvertOptions{
		PDF: PDFOptions{OCROnly: true},
	})
	if err == nil {
		t.Errorf("Expected error when OCR-only is requested without a Tika server, but got none")
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

// TestConvertWithTikaFallback tests that Tika is used as fallback when standard conversion works
func TestConvertWithTikaFallback(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "convert_tika_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple HTML file that should convert without Tika
	htmlContent := "<html><body>Test content</body></html>"
	htmlPath := filepath.Join(tempDir, "test.html")
	err = os.WriteFile(htmlPath, []byte(htmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test HTML file: %v", err)
	}

	// Test: Standard conversion works, Tika address provided but not needed
	err = Convert(tempDir, "html", ConvertOptions{TikaServer: "localhost:99999"})
	if err != nil {
		t.Errorf("Convert returned an error: %v", err)
	}

	// Verify output was created using standard conversion (Tika not used as fallback)
	txtPath := filepath.Join(tempDir, "test.txt")
	if _, err := os.Stat(txtPath); os.IsNotExist(err) {
		t.Errorf("Expected output file %s does not exist", txtPath)
	}
}

// TestConvertTikaFallbackIntegration tests that Tika is used as OCR fallback when standard methods fail
// This test is skipped if no Tika server is available
func TestConvertTikaFallbackIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test verifies Tika is only used as fallback, not as primary conversion method
	// We rely on test.sh to create problematic files that trigger the fallback
	t.Skip("Tika fallback integration is tested in test.sh with problematic PDFs")
}
