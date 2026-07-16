package prismaid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/open-and-sustainable/prismaid/conformance"
	"github.com/open-and-sustainable/prismaid/conversion"
	"github.com/open-and-sustainable/prismaid/download/list"
	"github.com/open-and-sustainable/prismaid/download/zotero"
	"github.com/open-and-sustainable/prismaid/revaise"
	"github.com/open-and-sustainable/prismaid/review/logic"
	screening "github.com/open-and-sustainable/prismaid/screening/logic"
)

// ConvertOptions exposes conversion options for the public API.
type ConvertOptions = conversion.ConvertOptions

// PDFOptions exposes PDF-specific conversion options for the public API.
type PDFOptions = conversion.PDFOptions

// Review processes a systematic literature review based on the provided TOML configuration.
//
// The tomlConfiguration parameter should contain a valid TOML string with all the
// necessary settings for the review process, including project details, LLM configuration,
// and review criteria. See the documentation for format details.
//
// Returns an error if the review process fails for any reason, such as invalid configuration,
// inaccessible files, or API errors.
func Review(tomlConfiguration string) (ReviewResult, error) {
	result, err := logic.Review(tomlConfiguration)
	if err != nil {
		return ReviewResult{}, err
	}
	return ReviewResult{
		OutputFile:           result.OutputFile,
		ManuscriptsProcessed: result.ManuscriptsProcessed,
		ReviewItems:          result.ReviewItems,
		Models:               result.Models,
	}, nil
}

// ReviewResult summarizes the outcome of a review run: the output file written,
// how many manuscripts were processed, how many review items were extracted, and
// which models were used. The detailed extraction output is in the output file.
type ReviewResult struct {
	OutputFile           string
	ManuscriptsProcessed int
	ReviewItems          int
	Models               []string
}

// DownloadZotero downloads PDF documents from a Zotero collection using TOML configuration.
//
// The tomlConfiguration parameter should contain a [zotero] table with user,
// api_key, group, and output_dir fields. It may also contain an optional
// [revaise] block to document the download in a RevAIse review record.
//
// Returns an error if the download process fails for any reason, such as invalid
// credentials, network issues, or file system permissions.
func DownloadZotero(tomlConfiguration string) (ZoteroResult, error) {
	client := &http.Client{}
	result, err := zotero.Download(client, tomlConfiguration)
	if err != nil {
		return ZoteroResult{}, err
	}
	return ZoteroResult{Downloaded: result.Downloaded, OutputDir: result.OutputDir}, nil
}

// ZoteroResult summarizes a Zotero download run: how many attachments were
// downloaded and the directory they were saved to.
type ZoteroResult struct {
	Downloaded int
	OutputDir  string
}

// DownloadURLList downloads files from a list of URLs specified in a text file.
//
// The path parameter should point to a valid text file containing URLs, with one URL
// per line. Each URL will be downloaded to the current directory, preserving the
// filename from the URL.
//
// Returns an error if the function fails to open or read the input file,
// but continues processing even if individual URLs fail to download.
func DownloadURLList(path string) (URLListResult, error) {
	result, err := list.DownloadURLList(path)
	if err != nil {
		return URLListResult{}, err
	}
	return URLListResult{Total: result.Total, Downloaded: result.Downloaded, Failed: result.Failed}, nil
}

// URLListResult summarizes a URL-list download run: how many URLs were listed,
// how many were downloaded, and how many failed. Detailed per-URL outcomes are
// written to the "_download" report next to the input file.
type URLListResult struct {
	Total      int
	Downloaded int
	Failed     int
}

// Convert processes files in the specified directory and converts them to plain text format.
//
// Parameters:
//   - inputDir: Path to the directory containing files to be converted
//   - selectedFormats: Comma-separated list of formats to process (e.g., "pdf,docx,html")
//   - options: Conversion options including optional Apache Tika server address for OCR fallback
//     (e.g., ConvertOptions{TikaServer: "localhost:9998"}).
//
// The function will scan the input directory for files with extensions matching the
// selected formats and convert each to a corresponding .txt file with the same base name.
// Currently supported formats include "pdf", "docx", and "html" (which also processes .htm files).
//
// When options.TikaServer is provided and standard conversion methods fail, files are
// automatically sent to the Tika server for OCR-based text extraction as a fallback.
//
// Returns an error if the conversion process fails for any reason, such as inaccessible
// files, unsupported formats, or file system permission issues.
func Convert(inputDir, selectedFormats string, options conversion.ConvertOptions) (ConvertResult, error) {
	result, err := conversion.Convert(inputDir, selectedFormats, options)
	if err != nil {
		return ConvertResult{}, err
	}
	return *result, nil
}

// ConvertResult exposes the conversion result summary for the public API.
type ConvertResult = conversion.ConvertResult

