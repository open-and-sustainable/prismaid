package results

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/open-and-sustainable/alembica/utils/logger"
)

// startJSONArray begins a new JSON array in the specified output file.
// This function writes the opening bracket for an array to indicate the start of a JSON list.
//
// Arguments:
// - outputFile: A pointer to an os.File where the JSON array will be written.
//
// Returns:
// - An error if writing to the file fails, otherwise returns nil.
func startJSONArray(outputFile *os.File) error {
	_, err := outputFile.WriteString("[\n")
	if err != nil {
		logger.Error("Error starting JSON array:", err)
		return err
	}
	return nil
}

// writeJSONData writes the given JSON response string to the specified output file.
// This function cleans up the response by removing any unnecessary code fences,
// enhances the JSON data by adding the filename field, and formats it with indentation.
//
// Arguments:
// - response: A string containing the JSON data to be written.
// - filename: The name of the file being processed. This is added to the JSON data as a "filename" field.
// - outputFile: A pointer to an os.File where the JSON content will be written.
//
// The function logs any errors encountered but does not return them.
// This function does not automatically close or flush the file; these operations should be handled separately.
func writeJSONData(response string, filename string, outputFile *os.File) {
	// Strip out markdown code fences (```json ... ```) if present
	response = cleanJSON(response)

	// Unmarshal the JSON string into a map to modify it
	var data map[string]interface{}
	err := json.Unmarshal([]byte(response), &data)
	if err != nil {
		logger.Error("Error unmarshaling JSON:", err)
		return
	}

	// Add the filename to the JSON data
	data["filename"] = filename

	// Marshal the modified data back into a JSON string
	modifiedJSON, err := json.MarshalIndent(data, "", "    ") // Indent for pretty JSON output
	if err != nil {
		logger.Error("Error marshaling modified JSON:", err)
		return
	}

	// Write the modified JSON to the file
	_, err = outputFile.WriteString(string(modifiedJSON))
	if err != nil {
		logger.Error("Error writing JSON to file:", err)
	}
}

// writeCommaInJSONArray writes a comma to the JSON file to separate individual elements in a JSON array.
// This function should be called between writing separate JSON objects to maintain valid JSON syntax.
//
// Arguments:
// - outputFile: A pointer to an os.File where the comma will be written.
//
// Returns:
// - An error if writing to the file fails, otherwise returns nil.
func writeCommaInJSONArray(outputFile *os.File) error {
	_, err := outputFile.WriteString(",\n")
	if err != nil {
		logger.Error("Error writing comma in JSON array:", err)
		return err
	}
	return nil
}

// closeJSONArray writes the closing bracket for a JSON array, indicating the end of the list.
// This function should be called after all elements in the JSON array have been written.
// It adds a newline before the closing bracket for proper formatting.
//
// Arguments:
// - outputFile: A pointer to an os.File where the closing bracket will be written.
//
// Returns:
// - An error if writing to the file fails, otherwise returns nil.
func closeJSONArray(outputFile *os.File) error {
	// Write the closing bracket
	_, err := outputFile.WriteString("\n]")
	if err != nil {
		logger.Error("Error closing JSON array:", err)
		return err
	}

	return nil
}

// cleanJSON strips out markdown code fences from a string containing JSON data.
// This function processes strings that might be wrapped in markdown-style code
// blocks (e.g., ```json ... ```) and removes these markers to extract the pure JSON content.
//
// Arguments:
// - response: A string potentially containing JSON data wrapped in markdown code fences.
//
// Returns:
// - A cleaned string containing only the JSON data with surrounding whitespace removed.
func cleanJSON(response string) string {
	// Remove triple backticks and the "json" part (if present)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimSuffix(response, "```")
	return strings.TrimSpace(response) // Trim any extra whitespace
}
