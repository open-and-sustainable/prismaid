package init

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	prompt "github.com/cqroot/prompt"
	choose "github.com/cqroot/prompt/choose"
	input "github.com/cqroot/prompt/input"
)

// ReviewItem stores a single review item's key and associated values
type ReviewItem struct {
	Key    string
	Values []string
}

// ModelItem stores a single model configuration
type ModelItem struct {
	Provider    string
	APIKey      string
	Model       string
	Temperature string
	TpmLimit    string
	RpmLimit    string
}

// RunInteractiveConfig launches an interactive terminal session to collect project configuration information
// from the user. It utilizes advanced prompt features to provide a user-friendly way of setting key project
// settings.
//
// This function offers different types of prompts such as single and multiple choices, allowing the user
// to customize configurations based on project requirements.
func RunInteractiveConfigCreation() {
	fmt.Println("Running interactive project configuration initialization...")

	// Ask for file path to save the configuration
	filePath, err := prompt.New().Ask("Enter file path to save the configuration:").Input(
		"./config.toml", input.WithHelp(true), input.WithValidateFunc(validatePath))
	checkErr(err)

	// Prompt for project name with help text
	projectName, err := prompt.New().Ask("Enter project name:").Input(
		"Test project",
		input.WithHelp(true),
	)
	checkErr(err)

	// Prompt for author name with help
	author, err := prompt.New().Ask("Enter author name:").Input(
		"Name Lastname",
		input.WithHelp(true),
	)
	checkErr(err)

	// Prompt for version
	version, err := prompt.New().Ask("Enter project version:").Input(
		"0.1",
		input.WithHelp(true),
	)
	checkErr(err)

	inputDir := ""

	// Configuration details with help for each choice
	inputDir, err = prompt.New().Ask("Enter input directory (must exist):").Input(
		"./",
		input.WithHelp(true), input.WithValidateFunc(validateDirectory))
	checkErr(err)

	resultsFileName, err := prompt.New().Ask("Enter results directory (must exist):").Input(
		"./",
		input.WithHelp(true), input.WithValidateFunc(validateDirectory))
	checkErr(err)

	// Output format
	outputFormat, err := prompt.New().Ask("Choose output format:").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "csv", Note: "Comma-separated values format for easier readability."},
				{Text: "json", Note: "JavaScript Object Notation format for structured data."},
			},
			choose.WithHelp(true))
	checkErr(err)

	// Log level with help
	logLevel, err := prompt.New().Ask("Choose log level:").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "low", Note: "Low verbosity: minimal logging."},
				{Text: "medium", Note: "High verbosity: logs displayed on stdout."},
				{Text: "high", Note: "High verbosity: logs saved to a file for detailed review."},
			},
			choose.WithHelp(true))
	checkErr(err)

	// Duplication option with help
	duplication, err := prompt.New().Ask("Enable duplication (for debugging)?").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "no", Note: "Do not duplicate reviews."},
				{Text: "yes", Note: "Duplicate the manuscripts to review, and the cost, useful for consistency checks."},
			},
			choose.WithHelp(true))
	checkErr(err)

	// Chain-of-thought justification option
	cotJustification, err := prompt.New().Ask("Enable chain-of-thought justification (saved on file)?").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "no", Note: "Do not enable chain-of-thought justification."},
				{Text: "yes", Note: "Enable model justification for the answers in terms of chain of thought."},
			},
			choose.WithHelp(true))
	checkErr(err)

	// Manuscript summary
	summary, err := prompt.New().Ask("Enable document summary (saved on file)?").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "no", Note: "Do not enable document summary."},
				{Text: "yes", Note: "Enable the preparation fo a short summary for each document reviewed."},
			},
			choose.WithHelp(true))
	checkErr(err)

	// Build models object
	models_items := collectModelItems()
	models := ""
	if len(models_items) > 0 {
		models = generateModelToml(models_items)
	} else {
		fmt.Println("You will have to specify the LLM parameters in your project configuration file.")
	}

	// Prompt for persona part of prompt
	persona := ""
	choice_persona, err := prompt.New().Ask("Do you confirm the standard 'persona' part of the review prompt?").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "yes", Note: "'You are an experienced scientist working on a systematic review of the literature.'"},
				{Text: "no", Note: "I will ask you to provide a new text."},
			},
			choose.WithHelp(true))
	checkErr(err)
	if choice_persona == "yes" {
		persona = "You are an experienced scientist working on a systematic review of the literature."
	} else {
		persona, err = prompt.New().Ask("Enter your persona description:").Input("", input.WithHelp(true))
		checkErr(err)
	}
	fmt.Printf("You selected: %s\n", persona)

	// Prompt for task part of prompt
	task := ""
	choice_task, err := prompt.New().Ask("Do you confirm the standard 'task' part of the review prompt?").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "yes", Note: "'You are asked to map the concepts discussed in a scientific paper attached here.'"},
				{Text: "no", Note: "I will ask you to provide a new text."},
			},
			choose.WithHelp(true))
	checkErr(err)
	if choice_task == "yes" {
		task = "You are asked to map the concepts discussed in a scientific paper attached here."
	} else {
		task, err = prompt.New().Ask("Enter your task description:").Input("", input.WithHelp(true))
		checkErr(err)
	}
	fmt.Printf("You selected: %s\n", task)

	// Prompt for expected_result part of prompt
	expected_result := ""
	choice_exp_result, err := prompt.New().Ask("Do you confirm the standard 'expected_result' part of the review prompt?").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "yes", Note: "'You should output a JSON object with the following keys and possible values:'"},
				{Text: "no", Note: "I will ask you to provide a new text."},
			},
			choose.WithHelp(true))
	checkErr(err)
	if choice_exp_result == "yes" {
		expected_result = "You should output a JSON object with the following keys and possible values:"
	} else {
		expected_result, err = prompt.New().Ask("Enter your expected_result description:").Input("", input.WithHelp(true))
		checkErr(err)
	}
	fmt.Printf("You selected: %s\n", expected_result)

	review := ""
	definitions := ""
	example := ""

	// Build answer object
	review_items := collectReviewItems()
	if len(review_items) > 0 {
		review = generateReviewToml(review_items)

		// Build definitions object
		definitions = collectDefinitions(review_items)

		// Build example object
		// Prompt for failsafe part of prompt
		example = ""
		choice_example, err := prompt.New().Ask("Do you want to provide examples for the review items?").
			AdvancedChoose(
				[]choose.Choice{
					{Text: "no", Note: "This section of the prompt will be left empty."},
					{Text: "yes, one by one", Note: "I will ask you to provide an example for each item separately."},
					{Text: "yes, as a whole", Note: "I will ask you to provide a single text example."},
				},
				choose.WithHelp(true))
		checkErr(err)
		if choice_example == "yes, one by one" {
			example = collectExamples(review_items)
		} else {
			if choice_example == "yes, as a whole" {
				example, err = prompt.New().Ask("Enter your example:").Input("The text 'Lorem ipsum' once reviewed should provide the JSON object [language = \"latin\", if_empty = \"yes\"]", input.WithHelp(true))
				checkErr(err)
			}
		}
	} else {
		fmt.Println("You will have to fill in review items, definitions and examples in your project configuration file.")
	}

	// Prompt for failsafe part of prompt
	failsafe := ""
	choice_failsafe, err := prompt.New().Ask("Do you confirm the standard 'failsafe' part of the review prompt?").
		AdvancedChoose(
			[]choose.Choice{
				{Text: "yes", Note: "'If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value.'"},
				{Text: "no", Note: "I will ask you to provide a new text."},
			},
			choose.WithHelp(true))
	checkErr(err)
	if choice_failsafe == "yes" {
		failsafe = "If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value."
	} else {
		failsafe, err = prompt.New().Ask("Enter your task description:").Input("", input.WithHelp(true))
		checkErr(err)
	}
	fmt.Printf("You selected: %s\n", failsafe)

	// Generate TOML config from user inputs
	config := generateTomlConfig(
		projectName, author, version,
		inputDir, resultsFileName, outputFormat, logLevel,
		duplication, cotJustification, summary, models,
		persona, task, expected_result,
		failsafe, definitions, example, review,
	)

	// Write the configuration to file
	err = writeTomlConfigToFile(config, filePath)
	if err != nil {
		fmt.Println("Error writing configuration file:", err)
	} else {
		fmt.Println("Configuration file created successfully at:", filePath)
	}
}

