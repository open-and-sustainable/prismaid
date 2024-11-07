package convert

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

)

func TestConvertHTML(t *testing.T) {
    // Step 1: Create a temporary directory
    tempDir, err := os.MkdirTemp("", "convert_test")
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
    <p>This is a <strong>test</strong> HTML file.</p>
</body>
</html>`
    htmlFilePath := filepath.Join(tempDir, "testfile.html")
    err = os.WriteFile(htmlFilePath, []byte(htmlContent), 0644)
    if err != nil {
        t.Fatalf("Failed to write test HTML file: %v", err)
    }

    // Step 4: Run the Convert function
    err = Convert(tempDir, "html")
    if err != nil {
        t.Errorf("Convert returned an error: %v", err)
    }

    // Step 5: Verify the output
    txtFilePath := filepath.Join(tempDir, "testfile.txt")
    content, err := os.ReadFile(txtFilePath)
    if err != nil {
        t.Errorf("Expected output file %s does not exist: %v", txtFilePath, err)
    } else {
        // Check that the content contains the expected text
        expectedText := "This is a test. HTML file."
        actualText := strings.TrimSpace(string(content))
        if !strings.Contains(actualText, expectedText) {
            t.Errorf("Converted text does not contain expected content.\nExpected to find: %s\nActual content: %s", expectedText, actualText)
        }
    }

    // Step 6: Clean-up is handled by defer os.RemoveAll(tempDir)
}
