// Package conversion provides utilities to convert various document formats (PDF, DOCX, HTML) into plain text format.
// It exposes functions to process and extract textual content from these document types.
//
// # Overview
//
// The `conversion` package is designed to convert a variety of document formats into plain text.
// It supports the following formats:
//   - PDF: Extracts text from PDF files using the `github.com/ledongthuc/pdf` library.
//   - DOCX: Converts DOCX files into plain text using the `github.com/fumiama/go-docx` library.
//   - HTML: Strips HTML tags and extracts textual content using the `jaytaylor.com/html2text` package.
//
// # Exported Functions
//
// Convert: Processes files from the input directory and converts them into plain text format.
// The function accepts an input directory path and a comma-separated list of formats to convert.
//
// Example:
//
//	> err := conversion.Convert("/path/to/files", "pdf,docx,html")
//	> if err != nil {
//	>     log.Fatalf("Conversion failed: %v", err)
//	> }
package conversion
