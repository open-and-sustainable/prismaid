package conversion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/open-and-sustainable/alembica/utils/logger"

	"github.com/open-and-sustainable/prismaid/conversion/doc"
	"github.com/open-and-sustainable/prismaid/conversion/html"
	"github.com/open-and-sustainable/prismaid/conversion/ocr"
	"github.com/open-and-sustainable/prismaid/conversion/pdf"
)

// Convert processes files from the specified input directory and converts them to text format.
// If standard conversion methods fail and a Tika server address is provided, it falls back to OCR.
//
// It scans the input directory for files with extensions matching the provided formats
// (comma-separated) and converts them to .txt files. Special handling is provided for
// .htm files when the html format is specified. When tikaAddress is provided and other
// methods fail, it uses Apache Tika server with OCR support as a fallback.
//
// Parameters:
//   - inputDir: Path to the directory containing files to convert.
//   - selectedFormats: Comma-separated list of formats to process (e.g., "pdf,docx,html").
//   - options: Conversion options including Tika server and PDF-specific behavior.
//
// Returns:
//   - error: An error if directory reading fails; individual file conversion errors are logged but don't stop processing.
//
// Example:
//
//	// Without Tika OCR fallback
//	err := Convert("/path/to/documents", "pdf,docx,html", ConvertOptions{})
//	if err != nil {
//	    log.Fatalf("Conversion failed: %v", err)
//	}
//
//	// With Tika OCR fallback
//	err := Convert("/path/to/documents", "pdf,docx,html", ConvertOptions{
//	    TikaServer: "localhost:9998",
//	})
//	if err != nil {
//	    log.Fatalf("Conversion failed: %v", err)
//	}
// PDFOptions controls PDF-specific conversion behavior.
type PDFOptions struct {
	SingleFile string
	OCROnly    bool
}

// ConvertOptions controls format-specific conversion behavior.
type ConvertOptions struct {
	TikaServer string
	PDF        PDFOptions
}

// Convert processes files from the specified input directory and converts them to text format.
// If PDF.OCROnly is true, it skips standard PDF conversion and uses Tika OCR directly.
// If PDF.SingleFile is set, only that PDF is processed for the pdf format.
func Convert(inputDir, selectedFormats string, options ConvertOptions) error {
	useTika := resolveUseTika(options.TikaServer)
	if options.PDF.OCROnly && !useTika {
		return fmt.Errorf("ocr-only requested but Tika server not available at %s", options.TikaServer)
	}

	// Load files from the input directory
	files, err := os.ReadDir(inputDir)
	if err != nil {
		logger.Error("Error: ", err)
		return fmt.Errorf("error reading input directory: %v", err)
	}

	// formats
	formats := strings.Split(selectedFormats, ",")

	// parse files
	for _, format := range formats { // FIXED: use value, not index
		format = strings.TrimSpace(format)
		if format == "" {
			continue
		}
		switch format {
		case "pdf":
			if err := convertPDF(inputDir, files, useTika, options.PDF, options.TikaServer); err != nil {
				return err
			}
		case "docx":
			if err := convertDOCX(inputDir, files, useTika, options.TikaServer); err != nil {
				return err
			}
		case "html":
			if err := convertHTML(inputDir, files, useTika, options.TikaServer); err != nil {
				return err
			}
		default:
			logger.Error("Unsupported document type: ", format)
			return fmt.Errorf("unsupported document type: %s", format)
		}
	}
	return nil
}

func resolveUseTika(tikaAddress string) bool {
	if tikaAddress == "" {
		return false
	}
	if ocr.IsTikaAvailable(tikaAddress) {
		logger.Info("Tika server available at %s - will use as OCR fallback", tikaAddress)
		return true
	}
	logger.Info("Tika server not available at %s - OCR fallback disabled", tikaAddress)
	return false
}

func convertPDF(inputDir string, files []os.DirEntry, useTika bool, options PDFOptions, tikaAddress string) error {
	if options.SingleFile != "" {
		ext := filepath.Ext(options.SingleFile)
		if !strings.EqualFold(ext, ".pdf") {
			return fmt.Errorf("file extension %s does not match format pdf", ext)
		}
		return convertSingle(options.SingleFile, "pdf", ext, useTika, tikaAddress, options.OCROnly)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if !strings.EqualFold(ext, ".pdf") {
			continue
		}
		fullPath := filepath.Join(inputDir, file.Name())
		if err := convertSingle(fullPath, "pdf", ext, useTika, tikaAddress, options.OCROnly); err != nil {
			return err
		}
	}
	return nil
}

