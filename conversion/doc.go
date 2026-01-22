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
//   - OCR Fallback: When standard methods fail, optionally uses Apache Tika server for OCR-based text extraction.
//
// # Exported Functions
//
// Convert: Processes files from the input directory and converts them into plain text format.
// The function accepts an input directory path, a comma-separated list of formats to convert,
// and conversion options including an optional Tika server address for OCR fallback.
//
// When a Tika server address is provided (e.g., "localhost:9998"), files that fail standard conversion
// will automatically be sent to the Tika server for OCR-based text extraction as a fallback.
// Leave TikaServer empty to disable OCR fallback.
//
// Example:
//
//	> // Without Tika OCR fallback
//	> err := conversion.Convert("/path/to/files", "pdf,docx,html", conversion.ConvertOptions{})
//	> if err != nil {
//	>     log.Fatalf("Conversion failed: %v", err)
//	> }
//
//	> // With Tika OCR fallback
//	> err := conversion.Convert("/path/to/files", "pdf,docx,html", conversion.ConvertOptions{
//	>     TikaServer: "localhost:9998",
//	> })
//	> if err != nil {
//	>     log.Fatalf("Conversion failed: %v", err)
//	> }
package conversion
