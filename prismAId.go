package prismaid

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/open-and-sustainable/prismaid/conformance"
	"github.com/open-and-sustainable/prismaid/conversion"
	"github.com/open-and-sustainable/prismaid/download/list"
	"github.com/open-and-sustainable/prismaid/download/zotero"
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
// values, and an unknown protocol is reported as an error. The check is
// read-only and offline, using shapes and context vendored with prismAId.
func CheckConformance(recordJSON, protocol string) (ConformanceReport, error) {
	report, err := conformance.Check(recordJSON, protocol)
	if err != nil {
		return ConformanceReport{}, err
	}
	return *report, nil
}

// ConformanceProtocols returns the protocol identifiers accepted by
// CheckConformance.
func ConformanceProtocols() []string {
	return conformance.AvailableProtocols()
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