func collectModelItems() []ModelItem {
	var modelItems []ModelItem
	count := 1

	for {
		// Ask if the user wants to define a review item
		addItem, err := prompt.New().Ask(fmt.Sprintf("Do you want to add the configuration of generative AI model #%d? (yes/no)", count)).
			Choose([]string{"yes", "no"},
				choose.WithHelp(true))
		checkErr(err)

		// Break the loop if the user doesn't want to add more models
		if addItem == "no" {
			break
		}

		// LLM provider selection with help
		provider, err := prompt.New().Ask("Choose LLM provider:").
			AdvancedChoose(
				[]choose.Choice{
					{Text: "OpenAI", Note: "OpenAI GPT-3 or GPT-4 models."},
					{Text: "GoogleAI", Note: "GoogleAI Gemini models."},
					{Text: "Cohere", Note: "Cohere language models."},
					{Text: "Anthropic", Note: "Anthropic Claude models."},
					{Text: "DeepSeek", Note: "DeepSeek models."},
				},
				choose.WithHelp(true))
		checkErr(err)

		// Prompt for API key with input mask (for security)
		apiKey, err := prompt.New().Ask("Enter LLM API key (leave it empty to use environment variable):").Input("", input.WithEchoMode(input.EchoPassword))
		checkErr(err)

		// Model choice for the selected LLM provider
		model := ""
		if provider == "OpenAI" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "gpt-3.5-turbo", Note: "GPT-3.5 Turbo."},
					{Text: "gpt-4-turbo", Note: "GPT-4 Turbo."},
					{Text: "gpt-4o", Note: "GPT-4 Omni."},
					{Text: "gpt-4o-mini", Note: "GPT-4 Omni Mini."},
				},
				choose.WithHelp(true))

		} else if provider == "GoogleAI" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "gemini-1.0-pro", Note: "Gemini 1.0 Pro."},
					{Text: "gemini-1.5-pro", Note: "Gemini 1.5 Pro."},
					{Text: "gemini-1.5-flash", Note: "Gemini 1.5 Flash."},
				},
				choose.WithHelp(true))
		} else if provider == "Cohere" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "command", Note: "Command."},
					{Text: "command-light", Note: "Command Light."},
					{Text: "command-r", Note: "Command R."},
					{Text: "command-r-plus", Note: "Command R+."},
					{Text: "command-r7b-12-2024", Note: "Command R7B."},
				},
				choose.WithHelp(true))
		} else if provider == "Anthropic" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "claude-3-haiku", Note: "Claude 3 Haiku."},
					{Text: "claude-3-sonnet", Note: "Claude 3 Sonnet."},
					{Text: "claude-3-opus", Note: "Claude 3 Opus."},
					{Text: "claude-3-5-haiku", Note: "Claude 3.5 Haiku."},
					{Text: "claude-3-5-sonnet", Note: "Claude 3.5 Sonnet."},
				},
				choose.WithHelp(true))
		} else if provider == "DeepSeek" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "deepseek-chat", Note: "DeepSeek Chat - v3."},
				},
				choose.WithHelp(true))
		}
		checkErr(err)

		// Prompt for model temperature
		temperature, err := prompt.New().Ask("Enter model temperature (usually between 0 and 1 or 2):").Input(
			"0",
			input.WithHelp(true), input.WithValidateFunc(validateNonNegative))
		checkErr(err)

		// Prompt for tpm limit
		tpmLimit, err := prompt.New().Ask("Enter maximum token per minute (0 to disable):").Input(
			"0",
			input.WithHelp(true), input.WithValidateFunc(validateNonNegative))
		checkErr(err)

		// Prompt for rpm limit
		rpmLimit, err := prompt.New().Ask("Enter maximum request per minute (0 to disable):").Input(
			"0",
			input.WithHelp(true), input.WithValidateFunc(validateNonNegative))
		checkErr(err)

		// Create a new ReviewItem and append it to the list
		modelItems = append(modelItems, ModelItem{
			Provider:    provider,
			APIKey:      apiKey,
			Model:       model,
			Temperature: temperature,
			TpmLimit:    tpmLimit,
			RpmLimit:    rpmLimit,
		})

		count++
	}

	return modelItems
}

