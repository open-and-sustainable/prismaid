package revaise

// RecordSeed carries the metadata to seed a new RevAIse review record. It embeds
// ReviewSeed (the review header) and controls whether empty stubs for the manual
// stages are added.
type RecordSeed struct {
	ReviewSeed

	// IncludeManualStageStubs, when true, appends empty placeholder stages for
	// the stages prismAId does not perform — registration, search, risk of bias,
	// and synthesis — so a reviewer can document them by hand and conformance can
	// track them as present-but-incomplete rather than missing.
	IncludeManualStageStubs bool
}

// manualStageStubs are the stages prismAId does not perform; a fresh record can
// be seeded with empty placeholders for them.
var manualStageStubs = []struct {
	stageType string
	label     string
}{
	{"registration", "Registration"},
	{"search", "Search"},
	{"risk_of_bias", "Risk of bias assessment"},
	{"synthesis_narrative", "Synthesis"},
}

// NewRecord builds a seed RevAIse review record from the seed. It always produces
// a valid review header; when IncludeManualStageStubs is set it also appends
// empty stubs for the stages prismAId does not perform, each ready to fill in.
func NewRecord(seed RecordSeed) Record {
	record := Record{}
	ensureRoot(record, Config{}, seed.ReviewSeed)

	if seed.IncludeManualStageStubs {
		stages := make([]any, 0, len(manualStageStubs))
		for _, stub := range manualStageStubs {
			stages = append(stages, map[string]any{
				"stage_type":  stub.stageType,
				"stage_label": stub.label,
			})
		}
		record["stages"] = stages
	}

	return record
}
