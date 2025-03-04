package results

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"os"

	"github.com/open-and-sustainable/prismaid/config"
	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/alembica/utils/logger"
)

func Save(config *config.Config, results string, filenames []string, keys []string) error {
	resultsFileName := config.Project.Configuration.ResultsFileName
	outputFormat := config.Project.Configuration.OutputFormat
	outputFilePath := resultsFileName + "." + outputFormat

	saveJustificationsAndSummaries(config, resultsFileName, results, filenames)

	if outputFormat == "json" {
		return saveJSON(outputFilePath, results, filenames)
	} else if outputFormat == "csv" {
		return saveCSV(outputFilePath, results, filenames, keys)
	} else {
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

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

func saveCSV(filePath string, resultsString string, filenames []string, keys []string) error {
    outputFile, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer outputFile.Close()

    writer := createCSVWriter(outputFile, keys)
    defer writer.Flush()

    var parsedResults definitions.Output
    if err := json.Unmarshal([]byte(resultsString), &parsedResults); err != nil {
        return err
    }

    for i, response := range parsedResults.Responses {
        filenameIndex := i % len(filenames) // Repeat filenames if needed

        for _, modelResponse := range response.ModelResponses {
            writeCSVData(modelResponse, filenames[filenameIndex], response.Provider, response.Model, writer, keys)
        }
    }

    return nil
}

func GetDirectoryPath(resultsFileName string) string {
	dir := filepath.Dir(resultsFileName)

	// If the directory is ".", return an empty string
	if dir == "." {
		return ""
	}
	return dir
}

// saveJustificationsAndSummaries extracts provider, model, and content from resultsString
func saveJustificationsAndSummaries(config *config.Config, resultsFileName string, resultsString string, filenames []string) error {
	// Unmarshal results JSON
	var parsedResults definitions.Output
	if err := json.Unmarshal([]byte(resultsString), &parsedResults); err != nil {
		logger.Error("Error parsing results JSON:", err)
		return err
	}

	// Loop through responses and save justifications & summaries
	for i, response := range parsedResults.Responses {
		filenameIndex := i % len(filenames) // Prevent index out-of-range
		filename := filenames[filenameIndex]
		provider := response.Provider
		model := response.Model

		// Construct file paths with provider & model
		justificationFilePath := GetDirectoryPath(resultsFileName) + "/" + filename + "_justification_" + provider + "_" + model + ".txt"
		summaryFilePath := GetDirectoryPath(resultsFileName) + "/" + filename + "_summary_" + provider + "_" + model + ".txt"

		// Extract justification & summary from JSON (modify keys as needed)
		justificationContent := "No justification found"
		summaryContent := "No summary found"

		// Check if JSON contains justification & summary
		if response.ModelResponses != nil && len(response.ModelResponses) > 0 {
			var responseData map[string]interface{}
			if err := json.Unmarshal([]byte(response.ModelResponses[0]), &responseData); err == nil {
				if val, exists := responseData["justification"]; exists {
					justificationContent = val.(string)
				}
				if val, exists := responseData["summary"]; exists {
					summaryContent = val.(string)
				}
			}
		}

		// Save Justifications
		if config.Project.Configuration.CotJustification == "yes" {
			err := os.WriteFile(justificationFilePath, []byte(justificationContent), 0644)
			if err != nil {
				logger.Error("Error writing justification file:", err)
				return err
			}
		}

		// Save Summaries
		if config.Project.Configuration.Summary == "yes" {
			err := os.WriteFile(summaryFilePath, []byte(summaryContent), 0644)
			if err != nil {
				logger.Error("Error writing summary file:", err)
				return err
			}
		}
	}

	logger.Info("Justifications and Summaries saved successfully")
	return nil
}
