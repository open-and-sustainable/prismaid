package results

import (
	"os"
	"strings"
	"testing"
)

func TestStartJSONArray(t *testing.T) {
    outputFile, err := os.CreateTemp("", "json_start")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(outputFile.Name()) // Clean up

    err = startJSONArray(outputFile)
    if err != nil {
        t.Fatalf("Failed to start JSON array: %v", err)
    }

    // Check the contents of the file
    content, err := os.ReadFile(outputFile.Name())
    if err != nil {
        t.Fatalf("Failed to read temp file: %v", err)
    }

    expectedContent := "[\n"
    if string(content) != expectedContent {
        t.Errorf("Expected %q, got %q", expectedContent, string(content))
    }
}

func TestWriteJSONData(t *testing.T) {
    outputFile, err := os.CreateTemp("", "json_data")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(outputFile.Name()) // Clean up

    response := `{"key": "value"}`
    filename := "testfile"

    writeJSONData(response, filename, outputFile)

    // Check the contents of the file
    content, err := os.ReadFile(outputFile.Name())
    if err != nil {
        t.Fatalf("Failed to read temp file: %v", err)
    }

    // Check for modifications like filename addition in JSON
    if !strings.Contains(string(content), `"filename": "testfile"`) {
        t.Errorf("JSON data does not contain expected filename entry")
    }
}

func TestWriteCommaAndCloseJSONArray(t *testing.T) {
    outputFile, err := os.CreateTemp("", "json_comma_close")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(outputFile.Name()) // Clean up

    // Start array
    startJSONArray(outputFile)

    // Write some data
    writeJSONData(`{"key": "value"}`, "testfile", outputFile)

    // Write a comma
    writeCommaInJSONArray(outputFile)

    // Write another data element
    writeJSONData(`{"key2": "value2"}`, "testfile2", outputFile)

    // Close the array
    closeJSONArray(outputFile)

    // Check the contents of the file
    content, err := os.ReadFile(outputFile.Name())
    if err != nil {
        t.Fatalf("Failed to read temp file: %v", err)
    }

    expectedContent := `[
{
    "filename": "testfile",
    "key": "value"
},
{
    "filename": "testfile2",
    "key2": "value2"
}
]
`
    if strings.TrimSpace(string(content)) != strings.TrimSpace(expectedContent) {
        t.Errorf("Expected %q, got %q", expectedContent, string(content))
    }
}
