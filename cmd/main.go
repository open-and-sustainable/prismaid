package main

import (
	"flag"
	"os"
	"path/filepath"

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

	convertPDFDir := flag.String("convert-pdf", "", "Directory containing PDF files to convert")
	convertDOCXDir := flag.String("convert-docx", "", "Directory containing DOCX files to convert")
	convertHTMLDir := flag.String("convert-html", "", "Directory containing HTML files to convert")

	flag.Parse()

	if flag.Arg(0) == "-help" || flag.Arg(0) == "--help" {
		flag.Usage()
		return
	}

	// PDF conversion
	if *convertPDFDir != "" {
		logger.SetupLogging(logger.Stdout, "")
		handleConversion(*convertPDFDir, "pdf")
	}

	// DOCX conversion
	if *convertDOCXDir != "" {
		logger.SetupLogging(logger.Stdout, "")
		handleConversion(*convertDOCXDir, "docx")
	}

	// HTML conversion
	if *convertHTMLDir != "" {
		logger.SetupLogging(logger.Stdout, "")
		handleConversion(*convertHTMLDir, "html")
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

	// Review project
	if *projectConfigPath != "" {
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

	if *projectConfigPath == "" && !*initFlag && *downloadURLPath == "" && *downloadZoteroPath == "" && *convertPDFDir == "" && *convertDOCXDir == "" && *convertHTMLDir == "" {
		logger.Error("No valid options provided. Use -help for usage information.")
		os.Exit(1)
	}
}

// handleConversion processes files in the specified input directory
// and converts them to text format based on the given source format.
//
// It calls the conversion.Convert function to perform the actual conversion
// and handles any errors that may occur during the process. If conversion
// fails, it logs an error message and exits the program with status code 1.
// On success, it logs an informational message.
//
// Parameters:
//   - inputDir: The directory containing files to be converted
//   - format: The source format of the files (e.g., "pdf", "docx", "html")
//
// The function doesn't return anything as it handles errors internally
// and terminates the program on failure.
func handleConversion(inputDir, format string) {
	err := conversion.Convert(inputDir, format)
	if err != nil {
		logger.Error("Error converting files in %s to %s: %v\n", inputDir, format, err)
		os.Exit(1)
	}
	logger.Info("Successfully converted files in %s to %s format\n", inputDir, format)
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
