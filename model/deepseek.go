package model

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"github.com/open-and-sustainable/prismaid/review"

	deepseek "github.com/cohesion-org/deepseek-go"
)

func queryDeepSeek(prompt string, llm review.Model, options review.Options) (string, string, string, error) {
	justification := ""
	summary := ""

	// Create a new deepseek client
	client := deepseek.NewClient(llm.APIKey)

	// Define your input data and create a prompt.
	messages := []deepseek.ChatCompletionMessage{{Role: deepseek.ChatMessageRoleUser, Content: prompt}}

	completionParams := &deepseek.ChatCompletionRequest{
		Model:    llm.Model,
		Messages: messages,
		ResponseFormat: &deepseek.ResponseFormat{Type: "json_object",},
		TopP:        float32(1.0),
		MaxTokens:   8192,
		Temperature: float32(llm.Temperature),
	}

	// Make the API call
	resp, err := client.CreateChatCompletion(context.Background(), completionParams)
	if err != nil {
		if apiErr, ok := err.(*deepseek.APIError); ok {
			log.Printf("API Error: StatusCode: %d, Message: %s", apiErr.StatusCode, apiErr.ErrorMsg)
		} else {
			log.Printf("Unexpected error: %v", err)
		}
		return "", "", "", fmt.Errorf("no response from deepseek: %v", err)
	}

	// Print the entire response object on log
	respJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Println("Failed to marshal response:", err)
		return "", "", "", err
	}
	log.Printf("Full deepseek response: %s\n", string(respJSON))

	// Assuming the content response is what you typically use:
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		log.Println("No content found in response")
		return "", "", "", fmt.Errorf("no content in response")
	}

	answer := resp.Choices[0].Message.Content

	if options.Justification {
		// Continue the conversation to ask for justification within the same chat
		messages = append(messages, deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleUser, Content: justification_query})

		justificationParams := &deepseek.ChatCompletionRequest{
			Model:       llm.Model,
			Messages:    messages, // Continue with the same conversation
			Temperature: float32(llm.Temperature),
		}

		justificationResp, err := client.CreateChatCompletion(context.Background(), justificationParams)
		if err != nil || len(justificationResp.Choices) != 1 {
			log.Printf("Justification error: err:%v len(choices):%v\n", err, len(justificationResp.Choices))
			return answer, "", "", fmt.Errorf("no justification response from deepseek: %v", err)
		}
		
		// Assign the justification content
		if len(justificationResp.Choices) > 0 {
			justification = justificationResp.Choices[0].Message.Content
		} else {
			log.Println("No content found in justification response")
		}
	}

	if options.Summary {
		// Continue the conversation to ask for summary within the same chat
		messages = append(messages, deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleUser, Content: summary_query})

		summaryParams := &deepseek.ChatCompletionRequest{
			Model:       llm.Model,
			Messages:    messages, // Continue with the same conversation
			Temperature: float32(llm.Temperature),
		}

		summaryResp, err := client.CreateChatCompletion(context.Background(), summaryParams)
		if err != nil || len(summaryResp.Choices) != 1 {
			log.Printf("Summary error: err:%v len(choices):%v\n", err, len(summaryResp.Choices))
			return answer, "", "", fmt.Errorf("no summary response from deepseek: %v", err)
		}
		
		// Assign the justification content
		if len(summaryResp.Choices) > 0 {
			summary = summaryResp.Choices[0].Message.Content
		} else {
			log.Println("No content found in summary response")
		}
	}

	return answer, justification, summary, nil
}
