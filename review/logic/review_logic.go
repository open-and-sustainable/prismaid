package logic

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/open-and-sustainable/alembica/extraction"
	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid/review/config"
	"github.com/open-and-sustainable/prismaid/review/debug"
	"github.com/open-and-sustainable/prismaid/review/prompt"
	"github.com/open-and-sustainable/prismaid/review/results"
)

const (
	// Define a specific exit code for wrong command call
	ExitCodeWrongCommandCall = 1
	// Define a specific exit code for review logic errors
	ExitCodeErrorInReviewLogic = 2
	// Define a specific exit code for supplier model errors
	ExitCodeInputSupplierModelError = 3
	// Define a specific exit code for input token errors
	ExitCodeInputTokenError = 4
)

var exitFunc = os.Exit

func exit(code int) {
	exitFunc(code)
}

// Global variable to store the timestamps of requests
var requestTimestamps []time.Time
var mutex sync.Mutex

// Review is the main function responsible for orchestrating the systematic review process.
// It takes a TOML string as input, which defines the configuration for the review, and executes
// the steps to carry out the review process, including configuration loading, prompt generation,
// extraction, and saving results.
//
// Parameters:
//   - tomlConfiguration: A string containing the TOML configuration data for the review project.
//
// Returns:
//   - An error if any step in the review process fails, or nil if the process completes successfully.
//
// The function performs the following steps:
//
// 1. **Load Configuration**:
//   - The TOML configuration string is passed to the LoadConfig function, which parses the TOML
//     and populates a Config structure.
//   - The configuration contains details such as the project settings, input/output settings,
//     logging levels, and debugging options.
//   - If the TOML data is invalid or an error occurs during parsing, the function logs the error and returns it.
//
// 2. **Setup Logging**:
//   - Based on the log level specified in the configuration (high, medium, or low), the function
//     sets up logging accordingly using the logger package.
//   - Logging can be written to a file, stdout, or be silent, depending on the log level. Logs are saved
//     in the directory specified by the ResultsFileName.
//
// 3. **Debugging Features Setup**:
//   - If the Duplication feature is enabled (`Duplication == "yes"`), it duplicates the input files for debugging purposes,
//     allowing the system to run the extraction twice on the same data for testing and comparison purposes.
//
// 4. **Prompt Generation**:
//   - Prompts are generated using the PrepareInput function, based on the parameters defined in the TOML configuration.
//   - The function logs the number of files found for review.
//
// 5. **Run Extraction**:
//   - The function calls extraction.Extract with the prepared JSON string to perform the actual review process.
//   - The extraction results are logged.
//
// 6. **Save Results**:
//   - Results are saved using the Save function, with review keys sorted alphabetically.
//   - If saving the results fails, an error is logged and returned.
//
// 7. **Cleanup**:
//   - If the Duplication feature was enabled for debugging, the function removes the duplicated input files created earlier.
//   - Finally, it logs "Done!" to indicate the successful completion of the review.
//
// 8. **Error Handling**:
//   - If any step in the review process encounters an error, the function logs the error and returns it to the caller.
//
// The Review function is the primary entry point for executing the entire review process, based on the user-provided TOML configuration string.
// It orchestrates the different stages of the review process, including input parsing, prompt generation, extraction, and results handling.
func Review(tomlConfiguration string) error {
	// load project configuration
	config, err := config.LoadConfig(tomlConfiguration, config.RealEnvReader{})
	if err != nil {
		fmt.Println("Error loading project configuration:", err) // here the logging function is not implemented yet
		return err
	}

	// setup logging
	if config.Project.Configuration.LogLevel == "high" {
		logger.SetupLogging(logger.File, config.Project.Configuration.ResultsFileName)
	} else if config.Project.Configuration.LogLevel == "medium" {
		logger.SetupLogging(logger.Stdout, config.Project.Configuration.ResultsFileName)
	} else {
		logger.SetupLogging(logger.Silent, config.Project.Configuration.ResultsFileName) // default value
	}

	// setup other debugging features
	if config.Project.Configuration.Duplication == "yes" {
		debug.DuplicateInput(config)
	}

	// generate prompts
	jsonString, filenames, err := prompt.PrepareInput(config)
	if err != nil {
		logger.Error("Error generating prompts:", err)
		return err
	}
	logger.Info("Found", len(filenames), "files")

	// run review
	reviewResults, err := extraction.Extract(jsonString)

	logger.Info("Results:\n%s", reviewResults)

	// save results
	keys := prompt.SortReviewKeysAlphabetically(config)
	err = results.Save(config, reviewResults, filenames, keys)
	if err != nil {
		logger.Error("Error saving results:", err)
		return err
	}

	// cleanup eventual debugging temporary files
	if config.Project.Configuration.Duplication == "yes" {
		debug.RemoveDuplicateInput(config)
	}

	logger.Info("Done!")
	return nil
}
