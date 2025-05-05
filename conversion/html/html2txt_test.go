package html

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadHtml(t *testing.T) {
	// Step 1: Create a temporary directory
	tempDir, err := os.MkdirTemp("", "readhtml_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	// Clean up the temporary directory after the test
	defer os.RemoveAll(tempDir)

	// Step 2: Create a fake HTML file in the temporary directory
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Test HTML File</title>
</head>
<body>
    <p>This is a test HTML file.</p>
</body>
</html>`
	htmlFilePath := filepath.Join(tempDir, "testfile.html")
	err = os.WriteFile(htmlFilePath, []byte(htmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test HTML file: %v", err)
	}

	// Step 3: Call the ReadHtml function
	result, err := ReadHtml(htmlFilePath)
	if err != nil {
		t.Errorf("ReadHtml returned an error: %v", err)
		return
	}

	// Step 4: Verify the output
	expectedText := "This is a test HTML file."
	if !strings.Contains(result, expectedText) {
		t.Errorf("Converted text does not contain expected content.\nExpected to find: %s\nActual content: %s", expectedText, result)
	}
}
