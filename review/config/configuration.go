package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// EnvReader is an interface for accessing environment variables.
type EnvReader interface {
	GetEnv(key string) string
}

type RealEnvReader struct{}

func (r RealEnvReader) GetEnv(key string) string {
	return os.Getenv(key)
}

// Config defines the top-level configuration structure, matching the TOML file layout.
type Config struct {
	Project ProjectConfig         `toml:"project"`
	Prompt  PromptConfig          `toml:"prompt"`
	Review  map[string]ReviewItem `toml:"review"`
}

// ProjectConfig holds details about the project, its metadata, and settings.
type ProjectConfig struct {
	Name          string               `toml:"name"`
	Author        string               `toml:"author"`
	Version       string               `toml:"version"`
	Configuration ProjectConfiguration `toml:"configuration"`
	LLM           map[string]LLMItem   `toml:"llm"`
}

// ProjectConfiguration defines various settings related to project input and output.
type ProjectConfiguration struct {
	InputDirectory   string `toml:"input_directory"`
	ResultsFileName  string `toml:"results_file_name"`
	OutputFormat     string `toml:"output_format"`
	LogLevel         string `toml:"log_level"`
	CotJustification string `toml:"cot_justification"`
	Duplication      string `toml:"duplication"`
	Summary          string `toml:"summary"`
}

// LLMConfig holds the configuration settings specific to the AI model being used.
type LLMItem struct {
	Provider     string  `toml:"provider"`
	ApiKey       string  `toml:"api_key"`
	Model        string  `toml:"model"`
	Temperature  float64 `toml:"temperature"`
	TpmLimit     int64   `toml:"tpm_limit"`
	RpmLimit     int64   `toml:"rpm_limit"`
	BaseURL      string  `toml:"base_url,omitempty"`      // For self-hosted OpenAI-compatible endpoints
	EndpointType string  `toml:"endpoint_type,omitempty"` // For cloud providers (AWS Bedrock, Azure, Vertex)
	Region       string  `toml:"region,omitempty"`        // For AWS Bedrock
	ProjectID    string  `toml:"project_id,omitempty"`    // For Vertex AI
	Location     string  `toml:"location,omitempty"`      // For Vertex AI
	APIVersion   string  `toml:"api_version,omitempty"`   // For Azure AI
}

// PromptConfig specifies the configurations related to task prompting.
type PromptConfig struct {
	Persona        string `toml:"persona"`
	Task           string `toml:"task"`
	ExpectedResult string `toml:"expected_result"`
	Failsafe       string `toml:"failsafe"`
	Definitions    string `toml:"definitions"`
	Example        string `toml:"example"`
}

// ReviewItem defines key-value pairs for review configurations.
type ReviewItem struct {
	Key    string   `toml:"key"`
	Values []string `toml:"values"`
}

// LoadConfig parses the given TOML configuration string and populates a Config structure.
// It also checks for missing API keys in the configuration and attempts to load them
// from environment variables using the provided EnvReader. Additionally, it sets
// default values for various configuration fields if they are not specified.
//
// Parameters:
//   - tomlConfiguration: A string containing the TOML configuration data.
//   - envReader: An instance of EnvReader, used to read environment variables for API keys.
//
// Returns:
//   - A pointer to a Config structure populated with the parsed configuration data.
//   - An error if the TOML data cannot be decoded or any other processing error occurs.
//
// The function handles the following:
//  1. Decoding the TOML configuration into the Config structure.
//  2. Checking for missing API keys and attempting to retrieve them from environment variables
//     based on the provider (OpenAI, GoogleAI, Cohere, Anthropic, DeepSeek).
//  3. Setting default values for missing or invalid configuration fields, such as
//     OutputFormat, LogLevel, CotJustification, Summary, and Duplication.
//  4. Ensuring that LLM configuration parameters like Temperature, TpmLimit, and RpmLimit are
//     non-negative by applying minimum value constraints.
func LoadConfig(tomlConfiguration string, envReader EnvReader) (*Config, error) {
	var config Config

	// Decode the TOML data
	if _, err := toml.Decode(tomlConfiguration, &config); err != nil {
		return nil, err
	}

	for key, llm := range config.Project.LLM {
		if llm.ApiKey == "" { // If API key is empty, look for it in environment variables
			switch llm.Provider {
			case "OpenAI":
				llm.ApiKey = envReader.GetEnv("OPENAI_API_KEY")
			case "GoogleAI":
				llm.ApiKey = envReader.GetEnv("GOOGLE_AI_API_KEY")
			case "Cohere":
				llm.ApiKey = envReader.GetEnv("CO_API_KEY")
			case "Anthropic":
				llm.ApiKey = envReader.GetEnv("ANTHROPIC_API_KEY")
			case "DeepSeek":
				llm.ApiKey = envReader.GetEnv("DEEPSEEK_API_KEY")
			case "Perplexity":
				llm.ApiKey = envReader.GetEnv("PERPLEXITY_API_KEY")
			case "AWS Bedrock":
				llm.ApiKey = envReader.GetEnv("AWS_ACCESS_KEY_ID")
			case "Azure AI":
				llm.ApiKey = envReader.GetEnv("AZURE_OPENAI_API_KEY")
			case "Vertex AI":
				llm.ApiKey = envReader.GetEnv("GOOGLE_APPLICATION_CREDENTIALS")
			case "SelfHosted":
				llm.ApiKey = envReader.GetEnv("SELF_HOSTED_API_KEY")
			}
		}

		if llm.Temperature < 0 {
			llm.Temperature = 0
		}
		if llm.TpmLimit < 0 {
			llm.TpmLimit = 0
		}
		if llm.RpmLimit < 0 {
			llm.RpmLimit = 0
		}
		// Update the map directly with the modified llm
		config.Project.LLM[key] = llm
	}

	if config.Project.Configuration.OutputFormat == "" {
		config.Project.Configuration.OutputFormat = "csv"
	}

	if config.Project.Configuration.LogLevel == "" {
		config.Project.Configuration.LogLevel = "low"
	}

	if config.Project.Configuration.CotJustification == "" {
		config.Project.Configuration.CotJustification = "no"
	}

	if config.Project.Configuration.Summary == "" {
		config.Project.Configuration.Summary = "no"
	}

	if config.Project.Configuration.Duplication == "" {
		config.Project.Configuration.Duplication = "no"
	}

	return &config, nil
}
