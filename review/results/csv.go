package results

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/open-and-sustainable/alembica/utils/logger"
)

// createCSVWriter initializes and returns a CSV writer for the specified output file.
// The writer generates a header row including provider, model, file name, and provided keys.
//
// Arguments:
// - outputFile: A pointer to an os.File where the CSV content will be written.
// - keys: A slice of strings representing the column headers for the CSV.
//
// Returns:
// - A pointer to a csv.Writer that is ready to write rows to the output file.
func createCSVWriter(outputFile *os.File, keys []string) *csv.Writer {
	writer := csv.NewWriter(outputFile)

	// Include provider and model in the header
	fullKeys := append([]string{"Provider", "Model", "File Name"}, keys...)

	writer.Write(fullKeys)
	return writer
}

// WriteCSVData writes a row of data to the provided CSV writer, ensuring provider, model, and file name are included.
//
// Arguments:
// - response: JSON string containing key-value pairs corresponding to the CSV header.
// - filename: The name of the file being processed (written as the third column).
// - provider: The name of the LLM provider (first column).
// - model: The model name used (second column).
// - writer: A pointer to a csv.Writer to which the data will be written.
// - keys: A slice of strings representing the column headers.
func writeCSVData(response string, filename string, provider string, model string, writer *csv.Writer, keys []string) {
	// Clean the response
	response = cleanJSON(response)

	// Unmarshal JSON
	var data map[string]any
	err := json.Unmarshal([]byte(response), &data)
	if err != nil {
		logger.Error("Error parsing JSON:", err)
		logger.Error("Raw response:", response) // Debug output
		return
	}

	// Prepare CSV row
	row := make([]string, len(keys)+3)
	row[0] = provider
	row[1] = model
	row[2] = filename

	// Map values to the correct columns
	for i, key := range keys {
		if val, exists := data[key]; exists {
			row[i+3] = fmt.Sprintf("%v", val)
		} else {
			row[i+3] = "" // Empty field if key is missing
		}
	}

	// Write row to CSV
	if err := writer.Write(row); err != nil {
		logger.Error("Error writing to CSV:", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		logger.Error("Error flushing CSV:", err)
	}
}
