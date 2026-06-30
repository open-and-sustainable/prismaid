package revaise

import (
	"fmt"
	"strings"
)

// ScreeningContribution contains the prismAId screening output to merge into a
// RevAIse ScreeningStage.
type ScreeningContribution struct {
	Review                   ReviewSeed
	Records                  []ScreeningRecord
	TotalRecords             int
	IncludedRecords          int
	ExcludedRecords          int
	Statistics               map[string]int
	InclusionCriteria        []string
	ExclusionCriteria        []string
	LanguageCriteria         []string
	ExcludeSystematicReviews bool
	PrimarySourcesOnly       bool
	ResultPath               string
	ResultFormat             string
	AIAssistance             *AIAssistance
}

// ScreeningRecord represents a screened manuscript and its decision state.
type ScreeningRecord struct {
	ID              string
	OriginalData    map[string]string
	Tags            map[string]any
	Include         bool
	ExclusionReason string
}

// AIAssistance describes AI model metadata to record in RevAIse stages.
type AIAssistance struct {
	ID          string
	Provider    string
	Model       string
	Version     string
	Purpose     []string
	Temperature string
	TPMLimit    string
	RPMLimit    string
}

// UpdateScreening merges screening output into a RevAIse Review record.
//
// The configured screening_round.round_id identifies the screening round to
// update. Reusing the same round ID updates that round, while a new round ID
// appends a new round to the configured screening stage.
func UpdateScreening(cfg Config, contribution ScreeningContribution) error {
	if !cfg.IsEnabled() {
		return nil
	}
	if cfg.Screening.RoundID == "" {
		return fmt.Errorf("revaise screening_round.round_id is required when revaise is enabled for screening")
	}
	return Update(cfg, contribution.Review, func(record Record) error {
		applyScreening(record, cfg, contribution)
		return nil
	})
}

func applyScreening(record Record, cfg Config, contribution ScreeningContribution) {
	literatureRecords := make([]map[string]any, 0, len(contribution.Records))
	for _, source := range contribution.Records {
		literatureRecords = append(literatureRecords, literatureRecordFromData(source.ID, source.OriginalData))
	}
	upsertLiteratureRecords(record, literatureRecords)

	stage := stageFromConfig(cfg, "screening_title_abstract", "Screening")
	stage["screening_criteria"] = screeningCriteria(cfg, contribution)
	stage["screening_protocol"] = map[string]any{
		"base_protocol_id":             cfg.Screening.RoundID + "_protocol",
		"minimum_reviewers_per_record": 1,
		"require_consensus":            false,
		"blind_screening":              false,
	}
	stage["overall_statistics"] = map[string]any{
		"base_stats_id":           firstNonEmpty(cfg.Screening.RoundID, "screening") + "_overall_stats",
		"base_total_items":        contribution.TotalRecords,
		"initial_records_count":   contribution.TotalRecords,
		"final_included_count":    contribution.IncludedRecords,
		"overall_inclusion_rate":  ratio(contribution.IncludedRecords, contribution.TotalRecords),
		"overall_exclusion_rate":  ratio(contribution.ExcludedRecords, contribution.TotalRecords),
		"title_abstract_excluded": contribution.ExcludedRecords,
	}
	if contribution.ResultPath != "" {
		upsertStageOutputs(stage, []map[string]any{
			stageOutput("screening_decisions", fileURI(contribution.ResultPath), contribution.ResultFormat),
		})
	}

	stage = upsertStage(record, stage)
	upsertScreeningRound(stage, cfg, contribution, literatureRecords)
}

func screeningCriteria(cfg Config, contribution ScreeningContribution) map[string]any {
	criteria := map[string]any{
		"criteria_id":        firstNonEmpty(cfg.Screening.RoundID+"_criteria", "screening_criteria"),
		"inclusion_criteria": nonEmptyStrings(contribution.InclusionCriteria, "Records passing prismAId screening filters"),
		"exclusion_criteria": nonEmptyStrings(contribution.ExclusionCriteria, "Records excluded by prismAId screening filters"),
	}
	if len(contribution.LanguageCriteria) > 0 {
		criteria["language_criteria"] = contribution.LanguageCriteria
	}
	if contribution.ExcludeSystematicReviews {
		criteria["exclude_systematic_reviews"] = true
	}
	if contribution.PrimarySourcesOnly {
		criteria["primary_sources_only"] = true
	}
	return criteria
}

