package prompt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/open-and-sustainable/prismaid/config"

	"github.com/open-and-sustainable/alembica/utils/logger"
)

// ParsePrompts reads the configuration and generates a list of prompts along with their corresponding filenames.
// The function combines different parts of the prompts to create a structured list of inputs.
//
// Arguments:
// - config: A pointer to the application's configuration which specifies how prompts should be parsed and organized.
//
// Returns:
// - Two slices of strings: 
//   - The first slice contains the generated prompts.
//   - The second slice contains the filenames associated with each prompt.
func ParsePrompts(config *config.Config) ([]string, []string) {
	// This slice will store all combined prompts
	var prompts []string
	// This slice will store the filenames corresponding to each prompt
	var filenames []string

	// The common part of prompts
	expected_result := parseExpectedResults(config)
	common_part := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		config.Prompt.Persona, config.Prompt.Task, expected_result,
		config.Prompt.Failsafe, config.Prompt.Definitions, config.Prompt.Example)

	// Load text files
	files, err := os.ReadDir(config.Project.Configuration.InputDirectory)
	if err != nil {
		logger.Error(err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".txt" {
			filePath := filepath.Join(config.Project.Configuration.InputDirectory, file.Name())
			documentText, err := os.ReadFile(filePath)
			if err != nil {
				logger.Error("Error reading file:", err)
				return nil, nil
			}

			// Combine prompt elements
			prompt := fmt.Sprintf("%s \n\n%s", common_part, documentText)
			// Append the combined text to the slice
			prompts = append(prompts, prompt)

			// Get the filename without extension
			fileNameWithoutExt := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			// Append the filename to the filenames slice
			filenames = append(filenames, fileNameWithoutExt)
		}
	}

	return prompts, filenames
}

func parseExpectedResults(config *config.Config) string {
	expectedResult := config.Prompt.ExpectedResult
	keys := GetReviewKeysByEntryOrder(config)

	// Build a map from sorted keys using descriptive keys
	sortedReviewItems := make(map[string][]string)
	for _, numericKey := range keys {
		item := config.Review[numericKey]
		sortedReviewItems[item.Key] = item.Values // Use the descriptive key for the JSON output
	}

	// Convert sorted map to JSON
	reviewJSON, err := json.Marshal(sortedReviewItems)
	if err != nil {
		logger.Error("Error marshalling review items to JSON: %v", err)
	}

	// Combine the expected result with the JSON-formatted review items
	fullSummary := fmt.Sprintf("%s %s", expectedResult, string(reviewJSON))
	return fullSummary
}

// GetReviewKeysByEntryOrder retrieves the keys from the review configuration in the order they appear
// in the configuration file. This function ensures that the keys are returned in a consistent order based
// on their entry sequence, which is useful for processing that relies on the sequence of entries such as
// when maintaining the original configuration order is necessary.
//
// Arguments:
// - config: A pointer to the application's configuration, which specifies the review keys to be retrieved.
//
// Returns:
// - A slice of strings containing the ordered review keys based on their entry order in the configuration file.
//
// This function is particularly useful in scenarios where the order of review items as defined in the 
// configuration impacts the workflow or results, such as generating reports or processing data in the 
// sequence of configuration.
func GetReviewKeysByEntryOrder(config *config.Config) []string {
	keys := make([]string, 0, len(config.Review))
	for key := range config.Review {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Sort keys to maintain the order of entries as in the configuration
	return keys
}

// SortReviewKeysAlphabetically retrieves and sorts the descriptive keys (not the TOML entry keys) from the 
// review configuration alphabetically. This sorting approach focuses on the descriptive aspects of the keys 
// rather than their position in the configuration file, making it useful for user interfaces or outputs where 
// alphabetical ordering facilitates better readability and accessibility.
//
// Arguments:
// - config: A pointer to the application's configuration that contains the review items.
//
// Returns:
// - A slice of strings containing the review keys sorted alphabetically by their descriptive labels.
//
// This function is ideal for scenarios where the logical grouping or alphabetical presentation of review items 
// is critical, such as in user interfaces, alphabetical listings in documentation, or any application where
// the user benefits from sorting by topic names rather than the order of entries.
func SortReviewKeysAlphabetically(config *config.Config) []string {
	keys := make([]string, 0)
	for _, item := range config.Review {
		keys = append(keys, item.Key)
	}
	sort.Strings(keys) // Sort keys alphabetically for consistent and logical output
	return keys
}
