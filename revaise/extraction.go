package revaise

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ExtractionContribution contains prismAId review/extraction output to merge
// into a RevAIse ExtractionStage.
type ExtractionContribution struct {
	Review       ReviewSeed
	Results      string
	Filenames    []string
	Fields       []string
	Models       []AIAssistance
	ResultPath   string
	ResultFormat string
}

type extractionOutput struct {
	Responses []extractionResponse `json:"responses"`
}

type extractionResponse struct {
	Provider       string   `json:"provider"`
	Model          string   `json:"model"`
	SequenceID     string   `json:"sequence_id"`
	SequenceNumber int      `json:"sequence_number"`
	ModelResponses []string `json:"model_responses"`
}

// UpdateExtraction merges review/extraction output into a RevAIse Review
// record.
//
// The configured extraction_run.run_id identifies the extraction run. Reusing
// the same run identity updates matching extracted studies for that run.
func UpdateExtraction(cfg Config, contribution ExtractionContribution) error {
	if !cfg.IsEnabled() {
		return nil
	}
	if cfg.Extraction.RunID == "" {
		return fmt.Errorf("revaise extraction_run.run_id is required when revaise is enabled for review/extraction")
	}
	return Update(cfg, contribution.Review, func(record Record) error {
		return applyExtraction(record, cfg, contribution)
	})
}

func applyExtraction(record Record, cfg Config, contribution ExtractionContribution) error {
	formID := firstNonEmpty(cfg.Extraction.FormID, "prismaid_extraction_form")
	formName := firstNonEmpty(cfg.Extraction.FormName, "prismAId extraction form")
	formVersion := firstNonEmpty(cfg.Extraction.FormVersion, "1")
	extractorID := firstNonEmpty(cfg.Extraction.ExtractorID, "prismaid")
	completedAt := firstNonEmpty(cfg.Extraction.CompletedAt, nowRFC3339())

	stageLabel := firstNonEmpty(cfg.Stage.Label, cfg.Extraction.Label, "Data extraction")
	stageCfg := cfg
	stageCfg.Stage.Label = stageLabel
	stage := stageFromConfig(stageCfg, "data_extraction", stageLabel)
	stage["extraction_protocol"] = map[string]any{
		"base_protocol_id":             cfg.Extraction.RunID + "_protocol",
		"minimum_extractors_per_study": 1,
		"independent_extraction":       false,
	}
	stage["extraction_forms"] = upsertForm(list(stage, "extraction_forms"), map[string]any{
		"form_id":      formID,
		"form_name":    formName,
		"form_version": formVersion,
	})
	if len(contribution.Models) > 0 {
		stage["ai_assistance_config"] = aiAssistanceObject(combinedAI(contribution.Models), []string{"EXTRACTION"}, cfg.HumanOversight)
	}
	if contribution.ResultPath != "" {
		upsertStageOutputs(stage, []map[string]any{
			stageOutput("extraction_table", fileURI(contribution.ResultPath), contribution.ResultFormat),
		})
	}

	literatureRecords := make([]map[string]any, 0, len(contribution.Filenames))
	for _, filename := range contribution.Filenames {
		literatureRecords = append(literatureRecords, literatureRecordFromData(filename, map[string]string{"title": filename}))
	}
	upsertLiteratureRecords(record, literatureRecords)

	studies, err := extractedStudies(cfg, contribution, formID, extractorID, completedAt)
	if err != nil {
		return err
	}
	stage["extracted_studies"] = upsertExtractedStudies(list(stage, "extracted_studies"), studies)
	stage["extraction_statistics"] = map[string]any{
		"base_stats_id":           cfg.Extraction.RunID + "_extraction_stats",
		"base_total_items":        len(contribution.Filenames),
		"total_studies_extracted": len(studies),
		"total_data_points":       totalDataPoints(studies),
	}

	upsertStage(record, stage)
	return nil
}

