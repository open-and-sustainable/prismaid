package model

import (
	"fmt"

	"github.com/open-and-sustainable/prismaid/review"
)

const justification_query = "For each one of the keys and answers you provided, provide a justification for your answer as a chain of thought. In particular, I want a textual description of the few stages of the chin of thought that lead you to the answer you provided and the sentences in the text you analyzes that support your decision. If the value of a key was 'no' or empty '' because of lack of information on that topic in the text analyzed, explicitly report this reason. Please provide only th einformation requested, neither introductory nor concluding remarks."
const summary_query = "Summarize in very few sentences the text provided to you before for your review."

// QueryService defines an interface for querying Large Language Models (LLMs).
// It allows for sending prompts to different LLMs and retrieving structured responses.
type QueryService interface {
    QueryLLM(prompt string, llm review.Model, options review.Options) (justification string, summary string, fullResponse string, err error)
}

type DefaultQueryService struct{}

// QueryLLM sends a prompt to a specified LLM (Large Language Model) and retrieves the model's response,
// including justifications and summaries if applicable.
//
// Arguments:
// - prompt: A string containing the input prompt to be processed.
// - llm: A pointer to the LLM configuration being used.
// - options: A pointer to the review.Options configuration that specifies additional processing options.
//
// Returns:
// - Three strings representing the justification, summary, and full response from the model.
// - An error if the interaction with the LLM fails or if processing issues occur.
func (dqs DefaultQueryService) QueryLLM(prompt string, llm review.Model, options review.Options) (string, string, string, error) {
    var queryFunc func(string, review.Model, review.Options) (string, string, string, error)

    switch llm.Provider {
    case "OpenAI":
        queryFunc = queryOpenAI
    case "GoogleAI":
        queryFunc = queryGoogleAI
    case "Cohere":
        queryFunc = queryCohere
    case "Anthropic":
        queryFunc = queryAnthropic
    case "DeepSeek":
        queryFunc = queryDeepSeek
    default:
        return "", "", "", fmt.Errorf("unsupported LLM provider: %s", llm.Provider)
    }

    return queryFunc(prompt, llm, options)
}

