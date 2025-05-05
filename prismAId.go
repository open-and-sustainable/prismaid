package prismaid

import (
	"net/http"

	"github.com/open-and-sustainable/prismaid/conversion"
	"github.com/open-and-sustainable/prismaid/download/list"
	"github.com/open-and-sustainable/prismaid/download/zotero"
	"github.com/open-and-sustainable/prismaid/review/logic"
)

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

// DownloadZoteroPDFs downloads PDF documents from a specified Zotero collection.
//
// Parameters:
//   - username: The Zotero username or user ID
//   - apiKey: The Zotero API key for authentication
//   - collectionName: The name of the collection to download PDFs from
//   - parentDir: The directory path where downloaded PDFs will be saved
//
// Returns an error if the download process fails for any reason, such as invalid
// credentials, network issues, or file system permissions.
func DownloadZoteroPDFs(username, apiKey, collectionName, parentDir string) error {
	client := &http.Client{}
	return zotero.DownloadZoteroPDFs(client, username, apiKey, collectionName, parentDir)
}

// DownloadURLList downloads files from a list of URLs specified in a text file.
//
// The path parameter should point to a valid text file containing URLs, with one URL
// per line. Each URL will be downloaded to the current directory, preserving the
// filename from the URL.
//
// This function does not return any value. Download failures for individual URLs
// are logged but do not stop the overall process.
func DownloadURLList(path string) {
	list.DownloadURLList(path)
	return
}

// Convert processes files in the specified directory and converts them to plain text format.
//
// Parameters:
//   - inputDir: Path to the directory containing files to be converted
//   - selectedFormats: Comma-separated list of formats to process (e.g., "pdf,docx,html")
//
// The function will scan the input directory for files with extensions matching the
// selected formats and convert each to a corresponding .txt file with the same base name.
// Currently supported formats include "pdf", "docx", and "html" (which also processes .htm files).
//
// Returns an error if the conversion process fails for any reason, such as inaccessible
// files, unsupported formats, or file system permission issues.
func Convert(inputDir, selectedFormats string) error {
	return conversion.Convert(inputDir, selectedFormats)
}
