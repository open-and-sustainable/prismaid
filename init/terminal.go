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
	Provider     string
	APIKey       string
	Model        string
	Temperature  string
	TpmLimit     string
	RpmLimit     string
	BaseURL      string
	EndpointType string
	Region       string
	ProjectID    string
	Location     string
	APIVersion   string
}

// RunInteractiveConfigCreation launches an interactive terminal session to collect project configuration
// information from the user. It guides the user through a comprehensive setup process using a variety
// of prompt styles to create a complete TOML configuration file.
//
// The function collects multiple types of configuration data:
// 1. Project metadata (name, author, version)
// 2. File system settings (input/output directories and formats)
// 3. Processing options (logging level, duplication, summaries)
// 4. LLM model configurations (providers, API keys, model selections)
// 5. Prompt components (persona, task descriptions, expected results)
// 6. Review criteria (items to review, definitions, examples)
//
// All collected information is structured into a TOML configuration file and saved to the
// user-specified location. If the user opts to skip certain sections, appropriate notices
// are displayed about manual configuration requirements.
//
// The function uses error checking throughout to ensure valid input, particularly for
// file paths and numeric values.
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

// collectModelItems interactively prompts the user to define LLM model configurations
// for the project. It repeatedly asks the user if they want to add model configurations,
// collecting details such as provider, API key, model name, temperature, and rate limits.
//
// The function uses an interactive command-line interface to:
// 1. Ask if the user wants to add a model configuration
// 2. If yes, prompt for a provider selection (OpenAI, GoogleAI, Cohere, etc.)
// 3. Collect an API key (with password masking for security)
// 4. Present provider-specific model choices
// 5. Collect temperature and rate limit settings (tpm/rpm)
// 6. Continue until the user chooses not to add more models
//
// Returns:
//   - A slice of ModelItem structures, each containing provider, API key,
//     model name, temperature, and rate limit configurations
//
// Each ModelItem will later be converted into a section in the TOML configuration
// file under the [project.llm] section.
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
					{Text: "Perplexity", Note: "Perplexity Sonar models."},
					{Text: "AWS Bedrock", Note: "AWS Bedrock cloud models."},
					{Text: "Azure AI", Note: "Azure OpenAI Service."},
					{Text: "Vertex AI", Note: "Google Cloud Vertex AI."},
					{Text: "SelfHosted", Note: "Self-hosted OpenAI-compatible endpoint."},
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
					{Text: "gpt-4.1", Note: "GPT-4.1."},
					{Text: "gpt-4.1-mini", Note: "GPT-4.1 Mini."},
					{Text: "gpt-4.1-nano", Note: "GPT-4.1 Nano."},
					{Text: "gpt-5", Note: "GPT-5."},
					{Text: "gpt-5.1", Note: "GPT-5.1."},
					{Text: "gpt-5.2", Note: "GPT-5.2."},
					{Text: "gpt-5-mini", Note: "GPT-5 Mini."},
					{Text: "gpt-5-nano", Note: "GPT-5 Nano."},
					{Text: "o1", Note: "o1."},
					{Text: "o1-mini", Note: "o1 Mini."},
					{Text: "o3", Note: "o3."},
					{Text: "o3-mini", Note: "o3 Mini."},
					{Text: "o4-mini", Note: "o4 Mini."},
				},
				choose.WithHelp(true))

		} else if provider == "GoogleAI" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "gemini-1.5-pro", Note: "Gemini 1.5 Pro."},
					{Text: "gemini-1.5-flash", Note: "Gemini 1.5 Flash."},
					{Text: "gemini-2.0-flash", Note: "Gemini 2.0 Flash."},
					{Text: "gemini-2.0-flash-lite", Note: "Gemini 2.0 Flash Lite."},
					{Text: "gemini-2.5-pro", Note: "Gemini 2.5 Pro."},
					{Text: "gemini-2.5-flash", Note: "Gemini 2.5 Flash."},
					{Text: "gemini-2.5-flash-lite", Note: "Gemini 2.5 Flash Lite."},
					{Text: "gemini-3-pro-preview", Note: "Gemini 3 Pro Preview."},
					{Text: "gemini-3-flash-preview", Note: "Gemini 3 Flash Preview."},
				},
				choose.WithHelp(true))
		} else if provider == "Cohere" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "command", Note: "Command."},
					{Text: "command-light", Note: "Command Light."},
					{Text: "command-r", Note: "Command R."},
					{Text: "command-r-08-2024", Note: "Command R August 2024."},
					{Text: "command-r-plus", Note: "Command R+."},
					{Text: "command-r7b-12-2024", Note: "Command R7B."},
					{Text: "command-a-03-2025", Note: "Command A."},
					{Text: "command-a-reasoning-08-2025", Note: "Command A Reasoning."},
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
					{Text: "claude-3-7-sonnet", Note: "Claude 3.7 Sonnet."},
					{Text: "claude-4-0-sonnet", Note: "Claude 4.0 Sonnet."},
					{Text: "claude-4-0-opus", Note: "Claude 4.0 Opus."},
					{Text: "claude-4-5-opus", Note: "Claude 4.5 Opus."},
					{Text: "claude-4-5-sonnet", Note: "Claude 4.5 Sonnet."},
					{Text: "claude-4-5-haiku", Note: "Claude 4.5 Haiku."},
				},
				choose.WithHelp(true))
		} else if provider == "DeepSeek" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "deepseek-chat", Note: "DeepSeek Chat - v3."},
					{Text: "deepseek-reasoner", Note: "DeepSeek Reasoner - v3."},
				},
				choose.WithHelp(true))
		} else if provider == "Perplexity" {
			model, err = prompt.New().Ask("Enter model to be used:").AdvancedChoose(
				[]choose.Choice{
					{Text: "", Note: "Model chosen automatically to minimize costs."},
					{Text: "sonar", Note: "Sonar."},
					{Text: "sonar-pro", Note: "Sonar Pro."},
					{Text: "sonar-reasoning-pro", Note: "Sonar Reasoning Pro."},
					{Text: "sonar-deep-research", Note: "Sonar Deep Research."},
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
		fmt.Printf("Added model: %s %s\n", provider, model)

		// Collect optional fields for cloud/self-hosted providers
		var baseURL, endpointType, region, projectID, location, apiVersion string

		if provider == "SelfHosted" {
			baseURL, err = prompt.New().Ask("Enter base URL (e.g., http://localhost:8000/v1):").Input("")
			checkErr(err)
		} else if provider == "AWS Bedrock" {
			endpointType = "bedrock"
			region, err = prompt.New().Ask("Enter AWS region (e.g., us-east-1):").Input("us-east-1")
			checkErr(err)
		} else if provider == "Azure AI" {
			endpointType = "azure"
			baseURL, err = prompt.New().Ask("Enter Azure OpenAI endpoint (e.g., https://your-resource.openai.azure.com):").Input("")
			checkErr(err)
			apiVersion, err = prompt.New().Ask("Enter API version (e.g., 2024-02-15-preview):").Input("2024-02-15-preview")
			checkErr(err)
		} else if provider == "Vertex AI" {
			endpointType = "vertex"
			projectID, err = prompt.New().Ask("Enter Google Cloud project ID:").Input("")
			checkErr(err)
			location, err = prompt.New().Ask("Enter location (e.g., us-central1):").Input("us-central1")
			checkErr(err)
		}

		modelItems = append(modelItems, ModelItem{
			Provider:     provider,
			APIKey:       apiKey,
			Model:        model,
			Temperature:  temperature,
			TpmLimit:     tpmLimit,
			RpmLimit:     rpmLimit,
			BaseURL:      baseURL,
			EndpointType: endpointType,
			Region:       region,
			ProjectID:    projectID,
			Location:     location,
			APIVersion:   apiVersion,
		})

		count++
	}

	return modelItems
}

// generateModelToml creates a formatted TOML string from a slice of ModelItem structures.
// It converts each model configuration into a TOML section with properties like provider,
// API key, model name, temperature, and rate limits. Each model is assigned a sequential
// number in the TOML output.
//
// Parameters:
//   - modelsItems: A slice of ModelItem structures containing model configurations
//
// Returns:
//   - A formatted string containing TOML configuration for the [project.llm] section
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
		if item.BaseURL != "" {
			tomlModelsSection.WriteString(fmt.Sprintf("base_url = \"%s\"\n", item.BaseURL))
		}
		if item.EndpointType != "" {
			tomlModelsSection.WriteString(fmt.Sprintf("endpoint_type = \"%s\"\n", item.EndpointType))
		}
		if item.Region != "" {
			tomlModelsSection.WriteString(fmt.Sprintf("region = \"%s\"\n", item.Region))
		}
		if item.ProjectID != "" {
			tomlModelsSection.WriteString(fmt.Sprintf("project_id = \"%s\"\n", item.ProjectID))
		}
		if item.Location != "" {
			tomlModelsSection.WriteString(fmt.Sprintf("location = \"%s\"\n", item.Location))
		}
		if item.APIVersion != "" {
			tomlModelsSection.WriteString(fmt.Sprintf("api_version = \"%s\"\n", item.APIVersion))
		}
		tomlModelsSection.WriteString("\n")
	}

	return tomlModelsSection.String()
}

