package prismaid

import (
	"strings"
	"testing"
)

func sampleReviewConfigParams() ReviewConfigParams {
	return ReviewConfigParams{
		Name:             "Test project",
		Author:           "Tester",
		Version:          "1.0",
		InputDirectory:   "/tmp/in",
		ResultsFileName:  "/tmp/out/results",
		OutputFormat:     "json",
		LogLevel:         "low",
		Duplication:      "no",
		CotJustification: "no",
		Summary:          "no",
		LLMs: []ReviewLLM{
			{Provider: "OpenAI", Model: "gpt-4o-mini", Temperature: 0.01},
		},
		Persona:        "You are an experienced scientist.",
		Task:           "Map the concepts discussed in the paper.",
		ExpectedResult: "A JSON object with the requested keys.",
		ReviewItems: []ReviewItem{
			{Key: "interest rate", Values: []string{""}},
		},
	}
}

// TestGenerateReviewConfigValidates checks that a complete set of params yields
// TOML that passes review validation, and that numeric fields are unquoted.
func TestGenerateReviewConfigValidates(t *testing.T) {
	toml := GenerateReviewConfig(sampleReviewConfigParams())

	if err := ValidateConfig("review", toml); err != nil {
		t.Fatalf("generated review config failed validation: %v\n---\n%s", err, toml)
	}
	if !strings.Contains(toml, "temperature = 0.01") {
		t.Errorf("expected unquoted temperature, got:\n%s", toml)
	}
	if strings.Contains(toml, `temperature = "`) {
		t.Errorf("temperature must not be quoted:\n%s", toml)
	}
	if !strings.Contains(toml, "tpm_limit = 0") || !strings.Contains(toml, "rpm_limit = 0") {
		t.Errorf("expected unquoted token/request limits:\n%s", toml)
	}
}

// TestGenerateReviewConfigWithRevAIse checks that the optional [revaise] block
// is emitted and the whole config still validates.
func TestGenerateReviewConfigWithRevAIse(t *testing.T) {
	params := sampleReviewConfigParams()
	params.RevAIse = &ReviewRevAIse{
		RecordFile:     "review.revaise.json",
		Format:         "json",
		SchemaVersion:  "0.7.1",
		HumanOversight: "NONE",
		StageLabel:     "AI-assisted extraction",
		RunID:          "full_extraction_001",
		RunLabel:       "Full extraction",
		FormID:         "extraction_form_v1",
		FormName:       "Extraction form",
		FormVersion:    "1",
		ExtractorID:    "prismaid",
	}
	toml := GenerateReviewConfig(params)

	if !strings.Contains(toml, "[revaise]") {
		t.Errorf("expected a [revaise] block:\n%s", toml)
	}
	if err := ValidateConfig("review", toml); err != nil {
		t.Fatalf("generated review config with revaise failed validation: %v\n---\n%s", err, toml)
	}
}

// TestGenerateReviewConfigIsSeparateFromValidation confirms generation and
// validation are independent: a partial config is still produced, and it is
// validation that reports the missing pieces.
func TestGenerateReviewConfigIsSeparateFromValidation(t *testing.T) {
	params := sampleReviewConfigParams()
	params.ReviewItems = nil // incomplete: no review items

	toml := GenerateReviewConfig(params)
	if !strings.Contains(toml, "[project]") {
		t.Fatalf("expected a config to still be generated:\n%s", toml)
	}
	if err := ValidateConfig("review", toml); err == nil {
		t.Fatal("expected validation to flag the missing review items")
	}
}
