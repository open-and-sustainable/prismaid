package revaise

// OutputContribution updates a RevAIse stage with produced artifacts.
type OutputContribution struct {
	Review       ReviewSeed
	StageType    string
	StageLabel   string
	StageOutputs []StageOutput
}

// StageOutput describes one artifact produced by a workflow stage.
type StageOutput struct {
	Kind        string
	ResourceURI string
	Format      string
}

// UpdateStageOutputs merges stage output artifact references into a RevAIse
// Review record.
func UpdateStageOutputs(cfg Config, contribution OutputContribution) error {
	if !cfg.IsEnabled() {
		return nil
	}
	return Update(cfg, contribution.Review, func(record Record) error {
		stageCfg := cfg
		if stageCfg.Stage.Type == "" {
			stageCfg.Stage.Type = contribution.StageType
		}
		if stageCfg.Stage.Label == "" {
			stageCfg.Stage.Label = contribution.StageLabel
		}
		stage := upsertStage(record, stageFromConfig(stageCfg, contribution.StageType, contribution.StageLabel))
		outputs := make([]map[string]any, 0, len(contribution.StageOutputs))
		for _, output := range contribution.StageOutputs {
			outputs = append(outputs, stageOutput(output.Kind, fileURI(output.ResourceURI), output.Format))
		}
		upsertStageOutputs(stage, outputs)
		return nil
	})
}
