package logic

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/open-and-sustainable/alembica/extraction"
	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid/revaise"
	"github.com/open-and-sustainable/prismaid/review/chunking"
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

var extractReview = extraction.Extract

func exit(code int) {
	exitFunc(code)
}

// Global variable to store the timestamps of requests
var requestTimestamps []time.Time
var mutex sync.Mutex

// emptyEnvReader resolves every variable to an empty string so that
// configuration validation never reads environment variables or resolves API
// keys.
type emptyEnvReader struct{}

func (emptyEnvReader) GetEnv(string) string { return "" }

// ValidateConfig parses and validates a review TOML configuration without
// running the review, accessing the network, or resolving API keys. It returns
// nil if the configuration is valid, or an error describing the problem found.
func ValidateConfig(tomlConfiguration string) error {
	_, err := config.LoadConfig(tomlConfiguration, emptyEnvReader{})
	return err
}

// ReviewResult summarizes a completed review run: where results were written,
// how many manuscripts were processed, how many review items were extracted, and
// which models were used. The detailed extraction output is in the results file.
type ReviewResult struct {
	OutputFile           string
	ChunkingReportFile   string
	ManuscriptsProcessed int
	ManuscriptsSucceeded int
	ManuscriptsFailed    int
	ReviewItems          int
	Models               []string
}

// PartialReviewError reports a run in which results were successfully saved
// for one or more documents but one or more document artifacts failed.
type PartialReviewError struct {
	OutputFile         string
	ChunkingReportFile string
	FailedDocuments    int
}

