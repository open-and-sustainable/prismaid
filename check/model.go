package check

import (
	"log"
	"github.com/open-and-sustainable/prismaid/tokens"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	openai "github.com/sashabaranov/go-openai"
)

// GetModel selects the appropriate model for the given provider based on user input and internal logic.
//
// Parameters:
//   - prompt: The input prompt provided by the user.
//   - providerName: The name of the AI provider (e.g., "OpenAI", "Cohere").
//   - modelName: The name of the specific model, if any. If empty, cost optimization is attempted.
//   - key: A string representing a key for the provider's service.
//
// Returns:
//   - A string representing the selected model name. An empty string is returned if the model is unsupported.
//
// Example:
//   > selectedModel := GetModel("some prompt", "OpenAI", "gpt-4-turbo", "api-key")
//   > if selectedModel == "" {
//   >     log.Println("No supported model selected")
//   > }
func GetModel(prompt string, providerName string, modelName string, key string) string {
	var modelFunc func(string, string, string) string
	switch providerName {
	case "OpenAI":
		modelFunc = getOpenAIModel
	case "GoogleAI":
		modelFunc = getGoogleAIModel
	case "Cohere":
		modelFunc = getCohereModel
	case "Anthropic":
		modelFunc = getAnthropicModel
	default:
		log.Println("Unsupported LLM provider: ", providerName)
		return ""
	}
	return modelFunc(prompt, modelName, key)
}

func getOpenAIModel(prompt string, modelName string, key string) string {
	model := openai.GPT4oMini
	switch modelName {
	case "": // cost optimization
		// old code before GPT 4 Omni mini model availability -- now the only solution to minimize the cost
		/*numTokens := numTokensFromMessages([]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}}, model)
		if numTokens > 16385 {
			model = openai.GPT4o
		}*/
	case "gpt-3.5-turbo":
		model = openai.GPT3Dot5Turbo
	case "gpt-4-turbo":
		model = openai.GPT4Turbo
	case "gpt-4o":
		model = openai.GPT4o
	case "gpt-4o-mini":
		model = openai.GPT4oMini
	default:
		log.Println("Unsopported model: ", modelName)
		return ""
	}
	return model
}

func getGoogleAIModel(prompt string, modelName string, key string) string {
	model := "gemini-1.0-pro"
	switch modelName {
	case "": // cost optimization, input token limit values: gemini-1.0-pro 30720, gemini-1.5-flash 1048576, gemini-1.5-pro 2097152
		counter := tokens.RealTokenCounter{}
		numTokens := counter.GetNumTokensFromPrompt(prompt, "GoogleAI", modelName, key)
		if numTokens > 30720 && numTokens <= 1048576 {
			model = "gemini-1.5-flash"
		} else if numTokens > 1048576 {
			model = "gemini-1.5-pro"
		}
	case "gemini-1.0-pro": // leave the model selected by the user, but chek if supported
		model = modelName
	case "gemini-1.5-flash":
		model = modelName
	case "gemini-1.5-pro":
		model = modelName
	default:
		log.Println("Unsopported model: ", modelName)
		return ""
	}
	return model
}

func getCohereModel(prompt string, modelName string, key string) string {
	model := "command-r"
	switch modelName {
	case "": 
		// cost optimization, command-r is currently the cheapest and with the most input tokens allowed
	case "command": // leave the model selected by the user, but chek if supported
		model = modelName
	case "command-light":
		model = modelName
	case "command-r":
		model = modelName
	case "command-r-plus":
		model = modelName
	case "command-r7b-12-2024":
		model = modelName			
	default:
		log.Println("Unsopported model: ", modelName)
		return ""
	}
	return model
}

func getAnthropicModel(prompt string, modelName string, key string) string {
	model := anthropic.ModelClaude_3_Haiku_20240307
	switch modelName {
	case "": // cost optimization
		// all models have the same context window size, hence leave to haiku as the cheapest
	case "claude-3-5-sonnet":
		model = anthropic.ModelClaude3_5SonnetLatest
	case "claude-3-5-haiku":
		model = anthropic.ModelClaude3_5HaikuLatest
	case "claude-3-opus":
		model = anthropic.ModelClaude3OpusLatest
	case "claude-3-sonnet":
		model = anthropic.ModelClaude_3_Sonnet_20240229
	case "claude-3-haiku":
		model = anthropic.ModelClaude_3_Haiku_20240307
	default:
		log.Println("Unsopported model: ", modelName)
		return ""
	}
	return model
}