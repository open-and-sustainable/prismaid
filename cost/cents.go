package cost

import (
	"log"

	"github.com/shopspring/decimal"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	openai "github.com/sashabaranov/go-openai"
)

// Define a map to hold the rates for each model
var modelRates = map[string]decimal.Decimal{ // dollar prices per input M token
	openai.GPT4oMini:         decimal.NewFromFloat(0.15).Div(decimal.NewFromInt(1000000)),
	openai.GPT4o:             decimal.NewFromFloat(5).Div(decimal.NewFromInt(1000000)),
	openai.GPT4Turbo:         decimal.NewFromFloat(10).Div(decimal.NewFromInt(1000000)),
	openai.GPT4:              decimal.NewFromFloat(30).Div(decimal.NewFromInt(1000000)),
	openai.GPT432K:           decimal.NewFromFloat(60).Div(decimal.NewFromInt(1000000)),
	openai.GPT3Dot5Turbo:     decimal.NewFromFloat(0.5).Div(decimal.NewFromInt(1000000)),
	openai.GPT3Dot5TurboInstruct: decimal.NewFromFloat(1.5).Div(decimal.NewFromInt(1000000)),
	"gemini-1.5-flash":       decimal.NewFromFloat(0.15).Div(decimal.NewFromInt(1000000)), // the rate is halved if <= 128K input tokens, fixed below
	"gemini-1.5-pro":         decimal.NewFromFloat(2.5).Div(decimal.NewFromInt(1000000)),    // the rate is halved if <= 128K input tokens, fixed below
	"gemini-1.0-pro":         decimal.NewFromFloat(0.5).Div(decimal.NewFromInt(1000000)),
	"command-r7b-12-2024":    decimal.NewFromFloat(0.0375).Div(decimal.NewFromInt(1000000)),
	"command-r-plus":         decimal.NewFromFloat(2.5).Div(decimal.NewFromInt(1000000)),
	"command-r":              decimal.NewFromFloat(0.15).Div(decimal.NewFromInt(1000000)),
	"command-light":          decimal.NewFromFloat(0.3).Div(decimal.NewFromInt(1000000)),
	"command":                decimal.NewFromFloat(1).Div(decimal.NewFromInt(1000000)),
	anthropic.ModelClaude3_5SonnetLatest:      decimal.NewFromFloat(3).Div(decimal.NewFromInt(1000000)),
	anthropic.ModelClaude3_5HaikuLatest:        decimal.NewFromFloat(1).Div(decimal.NewFromInt(1000000)),
	anthropic.ModelClaude3OpusLatest:          decimal.NewFromFloat(15).Div(decimal.NewFromInt(1000000)),
	anthropic.ModelClaude_3_Sonnet_20240229:        decimal.NewFromFloat(3).Div(decimal.NewFromInt(1000000)),
	anthropic.ModelClaude_3_Haiku_20240307:         decimal.NewFromFloat(0.25).Div(decimal.NewFromInt(1000000)),
	"deepseek-chat":           decimal.NewFromFloat(0.14).Div(decimal.NewFromInt(1000000)),
}

func numCentsFromTokens(numTokens int, model string) decimal.Decimal {
	rate, ok := modelRates[model]
	if !ok {
		rate = decimal.Zero
		log.Println("Cost estimation unavailable because model not found:", model)
	}
	// halve the rate if the number of tokens is less than 128K and using Google AI Gemini 1.5 flash
	if numTokens <= 128000 && ((model == "gemini-1.5-flash") || (model == "gemini-1.5-pro")) {
		//if model == "gemini-1.5-flash" {
			rate = rate.Div(decimal.NewFromInt(2))	
		//} else if model == "gemini-1.5-pro" {
		//	rate = rate.Div(decimal.NewFromInt(4))
		//}
	}
	// Calculate the total cost in cents
	costInCents := decimal.NewFromInt(int64(numTokens)).Mul(rate)

	return costInCents
}
