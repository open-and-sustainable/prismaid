package prismaid

import (
	"strings"
	"testing"
)

func sampleZoteroConfigParams() ZoteroConfigParams {
	return ZoteroConfigParams{
		User:      "your_username",
		APIKey:    "your_api_key",
		Group:     "your_collection_name",
		OutputDir: "papers/zotero",
	}
}

// TestGenerateZoteroConfigValidates checks that a complete set of params yields
// TOML that passes Zotero validation.
func TestGenerateZoteroConfigValidates(t *testing.T) {
	toml := GenerateZoteroConfig(sampleZoteroConfigParams())

	if err := ValidateConfig("zotero", toml); err != nil {
		t.Fatalf("generated Zotero config failed validation: %v\n---\n%s", err, toml)
	}
	if strings.Contains(toml, "[revaise]") {
		t.Errorf("did not expect a [revaise] block without params:\n%s", toml)
	}
}

// TestGenerateZoteroConfigWithRevAIse checks the optional [revaise] search-stage
// block and that the config still validates.
func TestGenerateZoteroConfigWithRevAIse(t *testing.T) {
	params := sampleZoteroConfigParams()
	params.RevAIse = &ZoteroRevAIse{
		RecordFile:    "review.revaise.json",
		Format:        "json",
		SchemaVersion: "0.7.1",
		StageLabel:    "Zotero full-text download",
	}
	toml := GenerateZoteroConfig(params)

	if !strings.Contains(toml, `stage_type = "search"`) {
		t.Errorf("expected a search stage in the [revaise] block:\n%s", toml)
	}
	if err := ValidateConfig("zotero", toml); err != nil {
		t.Fatalf("generated Zotero config with revaise failed validation: %v\n---\n%s", err, toml)
	}
}

// TestGenerateZoteroConfigIsSeparateFromValidation confirms an incomplete config
// is still produced but flagged by validation.
func TestGenerateZoteroConfigIsSeparateFromValidation(t *testing.T) {
	params := sampleZoteroConfigParams()
	params.APIKey = "" // missing required field

	toml := GenerateZoteroConfig(params)
	if !strings.Contains(toml, "[zotero]") {
		t.Fatalf("expected a config to still be generated:\n%s", toml)
	}
	if err := ValidateConfig("zotero", toml); err == nil {
		t.Fatal("expected validation to flag the missing api_key")
	}
}