// ValidateConfig validates a prismAId configuration without executing it.
//
// configType selects which configuration schema to validate against and must be
// one of "review", "screening", or "zotero" (case-insensitive). The
// tomlConfiguration parameter is the TOML configuration string.
//
// Validation is read-only: it parses the configuration and checks required
// fields and value constraints, including any optional [revaise] block. It
// performs no network access, file reads, or API-key resolution, so it is safe
// to call on draft configurations.
//
// It returns nil if the configuration is valid, or an error describing the
// problem found. An empty or unrecognized configType is itself reported as an
// error.
func ValidateConfig(configType, tomlConfiguration string) error {
	switch strings.ToLower(strings.TrimSpace(configType)) {
	case "review":
		return logic.ValidateConfig(tomlConfiguration)
	case "screening":
		return screening.ValidateConfig(tomlConfiguration)
	case "zotero":
		return zotero.ValidateConfig(tomlConfiguration)
	default:
		return fmt.Errorf("unknown config type %q: must be \"review\", \"screening\", or \"zotero\"", configType)
	}
}

// ConformanceReport summarizes a protocol conformance check: which protocol was
// applied, whether the record conforms, and the unmet constraints (each carrying
// the protocol's own message).
type ConformanceReport = conformance.Report

// CheckConformance validates a RevAIse review-record JSON string against a
// reporting protocol's SHACL shapes (for example "prisma-2020"). The verdict and
// the per-constraint messages come from the protocol's shapes, so conformance is
// decided symbolically rather than asserted by prismAId.
//
// The protocol is selected by name; ConformanceProtocols lists the accepted
// values, and an unknown protocol is reported as an error. The shapes are pulled
// from the latest version RevAIse publishes on GitHub Pages, so the check
// requires network access; nothing is vendored with prismAId.
func CheckConformance(recordJSON, protocol string) (ConformanceReport, error) {
	report, err := conformance.Check(recordJSON, protocol)
	if err != nil {
		return ConformanceReport{}, err
	}
	return *report, nil
}

// ConformanceProtocols returns the protocol identifiers accepted by
// CheckConformance. It reads the catalogue RevAIse publishes on GitHub Pages, so
// it requires network access and returns an error if the catalogue is
// unreachable.
func ConformanceProtocols() ([]string, error) {
	return conformance.AvailableProtocols()
}

// RevAIseRecordParams are the inputs for GenerateRevAIseRecord: the review
// header, and whether to include empty stubs for the stages prismAId does not
// perform. Type, Status, and Authors default to sensible values when omitted.
type RevAIseRecordParams struct {
	ID       string   `json:"id,omitempty"`
	Title    string   `json:"title,omitempty"`
	Type     string   `json:"type,omitempty"`
	Status   string   `json:"status,omitempty"`
	Version  string   `json:"version,omitempty"`
	Language string   `json:"language,omitempty"`
	Country  string   `json:"country,omitempty"`
	Authors  []string `json:"authors,omitempty"`

	// IncludeManualStageStubs adds empty placeholder stages for the stages
	// prismAId does not perform (registration, search, risk of bias, synthesis),
	// so they can be documented by hand and tracked by CheckConformance.
	IncludeManualStageStubs bool `json:"include_manual_stage_stubs,omitempty"`
}