func convertDOCX(inputDir string, files []os.DirEntry, useTika bool, tikaAddress string) error {
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if !strings.EqualFold(ext, ".docx") {
			continue
		}
		fullPath := filepath.Join(inputDir, file.Name())
		if err := convertSingle(fullPath, "docx", ext, useTika, tikaAddress, false); err != nil {
			return err
		}
	}
	return nil
}

func convertHTML(inputDir string, files []os.DirEntry, useTika bool, tikaAddress string) error {
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if !strings.EqualFold(ext, ".html") && !strings.EqualFold(ext, ".htm") {
			continue
		}
		fullPath := filepath.Join(inputDir, file.Name())
		if err := convertSingle(fullPath, "html", ext, useTika, tikaAddress, false); err != nil {
			return err
		}
	}
	return nil
}

func convertSingle(fullPath, format, ext string, useTika bool, tikaAddress string, ocrOnly bool) error {
	start := time.Now()
	usedTika := false
	logger.Info("Starting conversion: %s (format=%s)", fullPath, format)

	var txtContent string
	var err error
	if ocrOnly {
		usedTika = true
		txtContent, err = ocr.ReadWithTika(fullPath, tikaAddress)
	} else {
		txtContent, err = readText(fullPath, format)
	}

	// Try Tika OCR fallback if standard methods failed or returned empty text
	if !ocrOnly && (err != nil || txtContent == "") && useTika {
		usedTika = true
		logger.Info("Standard conversion failed for %s, attempting Tika OCR fallback", filepath.Base(fullPath))
		txtContent, err = ocr.ReadWithTika(fullPath, tikaAddress)
	}

	if err == nil {
		fileNameWithoutExt := strings.TrimSuffix(filepath.Base(fullPath), ext)
		txtPath := filepath.Join(filepath.Dir(fullPath), fileNameWithoutExt+".txt")

		err = writeText(txtContent, txtPath)
		if err != nil {
			logger.Error("Error: ", err)
			return fmt.Errorf("error writing to file: %v", err)
		}
		logger.Info("Finished conversion: %s (format=%s, tika=%t, duration=%s)", fullPath, format, usedTika, time.Since(start))
	} else {
		logger.Error("Failed to convert %s (tika=%t, duration=%s): %v", fullPath, usedTika, time.Since(start), err)
	}
	return nil
}

// readText extracts text content from a file based on its format.
//
// It determines the appropriate reading function based on the specified format
// and uses it to extract text from the given file. Supported formats include
// "pdf", "docx", and "html".
//
// Parameters:
//   - file: The path to the file to read.
//   - format: The format of the file ("pdf", "docx", or "html").
//
// Returns:
//   - string: The extracted text content from the file.
//   - error: An error if the format is unsupported or if reading fails.
func readText(file string, format string) (string, error) {
	var modelFunc func(string) (string, error)
	switch format {
	case "pdf":
		modelFunc = pdf.ReadPdf
	case "docx":
		modelFunc = doc.ReadDocx
	case "html":
		modelFunc = html.ReadHtml
	default:
		logger.Error("Unsupported document type: ", format)
		return "", fmt.Errorf("unsupported document type: %s", format)
	}
	return modelFunc(file)
}

// writeText writes the provided text content to a file at the specified path.
//
// It creates a new file if it doesn't exist, or truncates an existing file,
// then writes the given text to that file. The file permissions are set to 0644.
//
// Parameters:
//   - text: The string content to write to the file.
//   - txtPath: The file path where the text should be written.
//
// Returns:
//   - error: An error if file creation, opening, or writing fails; nil otherwise.
func writeText(text string, txtPath string) error {
	// Open the file for writing. If the file doesn't exist, it will be created.
	// The os.O_WRONLY flag opens the file for writing, and os.O_CREATE creates the file if it doesn't exist.
	// os.O_TRUNC truncates the file if it already exists.
	file, err := os.OpenFile(txtPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening or creating file: %v", err)
	}
	defer file.Close() // Ensure that the file is properly closed after writing

	// Write the text to the file
	_, err = file.WriteString(text)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	logger.Info("Successfully wrote to %s", txtPath)
	return nil
}
