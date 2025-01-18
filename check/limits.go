package check

import (
	"fmt"

    "github.com/open-and-sustainable/prismaid/tokens"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	openai "github.com/sashabaranov/go-openai"
)

const (
    // OpenAI Models
    GPT4MiniMaxTokens      = 128000
    GPT4MaxTokens          = 128000
    GPT4TurboMaxTokens     = 128000
    GPT35TurboMaxTokens    = 16385
    // GoogleAI Models
    Gemini15FlashMaxTokens = 1048576
    Gemini15ProMaxTokens   = 2097152
    Gemini10ProMaxTokens   = 32760
	// Cohere Models
	CommandMaxTokens = 4096
	CommandLightMaxTokens = 4096
	CommandRMaxTokens = 128000
	CommandRPlusMaxTokens = 128000
    // Anthropic Models
    AnthropicMaxTokens = 200000
    // DeepSeek Models
    DeepSeekChatMaxTokens = 64000
)

var ModelMaxTokens = map[string]int{
    openai.GPT4oMini:   GPT4MiniMaxTokens,
    openai.GPT4o:       GPT4MaxTokens,
    openai.GPT4Turbo:   GPT4TurboMaxTokens,
    openai.GPT3Dot5Turbo:    GPT35TurboMaxTokens,
    "gemini-1.5-flash": Gemini15FlashMaxTokens,
    "gemini-1.5-pro":   Gemini15ProMaxTokens,
    "gemini-1.0-pro":   Gemini10ProMaxTokens,
    "command-r7b-12-2024": CommandRMaxTokens,
	"command-r-plus":   CommandRPlusMaxTokens,
	"command-r":        CommandRMaxTokens,
	"command-light":    CommandLightMaxTokens,
	"command":          CommandMaxTokens,
    anthropic.ModelClaude3_5SonnetLatest:      AnthropicMaxTokens,
    anthropic.ModelClaude3_5HaikuLatest:      AnthropicMaxTokens,
    anthropic.ModelClaude3OpusLatest:      AnthropicMaxTokens,
    anthropic.ModelClaude_3_Sonnet_20240229:      AnthropicMaxTokens,
    anthropic.ModelClaude_3_Haiku_20240307:      AnthropicMaxTokens,
    "deepseek-chat": DeepSeekChatMaxTokens,
}

// RunInputLimitsCheck verifies if the number of tokens in given prompts exceed the allowed limits for a specified model.
//
// Parameters:
//   - prompt: A string containing the prompt to be checked.
//   - provider: The name of the AI provider (e.g., "OpenAI", "GoogleAI").
//   - model: The name of the model to use for limit checks.
//   - key: A string representing a key for the provider's service.
//
// Returns:
//   - A string indicating the problem if a token limit is exceeded or an error occurred, otherwise an empty string.
//   - An error if any token limit is exceeded or if the model is not found.
func RunInputLimitsCheck(prompt string, provider string, model string, key string, counter tokens.TokenCounter) error {
    nofTokens := counter.GetNumTokensFromPrompt(prompt, provider, model, key)
    errOnLimits := checkIfTokensExceedsLimits(nofTokens, model)
    if errOnLimits != nil {
        return errOnLimits
    }
	return nil
}

func checkIfTokensExceedsLimits(nofTokens int, model string) error {
    maxTokens, exists := ModelMaxTokens[model]
    if !exists {
        return fmt.Errorf("model '%s' not found", model)
    }
    if nofTokens > maxTokens {
        return fmt.Errorf("number of tokens in prompt (%d) exceeds limits for model '%s' (max allowed: %d)", nofTokens, model, maxTokens)
    }
    return nil
}