func extractedStudies(cfg Config, contribution ExtractionContribution, formID, extractorID, timestamp string) ([]map[string]any, error) {
	var output extractionOutput
	if err := json.Unmarshal([]byte(contribution.Results), &output); err != nil {
		return nil, err
	}

	byStudy := make(map[string]map[string]any)
	for _, response := range output.Responses {
		if response.SequenceNumber != 1 {
			continue
		}
		index, err := strconv.Atoi(response.SequenceID)
		if err != nil || index < 1 || index > len(contribution.Filenames) {
			continue
		}
		sourceID := contribution.Filenames[index-1]
		study := byStudy[sourceID]
		if study == nil {
			study = map[string]any{
				"extracted_study_id": cfg.Extraction.RunID + "_" + sourceID,
				"source_record_id":   sourceID,
				"extraction_form_id": formID,
				"extraction_status":  "COMPLETED",
				"extracted_data":     []any{},
				"ai_extraction_session": map[string]any{
					"ai_session_id":      cfg.Extraction.RunID + "_" + sourceID,
					"ai_config_id":       firstNonEmpty(slug(response.Provider+"_"+response.Model), "prismaid_ai"),
					"session_timestamp":  timestamp,
					"item_ids_processed": []any{sourceID},
				},
			}
			byStudy[sourceID] = study
		}
		dataPoints := list(study, "extracted_data")
		for _, modelResponse := range response.ModelResponses {
			points := dataPointsFromResponse(cfg.Extraction.RunID, sourceID, response, modelResponse, extractorID, timestamp)
			dataPoints = append(dataPoints, points...)
		}
		study["extracted_data"] = dataPoints
		study["final_extracted_values"] = dataPoints
	}

	studies := make([]map[string]any, 0, len(byStudy))
	for _, study := range byStudy {
		studies = append(studies, study)
	}
	return studies, nil
}

func dataPointsFromResponse(runID, sourceID string, response extractionResponse, modelResponse, extractorID, timestamp string) []any {
	parsed := make(map[string]any)
	if err := json.Unmarshal([]byte(modelResponse), &parsed); err != nil {
		return []any{
			map[string]any{
				"datapoint_id":         slug(runID + "_" + sourceID + "_" + response.Provider + "_" + response.Model + "_raw_response"),
				"extracted_value":      modelResponse,
				"extraction_timestamp": timestamp,
				"extractor_id":         extractorID,
				"ai_assisted":          true,
			},
		}
	}

	points := make([]any, 0, len(parsed))
	for key, value := range parsed {
		points = append(points, map[string]any{
			"datapoint_id":         slug(runID + "_" + sourceID + "_" + response.Provider + "_" + response.Model + "_" + key),
			"extracted_value":      valueToString(value),
			"extraction_timestamp": timestamp,
			"extractor_id":         extractorID,
			"ai_assisted":          true,
		})
	}
	return points
}

func valueToString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(value)
		}
		return string(data)
	}
}

func upsertForm(items []any, form map[string]any) []any {
	formID := stringValue(form, "form_id")
	return upsertByKey(items, form, func(existing map[string]any) bool {
		return stringValue(existing, "form_id") == formID
	})
}

func upsertExtractedStudies(items []any, studies []map[string]any) []any {
	for _, study := range studies {
		sourceID := stringValue(study, "source_record_id")
		formID := stringValue(study, "extraction_form_id")
		items = upsertByKey(items, study, func(existing map[string]any) bool {
			return stringValue(existing, "source_record_id") == sourceID &&
				stringValue(existing, "extraction_form_id") == formID
		})
	}
	return items
}

func totalDataPoints(studies []map[string]any) int {
	total := 0
	for _, study := range studies {
		total += len(list(study, "extracted_data"))
	}
	return total
}

func combinedAI(models []AIAssistance) AIAssistance {
	if len(models) == 1 {
		return models[0]
	}
	names := make([]string, 0, len(models))
	providers := make([]string, 0, len(models))
	for _, model := range models {
		if model.Model != "" {
			names = append(names, model.Model)
		}
		if model.Provider != "" {
			providers = append(providers, model.Provider)
		}
	}
	return AIAssistance{
		ID:       "prismaid_extraction_ai",
		Provider: strings.Join(providers, ", "),
		Model:    strings.Join(names, ", "),
		Version:  "multiple",
		Purpose:  []string{"EXTRACTION"},
	}
}