// collectReviewItems interactively prompts the user to define review criteria
// for project configuration. It repeatedly asks the user if they want to add
// review items, collecting a key and a set of possible values for each one.
//
// The function uses an interactive command-line interface to:
// 1. Ask if the user wants to add a review item
// 2. If yes, prompt for a key name
// 3. Collect possible values as a comma-separated list
// 4. Convert the input into a ReviewItem structure
// 5. Continue until the user chooses not to add more items
//
// Returns:
//   - A slice of ReviewItem structures, each containing a key string and
//     a slice of possible values for that key
//
// Each ReviewItem will later be converted into a section in the TOML configuration
// file under the [review] section.
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

// generateReviewToml creates a formatted TOML string from a slice of ReviewItem structures.
// It converts each review item into a TOML section with a key and an array of possible values.
// Each review item is assigned a sequential number in the TOML output.
//
// Parameters:
//   - reviewItems: A slice of ReviewItem structures containing keys and their possible values
//
// Returns:
//   - A formatted string containing TOML configuration for the [review] section
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

// collectDefinitions interactively prompts the user to provide definitions for review items.
// It iterates through the provided review items, asking if the user wants to define
// each one. For each confirmed item, it collects a descriptive definition text.
//
// Parameters:
//   - reviewItems: A slice of ReviewItem structures containing keys and their possible values
//
// Returns:
//   - A concatenated string containing all the provided definitions with spaces in between
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

