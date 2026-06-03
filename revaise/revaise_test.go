package revaise

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateScreeningCreatesRecordAndUpdatesRound(t *testing.T) {
	tmp := t.TempDir()
	recordPath := filepath.Join(tmp, "review.revaise.json")
	cfg := Config{
		Enabled:    true,
		RecordFile: recordPath,
		Screening: ScreeningRoundConfig{
			RoundID:     "ta_pilot_001",
			RoundType:   "TITLE_ABSTRACT",
			RoundNumber: 1,
			RoundLabel:  "Pilot title and abstract screening",
		},
	}
	contribution := sampleScreeningContribution()

	if err := UpdateScreening(cfg, contribution); err != nil {
		t.Fatalf("UpdateScreening failed: %v", err)
	}
	if err := UpdateScreening(cfg, contribution); err != nil {
		t.Fatalf("second UpdateScreening failed: %v", err)
	}

	record := readRecord(t, recordPath)
	if record["review_id"] == "" {
		t.Fatal("expected review_id to be populated")
	}
	stages := record["stages"].([]any)
	if len(stages) != 1 {
		t.Fatalf("expected one stage, got %d", len(stages))
	}
	stage := stages[0].(map[string]any)
	rounds := stage["screening_rounds"].([]any)
	if len(rounds) != 1 {
		t.Fatalf("expected one screening round after rerun, got %d", len(rounds))
	}
	decisions := rounds[0].(map[string]any)["screening_decisions"].([]any)
	if len(decisions) != 2 {
		t.Fatalf("expected two screening decisions, got %d", len(decisions))
	}
}

func TestUpdateScreeningAppendsSecondRoundAndBacksUp(t *testing.T) {
	tmp := t.TempDir()
	recordPath := filepath.Join(tmp, "review.revaise.json")
	cfg := Config{
		Enabled:    true,
		RecordFile: recordPath,
		Screening: ScreeningRoundConfig{
			RoundID:     "ta_pilot_001",
			RoundType:   "TITLE_ABSTRACT",
			RoundNumber: 1,
		},
	}
	if err := UpdateScreening(cfg, sampleScreeningContribution()); err != nil {
		t.Fatalf("UpdateScreening failed: %v", err)
	}

	cfg.Screening.RoundID = "ta_full_001"
	cfg.Screening.RoundNumber = 2
	if err := UpdateScreening(cfg, sampleScreeningContribution()); err != nil {
		t.Fatalf("second round UpdateScreening failed: %v", err)
	}

	record := readRecord(t, recordPath)
	stage := record["stages"].([]any)[0].(map[string]any)
	rounds := stage["screening_rounds"].([]any)
	if len(rounds) != 2 {
		t.Fatalf("expected two screening rounds, got %d", len(rounds))
	}

	backups, err := filepath.Glob(filepath.Join(tmp, ".revaise-backups", "*.bak.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one backup, got %d", len(backups))
	}
}

func TestBackupCanBeDisabled(t *testing.T) {
	tmp := t.TempDir()
	recordPath := filepath.Join(tmp, "review.revaise.json")
	backup := false
	cfg := Config{
		Enabled:    true,
		RecordFile: recordPath,
		Backup:     &backup,
		Screening:  ScreeningRoundConfig{RoundID: "round_001"},
	}
	if err := UpdateScreening(cfg, sampleScreeningContribution()); err != nil {
		t.Fatalf("UpdateScreening failed: %v", err)
	}
	if err := UpdateScreening(cfg, sampleScreeningContribution()); err != nil {
		t.Fatalf("second UpdateScreening failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".revaise-backups")); !os.IsNotExist(err) {
		t.Fatalf("expected no backup directory, got err=%v", err)
	}
}

func TestUpdateExtractionCreatesExtractedStudy(t *testing.T) {
	tmp := t.TempDir()
	recordPath := filepath.Join(tmp, "review.revaise.json")
	cfg := Config{
		Enabled:    true,
		RecordFile: recordPath,
		Extraction: ExtractionRunConfig{
			RunID:       "pilot_extraction_001",
			Label:       "Pilot extraction",
			FormID:      "form_001",
			FormName:    "Pilot form",
			FormVersion: "1",
		},
	}
	results := `{
		"responses": [
			{
				"provider": "OpenAI",
				"model": "test-model",
				"sequence_id": "1",
				"sequence_number": 1,
				"model_responses": ["{\"population\":\"adults\",\"outcome\":\"access\"}"]
			}
		]
	}`
	if err := UpdateExtraction(cfg, ExtractionContribution{
		Review:    ReviewSeed{ID: "review", Title: "Review", Authors: []string{"Tester"}},
		Results:   results,
		Filenames: []string{"paper_001"},
		Fields:    []string{"population", "outcome"},
		Models: []AIAssistance{
			{Provider: "OpenAI", Model: "test-model"},
		},
	}); err != nil {
		t.Fatalf("UpdateExtraction failed: %v", err)
	}

	record := readRecord(t, recordPath)
	stage := record["stages"].([]any)[0].(map[string]any)
	studies := stage["extracted_studies"].([]any)
	if len(studies) != 1 {
		t.Fatalf("expected one extracted study, got %d", len(studies))
	}
	data := studies[0].(map[string]any)["extracted_data"].([]any)
	if len(data) != 2 {
		t.Fatalf("expected two extracted data points, got %d", len(data))
	}
}

func sampleScreeningContribution() ScreeningContribution {
	return ScreeningContribution{
		Review:          ReviewSeed{ID: "review", Title: "Review", Authors: []string{"Tester"}},
		TotalRecords:    2,
		IncludedRecords: 1,
		ExcludedRecords: 1,
		Records: []ScreeningRecord{
			{
				ID:           "rec_001",
				OriginalData: map[string]string{"title": "Included paper", "doi": "10.1/example"},
				Tags:         map[string]any{},
				Include:      true,
			},
			{
				ID:              "rec_002",
				OriginalData:    map[string]string{"title": "Excluded paper"},
				Tags:            map[string]any{},
				Include:         false,
				ExclusionReason: "Duplicate of rec_001",
			},
		},
		InclusionCriteria: []string{"Include relevant papers"},
		ExclusionCriteria: []string{"Exclude duplicates"},
	}
}

func readRecord(t *testing.T, path string) Record {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	record := Record{}
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}
