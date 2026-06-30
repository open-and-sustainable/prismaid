package revaise

// Config controls optional RevAIse review-record updates.
//
// RevAIse support is disabled unless Enabled is true. When enabled, RecordFile
// points to the Review document to create or update. Backups are enabled by
// default and must be explicitly disabled with backup = false.
type Config struct {
	Enabled       bool   `toml:"enabled" json:"enabled,omitempty" yaml:"enabled,omitempty"`
	RecordFile    string `toml:"record_file" json:"record_file,omitempty" yaml:"record_file,omitempty"`
	Format        string `toml:"format" json:"format,omitempty" yaml:"format,omitempty"`
	SchemaVersion string `toml:"schema_version" json:"schema_version,omitempty" yaml:"schema_version,omitempty"`
	// HumanOversight records how much a human reviewed the AI output, written to
	// every AIAssistance entry. It must be a RevAIse HumanOversightLevel value:
	// FULL_REVIEW, SAMPLE_REVIEW, CONFIDENCE_BASED, EXCEPTION_ONLY, MINIMAL, or
	// NONE. Defaults to NONE, since prismAId itself performs no human review.
	HumanOversight string               `toml:"human_oversight_level" json:"human_oversight_level,omitempty" yaml:"human_oversight_level,omitempty"`
	Backup         *bool                `toml:"backup" json:"backup,omitempty" yaml:"backup,omitempty"`
	BackupDir      string               `toml:"backup_dir" json:"backup_dir,omitempty" yaml:"backup_dir,omitempty"`
	Review         ReviewConfig         `toml:"review" json:"review,omitempty" yaml:"review,omitempty"`
	Stage          StageConfig          `toml:"stage" json:"stage,omitempty" yaml:"stage,omitempty"`
	Screening      ScreeningRoundConfig `toml:"screening_round" json:"screening_round,omitempty" yaml:"screening_round,omitempty"`
	Extraction     ExtractionRunConfig  `toml:"extraction_run" json:"extraction_run,omitempty" yaml:"extraction_run,omitempty"`
}

// ReviewConfig provides root Review metadata when a new RevAIse document is
// created. Existing root fields are preserved when omitted.
type ReviewConfig struct {
	ID       string   `toml:"id" json:"id,omitempty" yaml:"id,omitempty"`
	Title    string   `toml:"title" json:"title,omitempty" yaml:"title,omitempty"`
	Type     string   `toml:"type" json:"type,omitempty" yaml:"type,omitempty"`
	Status   string   `toml:"status" json:"status,omitempty" yaml:"status,omitempty"`
	Version  string   `toml:"version" json:"version,omitempty" yaml:"version,omitempty"`
	Language string   `toml:"language" json:"language,omitempty" yaml:"language,omitempty"`
	Country  string   `toml:"country" json:"country,omitempty" yaml:"country,omitempty"`
	Authors  []string `toml:"authors" json:"authors,omitempty" yaml:"authors,omitempty"`
}

// StageConfig identifies the stage execution to update.
type StageConfig struct {
	Type        string `toml:"stage_type" json:"stage_type,omitempty" yaml:"stage_type,omitempty"`
	Label       string `toml:"stage_label" json:"stage_label,omitempty" yaml:"stage_label,omitempty"`
	Description string `toml:"stage_description" json:"stage_description,omitempty" yaml:"stage_description,omitempty"`
	StartedAt   string `toml:"started_at" json:"started_at,omitempty" yaml:"started_at,omitempty"`
	EndedAt     string `toml:"ended_at" json:"ended_at,omitempty" yaml:"ended_at,omitempty"`
}

// ScreeningRoundConfig identifies a screening round or substage.
type ScreeningRoundConfig struct {
	RoundID     string `toml:"round_id" json:"round_id,omitempty" yaml:"round_id,omitempty"`
	RoundType   string `toml:"round_type" json:"round_type,omitempty" yaml:"round_type,omitempty"`
	RoundNumber int    `toml:"round_number" json:"round_number,omitempty" yaml:"round_number,omitempty"`
	RoundLabel  string `toml:"round_label" json:"round_label,omitempty" yaml:"round_label,omitempty"`
	ReviewerID  string `toml:"reviewer_id" json:"reviewer_id,omitempty" yaml:"reviewer_id,omitempty"`
	// ReviewerRole is the RevAIse ParticipantRole recorded for the screening
	// reviewer (for example SCREENER or DATA_EXTRACTOR). Defaults to SCREENER.
	ReviewerRole string `toml:"reviewer_role" json:"reviewer_role,omitempty" yaml:"reviewer_role,omitempty"`
	CompletedAt  string `toml:"completed_at" json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
}

// ExtractionRunConfig identifies a data-extraction run, such as a pilot or a
// full extraction pass.
type ExtractionRunConfig struct {
	RunID       string `toml:"run_id" json:"run_id,omitempty" yaml:"run_id,omitempty"`
	Label       string `toml:"label" json:"label,omitempty" yaml:"label,omitempty"`
	FormID      string `toml:"form_id" json:"form_id,omitempty" yaml:"form_id,omitempty"`
	FormName    string `toml:"form_name" json:"form_name,omitempty" yaml:"form_name,omitempty"`
	FormVersion string `toml:"form_version" json:"form_version,omitempty" yaml:"form_version,omitempty"`
	ExtractorID string `toml:"extractor_id" json:"extractor_id,omitempty" yaml:"extractor_id,omitempty"`
	CompletedAt string `toml:"completed_at" json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
}

// IsEnabled reports whether RevAIse updates should run for a workflow.
func (c Config) IsEnabled() bool {
	return c.Enabled
}

func (c Config) backupEnabled() bool {
	return c.Backup == nil || *c.Backup
}
