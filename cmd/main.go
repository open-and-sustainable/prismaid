package main

import (
	"encoding/csv"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid"
	"github.com/open-and-sustainable/prismaid/conversion"
	terminal "github.com/open-and-sustainable/prismaid/init"
)

// ZoteroConfig represents the configuration needed to download PDFs from Zotero.
// It contains user ID, API key, and the collection/group name.
type ZoteroConfig struct {
	User   string `toml:"user"`
	APIKey string `toml:"api_key"`
	Group  string `toml:"group"`
}

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

	flag.Parse()

	if flag.Arg(0) == "-help" || flag.Arg(0) == "--help" {
		flag.Usage()
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
		err = prismaid.Screening(string(data))
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
		err = prismaid.Review(string(data))
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
	err := conversion.Convert(inputDir, format, conversion.ConvertOptions{
		TikaServer: tikaServer,
		PDF: conversion.PDFOptions{
			OCROnly: ocrOnly && format == "pdf",
		},
	})
	if err != nil {
		logger.Error("Error converting files in %s to %s: %v\n", inputDir, format, err)
		os.Exit(1)
	}
	logger.Info("Successfully converted files in %s to txt (source=%s)\n", inputDir, format)
}

func handleConversionFile(filePath, tikaServer string, ocrOnly bool) {
	err := conversion.Convert(filepath.Dir(filePath), "pdf", conversion.ConvertOptions{
		TikaServer: tikaServer,
		PDF: conversion.PDFOptions{
			SingleFile: filePath,
			OCROnly:    ocrOnly,
		},
	})
	if err != nil {
		logger.Error("Error converting file %s to pdf: %v\n", filePath, err)
		os.Exit(1)
	}
	logger.Info("Successfully converted file %s to txt (source=pdf)\n", filePath)
}

func handleConversionIsolated(inputDir, format, tikaServer string, ocrOnly bool) {
	reportPath := filepath.Join(inputDir, "conversion_report.csv")
	reportFile, err := os.OpenFile(reportPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Error("Error opening report file: %v\n", err)
		os.Exit(1)
	}
	defer reportFile.Close()

	stat, err := reportFile.Stat()
	if err != nil {
		logger.Error("Error stating report file: %v\n", err)
		os.Exit(1)
	}
	writer := csv.NewWriter(reportFile)
	if stat.Size() == 0 {
		if err := writer.Write([]string{"file", "status", "error"}); err != nil {
			logger.Error("Error writing report header: %v\n", err)
			os.Exit(1)
		}
		writer.Flush()
	}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		logger.Error("Error reading input directory: %v\n", err)
		os.Exit(1)
	}

	exePath, err := os.Executable()
	if err != nil {
		logger.Error("Error resolving executable path: %v\n", err)
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
			logger.Error("Error writing report row for %s: %v\n", file.Name(), err)
			os.Exit(1)
		}
		writer.Flush()
	}
	logger.Info("Conversion report written to %s\n", reportPath)
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
// It reads the configuration file, validates that all required fields (user, api_key, and group)
// are present, and then calls the prismaid.DownloadZoteroPDFs function to perform the actual download.
// The PDFs are saved to the same directory as the configuration file.
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
func handleZoteroDownload(configPath string) {
	var config ZoteroConfig
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		logger.Error("Error reading Zotero configuration: %v\n", err)
		os.Exit(1)
	}

	if config.User == "" || config.APIKey == "" || config.Group == "" {
		logger.Error("Error: Zotero configuration must include user, api_key, and group")
		os.Exit(1)
	}

	configDir := filepath.Dir(configPath)
	err = prismaid.DownloadZoteroPDFs(config.User, config.APIKey, config.Group, configDir)
	if err != nil {
		logger.Error("Error downloading Zotero PDFs: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Successfully downloaded PDFs from Zotero")
}
