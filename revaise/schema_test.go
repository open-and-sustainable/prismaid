package revaise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// TestEmittedRecordConformsToSchema builds a RevAIse record covering the parts
// prismAId writes (a search stage from a Zotero download, a screening stage with
// a round, and a data-extraction stage) and validates the emitted document
// against the vendored RevAIse JSON Schema in testdata.
//
// It guards against drift: if a change to the record builders, or a refresh of
// the vendored schema to a new RevAIse version, makes prismAId's output stop
// conforming, this test fails and reports exactly which fields are wrong.
func TestEmittedRecordConformsToSchema(t *testing.T) {
	tmp := t.TempDir()
	recordPath := filepath.Join(tmp, "review.revaise.json")
	base := Config{Enabled: true, RecordFile: recordPath}
	seed := ReviewSeed{ID: "review", Title: "Review", Authors: []string{"Tester"}}

	if err := UpdateStageOutputs(base, OutputContribution{
		Review:     seed,
		StageType:  "search",
		StageLabel: "Zotero full-text download",
		StageOutputs: []StageOutput{
			{Kind: "fulltexts", ResourceURI: "papers/zotero", Format: "directory"},
		},
	}); err != nil {
		t.Fatalf("UpdateStageOutputs failed: %v", err)
	}

	screeningCfg := base
	screeningCfg.Screening = ScreeningRoundConfig{
		RoundID:     "ta_pilot_001",
		RoundType:   "TITLE_ABSTRACT",
		RoundNumber: 1,
		RoundLabel:  "Pilot title and abstract screening",
	}
	if err := UpdateScreening(screeningCfg, sampleScreeningContribution()); err != nil {
		t.Fatalf("UpdateScreening failed: %v", err)
	}

	extractionCfg := base
	extractionCfg.Extraction = ExtractionRunConfig{
		RunID:       "pilot_extraction_001",
		FormID:      "form_001",
		FormName:    "Pilot form",
		FormVersion: "1",
	}
	results := `{"responses":[{"provider":"OpenAI","model":"test-model","sequence_id":"1","sequence_number":1,"model_responses":["{\"population\":\"adults\"}"]}]}`
	if err := UpdateExtraction(extractionCfg, ExtractionContribution{
		Review:    seed,
		Results:   results,
		Filenames: []string{"paper_001"},
		Fields:    []string{"population"},
		Models:    []AIAssistance{{Provider: "OpenAI", Model: "test-model"}},
	}); err != nil {
		t.Fatalf("UpdateExtraction failed: %v", err)
	}

	schema := compileSchema(t)
	instance := loadJSON(t, recordPath)
	if err := schema.Validate(instance); err != nil {
		t.Fatalf("emitted record does not conform to the RevAIse schema:\n%v", err)
	}
}

func compileSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	const id = "revaise.schema.json"
	doc := loadJSON(t, filepath.Join("testdata", "revaise.schema.json"))
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(id, doc); err != nil {
		t.Fatalf("add schema resource: %v", err)
	}
	schema, err := compiler.Compile(id)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return schema
}

func loadJSON(t *testing.T, path string) any {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()
	doc, err := jsonschema.UnmarshalJSON(file)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return doc
}
