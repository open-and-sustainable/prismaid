package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/open-and-sustainable/prismaid/revaise"
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
	RevAIse revaise.Config        `toml:"revaise"`
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
	// Chunking is an optional, explicit plan for prompts exceeding a known
	// input-context limit. It is disabled unless Chunking.Enabled is true.
	Chunking ChunkingConfig `toml:"chunking"`
}

// ChunkingConfig defines an explicit plan for safely processing manuscripts
// that exceed a user-defined input-context limit. Enabled defaults to false.
// When enabled, InputContextTokens and one Merge rule for every configured
// ReviewItem key are required. OverlapTokens defaults to zero. The input limit
// applies to the complete generated prompt, not merely the source text.
type ChunkingConfig struct {
	Enabled            bool                 `toml:"enabled"`
	InputContextTokens int                  `toml:"input_context_tokens"`
	OverlapTokens      int                  `toml:"overlap_tokens"`
	Merge              map[string]MergeRule `toml:"merge"`
}

// MergeRule defines the user-selected rule for combining one review field
// across document chunks. The required parameters depend on Rule:
//
//   - union: Sentinels
//   - ordinal: Order from weakest to strongest
//   - categorical: Defaults and TieBreak (currently "first")
//   - unique_text: Separator and MaxLength
//   - numeric: Operation ("max", "mean", or "min")
//   - metadata: OnMismatch ("warn" or "error")
//
// Rules are intentionally not inferred from model output or review values.
type MergeRule struct {
	Rule       string   `toml:"rule"`
	Sentinels  []string `toml:"sentinels"`
	Defaults   []string `toml:"defaults"`
	TieBreak   string   `toml:"tie_break"`
	Order      []string `toml:"order"`
	Separator  string   `toml:"separator"`
	MaxLength  int      `toml:"max_length"`
	Operation  string   `toml:"operation"`
	OnMismatch string   `toml:"on_mismatch"`
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

	if err := validate(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validate checks that a parsed review configuration contains the fields that
// are required for a meaningful review. It does not check API keys, which may be
// supplied through environment variables at run time.
func validate(c *Config) error {
	if c.Project.Configuration.InputDirectory == "" {
		return fmt.Errorf("project.configuration.input_directory is required")
	}
	if c.Project.Configuration.ResultsFileName == "" {
		return fmt.Errorf("project.configuration.results_file_name is required")
	}
	if len(c.Project.LLM) == 0 {
		return fmt.Errorf("at least one [project.llm] model is required")
	}
	for key, llm := range c.Project.LLM {
		if llm.Provider == "" {
			return fmt.Errorf("project.llm.%s.provider is required", key)
		}
	}
	if c.Prompt.Task == "" {
		return fmt.Errorf("prompt.task is required")
	}
	if c.Prompt.ExpectedResult == "" {
		return fmt.Errorf("prompt.expected_result is required")
	}
	if len(c.Review) == 0 {
		return fmt.Errorf("at least one [review] item is required")
	}
	for key, item := range c.Review {
		if item.Key == "" {
			return fmt.Errorf("review.%s.key is required", key)
		}
	}
	if err := validateChunking(c); err != nil {
		return err
	}
	return nil
}

func validateChunking(c *Config) error {
	chunking := c.Project.Configuration.Chunking
	if !chunking.Enabled {
		return nil
	}

	issues := make([]string, 0)
	if chunking.InputContextTokens <= 0 {
		issues = append(issues, "project.configuration.chunking.input_context_tokens is required")
	}
	if chunking.OverlapTokens < 0 {
		issues = append(issues, "project.configuration.chunking.overlap_tokens cannot be negative")
	}

	reviewKeys := make(map[string]struct{}, len(c.Review))
	for _, item := range c.Review {
		reviewKeys[item.Key] = struct{}{}
		if rule, ok := chunking.Merge[item.Key]; !ok {
			issues = append(issues, fmt.Sprintf("project.configuration.chunking.merge.%s is required", item.Key))
		} else if err := validateMergeRule(item.Key, rule); err != nil {
			issues = append(issues, err.Error())
		}
	}
	for key := range chunking.Merge {
		if _, ok := reviewKeys[key]; !ok {
			issues = append(issues, fmt.Sprintf("project.configuration.chunking.merge.%s does not match a review field", key))
		}
	}
	if len(issues) > 0 {
		return fmt.Errorf("chunking is enabled but the plan is incomplete:\n- %s", strings.Join(issues, "\n- "))
	}
	return nil
}

func validateMergeRule(key string, rule MergeRule) error {
	prefix := "project.configuration.chunking.merge." + key
	switch rule.Rule {
	case "union":
		if rule.Sentinels == nil {
			return fmt.Errorf("%s.sentinels is required for the union rule", prefix)
		}
	case "ordinal":
		if len(rule.Order) == 0 {
			return fmt.Errorf("%s.order is required for the ordinal rule", prefix)
		}
	case "categorical":
		if rule.Defaults == nil {
			return fmt.Errorf("%s.defaults is required for the categorical rule", prefix)
		}
		if rule.TieBreak != "first" {
			return fmt.Errorf("%s.tie_break must be \"first\" for the categorical rule", prefix)
		}
	case "unique_text":
		if rule.Separator == "" {
			return fmt.Errorf("%s.separator is required for the unique_text rule", prefix)
		}
		if rule.MaxLength <= 0 {
			return fmt.Errorf("%s.max_length must be greater than zero for the unique_text rule", prefix)
		}
	case "numeric":
		if rule.Operation != "max" && rule.Operation != "mean" && rule.Operation != "min" {
			return fmt.Errorf("%s.operation must be \"max\", \"mean\", or \"min\" for the numeric rule", prefix)
		}
	case "metadata":
		if rule.OnMismatch != "warn" && rule.OnMismatch != "error" {
			return fmt.Errorf("%s.on_mismatch must be \"warn\" or \"error\" for the metadata rule", prefix)
		}
	default:
		return fmt.Errorf("%s.rule must be one of \"union\", \"ordinal\", \"categorical\", \"unique_text\", \"numeric\", or \"metadata\"", prefix)
	}
	return nil
}