// collectExamples interactively prompts the user to provide examples for each review item.
// It iterates through the provided review items, asking if the user wants to create an
// example for each one. For each confirmed item, it collects a descriptive example text.
//
// Parameters:
//   - reviewItems: A slice of ReviewItem structures containing keys and their possible values
//
// Returns:
//   - A concatenated string containing all the provided examples with spaces in between
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

// generateTomlConfig creates a formatted TOML configuration string from the provided parameters.
// It structures the configuration into sections for project metadata, operational settings,
// LLM configuration, prompt components, and review criteria.
//
// Parameters:
//   - projectName: Name of the project
//   - author: Author of the project
//   - version: Version number
//   - inputDir: Directory containing input files
//   - resultsFileName: Name of the file to store results
//   - outputFormat: Format for output data (e.g., "csv", "json")
//   - logLevel: Logging verbosity level
//   - duplication: Whether to enable duplication for debugging
//   - cotJustification: Whether to enable chain-of-thought justification
//   - summary: Whether to enable document summarization
//   - models: Pre-formatted TOML string for LLM model configurations
//   - persona: Description of the AI persona for the review prompt
//   - task: Description of the task for the review prompt
//   - expected_result: Description of expected results for the review prompt
//   - failsafe: Fallback instructions for the review prompt
//   - definitions: Definitions of key terms used in the review
//   - example: Example reviews to guide the LLM
//   - review: Pre-formatted TOML string for review criteria
//
// Returns:
//   - A formatted TOML configuration string with all whitespace trimmed
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

// writeTomlConfigToFile writes the generated TOML configuration string to a file
// at the specified path. It creates the file if it doesn't exist or truncates it
// if it already exists.
//
// Parameters:
//   - config: A string containing the complete TOML configuration content
//   - filePath: The path where the configuration file should be saved
//
// Returns:
//   - An error if file creation or writing fails; nil otherwise
func writeTomlConfigToFile(config, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(config)
	return err
}

// checkErr is a helper function that handles error checking by terminating
// the program when an error is encountered. It prints the error message to
// standard output before exiting with status code 1.
//
// Parameters:
//   - err: The error to check. If nil, the function returns normally.
//     If non-nil, the error message is printed and the program exits.
//
// Note that this function will terminate the entire program if an error
// is detected, making it appropriate for initialization code where errors
// are non-recoverable.
func checkErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// validatePath checks if the provided path is valid by verifying both the directory
// and file components. It separates the path into directory and filename parts,
// then validates each part individually.
//
// Parameters:
//   - path: A string representing the full file path to validate
//
// Returns:
//   - An error if either the directory doesn't exist or the filename contains
//     invalid characters; nil if the entire path is valid
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

// validateDirectory checks if the given directory path exists and is a valid directory.
// It verifies both that the path exists in the filesystem and that it points to a
// directory rather than a regular file or other file system object.
//
// Parameters:
//   - dir: A string representing the directory path to validate
//
// Returns:
//   - An error if the directory doesn't exist, cannot be accessed, or is not a directory;
//     nil if the directory is valid and accessible
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

// validateFileName checks if the provided filename is valid for use in a file system.
// It validates that the filename doesn't contain characters that are invalid
// in most file systems, ensures it's not empty, and confirms it has a .toml extension.
//
// Parameters:
//   - fileName: The filename string to validate
//
// Returns:
//   - An error if the filename contains invalid characters, is empty, or
//     doesn't have a .toml extension; nil otherwise
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

// validateNonNegative checks if the provided string value represents a non-negative number.
// It returns an error if the value is empty or starts with a negative sign.
//
// Parameters:
//   - value: The string to validate
//
// Returns:
//   - An error if the value is empty or negative, nil otherwise
func validateNonNegative(value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if value[0] == '-' {
		return fmt.Errorf("value cannot be negative")
	}
	return nil
}
