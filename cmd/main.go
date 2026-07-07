package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid"
	"github.com/open-and-sustainable/prismaid/conversion"
	terminal "github.com/open-and-sustainable/prismaid/init"
)

// main is the entry point for the PrismAId CLI application.
//
// It processes command-line arguments to perform various operations:
//   - Running a review project with a TOML configuration file
//   - Initializing a new project configuration file interactively
//   - Downloading files from a list of URLs
//   - Downloading PDFs from Zotero using credentials
//   - Converting files in various formats (PDF, DOCX, HTML) to text
//
// The function handles appropriate error logging and exits with
// non-zero status codes when operations fail.
//
// If no valid options are provided, it displays an error message
// and exits with status code 1.
func main() {
	projectConfigPath := flag.String("project", "", "Path to the project configuration file")
	initFlag := flag.Bool("init", false, "Run interactively to initialize a new project configuration file")
	downloadURLPath := flag.String("download-URL", "", "Path to a text file containing URLs to download")
	downloadZoteroPath := flag.String("download-zotero", "", "Path to a TOML file containing Zotero credentials")

	singleFilePath := flag.String("single-file", "", "Path to a single PDF file to convert")
	ocrOnly := flag.Bool("ocr-only", false, "Use Tika OCR only (PDF conversion)")
	convertPDFDir := flag.String("convert-pdf", "", "Directory containing PDF files to convert")
	convertDOCXDir := flag.String("convert-docx", "", "Directory containing DOCX files to convert")
	convertHTMLDir := flag.String("convert-html", "", "Directory containing HTML files to convert")
	tikaServer := flag.String("tika-server", "", "Tika server address for OCR fallback (e.g., 'localhost:9998' or '0.0.0.0:9998')")

	screeningConfigPath := flag.String("screening", "", "Path to the screening configuration TOML file")

	validateFlag := flag.Bool("validate", false, "Validate a configuration file without executing it; combine with -project, -screening, or -download-zotero")

	conformancePath := flag.String("conformance", "", "Path to a RevAIse review-record JSON file to check for protocol conformance")
	protocol := flag.String("protocol", "prisma-2020", "Protocol to check conformance against (used with -conformance)")

	flag.Parse()

	if flag.Arg(0) == "-help" || flag.Arg(0) == "--help" {
		flag.Usage()
		return
	}

	// Configuration validation (no execution)
	if *validateFlag {
		logger.SetupLogging(logger.Stdout, "")
		switch {
		case *projectConfigPath != "":
			handleValidate("review", *projectConfigPath)
		case *screeningConfigPath != "":
			handleValidate("screening", *screeningConfigPath)
		case *downloadZoteroPath != "":
			handleValidate("zotero", *downloadZoteroPath)
		default:
			logger.Error("Error: -validate requires one of -project, -screening, or -download-zotero with a configuration file path")
			os.Exit(1)
		}
		return
	}

	// Protocol conformance check (no execution)
	if *conformancePath != "" {
		logger.SetupLogging(logger.Stdout, "")
		handleConformance(*conformancePath, *protocol)
		return
	}

	// PDF conversion
	if *convertPDFDir != "" {
		logger.SetupLogging(logger.Stdout, "")
		if *singleFilePath != "" {
			handleConversionFile(*singleFilePath, *tikaServer, *ocrOnly)
		} else {
			handleConversionIsolated(*convertPDFDir, "pdf", *tikaServer, *ocrOnly)
		}
	}

	// DOCX conversion
	if *convertDOCXDir != "" {
		logger.SetupLogging(logger.Stdout, "")
		handleConversion(*convertDOCXDir, "docx", *tikaServer, false)
	}

	// HTML conversion
	if *convertHTMLDir != "" {
		logger.SetupLogging(logger.Stdout, "")
		handleConversion(*convertHTMLDir, "html", *tikaServer, false)
	}

	// Zotero PDF download
	if *downloadZoteroPath != "" {
		logger.SetupLogging(logger.Stdout, "")
		handleZoteroDownload(*downloadZoteroPath)
	}

	// URL download
	if *downloadURLPath != "" {
		logger.SetupLogging(logger.Stdout, "")
		prismaid.DownloadURLList(*downloadURLPath)
	}

	// Screening process
	if *screeningConfigPath != "" {
		logger.SetupLogging(logger.Stdout, "")
		data, err := os.ReadFile(*screeningConfigPath)
		if err != nil {
			logger.Error("Error reading Screening configuration:", err)
			os.Exit(1)
		}
		_, err = prismaid.Screening(string(data))
		if err != nil {
			logger.Error("Error running Screening logic:", err)
			os.Exit(1)
		}
	}

	// Review project
	if *projectConfigPath != "" {
		logger.SetupLogging(logger.Stdout, "")
		data, err := os.ReadFile(*projectConfigPath)
		if err != nil {
			logger.Error("Error reading Review configuration:", err)
			os.Exit(1)
		}
		_, err = prismaid.Review(string(data))
		if err != nil {
			logger.Error("Error running Review logic:", err)
			os.Exit(1)
		}
	}

	// Initiate project configuration
	if *initFlag {
		terminal.RunInteractiveConfigCreation()
	}

	if *singleFilePath != "" && *convertPDFDir == "" {
		logger.Error("Error: -single-file must be used with -convert-pdf")
		os.Exit(1)
	}

	if *projectConfigPath == "" && !*initFlag && *downloadURLPath == "" && *downloadZoteroPath == "" && *convertPDFDir == "" && *convertDOCXDir == "" && *convertHTMLDir == "" && *screeningConfigPath == "" {
		logger.Error("No valid options provided. Use -help for usage information.")
		os.Exit(1)
	}
}

