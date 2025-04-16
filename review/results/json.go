package results

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/open-and-sustainable/alembica/utils/logger"
)

// StartJSONArray begins a new JSON array in the specified output file. 
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

// WriteJSONData writes the given JSON response string to the specified output file.
// This function cleans up the response by removing any unnecessary code fences, 
// ensuring that the data is in a proper JSON format.
//
// Arguments:
// - response: A string containing the JSON data to be written.
// - filename: The name of the file being processed (used for logging or debugging purposes).
// - outputFile: A pointer to an os.File where the JSON content will be written.
//
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

// WriteCommaInJSONArray writes a comma to the JSON file to separate individual elements in a JSON array.
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

// CloseJSONArray writes the closing bracket for a JSON array, indicating the end of the list.
// This function should be called after all elements in the JSON array have been written.
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

// cleanJSON strips out the markdown code fences from the response if present.
func cleanJSON(response string) string {
	// Remove triple backticks and the "json" part (if present)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimSuffix(response, "```")
	return strings.TrimSpace(response) // Trim any extra whitespace
}