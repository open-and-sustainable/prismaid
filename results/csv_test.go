package results

import (
	"os"
	"testing"
)

func TestCreateCSVWriter(t *testing.T) {
    // Create a temporary file
    outputFile, err := os.CreateTemp("", "csv_test")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(outputFile.Name()) // Clean up

    // Keys to create headers
    keys := []string{"column1", "column2"}

    // Create CSV writer
    writer := createCSVWriter(outputFile, keys)
    writer.Flush()

    // Reopen the file to check contents
    fileContent, err := os.ReadFile(outputFile.Name())
    if err != nil {
        t.Fatalf("Failed to read temp file: %v", err)
    }

    // Check the headers
    expectedHeader := "Provider,Model,File Name,column1,column2\n"
    if string(fileContent) != expectedHeader {
        t.Errorf("Expected header %q, got %q", expectedHeader, string(fileContent))
    }
}

func TestWriteCSVData(t *testing.T) {
    // Setup
    outputFile, err := os.CreateTemp("", "csv_write_test")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(outputFile.Name()) // Clean up

    keys := []string{"column1", "column2"}
    writer := createCSVWriter(outputFile, keys)

    // Example data
    response := `{"column1": "value1", "column2": "value2"}`
    fileNameWithoutExt := "testfile"
    provider := "TestProvider"
    model := "TestModel"

    // Write data to CSV
    writeCSVData(response, fileNameWithoutExt, provider, model, writer, keys)
    writer.Flush()

    // Reopen the file to check contents
    fileContent, err := os.ReadFile(outputFile.Name())
    if err != nil {
        t.Fatalf("Failed to read temp file: %v", err)
    }

    // Check the contents, including the header and the data row
    expectedContent := "Provider,Model,File Name,column1,column2\n" +
                       "TestProvider,TestModel,testfile,value1,value2\n"
    if string(fileContent) != expectedContent {
        t.Errorf("Expected CSV content %q, got %q", expectedContent, string(fileContent))
    }
}