// handleConversion processes files in the specified input directory
// and converts them to text format based on the given source format.
// If tikaServer is provided, uses Apache Tika as OCR fallback for failed conversions.
//
// It calls the conversion.Convert function to perform the actual conversion
// and handles any errors that may occur during the process. If conversion
// fails, it logs an error message and exits the program with status code 1.
// On success, it logs an informational message.
//
// Parameters:
//   - inputDir: The directory containing files to be converted
//   - format: The source format of the files (e.g., "pdf", "docx", "html")
//   - tikaServer: Optional Tika server address (e.g., "localhost:9998"). Empty string disables OCR fallback.
//
// The function doesn't return anything as it handles errors internally
// and terminates the program on failure.
func handleConversion(inputDir, format, tikaServer string, ocrOnly bool) {
	_, err := conversion.Convert(inputDir, format, conversion.ConvertOptions{
		TikaServer: tikaServer,
		PDF: conversion.PDFOptions{
			OCROnly: ocrOnly && format == "pdf",
		},
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Error converting files in %s to %s: %v", inputDir, format, err))
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf("Successfully converted files in %s to txt (source=%s)", inputDir, format))
}

func handleConversionFile(filePath, tikaServer string, ocrOnly bool) {
	_, err := conversion.Convert(filepath.Dir(filePath), "pdf", conversion.ConvertOptions{
		TikaServer: tikaServer,
		PDF: conversion.PDFOptions{
			SingleFile: filePath,
			OCROnly:    ocrOnly,
		},
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Error converting file %s to pdf: %v", filePath, err))
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf("Successfully converted file %s to txt (source=pdf)", filePath))
}

func handleConversionIsolated(inputDir, format, tikaServer string, ocrOnly bool) {
	reportPath := filepath.Join(inputDir, "conversion_report.csv")
	reportFile, err := os.OpenFile(reportPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("Error opening report file: %v", err))
		os.Exit(1)
	}
	defer reportFile.Close()

	stat, err := reportFile.Stat()
	if err != nil {
		logger.Error(fmt.Sprintf("Error stating report file: %v", err))
		os.Exit(1)
	}
	writer := csv.NewWriter(reportFile)
	if stat.Size() == 0 {
		if err := writer.Write([]string{"file", "status", "error"}); err != nil {
			logger.Error(fmt.Sprintf("Error writing report header: %v", err))
			os.Exit(1)
		}
		writer.Flush()
	}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading input directory: %v", err))
		os.Exit(1)
	}

	exePath, err := os.Executable()
	if err != nil {
		logger.Error(fmt.Sprintf("Error resolving executable path: %v", err))
		os.Exit(1)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if !matchesFormat(ext, format) {
			continue
		}
		fullPath := filepath.Join(inputDir, file.Name())
		txtPath := filepath.Join(inputDir, strings.TrimSuffix(file.Name(), ext)+".txt")
		if info, err := os.Stat(txtPath); err == nil && info.Size() > 0 {
			continue
		}

		output, err := runConvertCommand(exePath, inputDir, fullPath, tikaServer, ocrOnly)
		status := "ok"
		errMsg := ""
		if err != nil {
			status = "error"
			errMsg = err.Error()
		}
		if len(output) > 0 {
			errMsg = strings.TrimSpace(errMsg + " " + string(output))
		}

		zeroOutput := isZeroSizeFile(txtPath)
		if (err != nil || zeroOutput) && tikaServer != "" && !ocrOnly {
			if zeroOutput {
				_ = os.Remove(txtPath)
			}
			retryOutput, retryErr := runConvertCommand(exePath, inputDir, fullPath, tikaServer, true)
			if retryErr != nil {
				status = "error"
				errMsg = strings.TrimSpace(errMsg + " ocr-only retry failed: " + retryErr.Error())
				if len(retryOutput) > 0 {
					errMsg = strings.TrimSpace(errMsg + " " + string(retryOutput))
				}
			} else {
				status = "ok"
				if len(retryOutput) > 0 {
					errMsg = strings.TrimSpace(errMsg + " ocr-only retry ok: " + string(retryOutput))
				}
			}
		} else if zeroOutput {
			status = "error"
			errMsg = strings.TrimSpace(errMsg + " output txt is zero bytes")
		}
		errMsg = strings.ReplaceAll(errMsg, "\n", " ")
		errMsg = truncateString(errMsg, 2000)

		if err := writer.Write([]string{file.Name(), status, errMsg}); err != nil {
			logger.Error(fmt.Sprintf("Error writing report row for %s: %v", file.Name(), err))
			os.Exit(1)
		}
		writer.Flush()
	}
	logger.Info(fmt.Sprintf("Conversion report written to %s", reportPath))
}