func upsertScreeningRound(stage map[string]any, cfg Config, contribution ScreeningContribution, literatureRecords []map[string]any) {
	roundID := cfg.Screening.RoundID
	roundType := firstNonEmpty(cfg.Screening.RoundType, "TITLE_ABSTRACT")
	roundNumber := cfg.Screening.RoundNumber
	if roundNumber == 0 {
		roundNumber = 1
	}
	roundLabel := firstNonEmpty(cfg.Screening.RoundLabel, roundID)
	reviewerID := firstNonEmpty(cfg.Screening.ReviewerID, "prismaid")
	reviewerRole := firstNonEmpty(cfg.Screening.ReviewerRole, "SCREENER")
	timestamp := firstNonEmpty(cfg.Screening.CompletedAt, nowRFC3339())

	decisions := make([]any, 0, len(contribution.Records))
	includedIDs := make([]any, 0)
	excludedIDs := make([]any, 0)
	for _, record := range contribution.Records {
		decision := screeningDecision(record)
		if record.Include {
			includedIDs = append(includedIDs, record.ID)
		} else {
			excludedIDs = append(excludedIDs, record.ID)
		}
		decisionObject := map[string]any{
			"decision_id":          roundID + "_" + record.ID,
			"record_id":            record.ID,
			"decision_reviewer_id": reviewerID,
			"decision":             decision,
			"decision_timestamp":   timestamp,
		}
		if record.ExclusionReason != "" {
			decisionObject["exclusion_reasons"] = []any{record.ExclusionReason}
		}
		if score, ok := record.Tags["topic_relevance_score"]; ok {
			decisionObject["ai_confidence_score"] = score
		}
		decisions = append(decisions, decisionObject)
	}

	round := map[string]any{
		"round_id":           roundID,
		"round_type":         roundType,
		"round_number":       roundNumber,
		"round_label":        roundLabel,
		"round_completed_at": timestamp,
		"input_records": map[string]any{
			"collection_id":   roundID + "_input",
			"collection_name": roundLabel + " input records",
			"records":         mapsToAny(literatureRecords),
			"record_count":    len(literatureRecords),
		},
		"screening_decisions": decisions,
		"included_record_ids": includedIDs,
		"excluded_record_ids": excludedIDs,
		"reviewers": []any{
			map[string]any{"name": reviewerID, "participant_role": []any{reviewerRole}},
		},
		"round_statistics": map[string]any{
			"total_records_screened": contribution.TotalRecords,
			"records_included":       contribution.IncludedRecords,
			"records_excluded":       contribution.ExcludedRecords,
			"inclusion_rate":         ratio(contribution.IncludedRecords, contribution.TotalRecords),
			"exclusion_rate":         ratio(contribution.ExcludedRecords, contribution.TotalRecords),
		},
	}
	if contribution.AIAssistance != nil {
		round["ai_assistance"] = aiAssistanceObject(*contribution.AIAssistance, []string{"CLASSIFICATION"}, cfg.HumanOversight)
	}

	rounds := list(stage, "screening_rounds")
	rounds = upsertByKey(rounds, round, func(existing map[string]any) bool {
		return stringValue(existing, "round_id") == roundID
	})
	stage["screening_rounds"] = rounds
}

func screeningDecision(record ScreeningRecord) string {
	if record.Include {
		return "INCLUDE"
	}
	reason := strings.ToLower(record.ExclusionReason)
	if strings.Contains(reason, "duplicate") {
		return "DUPLICATE"
	}
	if strings.Contains(reason, "language") {
		return "WRONG_LANGUAGE"
	}
	return "EXCLUDE"
}

func aiAssistanceObject(ai AIAssistance, defaultPurpose []string, oversight string) map[string]any {
	purpose := ai.Purpose
	if len(purpose) == 0 {
		purpose = defaultPurpose
	}
	object := map[string]any{
		"ai_id":                 firstNonEmpty(ai.ID, slug(ai.Provider+"_"+ai.Model)),
		"ai_model":              firstNonEmpty(ai.Model, "unspecified"),
		"ai_version":            firstNonEmpty(ai.Version, "unspecified"),
		"ai_provider":           ai.Provider,
		"ai_purpose":            purpose,
		"human_oversight_level": firstNonEmpty(oversight, "NONE"),
	}
	parameters := make([]any, 0)
	if ai.Temperature != "" {
		parameters = append(parameters, map[string]any{"parameter_name": "temperature", "parameter_value": ai.Temperature})
	}
	if ai.TPMLimit != "" {
		parameters = append(parameters, map[string]any{"parameter_name": "tpm_limit", "parameter_value": ai.TPMLimit})
	}
	if ai.RPMLimit != "" {
		parameters = append(parameters, map[string]any{"parameter_name": "rpm_limit", "parameter_value": ai.RPMLimit})
	}
	if len(parameters) > 0 {
		object["ai_parameters"] = parameters
	}
	return object
}

func nonEmptyStrings(values []string, fallback string) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	if len(result) == 0 && fallback != "" {
		result = append(result, fallback)
	}
	return result
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func mapsToAny(values []map[string]any) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}
