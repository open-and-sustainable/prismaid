package prismaid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/alembica/extraction"
	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid/config"
	"github.com/open-and-sustainable/prismaid/convert"
	"github.com/open-and-sustainable/prismaid/debug"
	"github.com/open-and-sustainable/prismaid/prompt"
	"github.com/open-and-sustainable/prismaid/results"
	"github.com/open-and-sustainable/prismaid/zotero"
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

// prompts for specific functionalities
const justification_query = "For each one of the keys and answers you provided, provide a justification for your answer as a chain of thought. In particular, I want a textual description of the few stages of the chin of thought that lead you to the answer you provided and the sentences in the text you analyzes that support your decision. If the value of a key was 'no' or empty '' because of lack of information on that topic in the text analyzed, explicitly report this reason. Please provide only th einformation requested, neither introductory nor concluding remarks."
const summary_query = "Summarize in very few sentences the text provided to you before for your review."


// RunReview is the main function responsible for orchestrating the systematic review process.
// It takes a TOML string as input, which defines the configuration for the review, and executes 
// the steps to carry out the review process, including model setup, input conversion, prompt generation, 
// and execution of the review logic.
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
//    - The TOML configuration string is passed to the LoadConfig function, which parses the TOML 
//      and populates a Config structure.
//    - The configuration contains details such as the project settings, LLM models, input/output settings, 
//      logging levels, and debugging options.
//    - If the TOML data is invalid or an error occurs during parsing, the function logs the error and returns it.
//
// 2. **Setup Logging**:
//    - Based on the log level specified in the configuration (high, medium, or low), the function 
//      sets up logging accordingly using the debug package.
//    - Logging can be written to a file, stdout, or be silent, depending on the log level. Logs are saved 
//      in the directory specified by the ResultsFileName.
//
// 3. **Input Conversion**:
//    - If the configuration specifies that input conversion is needed (e.g., converting PDF, DOCX files to text), 
//      the Convert function is called.
//    - If the conversion fails, an error is logged, and the process exits with a predefined error code.
//
// 4. **Debugging Features Setup**:
//    - If the Duplication feature is enabled (`Duplication == "yes"`), it duplicates the input files for debugging purposes, 
//      allowing the system to run the model queries twice on the same data for testing and comparison purposes.
//
// 5. **Prompt Generation**:
//    - Prompts are generated using the ParsePrompts function, based on the parameters defined in the TOML configuration. 
//      These include the persona, task, and other components needed for the systematic review.
//    - The function logs the number of files generated for review.
//
// 6. **Build Options Object**:
//    - The function creates an options object using the NewOptions function, passing in parameters such as 
//      the results file name, output format (e.g., CSV, JSON), and whether to include chain-of-thought justification 
//      and summaries in the results.
//    - If building the options fails, an error is returned.
//
// 7. **Build Query Object**:
//    - A query object is built using the NewQuery function, which organizes the parsed prompts 
//      and applies sorting logic based on the configuration (e.g., alphabetical order).
//    - If building the query fails, the function logs and returns the error.
//
// 8. **Model Setup and Execution**:
//    - The models object is built using the NewModels function, which loads the LLM models specified in the configuration.
//    - If there are multiple models in the configuration, the process is recognized as an ensemble review.
//    - The function runs each model individually by calling runSingleModelReview, passing in the model, options, query, and filenames.
//
// 9. **Ensemble Logic**:
//    - If multiple models are used (ensemble), the function logs that cost estimates are only available for single model reviews.
//    - For single models, it runs the review for each model, logging any errors encountered during the process.
//
// 10. **Cleanup**:
//    - If the Duplication feature was enabled for debugging, the function removes the duplicated input files created earlier.
//    - Finally, it logs "Done!" to indicate the successful completion of the review.
//
// 11. **Error Handling**:
//    - If any step in the review process encounters an error (e.g., loading configuration, input conversion, or review execution), 
//      the function logs the error and returns it to the caller.
//
// The RunReview function is the primary entry point for executing the entire review process, based on the user-provided TOML configuration string. 
// It orchestrates the different stages of the review process, including input parsing, prompt generation, model interaction, and output management.
func RunReview(tomlConfiguration string) error {
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

	// Zotero review logic
	if config.Project.Zotero.User != "" {
		client := &http.Client{}
		// downlaod pdfs
		err := zotero.DownloadPDFs(client, config.Project.Zotero.User, config.Project.Zotero.API, config.Project.Zotero.Group, results.GetDirectoryPath(config.Project.Configuration.ResultsFileName))
		if err != nil {
			logger.Error("Error:\n%v", err)
			return err
		}
		// convert pdfs
		err = convert.Convert(results.GetDirectoryPath(config.Project.Configuration.ResultsFileName)+"/zotero", "pdf")
		if err != nil {
			logger.Error("Error:\n%v", err)
			exit(ExitCodeErrorInReviewLogic)
		}
	} else {
		// run input conversion if needed and not a Zotero project
		if config.Project.Configuration.InputConversion != "no" {
			err := convert.Convert(config.Project.Configuration.InputDirectory, config.Project.Configuration.InputConversion)
			if err != nil {
				logger.Error("Error:\n%v", err)
				exit(ExitCodeErrorInReviewLogic)
			}
		}
	}

	// setup other debugging features
	if config.Project.Configuration.Duplication == "yes" {
		debug.DuplicateInput(config)
	}

	// generate prompts
	prompts, filenames := prompt.ParsePrompts(config)
	logger.Info("Found", len(prompts), "files")

	// convert to JSON format
	jsonString, err := convertToJSON(config, prompts)
	if err != nil {
		logger.Error("Error converting to JSON:", err)
		return err
	}

	// run review
	reviewResults, err := extraction.Extract(jsonString)

	logger.Info("Results:\n", reviewResults)
	
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

// Convert TOML config to JSON structure
func convertToJSON(config *config.Config, prompts []string) (string, error) {
	// Populate metadata
	jsonSchema := definitions.Input{
		Metadata: definitions.InputMetadata{
			Version:       config.Project.Version,
			SchemaVersion: "1.0", // Hardcoded since it's not in TOML
			Timestamp:     time.Now().Format(time.RFC3339),
		},
	}

	// Populate models
	for _, llm := range config.Project.LLM {
		jsonSchema.Models = append(jsonSchema.Models, definitions.Model{
			Provider:    llm.Provider,
			APIKey:      llm.ApiKey,
			Model:       llm.Model,
			Temperature: llm.Temperature,
			TPMLimit:    int(llm.TpmLimit),
			RPMLimit:    int(llm.RpmLimit),
		})
	}

	// Populate prompts
	for i, promptText := range prompts {
		y := 1
		jsonSchema.Prompts = append(jsonSchema.Prompts, definitions.Prompt{
			PromptContent:  promptText,
			SequenceID:     strconv.Itoa(i + 1),
			SequenceNumber: 1,
		})
		// add justifications
		if config.Project.Configuration.CotJustification == "yes" {
			jsonSchema.Prompts = append(jsonSchema.Prompts, definitions.Prompt{
				PromptContent:  justification_query,
				SequenceID:     strconv.Itoa(i + 1),
				SequenceNumber: y+1,
			})
			y++
		}
		// add summaries
		if config.Project.Configuration.Summary == "yes" {
			jsonSchema.Prompts = append(jsonSchema.Prompts, definitions.Prompt{
				PromptContent:  summary_query,
				SequenceID:     strconv.Itoa(i + 1),
				SequenceNumber: y+1,
			})
			y++
		}
	}

	// Convert to JSON string
	jsonData, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		logger.Error("Error marshaling JSON:", err)
		return "", err
	}

	return string(jsonData), nil
}