func matchesFormat(ext, format string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	format = strings.ToLower(format)
	if ext == format {
		return true
	}
	return ext == "htm" && format == "html"
}

func truncateString(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "…"
}

var runConvertCommand = func(exePath, inputDir, fullPath, tikaServer string, ocrOnly bool) ([]byte, error) {
	cmd := exec.Command(exePath, "--convert-pdf", inputDir, "--single-file", fullPath)
	if tikaServer != "" {
		cmd.Args = append(cmd.Args, "--tika-server", tikaServer)
	}
	if ocrOnly {
		cmd.Args = append(cmd.Args, "--ocr-only")
	}
	return cmd.CombinedOutput()
}

func isZeroSizeFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Size() == 0
}

// handleZoteroDownload processes a TOML configuration file containing Zotero credentials
// and downloads PDFs from the specified Zotero collection or group.
//
// It reads the configuration file and calls the prismaid.DownloadZotero function
// to parse the TOML and perform the actual download.
//
// The function handles any errors that may occur during configuration reading or PDF downloading,
// logging appropriate error messages and exiting the program with status code 1 if an error occurs.
// On success, it logs an informational message.
//
// Parameters:
//   - configPath: The path to the TOML configuration file containing Zotero credentials
//
// The function doesn't return anything as it handles errors internally
// and terminates the program on failure.
// handleValidate reads a configuration file and validates it without executing
// the corresponding tool. The configType selects the configuration schema and
// must be "review", "screening", or "zotero".
//
// It logs an error and exits with status code 1 if the file cannot be read or
// the configuration is invalid. On success, it logs a confirmation message.
func handleValidate(configType, configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading %s configuration: %v", configType, err))
		os.Exit(1)
	}

	if err := prismaid.ValidateConfig(configType, string(data)); err != nil {
		logger.Error(fmt.Sprintf("Invalid %s configuration: %v", configType, err))
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("Configuration is valid (%s)", configType))
}

// handleConformance reads a RevAIse review-record JSON file and checks it against
// a reporting protocol's shapes, printing the verdict and any unmet constraints.
// It exits with status code 1 on error or when the record does not conform, so
// the result is scriptable.
func handleConformance(recordPath, protocol string) {
	data, err := os.ReadFile(recordPath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading record file: %v", err))
		os.Exit(1)
	}

	report, err := prismaid.CheckConformance(string(data), protocol)
	if err != nil {
		logger.Error(fmt.Sprintf("Conformance check failed: %v", err))
		os.Exit(1)
	}

	if report.Conforms {
		logger.Info(fmt.Sprintf("Record conforms to %s", protocol))
		return
	}

	logger.Info(fmt.Sprintf("Record does NOT conform to %s (%d unmet constraints):", protocol, len(report.Violations)))
	for _, v := range report.Violations {
		logger.Info("  - " + v.Message)
	}
	os.Exit(1)
}

func handleZoteroDownload(configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading Zotero configuration: %v", err))
		os.Exit(1)
	}

	_, err = prismaid.DownloadZotero(string(data))
	if err != nil {
		logger.Error(fmt.Sprintf("Error downloading Zotero PDFs: %v", err))
		os.Exit(1)
	}

	logger.Info("Successfully downloaded PDFs from Zotero")
}