func generateModelToml(modelsItems []ModelItem) string {
	var tomlModelsSection strings.Builder

	// Loop through the review items and append each one to the TOML string
	for i, item := range modelsItems {
		tomlModelsSection.WriteString(fmt.Sprintf("[project.llm.%d]\n", i+1))
		tomlModelsSection.WriteString(fmt.Sprintf("provider = \"%s\"\n", item.Provider))
		tomlModelsSection.WriteString(fmt.Sprintf("api_key = \"%s\"\n", item.APIKey))
		tomlModelsSection.WriteString(fmt.Sprintf("model = \"%s\"\n", item.Model))
		tomlModelsSection.WriteString(fmt.Sprintf("temperature = \"%s\"\n", item.Temperature))
		tomlModelsSection.WriteString(fmt.Sprintf("tpm_limit = \"%s\"\n", item.TpmLimit))
		tomlModelsSection.WriteString(fmt.Sprintf("rpm_limit = \"%s\"\n", item.RpmLimit))
		tomlModelsSection.WriteString("\n")
	}

	return tomlModelsSection.String()
}

// Function to interactively collect review items and generate the [review] section of the TOML file
func collectReviewItems() []ReviewItem {
	var reviewItems []ReviewItem
	count := 1

	for {
		// Ask if the user wants to define a review item
		addItem, err := prompt.New().Ask(fmt.Sprintf("Do you want to add review item #%d? (yes/no)", count)).
			Choose([]string{"yes", "no"},
				choose.WithHelp(true))
		checkErr(err)

		// Break the loop if the user doesn't want to add more items
		if addItem == "no" {
			break
		}

		// Prompt for the key
		key, err := prompt.New().Ask(fmt.Sprintf("Enter key for review item #%d:", count)).Input("", input.WithHelp(true))
		checkErr(err)

		// Prompt for the list of values (comma-separated)
		valuesInput, err := prompt.New().Ask(fmt.Sprintf("Enter possible values for review item #%d (comma-separated, e.g.: '1, 2, 3'):", count)).Input("", input.WithHelp(true))
		checkErr(err)

		// Split the values by comma and store them in a slice
		values := strings.Split(valuesInput, ",")

		// Create a new ReviewItem and append it to the list
		reviewItems = append(reviewItems, ReviewItem{
			Key:    key,
			Values: values,
		})

		count++
	}

	return reviewItems
}

