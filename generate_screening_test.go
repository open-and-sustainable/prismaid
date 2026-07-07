package prismaid

import (
	"strings"
	"testing"
)

func sampleScreeningConfigParams() ScreeningConfigParams {
	return ScreeningConfigParams{
		Name:             "Test Screening",
		Author:           "Tester",
		Version:          "1.0",
		InputFile:        "/tmp/manuscripts.csv",
		OutputFile:       "/tmp/screening_output",
		TextColumn:       "abstract",
		IdentifierColumn: "doi",
		OutputFormat:     "csv",
		LogLevel:         "low",
		Language: ScreeningLanguage{
			Enabled:           true,
			AcceptedLanguages: []string{"en", "es"},
		},
	}
}

// TestGenerateScreeningConfigValidates checks that a complete set of params
// yields TOML that passes screening validation, with unquoted booleans/numbers.
func TestGenerateScreeningConfigValidates(t *testing.T) {
	toml := GenerateScreeningConfig(sampleScreeningConfigParams())

	if err := ValidateConfig("screening", toml); err != nil {
		t.Fatalf("generated screening config failed validation: %v\n---\n%s", err, toml)
	}
	if !strings.Contains(toml, "enabled = true") {
		t.Errorf("expected unquoted boolean, got:\n%s", toml)
	}
	if !strings.Contains(toml, `accepted_languages = ["en", "es"]`) {
		t.Errorf("expected inline string array, got:\n%s", toml)
	}
	// The single supported [filters.llm] table must not be present without an LLM.
	if strings.Contains(toml, "[filters.llm]") {
		t.Errorf("did not expect an [filters.llm] table without an LLM:\n%s", toml)
	}
}

// TestGenerateScreeningConfigWithLLMAndRevAIse checks the optional LLM table and
// [revaise] block, and that the result still validates.
func TestGenerateScreeningConfigWithLLMAndRevAIse(t *testing.T) {
	params := sampleScreeningConfigParams()
	params.LLM = &ScreeningLLM{Provider: "OpenAI", Model: "gpt-4o-mini", Temperature: 0.01}
	params.RevAIse = &ScreeningRevAIse{
		RecordFile:     "review.revaise.json",
		Format:         "json",
		SchemaVersion:  "0.7.1",
		HumanOversight: "NONE",
		StageLabel:     "Title and abstract screening",
		RoundID:        "ta_pilot_001",
		RoundType:      "TITLE_ABSTRACT",
		RoundNumber:    1,
		RoundLabel:     "Pilot screening",
		ReviewerID:     "prismaid",
		ReviewerRole:   "SCREENER",
	}
	toml := GenerateScreeningConfig(params)

	if !strings.Contains(toml, "[filters.llm]") {
		t.Errorf("expected an [filters.llm] table:\n%s", toml)
	}
	if !strings.Contains(toml, "[revaise.screening_round]") {
		t.Errorf("expected a [revaise.screening_round] table:\n%s", toml)
	}
	if err := ValidateConfig("screening", toml); err != nil {
		t.Fatalf("generated screening config failed validation: %v\n---\n%s", err, toml)
	}
}

// TestGenerateScreeningConfigIsSeparateFromValidation confirms a config with no
// enabled filter is still generated but flagged by validation.
func TestGenerateScreeningConfigIsSeparateFromValidation(t *testing.T) {
	params := sampleScreeningConfigParams()
	params.Language.Enabled = false // no filter enabled

	toml := GenerateScreeningConfig(params)
	if !strings.Contains(toml, "[project]") {
		t.Fatalf("expected a config to still be generated:\n%s", toml)
	}
	if err := ValidateConfig("screening", toml); err == nil {
		t.Fatal("expected validation to flag that no filter is enabled")
	}
}
