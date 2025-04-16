package main

import (
	"flag"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid"
	"github.com/open-and-sustainable/prismaid/convert/file"
	terminal "github.com/open-and-sustainable/prismaid/init"
)

type ZoteroConfig struct {
	User   string `toml:"user"`
	APIKey string `toml:"api_key"`
	Group  string `toml:"group"`
}

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

func handleConversion(inputDir, format string) {
	err := file.Convert(inputDir, format)
	if err != nil {
		logger.Error("Error converting files in %s to %s: %v\n", inputDir, format, err)
		os.Exit(1)
	}
	logger.Info("Successfully converted files in %s to %s format\n", inputDir, format)
}

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

	err = prismaid.DownloadZoteroPDFs(config.User, config.APIKey, config.Group, configPath)
	if err != nil {
		logger.Error("Error downloading Zotero PDFs: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Successfully downloaded PDFs from Zotero")
}
