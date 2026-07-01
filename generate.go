package prismaid

import (
	"fmt"
	"strconv"
	"strings"
)

// ReviewLLM describes one language-model entry for a review configuration.
type ReviewLLM struct {
	Provider     string
	APIKey       string
	Model        string
	Temperature  float64
	TpmLimit     int64
	RpmLimit     int64
	BaseURL      string
	EndpointType string
	Region       string
	ProjectID    string
	Location     string
	APIVersion   string
}

// ReviewItem describes one review item: a key and its allowed values.
type ReviewItem struct {
	Key    string
	Values []string
}

// ReviewRevAIse describes an optional [revaise] block documenting the review as
// a RevAIse data_extraction stage. Pass a nil *ReviewRevAIse to omit the block.
type ReviewRevAIse struct {
	RecordFile     string
	Format         string
	SchemaVersion  string
	HumanOversight string
	StageLabel     string
	RunID          string
	RunLabel       string
	FormID         string
	FormName       string
	FormVersion    string
	ExtractorID    string
}

// ReviewConfigParams holds the inputs for GenerateReviewConfig.
type ReviewConfigParams struct {
	Name    string
	Author  string
	Version string

	InputDirectory   string
	ResultsFileName  string
	OutputFormat     string
	LogLevel         string
	Duplication      string
	CotJustification string
	Summary          string

	LLMs []ReviewLLM

	Persona        string
	Task           string
	ExpectedResult string
	Failsafe       string
	Definitions    string
	Example        string

	ReviewItems []ReviewItem

	RevAIse *ReviewRevAIse
}

// GenerateReviewConfig builds a review-tool TOML configuration from params.
//
// It produces well-formed, deterministic TOML: string values are escaped, and
// numeric fields (temperature, token and request limits) are written as numbers
// rather than quoted strings. It does not check completeness; callers validate
// the result with ValidateConfig("review", ...). Generation and validation are
// kept as separate, composable steps so a partial configuration can still be
// produced as a starting point.
func GenerateReviewConfig(p ReviewConfigParams) string {
	var b strings.Builder

	b.WriteString("[project]\n")
	fmt.Fprintf(&b, "name = %q\n", p.Name)
	fmt.Fprintf(&b, "author = %q\n", p.Author)
	fmt.Fprintf(&b, "version = %q\n", p.Version)

	b.WriteString("\n[project.configuration]\n")
	fmt.Fprintf(&b, "input_directory = %q\n", p.InputDirectory)
	fmt.Fprintf(&b, "results_file_name = %q\n", p.ResultsFileName)
	fmt.Fprintf(&b, "output_format = %q\n", p.OutputFormat)
	fmt.Fprintf(&b, "log_level = %q\n", p.LogLevel)
	fmt.Fprintf(&b, "duplication = %q\n", p.Duplication)
	fmt.Fprintf(&b, "cot_justification = %q\n", p.CotJustification)
	fmt.Fprintf(&b, "summary = %q\n", p.Summary)

	b.WriteString("\n[project.llm]\n")
	b.WriteString(reviewLLMSection(p.LLMs))

	b.WriteString("[prompt]\n")
	fmt.Fprintf(&b, "persona = %q\n", p.Persona)
	fmt.Fprintf(&b, "task = %q\n", p.Task)
	fmt.Fprintf(&b, "expected_result = %q\n", p.ExpectedResult)
	fmt.Fprintf(&b, "failsafe = %q\n", p.Failsafe)
	fmt.Fprintf(&b, "definitions = %q\n", p.Definitions)
	fmt.Fprintf(&b, "example = %q\n", p.Example)

	b.WriteString("\n[review]\n")
	b.WriteString(reviewItemSection(p.ReviewItems))

	result := strings.TrimSpace(b.String())
	if p.RevAIse != nil {
		result += "\n\n" + reviewRevAIseSection(*p.RevAIse)
	}
	return result
}

func reviewLLMSection(llms []ReviewLLM) string {
	var b strings.Builder
	for i, m := range llms {
		fmt.Fprintf(&b, "[project.llm.%d]\n", i+1)
		fmt.Fprintf(&b, "provider = %q\n", m.Provider)
		fmt.Fprintf(&b, "api_key = %q\n", m.APIKey)
		fmt.Fprintf(&b, "model = %q\n", m.Model)
		fmt.Fprintf(&b, "temperature = %s\n", formatFloat(m.Temperature))
		fmt.Fprintf(&b, "tpm_limit = %d\n", m.TpmLimit)
		fmt.Fprintf(&b, "rpm_limit = %d\n", m.RpmLimit)
		if m.BaseURL != "" {
			fmt.Fprintf(&b, "base_url = %q\n", m.BaseURL)
		}
		if m.EndpointType != "" {
			fmt.Fprintf(&b, "endpoint_type = %q\n", m.EndpointType)
		}
		if m.Region != "" {
			fmt.Fprintf(&b, "region = %q\n", m.Region)
		}
		if m.ProjectID != "" {
			fmt.Fprintf(&b, "project_id = %q\n", m.ProjectID)
		}
		if m.Location != "" {
			fmt.Fprintf(&b, "location = %q\n", m.Location)
		}
		if m.APIVersion != "" {
			fmt.Fprintf(&b, "api_version = %q\n", m.APIVersion)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func reviewItemSection(items []ReviewItem) string {
	var b strings.Builder
	for i, item := range items {
		fmt.Fprintf(&b, "[review.%d]\n", i+1)
		fmt.Fprintf(&b, "key = %q\n", item.Key)
		b.WriteString("values = [")
		for j, value := range item.Values {
			if j > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%q", strings.TrimSpace(value))
		}
		b.WriteString("]\n")
	}
	return b.String()
}

func reviewRevAIseSection(r ReviewRevAIse) string {
	var b strings.Builder
	b.WriteString("[revaise]\n")
	b.WriteString("enabled = true\n")
	fmt.Fprintf(&b, "record_file = %q\n", r.RecordFile)
	fmt.Fprintf(&b, "format = %q\n", r.Format)
	fmt.Fprintf(&b, "schema_version = %q\n", r.SchemaVersion)
	fmt.Fprintf(&b, "human_oversight_level = %q\n", r.HumanOversight)
	b.WriteString("\n[revaise.stage]\n")
	b.WriteString("stage_type = \"data_extraction\"\n")
	fmt.Fprintf(&b, "stage_label = %q\n", r.StageLabel)
	b.WriteString("\n[revaise.extraction_run]\n")
	fmt.Fprintf(&b, "run_id = %q\n", r.RunID)
	fmt.Fprintf(&b, "label = %q\n", r.RunLabel)
	fmt.Fprintf(&b, "form_id = %q\n", r.FormID)
	fmt.Fprintf(&b, "form_name = %q\n", r.FormName)
	fmt.Fprintf(&b, "form_version = %q\n", r.FormVersion)
	fmt.Fprintf(&b, "extractor_id = %q\n", r.ExtractorID)
	return strings.TrimSpace(b.String())
}

// formatFloat renders a float as a TOML float literal, ensuring a decimal point
// so the value is decoded as a float rather than an integer.
func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}