// Helper function to generate the TOML configuration string for the [review] section
func generateReviewToml(reviewItems []ReviewItem) string {
	var tomlReviewSection strings.Builder

	// Loop through the review items and append each one to the TOML string
	for i, item := range reviewItems {
		tomlReviewSection.WriteString(fmt.Sprintf("[review.%d]\n", i+1))
		tomlReviewSection.WriteString(fmt.Sprintf("key = \"%s\"\n", item.Key))
		tomlReviewSection.WriteString("values = [")
		for j, value := range item.Values {
			tomlReviewSection.WriteString(fmt.Sprintf("\"%s\"", strings.TrimSpace(value)))
			if j < len(item.Values)-1 {
				tomlReviewSection.WriteString(", ")
			}
		}
		tomlReviewSection.WriteString("]\n")
	}

	return tomlReviewSection.String()
}

// Function to interactively collect definitions based on review items
func collectDefinitions(reviewItems []ReviewItem) string {
	definitions := ""
	for i, rev := range reviewItems {
		// Ask if the user wants to define a review item
		addItem, err := prompt.New().Ask(fmt.Sprintf("Do you want to add a definition for review item #%d with key = '%s'? (yes/no)", i, rev.Key)).
			Choose([]string{"yes", "no"})
		checkErr(err)

		// Break the loop if the user doesn't want to add more items
		if addItem == "no" {
			break
		}

		// Prompt for the example
		def, err := prompt.New().Ask(fmt.Sprintf("Enter '%s' definition:", rev.Key)).Input(fmt.Sprintf("As '%s' we intend ...", rev.Key), input.WithHelp(true))
		checkErr(err)

		// Add the definition to the definitions string
		definitions += def + " "
	}

	return definitions
}