// GenerateRevAIseRecord builds a seed RevAIse review record and returns it as an
// indented JSON string. It always produces a valid review header from the
// provided parameters; when IncludeManualStageStubs is set it also adds empty
// stubs for the stages prismAId does not perform (registration, search, risk of
// bias, synthesis), ready to fill in and track with CheckConformance.
func GenerateRevAIseRecord(params RevAIseRecordParams) (string, error) {
	record := revaise.NewRecord(revaise.RecordSeed{
		ReviewSeed: revaise.ReviewSeed{
			ID:       params.ID,
			Title:    params.Title,
			Type:     params.Type,
			Status:   params.Status,
			Version:  params.Version,
			Language: params.Language,
			Country:  params.Country,
			Authors:  params.Authors,
		},
		IncludeManualStageStubs: params.IncludeManualStageStubs,
	})
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RevAIseSchemaParams selects what GenerateRevAIseSchema returns: a description
// of a type, the list of types (when Type is empty), or a raw released artifact.
type RevAIseSchemaParams struct {
	// Type is a class or enum name to describe. Empty lists the available
	// classes and enums.
	Type string `json:"type,omitempty"`
	// Raw returns the full released JSON Schema instead of a description.
	Raw bool `json:"raw,omitempty"`
	// Context returns the released JSON-LD context instead of a description.
	Context bool `json:"context,omitempty"`
}

// RevAIseSchemaDescription describes the RevAIse data model or one of its types.
type RevAIseSchemaDescription = revaise.SchemaDescription

// RevAIseSchemaResult carries either a structured description or, when Raw or
// Context is requested, the raw released artifact as a JSON string.
type RevAIseSchemaResult struct {
	Description *revaise.SchemaDescription `json:"description,omitempty"`
	Raw         string                     `json:"raw,omitempty"`
}

// RevAIseSchema serves the RevAIse data model from the released, verified
// artifacts RevAIse publishes (JSON Schema and JSON-LD context), fetched live —
// nothing is vendored, and the LinkML source is never used. By default it
// describes a type (or lists the available classes and enums when Type is
// empty); Raw returns the full JSON Schema and Context returns the JSON-LD
// context. It requires network access.
func RevAIseSchema(params RevAIseSchemaParams) (RevAIseSchemaResult, error) {
	switch {
	case params.Raw:
		raw, err := revaise.FetchSchema()
		if err != nil {
			return RevAIseSchemaResult{}, err
		}
		return RevAIseSchemaResult{Raw: raw}, nil
	case params.Context:
		ctx, err := revaise.FetchContext()
		if err != nil {
			return RevAIseSchemaResult{}, err
		}
		return RevAIseSchemaResult{Raw: ctx}, nil
	default:
		desc, err := revaise.DescribeSchema(params.Type)
		if err != nil {
			return RevAIseSchemaResult{}, err
		}
		return RevAIseSchemaResult{Description: desc}, nil
	}
}

// MergeRecordStage merges a stage into an existing RevAIse review record and
// returns the updated record as JSON. The stage (a JSON object with at least a
// stage_type) fills a matching stub — matched by stage_type and stage_label — or
// is appended when none matches. This is the "append each stage to the record"
// step of the review lifecycle.
func MergeRecordStage(recordJSON, stageJSON string) (string, error) {
	return revaise.MergeStage(recordJSON, stageJSON)
}

// RecordValidation is the outcome of validating a RevAIse record against the
// data-model JSON Schema.
type RecordValidation = revaise.RecordValidation

// ValidateRecord validates a RevAIse review-record JSON string against the
// released RevAIse data-model JSON Schema, fetched live. This checks structural
// validity (field names, types, required slots) — distinct from CheckConformance,
// which checks a reporting protocol. It requires network access.
func ValidateRecord(recordJSON string) (RecordValidation, error) {
	result, err := revaise.ValidateRecord(recordJSON)
	if err != nil {
		return RecordValidation{}, err
	}
	return *result, nil
}

// ConformanceGuidance is the full set of requirements a protocol imposes,
// together with its metadata.
type ConformanceGuidance = conformance.Guidance

// ProtocolGuidance returns the requirement checklist a protocol imposes,
// extracted from the SHACL shapes RevAIse publishes, together with the protocol's
// metadata. It is advisory — it helps plan a conforming review before any record
// exists — and does not constrain the order in which prismAId's tools are used.
// It requires network access.
func ProtocolGuidance(protocol string) (ConformanceGuidance, error) {
	guidance, err := conformance.ProtocolGuidance(protocol)
	if err != nil {
		return ConformanceGuidance{}, err
	}
	return *guidance, nil
}

// Screening processes a list of manuscripts to identify items for exclusion based on various criteria.
//
// The tomlConfiguration parameter should contain a valid TOML string with all the necessary
// settings for the screening process, including input/output files, filter configurations,
// and optional LLM settings for AI-assisted screening.
//
// The screening tool can apply multiple filters:
//   - Deduplication: Identifies duplicate manuscripts using exact, fuzzy, or semantic matching
//   - Language detection: Filters manuscripts based on detected language
//   - Article type classification: Identifies and filters based on article types (reviews, editorials, etc.)
//   - Topic relevance: Scores manuscripts based on relevance to specified topics using keyword, concept, and field matching
//
// It returns a ScreeningResult summarizing the run, or an error if the screening
// process fails for any reason, such as invalid configuration, inaccessible
// files, or processing errors.
func Screening(tomlConfiguration string) (ScreeningResult, error) {
	result, err := screening.Screen(tomlConfiguration)
	if err != nil {
		return ScreeningResult{}, err
	}
	return ScreeningResult{
		TotalRecords:    result.TotalRecords,
		IncludedRecords: result.IncludedRecords,
		ExcludedRecords: result.ExcludedRecords,
		Statistics:      result.Statistics,
	}, nil
}

// ScreeningResult summarizes the outcome of a screening run: the total number of
// records processed, how many were included and excluded, and per-filter
// statistics. The detailed per-record output is written to the configured
// output file, not returned here.
type ScreeningResult struct {
	TotalRecords    int
	IncludedRecords int
	ExcludedRecords int
	Statistics      map[string]int
}