func (e *PartialReviewError) Error() string {
	return fmt.Sprintf("review completed with partial results: %d document(s) failed; results saved to %s; chunking report saved to %s", e.FailedDocuments, e.OutputFile, e.ChunkingReportFile)
}

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
func Review(tomlConfiguration string) (*ReviewResult, error) {
	// load project configuration
	config, err := config.LoadConfig(tomlConfiguration, config.RealEnvReader{})
	if err != nil {
		fmt.Println("Error loading project configuration:", err) // here the logging function is not implemented yet
		return nil, err
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
	preparedInput, err := prompt.PreparePlan(config)
	if err != nil {
		logger.Error("Error generating prompts:", err)
		return nil, err
	}
	logger.Info("Found", len(preparedInput.Filenames), "files")
	if config.Project.Configuration.Chunking.Enabled {
		for _, report := range preparedInput.ChunkReports {
			logger.Info("Chunking plan for", report.Filename,
				"counter:", report.CounterMethod,
				"full prompt tokens:", report.FullPromptTokens,
				"chunks:", report.ChunkCount,
				"chunk prompt tokens:", report.ChunkPromptTokens)
		}
	}

	// run review
	reviewResults, err := extractReview(preparedInput.JSON)
	if err != nil {
		logger.Error("Error extracting review results:", err)
		return nil, err
	}
	mergeReport := chunking.MergeReport{}
	if preparedInput.HasChunks {
		reviewResults, mergeReport, err = chunking.MergeWithExpected(reviewResults, preparedInput.Bindings, preparedInput.ExpectedGroups, config.Project.Configuration.Chunking)
		if err != nil {
			logger.Error("Error merging chunked review results:", err)
			return nil, err
		}
	}
	logMergeReport(mergeReport)

	logger.Info("Results:\n", reviewResults)

	// save results
	keys := prompt.SortReviewKeysAlphabetically(config)
	err = results.Save(config, reviewResults, preparedInput.Filenames, keys)
	if err != nil {
		logger.Error("Error saving results:", err)
		return nil, err
	}

	chunkingReportFile := ""
	if config.Project.Configuration.Chunking.Enabled {
		chunkingReportFile, err = results.SaveChunkingReport(config.Project.Configuration.ResultsFileName, buildChunkingReport(preparedInput.ChunkReports, mergeReport))
		if err != nil {
			logger.Error("Error saving chunking report:", err)
			return nil, err
		}
	}

	if err := updateRevAIseExtraction(config, reviewResults, preparedInput.Filenames, keys); err != nil {
		logger.Error("Error updating RevAIse record:", err)
		return nil, err
	}

	// cleanup eventual debugging temporary files
	if config.Project.Configuration.Duplication == "yes" {
		debug.RemoveDuplicateInput(config)
	}

	models := make([]string, 0, len(config.Project.LLM))
	for _, llm := range config.Project.LLM {
		models = append(models, llm.Provider+" "+llm.Model)
	}

	failedDocuments := failedDocumentCount(mergeReport.Failures)
	result := &ReviewResult{
		OutputFile:           config.Project.Configuration.ResultsFileName + "." + config.Project.Configuration.OutputFormat,
		ChunkingReportFile:   chunkingReportFile,
		ManuscriptsProcessed: len(preparedInput.Filenames),
		ManuscriptsSucceeded: len(preparedInput.Filenames) - failedDocuments,
		ManuscriptsFailed:    failedDocuments,
		ReviewItems:          len(keys),
		Models:               models,
	}
	if failedDocuments > 0 {
		logger.Error("Review completed with partial results; failed documents:", failedDocuments)
		return result, &PartialReviewError{OutputFile: result.OutputFile, ChunkingReportFile: chunkingReportFile, FailedDocuments: failedDocuments}
	}
	logger.Info("Done!")
	return result, nil
}

func logMergeReport(report chunking.MergeReport) {
	for _, coercion := range report.Coercions {
		logger.Info("WARNING: Chunk merge coercion for", coercion.Filename, "chunk", coercion.ChunkIndex+1, "field", coercion.Field+":", coercion.SourceType, "-", coercion.Action)
	}
	for _, conflict := range report.Conflicts {
		logger.Info("WARNING: Chunk merge conflict for", conflict.Filename, "field", conflict.Field+":", conflict.Resolution)
	}
	for _, failure := range report.Failures {
		logger.Error("Chunk merge failure for", failure.Filename, "artifact", failure.SequenceNumber, ":", failure.Message)
	}
}

func buildChunkingReport(plans []prompt.ChunkReport, mergeReport chunking.MergeReport) results.ChunkingReport {
	report := results.ChunkingReport{Documents: make([]results.ChunkingDocumentReport, 0, len(plans))}
	for index, plan := range plans {
		document := results.ChunkingDocumentReport{Filename: plan.Filename, CounterMethod: plan.CounterMethod, FullPromptTokens: plan.FullPromptTokens, ChunkPromptTokens: plan.ChunkPromptTokens, ChunkCount: plan.ChunkCount}
		for _, coercion := range mergeReport.Coercions {
			if coercion.DocumentIndex == index {
				document.Coercions = append(document.Coercions, coercion)
			}
		}
		for _, conflict := range mergeReport.Conflicts {
			if conflict.DocumentIndex == index {
				document.Conflicts = append(document.Conflicts, conflict)
			}
		}
		for _, failure := range mergeReport.Failures {
			if failure.DocumentIndex == index {
				document.Failures = append(document.Failures, failure)
			}
		}
		report.Documents = append(report.Documents, document)
	}
	return report
}

func failedDocumentCount(failures []chunking.Failure) int {
	documents := make(map[int]struct{}, len(failures))
	for _, failure := range failures {
		if failure.DocumentIndex >= 0 {
			documents[failure.DocumentIndex] = struct{}{}
		}
	}
	return len(documents)
}

func updateRevAIseExtraction(config *config.Config, reviewResults string, filenames, keys []string) error {
	if !config.RevAIse.IsEnabled() {
		return nil
	}

	models := make([]revaise.AIAssistance, 0, len(config.Project.LLM))
	for _, llm := range config.Project.LLM {
		models = append(models, revaise.AIAssistance{
			ID:          "prismaid_extraction_ai",
			Provider:    llm.Provider,
			Model:       llm.Model,
			Version:     "unspecified",
			Purpose:     []string{"EXTRACTION"},
			Temperature: fmt.Sprintf("%g", llm.Temperature),
			TPMLimit:    fmt.Sprintf("%d", llm.TpmLimit),
			RPMLimit:    fmt.Sprintf("%d", llm.RpmLimit),
		})
	}

	outputFormat := config.Project.Configuration.OutputFormat
	outputPath := config.Project.Configuration.ResultsFileName + "." + outputFormat
	return revaise.UpdateExtraction(config.RevAIse, revaise.ExtractionContribution{
		Review: revaise.ReviewSeed{
			ID:      config.Project.Name,
			Title:   config.Project.Name,
			Type:    "SYSTEMATIC_REVIEW",
			Status:  "IN_PROGRESS",
			Version: config.Project.Version,
			Authors: []string{
				config.Project.Author,
			},
		},
		Results:      reviewResults,
		Filenames:    filenames,
		Fields:       keys,
		Models:       models,
		ResultPath:   outputPath,
		ResultFormat: outputFormat,
	})
}
