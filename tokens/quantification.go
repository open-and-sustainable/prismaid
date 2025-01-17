package tokens

import (
	"log"
)

// TokenCounter defines an interface for counting tokens in text prompts.
// It requires an implementation that can handle different providers and models,
// using provider-specific logic to interact with APIs or SDKs.
type TokenCounter interface {
    GetNumTokensFromPrompt(prompt string, provider string, model string, key string) int
}

// RealTokenCounter is an implementation of the TokenCounter interface that uses actual APIs.
// It supports multiple providers by making HTTP requests to their respective APIs
// to calculate the number of tokens in given text prompts.
type RealTokenCounter struct{}

// GetNumTokensFromPrompt calculates the number of tokens in a given prompt.
// This method dispatches requests to various provider APIs based on the provider specified,
// and extracts the token count from the API's response.
//
// Arguments:
//   - prompt: The input text to be analyzed.
//   - provider: The name of the AI provider, such as "OpenAI", "Cohere", or "GoogleAI".
//   - model: The specific model to use for the token calculation, relevant to the provider.
//   - key: The API key used for authentication with the provider's services.
//
// Returns:
//   - An integer representing the number of tokens in the prompt, or zero if the provider is unsupported.
//
// The function logs an error and returns zero if the provider is not supported.
func (rtc RealTokenCounter) GetNumTokensFromPrompt(prompt string, provider string, model string, key string) int {
    var numTokens int
    switch provider {
    case "OpenAI":
        numTokens = numTokensFromPromptOpenAI(prompt, model, key)
    case "GoogleAI":
        numTokens = numTokensFromPromptGoogleAI(prompt, model, key)
    case "Cohere":
        numTokens = numTokensFromPromptCohere(prompt, model, key)
    case "Anthropic":
        numTokens = numTokensFromPromptOpenAI(prompt, "gpt-4o", key)
    case "DeepSeek":
        numTokens = numTokensFromPromptOpenAI(prompt, "gpt-4o", key)
    default:
        log.Println("Unsupported LLM provider: ", provider)
        return 0
    }
    return numTokens
}