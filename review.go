package prismaid

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"github.com/open-and-sustainable/prismaid/check"
	"github.com/open-and-sustainable/prismaid/config"
	"github.com/open-and-sustainable/prismaid/convert"
	"github.com/open-and-sustainable/prismaid/cost"
	"github.com/open-and-sustainable/prismaid/debug"
	"github.com/open-and-sustainable/prismaid/model"
	"github.com/open-and-sustainable/prismaid/prompt"
	"github.com/open-and-sustainable/prismaid/results"
	"github.com/open-and-sustainable/prismaid/review"
	"github.com/open-and-sustainable/prismaid/tokens"
	"github.com/open-and-sustainable/prismaid/zotero"
	"sync"
	"time"
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
		debug.SetupLogging(debug.File, config.Project.Configuration.ResultsFileName)
	} else if config.Project.Configuration.LogLevel == "medium" {
		debug.SetupLogging(debug.Stdout, config.Project.Configuration.ResultsFileName)
	} else {
		debug.SetupLogging(debug.Silent, config.Project.Configuration.ResultsFileName) // default value
	}

	// Zotero review logic
	if config.Project.Zotero.User != "" {
		// downlaod pdfs
		err := zotero.DownloadPDFs(config.Project.Zotero.User, config.Project.Zotero.API, config.Project.Zotero.Group, getDirectoryPath(config.Project.Configuration.ResultsFileName))
		if err != nil {
			log.Printf("Error:\n%v", err)
			return err
		}
		// convert pdfs
		err = convert.Convert(getDirectoryPath(config.Project.Configuration.ResultsFileName)+"/zotero", "pdf")
		if err != nil {
			log.Printf("Error:\n%v", err)
			exit(ExitCodeErrorInReviewLogic)
		}
	} else {
		// run input conversion if needed and not a Zotero project
		if config.Project.Configuration.InputConversion != "no" {
			err := convert.Convert(config.Project.Configuration.InputDirectory, config.Project.Configuration.InputConversion)
			if err != nil {
				log.Printf("Error:\n%v", err)
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
	log.Println("Found", len(prompts), "files")

	// build options object
	options, err := review.NewOptions(config.Project.Configuration.ResultsFileName, config.Project.Configuration.OutputFormat, config.Project.Configuration.CotJustification, config.Project.Configuration.Summary)
	if err != nil {
		log.Printf("Error:\n%v", err)
		return err
	}

	// build query object
	query, err := review.NewQuery(prompts, prompt.SortReviewKeysAlphabetically(config))
	if err != nil {
		log.Printf("Error:\n%v", err)
		return err
	}

	// build models object
	models, err := review.NewModels(config.Project.LLM)
	if err != nil {
		log.Printf("Error:\n%v", err)
		return err
	}
	
	// differentiate logic if simgle model review or ensemble
	ensemble := false
	if len(models) > 1 {ensemble = true}

	if ensemble {
		fmt.Println("Cost estimates are available for single model reviews only.")
	}
	
	for _, model := range models {
		if !ensemble {model.ID = ""}
		err = runSingleModelReview(model, options, query, filenames)
		if err != nil {
			log.Printf("Error:\n%v", err)
			return err
		}	
	}
	
	// cleanup eventual debugging temporary files
	if config.Project.Configuration.Duplication == "yes" {
		debug.RemoveDuplicateInput(config)
	}

	log.Println("Done!")
	return nil
}

// Method that returns the number of seconds to wait to respect TPM limits
func getWaitTime(prompt string, llm review.Model) int {
	// Locking to ensure thread-safety when accessing the requestTimestamps slice
	mutex.Lock()
	defer mutex.Unlock()

	// Clean up old timestamps (older than 60 seconds)
	now := time.Now()
	cutoff := now.Add(-60 * time.Second)
	validTimestamps := []time.Time{}
	for _, timestamp := range requestTimestamps {
		if timestamp.After(cutoff) {
			validTimestamps = append(validTimestamps, timestamp)
		}
	}
	requestTimestamps = validTimestamps

	// Add the current request timestamp
	requestTimestamps = append(requestTimestamps, now)

	// Get the current number of requests in the last 60 seconds
	numRequests := len(requestTimestamps)

	// Calculate the time to wait until the next minute
	remainingSeconds := 60 - now.Second()
	// Analyze TPM limits
	tpm_wait_seconds := 0
	// Get the TPM limit from the configuration
	tpmLimit := llm.TPM
	if tpmLimit > 0 {
		// Get the number of tokens from the prompt
		counter := tokens.RealTokenCounter{}
		tokens := counter.GetNumTokensFromPrompt(prompt, llm.Provider, llm.Model, llm.APIKey)
		tpm_wait_seconds = remainingSeconds
		// Calculate the number of tokens per second allowed
		tokensPerSecond := float64(tpmLimit) / 60.0
		// Calculate the required wait time in seconds to not exceed TPM limit
		requiredWaitTime := float64(tokens) / tokensPerSecond
		// Calculate the seconds to the next minute
		secondsToMinute := 0
		if int(requiredWaitTime) > 60 {
			secondsToMinute = 60 - int(requiredWaitTime)%60
		}
		// If required wait time is more than remaining seconds in the current minute, wait until next minute
		if requiredWaitTime > float64(remainingSeconds) {
			tpm_wait_seconds = remainingSeconds + int(requiredWaitTime) + secondsToMinute
		}
		// Otherwise, calculate the wait time based on tokens used
	}
	// Analyze RPM limits
	rpm_wait_seconds := 0
	rpmLimit := llm.RPM
	if rpmLimit > 0 {
		// If the number of requests risks to exceed the RPM limit, we need to wait
		if numRequests >= int(rpmLimit-1) {
			rpm_wait_seconds = remainingSeconds
		}
	}

	// Return the maximum of tpm_wait_seconds and rpm_wait_seconds
	if tpm_wait_seconds > rpm_wait_seconds {
		return tpm_wait_seconds
	} else {
		return rpm_wait_seconds
	}
}

func waitWithStatus(waitTime int) {
	ticker := time.NewTicker(1 * time.Second) // Ticks every second
	defer ticker.Stop()
	remainingTime := waitTime
	for range ticker.C {
		// Print the status only when the remaining time modulo 5 equals 0
		if remainingTime%5 == 0 {
			log.Printf("Waiting... %d seconds remaining\n", remainingTime)
		}
		remainingTime--
		// Break the loop when no time is left
		if remainingTime <= 0 {
			log.Println("Wait completed.")
			break
		}
	}
}

func getDirectoryPath(resultsFileName string) string {
	dir := filepath.Dir(resultsFileName)

	// If the directory is ".", return an empty string
	if dir == "." {
		return ""
	}
	return dir
}

func runSingleModelReview(llm review.Model, options review.Options, query review.Query, filenames []string) error {

	// start writer for results.. the file will be project_name[.csv or .json] in the path where the toml is
	resultsFileName := options.ResultsFileName
	outputFilePath := resultsFileName + "." + options.OutputFormat
	if llm.ID != "" {outputFilePath = resultsFileName + "_" + llm.ID + "." + options.OutputFormat}
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Println("Error creating output file:", err)
		return err
	}
	defer outputFile.Close() // Ensure the file is closed after all operations are done

	var writer *csv.Writer
	if options.OutputFormat == "csv" {
		writer = results.CreateCSVWriter(outputFile, query.Keys) // Pass the file to CreateWriter
		defer writer.Flush()                                // Ensure data is flushed after all writes	
	} else if options.OutputFormat == "json" {
		err = results.StartJSONArray(outputFile)
		if err != nil {
			log.Println("Error starting JSON array:", err)
			return err
		}
	}

	// Loop through the prompts
	for i, promptText := range query.Prompts {
		log.Println("File: ", filenames[i], " Prompt: ", promptText)

		// clean model names
		llm.Model = check.GetModel(promptText, llm.Provider, llm.Model, llm.APIKey)
		fmt.Println("Processing file "+fmt.Sprint(i+1)+"/"+fmt.Sprint(len(query.Prompts))+" "+filenames[i]+" with model "+llm.Model)
		
		// check if prompts resepct input tokens limits for selected models
		counter := tokens.RealTokenCounter{}
		checkInputLimits := check.RunInputLimitsCheck(promptText, llm.Provider, llm.Model, llm.APIKey, counter)
		if checkInputLimits != nil {
			fmt.Println("Error resepecting the max input tokens limits for the following manuscripts and models.")
			log.Printf("Error:\n%v", checkInputLimits)
			exit(ExitCodeInputTokenError)	
		}
		if llm.ID == "" {			
			// ask if continuing given the total cost
			check := check.RunUserCheck(cost.ComputeCosts(query.Prompts, llm.Provider, llm.Model, llm.APIKey), llm.Provider)
			if check != nil {
				log.Printf("Error:\n%v", check)
				exit(0) // if the user stops the execution it is still a success run, hence exit code = 0, but the reason for the exit may be different hence is logged
			}
		}

		// Query the LLM
		realQueryService := model.DefaultQueryService{}
		response, justification, summary, err := realQueryService.QueryLLM(promptText, llm, options)
		if err != nil {
			log.Println("Error querying LLM:", err)
			return err
		}

		// Handle the output format
		if options.OutputFormat == "json" {
			results.WriteJSONData(response, filenames[i], outputFile) // Write formatted JSON to file
			// add comma if it's not the last element
			if i < len(query.Prompts)-1 {
				results.WriteCommaInJSONArray(outputFile)
			}
		} else {
			if options.OutputFormat == "csv" {
				results.WriteCSVData(response, filenames[i], writer, query.Keys)
			}
		}
		// save justifications
		if options.Justification {
			justificationFilePath := getDirectoryPath(resultsFileName) + "/" + filenames[i] + "_justification.txt"
			if llm.ID != "" {justificationFilePath = getDirectoryPath(resultsFileName) + "/" + filenames[i] + "_justification_"+llm.ID+".txt"}
			err := os.WriteFile(justificationFilePath, []byte(justification), 0644)
			if err != nil {
				log.Println("Error writing justification file:", err)
				return err
			}
		}
		// save summaries
		if options.Summary {
			summaryFilePath := getDirectoryPath(resultsFileName) + "/" + filenames[i] + "_summary.txt"
			if llm.ID != "" {summaryFilePath = getDirectoryPath(resultsFileName) + "/" + filenames[i] + "_summary_"+llm.ID+".txt"}
			err := os.WriteFile(summaryFilePath, []byte(summary), 0644)
			if err != nil {
				log.Println("Error writing summary file:", err)
				return err
			}
		}

		// Sleep before the next prompt if it's not the last one
		if i < len(query.Prompts)-1 {
			waitWithStatus(getWaitTime(promptText, llm))
		}
	}

	// close JSON array if needed
	if options.OutputFormat == "json" {
		err = results.CloseJSONArray(outputFile)
		if err != nil {
			log.Println("Error closing JSON array:", err)
			return err
		}
	}	
	
	return nil
}
