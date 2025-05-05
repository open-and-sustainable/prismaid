package results

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid/review/config"
)

// Save writes processed model response data to a file in the configured format.
// It determines the appropriate output format based on the configuration (JSON or CSV)
// and dispatches to the corresponding save function. When CSV format is selected,
// it also extracts and saves justifications and summaries to separate text files.
//
// Parameters:
//   - config: Application configuration containing output settings
//   - results: JSON string containing all model responses
//   - filenames: List of input filenames that were processed
//   - keys: List of column headers to include in CSV output
//
// Returns:
//   - error: nil if successful, otherwise an error describing what failed
func Save(config *config.Config, results string, filenames []string, keys []string) error {
	resultsFileName := config.Project.Configuration.ResultsFileName
	outputFormat := config.Project.Configuration.OutputFormat
	outputFilePath := resultsFileName + "." + outputFormat

	// Save justifications & summaries ONLY if CSV format
	if outputFormat == "csv" {
		saveJustificationsAndSummaries(config, resultsFileName, results, filenames)
	}

	if outputFormat == "json" {
		return saveJSON(outputFilePath, results, filenames)
	} else if outputFormat == "csv" {
		return saveCSV(outputFilePath, results, filenames, keys)
	} else {
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// saveJSON creates and populates a JSON file with processed model responses.
// It formats the responses as an array of objects, each containing provider, model, and filename metadata
// along with the model's response content. The function preserves the original structure of each model response
// while adding consistent metadata fields.
//
// Parameters:
//   - filePath: The output file path for the JSON
//   - resultsString: JSON string containing all model responses
//   - filenames: List of input filenames that were processed
//
// Returns:
//   - error: nil if successful, otherwise an error describing what failed
func saveJSON(filePath string, resultsString string, filenames []string) error {
	outputFile, err := os.Create(filePath)
	if err != nil {
		logger.Error("Error creating JSON file:")
		return err
	}
	defer outputFile.Close()

	// Start JSON array
	if err := startJSONArray(outputFile); err != nil {
		return err
	}

	// Unmarshal JSON results
	var parsedResults definitions.Output
	if err := json.Unmarshal([]byte(resultsString), &parsedResults); err != nil {
		logger.Error("Error parsing JSON for structured output:", err)
		return err
	}

	// Debugging: Check response count
	fmt.Println("Total JSON responses:", len(parsedResults.Responses))

	// Write each response separately with provider & model metadata
	for i, response := range parsedResults.Responses {
		filenameIndex := i % len(filenames) // Prevent index out of range
		fmt.Println("Processing response", i+1, "/", len(parsedResults.Responses), "Filename:", filenames[filenameIndex])

		modifiedResponse := map[string]interface{}{
			"provider": response.Provider,
			"model":    response.Model,
			"filename": filenames[filenameIndex],
		}

		// Merge model response into the modified response map
		var responseData map[string]interface{}
		if err := json.Unmarshal([]byte(response.ModelResponses[0]), &responseData); err == nil {
			for key, value := range responseData {
				modifiedResponse[key] = value
			}
		}

		// Convert to JSON string and write it
		modifiedJSON, err := json.MarshalIndent(modifiedResponse, "", "    ")
		if err != nil {
			logger.Error("Error marshaling modified JSON:", err)
			return err
		}

		// Ensure newline between JSON objects
		if i > 0 {
			if err := writeCommaInJSONArray(outputFile); err != nil {
				return err
			}
		}

		_, err = outputFile.WriteString(string(modifiedJSON) + "\n") // Add newline for readability
		if err != nil {
			logger.Error("Error writing JSON to file:", err)
			return err
		}
	}

	// Close JSON array
	if err := closeJSONArray(outputFile); err != nil {
		return err
	}

	logger.Info("JSON results successfully saved to:", filePath)
	return nil
}

// saveCSV creates and populates a CSV file with processed model responses.
// It converts the JSON model responses into a tabular format with columns specified by keys.
// Only primary responses (SequenceNumber = 1) are included in the CSV; justifications and summaries are skipped.
//
// Parameters:
//   - filePath: The output file path for the CSV
//   - resultsString: JSON string containing all model responses
//   - filenames: List of input filenames that were processed
//   - keys: List of column headers to include in the CSV
//
// Returns:
//   - error: nil if successful, otherwise an error describing what failed
func saveCSV(filePath string, resultsString string, filenames []string, keys []string) error {
	outputFile, err := os.Create(filePath)
	if err != nil {
		logger.Error("Error creating CSV file: %v", err)
		return err
	}
	defer outputFile.Close()

	writer := createCSVWriter(outputFile, keys)
	defer writer.Flush()

	// Parse JSON results
	var parsedResults definitions.Output
	if err := json.Unmarshal([]byte(resultsString), &parsedResults); err != nil {
		logger.Error("Error parsing results JSON: %v", err)
		return err
	}

	// Process responses
	for i, response := range parsedResults.Responses {
		// Skip justifications & summaries (SequenceNumber > 1)
		if response.SequenceNumber > 1 {
			logger.Info("Skipping justification/summary in CSV (SeqNum: %d)", response.SequenceNumber)
			continue
		}

		// Ensure correct filename mapping
		filenameIndex := i % len(filenames) // Prevent index out of range

		// Write the main response data
		for _, modelResponse := range response.ModelResponses {
			writeCSVData(modelResponse, filenames[filenameIndex], response.Provider, response.Model, writer, keys)
		}
	}

	logger.Info("CSV results successfully saved to: %s", filePath)
	return nil
}

// GetDirectoryPath extracts the directory component from a file path.
// It returns an empty string if the directory is the current directory (".").
//
// Parameters:
//   - resultsFileName: The file path to extract the directory from
//
// Returns:
//   - string: The directory path, or empty string if the directory is "."
func GetDirectoryPath(resultsFileName string) string {
	dir := filepath.Dir(resultsFileName)

	// If the directory is ".", return an empty string
	if dir == "." {
		return ""
	}
	return dir
}

// saveJustificationsAndSummaries extracts and saves justification and summary content from model responses
// to separate text files. It only processes these files if the respective configuration options are enabled.
//
// Parameters:
//   - config: The application configuration containing justification and summary settings
//   - resultsFileName: Base name for result files (without extension)
//   - resultsString: JSON string containing all model responses
//   - filenames: List of input filenames that were processed
//
// Returns:
//   - error: nil if successful, otherwise an error describing what failed
func saveJustificationsAndSummaries(config *config.Config, resultsFileName string, resultsString string, filenames []string) error {
	justificationEnabled := config.Project.Configuration.CotJustification == "yes"
	summaryEnabled := config.Project.Configuration.Summary == "yes"

	if !justificationEnabled && !summaryEnabled {
		logger.Info("Skipping justification and summary saving as they are not enabled.")
		return nil
	}

	// Unmarshal results JSON
	var parsedResults definitions.Output
	if err := json.Unmarshal([]byte(resultsString), &parsedResults); err != nil {
		logger.Error("Error parsing results JSON: %v", err)
		return err
	}

	// Ensure we have filenames
	if len(filenames) == 0 {
		return fmt.Errorf("no filenames provided")
	}

	// Group responses by sequenceId AND provider AND model
	sequenceResponses := make(map[string][]definitions.Response)
	for _, response := range parsedResults.Responses {
		// Create a composite key that includes all three fields
		compositeKey := fmt.Sprintf("%s_%s_%s", response.SequenceID, response.Provider, response.Model)
		sequenceResponses[compositeKey] = append(sequenceResponses[compositeKey], response)
	}

	// Process grouped responses
	for compositeKey, responses := range sequenceResponses {
		if len(responses) < 2 {
			continue // Not enough responses to contain a justification or summary
		}

		// Sort responses by SequenceNumber
		sort.SliceStable(responses, func(i, j int) bool {
			return responses[i].SequenceNumber < responses[j].SequenceNumber
		})

		// Extract sequenceId from the composite key (assuming format "seqId_provider_model")
		parts := strings.Split(compositeKey, "_")
		seqID := parts[0]

		// Identify filename mapping
		seqIndex, err := strconv.Atoi(seqID)
		if err != nil || seqIndex < 1 || seqIndex > len(filenames) {
			logger.Error("Invalid sequence ID mapping for file: %s", seqID)
			continue
		}

		originalFilename := filenames[seqIndex-1]
		provider := responses[0].Provider
		model := responses[0].Model
		baseFilename := fmt.Sprintf("%s/%s_%s_%s", GetDirectoryPath(resultsFileName), originalFilename, provider, model)

		// Identify and save Justification (if enabled)
		if justificationEnabled {
			for _, response := range responses {
				if response.SequenceNumber == 2 && len(response.ModelResponses) > 0 {
					justificationFilePath := baseFilename + "_justification.txt"
					justificationContent := response.ModelResponses[0] // Correctly extract model output
					if err := os.WriteFile(justificationFilePath, []byte(justificationContent), 0644); err != nil {
						logger.Error("Error writing justification file: %v", err)
						return err
					}
					logger.Info("Saved justification to: %s", justificationFilePath)
					break
				}
			}
		}

		// Identify and save Summary (if enabled)
		if summaryEnabled {
			for _, response := range responses {
				if response.SequenceNumber == 3 && len(response.ModelResponses) > 0 {
					summaryFilePath := baseFilename + "_summary.txt"
					summaryContent := response.ModelResponses[0] // Correctly extract model output
					if err := os.WriteFile(summaryFilePath, []byte(summaryContent), 0644); err != nil {
						logger.Error("Error writing summary file: %v", err)
						return err
					}
					logger.Info("Saved summary to: %s", summaryFilePath)
					break
				}
			}
		}
	}

	logger.Info("Justifications and Summaries saved successfully")
	return nil
}
