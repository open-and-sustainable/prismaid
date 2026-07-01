package prismaid

import (
	"fmt"
	"strings"
)

// ScreeningDeduplication configures the deduplication filter.
type ScreeningDeduplication struct {
	Enabled       bool
	UseAI         bool
	CompareFields []string
}

// ScreeningLanguage configures the language filter.
type ScreeningLanguage struct {
	Enabled           bool
	AcceptedLanguages []string
	UseAI             bool
}

// ScreeningArticleType configures the article-type filter.
type ScreeningArticleType struct {
	Enabled            bool
	UseAI              bool
	ExcludeReviews     bool
	ExcludeEditorials  bool
	ExcludeLetters     bool
	ExcludeTheoretical bool
	ExcludeEmpirical   bool
	ExcludeMethods     bool
	ExcludeSingleCase  bool
	ExcludeSample      bool
	IncludeTypes       []string
}

// ScreeningTopicRelevance configures the topic-relevance filter and its scoring
// weights.
type ScreeningTopicRelevance struct {
	Enabled        bool
	UseAI          bool
	Topics         []string
	MinScore       float64
	KeywordWeight  float64
	ConceptWeight  float64
	FieldRelevance float64
}

// ScreeningLLM configures the single optional AI model for AI-assisted
// screening. A nil *ScreeningLLM omits the [filters.llm] table.
type ScreeningLLM struct {
	Provider    string
	APIKey      string
	Model       string
	Temperature float64
	TpmLimit    int64
	RpmLimit    int64
}

// ScreeningRevAIse describes an optional [revaise] block documenting the
// screening as a RevAIse screening round. A nil *ScreeningRevAIse omits it.
type ScreeningRevAIse struct {
	RecordFile     string
	Format         string
	SchemaVersion  string
	HumanOversight string
	StageLabel     string
	RoundID        string
	RoundType      string
	RoundNumber    int
	RoundLabel     string
	ReviewerID     string
	ReviewerRole   string
}

// ScreeningConfigParams holds the inputs for GenerateScreeningConfig.
type ScreeningConfigParams struct {
	Name             string
	Author           string
	Version          string
	InputFile        string
	OutputFile       string
	TextColumn       string
	IdentifierColumn string
	OutputFormat     string
	LogLevel         string

	Deduplication  ScreeningDeduplication
	Language       ScreeningLanguage
	ArticleType    ScreeningArticleType
	TopicRelevance ScreeningTopicRelevance
	LLM            *ScreeningLLM

	RevAIse *ScreeningRevAIse
}