// Function to interactively collect examples based on review items
func collectExamples(reviewItems []ReviewItem) string {
	examples := ""
	for i, rev := range reviewItems {
		// Ask if the user wants to make an example for a review item
		addItem, err := prompt.New().Ask(fmt.Sprintf("Do you want to make an example for review item #%d with key = '%s'? (yes/no)", i, rev.Key)).
			Choose([]string{"yes", "no"})
		checkErr(err)

		// Break the loop if the user doesn't want to add more items
		if addItem == "no" {
			break
		}

		// Prompt for the example
		exa, err := prompt.New().Ask(fmt.Sprintf("Enter '%s' example:", rev.Key)).Input(fmt.Sprintf("'%s' takes value .. if reviewing the sentence ..", rev.Key), input.WithHelp(true))
		checkErr(err)

		// Add the definition to the definitions string
		examples += exa + " "
	}

	return examples
}

// Helper function to generate the TOML configuration string
func generateTomlConfig(projectName, author, version, inputDir, resultsFileName, outputFormat,
	logLevel, duplication, cotJustification, summary, models,
	persona, task, expected_result, failsafe, definitions, example, review string) string {
	config := fmt.Sprintf(`
[project]
name = "%s"
author = "%s"
version = "%s"

[project.configuration]
input_directory = "%s"
results_file_name = "%s"
output_format = "%s"
log_level = "%s"
duplication = "%s"
cot_justification = "%s"
summary = "%s"

[project.llm]
%s
[prompt]
persona = "%s"
task = "%s"
expected_result = "%s"
failsafe = "%s"
definitions = "%s"
example = "%s"

[review]
%s
`, projectName, author, version, inputDir, resultsFileName, outputFormat,
		logLevel, duplication, cotJustification, summary, models,
		persona, task, expected_result, failsafe, definitions, example, review)
	return strings.TrimSpace(config)
}

// Helper function to write the TOML configuration to a file
func writeTomlConfigToFile(config, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(config)
	return err
}

// CheckErr is a helper to check and handle errors
func checkErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// validatePath checks if the directory in the given path exists and if the file name is valid.
func validatePath(path string) error {
	// Separate the directory from the file
	dir := filepath.Dir(path)
	file := filepath.Base(path)

	// Check if the directory path is valid
	if err := validateDirectory(dir); err != nil {
		return err
	}

	// Check if the file name contains invalid characters
	if err := validateFileName(file); err != nil {
		return err
	}

	// Path is valid
	return nil
}

// validateDirectory checks if the given directory is valid.
func validateDirectory(dir string) error {
	// Check if the directory exists and is a valid directory
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s: %w", dir, fmt.Errorf("invalid path"))
	} else if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Directory is valid
	return nil
}

// validateFileName checks if the file name contains invalid characters
func validateFileName(fileName string) error {
	// Define a regular expression for invalid characters in file names
	// For example, on Windows: <>:"/\|?*
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)

	if invalidChars.MatchString(fileName) {
		return fmt.Errorf("%s: %w", fileName, fmt.Errorf("invalid filename"))
	}

	// You can also check for empty filenames or other restrictions, like file extension:
	if fileName == "" {
		return fmt.Errorf("filename cannot be empty: %w", fmt.Errorf("invalid filename"))
	}

	// Check .tom;:
	if !strings.HasSuffix(fileName, ".toml") {
		return fmt.Errorf("filename must have a .toml extension")
	}

	return nil
}

func validateNonNegative(value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if value[0] == '-' {
		return fmt.Errorf("value cannot be negative")
	}
	return nil
}
