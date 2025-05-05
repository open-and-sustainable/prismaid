package config

import (
	"reflect"
	"testing"
)

// MockEnvReader implements the EnvReader interface for testing.
type MockEnvReader struct {
	values map[string]string
}

func (m *MockEnvReader) GetEnv(key string) string {
	return m.values[key]
}

// TestLoadConfig tests the LoadConfig function using the provided TOML configuration.
func TestLoadConfig(t *testing.T) {
	tomlContent := `
### The [project] section contains some basic information for internal reference
[project]
name = "Use of LLM for systematic review"   # Project title
author = "John Doe"                         # Project author
version = "1.0"                             # Project configuration version

### The [project.configuration] section contains the main parameters and options defining the review project
[project.configuration]
input_directory = "/path/to/txt/files"      # The location of the manuscript to be reviewed
results_file_name = "/path/to/save/results" # Location and filename for storing outputs, the path must exists, file extension will be added
output_format = "json"                      # Can be "csv" [default] or "json"
log_level = "low"                           # Can be "low" [default], "medium" showing entries on stdout, or "high" saving entries on file, see user manual for details
duplication = "no"                          # Can be "yes" or "no" [default]. It duplicates the manuscripts to review, hence running model queries twice, for debugging.
cot_justification = "no"                    # Can be "yes" or "no" [default]. It requests and saves the model justification in terms of chain of thought for the answers provided.
summary = "no"                              # Can be "yes" or "no" [default]. If positive, manuscript summaries will be generated an saved.

### The [project.llm] section, if more than 1 will be an ensemble project
[project.llm]
[project.llm.1]
provider = "OpenAI"
api_key = ""
model = "gpt-4o-mini"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[project.llm.2]
provider = "GoogleAI"
api_key = ""
model = "gemini-1.5-flash"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[project.llm.3]
provider = "Cohere"
api_key = ""
model = "command-r"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[project.llm.4]
provider = "Anthropic"
api_key = ""
model = "claude-3-haiku"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

### The [prompt] section defines the main components of the prompt for reviews
[prompt]
persona = "You are an experienced scientist working on a systematic review of the literature."
task = "You are asked to map the concepts discussed in a scientific paper attached here."
expected_result = "You should output a JSON object with the following keys and possible values: "
definitions = "'Interest rate' is the percentage charged by a lender for borrowing money or earned by an investor on a deposit over a specific period, typically expressed annually."
example = ""
failsafe = "If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value."

### The [review] section defines the JSON object storing the review items, i.e., the knowledge map that needs to be filled in
[review]
[review.1]
key = "interest rate"
values = [""]

[review.2]
key = "regression models"
values = ["yes", "no"]

[review.3]
key = "geographical scale"
values = ["world", "continent", "river basin"]
`

	// Create a mock environment reader
	mockEnvReader := &MockEnvReader{
		values: map[string]string{
			"OPENAI_API_KEY":    "env12345",
			"GOOGLE_AI_API_KEY": "env67890",
			"CO_API_KEY":        "env13579",
			"ANTHROPIC_API_KEY": "env24680",
		},
	}

	// Call LoadConfig with mock TOML configuration and mock environment reader
	config, err := LoadConfig(tomlContent, mockEnvReader)
	if err != nil {
		t.Fatalf("LoadConfig returned an unexpected error: %v", err)
	}

	// Define the expected configuration
	expectedConfig := &Config{
		Project: ProjectConfig{
			Name:    "Use of LLM for systematic review",
			Author:  "John Doe",
			Version: "1.0",
			Configuration: ProjectConfiguration{
				InputDirectory:   "/path/to/txt/files",
				ResultsFileName:  "/path/to/save/results",
				OutputFormat:     "json",
				LogLevel:         "low",
				Duplication:      "no",
				CotJustification: "no",
				Summary:          "no",
			},
			LLM: map[string]LLMItem{
				"1": {
					Provider:    "OpenAI",
					ApiKey:      "env12345", // Loaded from environment
					Model:       "gpt-4o-mini",
					Temperature: 0.01,
					TpmLimit:    0,
					RpmLimit:    0,
				},
				"2": {
					Provider:    "GoogleAI",
					ApiKey:      "env67890", // Loaded from environment
					Model:       "gemini-1.5-flash",
					Temperature: 0.01,
					TpmLimit:    0,
					RpmLimit:    0,
				},
				"3": {
					Provider:    "Cohere",
					ApiKey:      "env13579", // Loaded from environment
					Model:       "command-r",
					Temperature: 0.01,
					TpmLimit:    0,
					RpmLimit:    0,
				},
				"4": {
					Provider:    "Anthropic",
					ApiKey:      "env24680", // Loaded from environment
					Model:       "claude-3-haiku",
					Temperature: 0.01,
					TpmLimit:    0,
					RpmLimit:    0,
				},
			},
		},
		Prompt: PromptConfig{
			Persona:        "You are an experienced scientist working on a systematic review of the literature.",
			Task:           "You are asked to map the concepts discussed in a scientific paper attached here.",
			ExpectedResult: "You should output a JSON object with the following keys and possible values: ",
			Definitions:    "'Interest rate' is the percentage charged by a lender for borrowing money or earned by an investor on a deposit over a specific period, typically expressed annually.",
			Example:        "",
			Failsafe:       "If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value.",
		},
		Review: map[string]ReviewItem{
			"1": {
				Key:    "interest rate",
				Values: []string{""},
			},
			"2": {
				Key:    "regression models",
				Values: []string{"yes", "no"},
			},
			"3": {
				Key:    "geographical scale",
				Values: []string{"world", "continent", "river basin"},
			},
		},
	}

	// Compare the loaded configuration with the expected configuration
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("Loaded config does not match expected config.\nExpected: %+v\nGot: %+v", expectedConfig, config)
	}
}