// GenerateScreeningConfig builds a screening-tool TOML configuration from
// params. Like GenerateReviewConfig it produces well-formed, deterministic TOML
// and leaves completeness to ValidateConfig("screening", ...). It emits the
// single supported [filters.llm] table only when an LLM is provided.
func GenerateScreeningConfig(p ScreeningConfigParams) string {
	var b strings.Builder

	b.WriteString("[project]\n")
	fmt.Fprintf(&b, "name = %q\n", p.Name)
	fmt.Fprintf(&b, "author = %q\n", p.Author)
	fmt.Fprintf(&b, "version = %q\n", p.Version)
	fmt.Fprintf(&b, "input_file = %q\n", p.InputFile)
	fmt.Fprintf(&b, "output_file = %q\n", p.OutputFile)
	fmt.Fprintf(&b, "text_column = %q\n", p.TextColumn)
	fmt.Fprintf(&b, "identifier_column = %q\n", p.IdentifierColumn)
	fmt.Fprintf(&b, "output_format = %q\n", p.OutputFormat)
	fmt.Fprintf(&b, "log_level = %q\n", p.LogLevel)

	b.WriteString("\n[filters.deduplication]\n")
	fmt.Fprintf(&b, "enabled = %t\n", p.Deduplication.Enabled)
	fmt.Fprintf(&b, "use_ai = %t\n", p.Deduplication.UseAI)
	fmt.Fprintf(&b, "compare_fields = %s\n", tomlStringArray(p.Deduplication.CompareFields))

	b.WriteString("\n[filters.language]\n")
	fmt.Fprintf(&b, "enabled = %t\n", p.Language.Enabled)
	fmt.Fprintf(&b, "accepted_languages = %s\n", tomlStringArray(p.Language.AcceptedLanguages))
	fmt.Fprintf(&b, "use_ai = %t\n", p.Language.UseAI)

	at := p.ArticleType
	b.WriteString("\n[filters.article_type]\n")
	fmt.Fprintf(&b, "enabled = %t\n", at.Enabled)
	fmt.Fprintf(&b, "use_ai = %t\n", at.UseAI)
	fmt.Fprintf(&b, "exclude_reviews = %t\n", at.ExcludeReviews)
	fmt.Fprintf(&b, "exclude_editorials = %t\n", at.ExcludeEditorials)
	fmt.Fprintf(&b, "exclude_letters = %t\n", at.ExcludeLetters)
	fmt.Fprintf(&b, "exclude_theoretical = %t\n", at.ExcludeTheoretical)
	fmt.Fprintf(&b, "exclude_empirical = %t\n", at.ExcludeEmpirical)
	fmt.Fprintf(&b, "exclude_methods = %t\n", at.ExcludeMethods)
	fmt.Fprintf(&b, "exclude_single_case = %t\n", at.ExcludeSingleCase)
	fmt.Fprintf(&b, "exclude_sample = %t\n", at.ExcludeSample)
	fmt.Fprintf(&b, "include_types = %s\n", tomlStringArray(at.IncludeTypes))

	tr := p.TopicRelevance
	b.WriteString("\n[filters.topic_relevance]\n")
	fmt.Fprintf(&b, "enabled = %t\n", tr.Enabled)
	fmt.Fprintf(&b, "use_ai = %t\n", tr.UseAI)
	fmt.Fprintf(&b, "topics = %s\n", tomlStringArray(tr.Topics))
	fmt.Fprintf(&b, "min_score = %s\n", formatFloat(tr.MinScore))
	b.WriteString("\n[filters.topic_relevance.score_weights]\n")
	fmt.Fprintf(&b, "keyword_match = %s\n", formatFloat(tr.KeywordWeight))
	fmt.Fprintf(&b, "concept_match = %s\n", formatFloat(tr.ConceptWeight))
	fmt.Fprintf(&b, "field_relevance = %s\n", formatFloat(tr.FieldRelevance))

	if p.LLM != nil {
		b.WriteString("\n[filters.llm]\n")
		fmt.Fprintf(&b, "provider = %q\n", p.LLM.Provider)
		fmt.Fprintf(&b, "api_key = %q\n", p.LLM.APIKey)
		fmt.Fprintf(&b, "model = %q\n", p.LLM.Model)
		fmt.Fprintf(&b, "temperature = %s\n", formatFloat(p.LLM.Temperature))
		fmt.Fprintf(&b, "tpm_limit = %d\n", p.LLM.TpmLimit)
		fmt.Fprintf(&b, "rpm_limit = %d\n", p.LLM.RpmLimit)
	}

	result := strings.TrimSpace(b.String())
	if p.RevAIse != nil {
		result += "\n\n" + screeningRevAIseSection(*p.RevAIse)
	}
	return result
}

func screeningRevAIseSection(r ScreeningRevAIse) string {
	var b strings.Builder
	b.WriteString("[revaise]\n")
	b.WriteString("enabled = true\n")
	fmt.Fprintf(&b, "record_file = %q\n", r.RecordFile)
	fmt.Fprintf(&b, "format = %q\n", r.Format)
	fmt.Fprintf(&b, "schema_version = %q\n", r.SchemaVersion)
	fmt.Fprintf(&b, "human_oversight_level = %q\n", r.HumanOversight)
	b.WriteString("\n[revaise.stage]\n")
	b.WriteString("stage_type = \"screening_title_abstract\"\n")
	fmt.Fprintf(&b, "stage_label = %q\n", r.StageLabel)
	b.WriteString("\n[revaise.screening_round]\n")
	fmt.Fprintf(&b, "round_id = %q\n", r.RoundID)
	fmt.Fprintf(&b, "round_type = %q\n", r.RoundType)
	fmt.Fprintf(&b, "round_number = %d\n", r.RoundNumber)
	fmt.Fprintf(&b, "round_label = %q\n", r.RoundLabel)
	fmt.Fprintf(&b, "reviewer_id = %q\n", r.ReviewerID)
	fmt.Fprintf(&b, "reviewer_role = %q\n", r.ReviewerRole)
	return strings.TrimSpace(b.String())
}

// tomlStringArray renders a slice of strings as a TOML inline array.
func tomlStringArray(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
