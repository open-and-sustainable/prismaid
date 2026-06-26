package prismaid

import (
	"fmt"
	"net/http"
	"strings"

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
func Review(tomlConfiguration string) error {
	return logic.Review(tomlConfiguration)
}

// DownloadZotero downloads PDF documents from a Zotero collection using TOML configuration.
//
// The tomlConfiguration parameter should contain a [zotero] table with user,
// api_key, group, and output_dir fields. It may also contain an optional
// [revaise] block to document the download in a RevAIse review record.
//
// Returns an error if the download process fails for any reason, such as invalid
// credentials, network issues, or file system permissions.
func DownloadZotero(tomlConfiguration string) error {
	client := &http.Client{}
	return zotero.Download(client, tomlConfiguration)
}

// DownloadURLList downloads files from a list of URLs specified in a text file.
//
// The path parameter should point to a valid text file containing URLs, with one URL
// per line. Each URL will be downloaded to the current directory, preserving the
// filename from the URL.
//
// Returns an error if the function fails to open or read the input file,
// but continues processing even if individual URLs fail to download.
func DownloadURLList(path string) error {
	return list.DownloadURLList(path)
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
func Convert(inputDir, selectedFormats string, options conversion.ConvertOptions) error {
	return conversion.Convert(inputDir, selectedFormats, options)
}

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
// Returns an error if the screening process fails for any reason, such as invalid configuration,
// inaccessible files, or processing errors.
func Screening(tomlConfiguration string) error {
	return screening.Screen(tomlConfiguration)
}
