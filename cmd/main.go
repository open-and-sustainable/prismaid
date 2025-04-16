package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid"
	terminal "github.com/open-and-sustainable/prismaid/init"
)

// ZoteroConfig holds the configuration for downloading PDFs from Zotero
type ZoteroConfig struct {
	User   string `toml:"user"`
	APIKey string `toml:"api_key"`
	Group  string `toml:"group"`
}

// Main function
func main() {
	// Define flags for the project configuration file, the init option, and the download option
	projectConfigPath := flag.String("project", "", "Path to the project configuration file")
	initFlag := flag.Bool("init", false, "Run interactively to initialize a new project configuration file")
	downloadURLPath := flag.String("download-URL", "", "Path to a text file containing URLs to download")
	downloadZoteroPath := flag.String("download-zotero", "", "Path to a TOML file containing Zotero credentials")

	// Parse the flags
	flag.Parse()

	// Check if the user requested help
	if flag.Arg(0) == "-help" || flag.Arg(0) == "--help" {
		flag.Usage()
		return
	}

	// Handle Zotero download logic if -download-zotero flag is provided
	if *downloadZoteroPath != "" {
		handleZoteroDownload(*downloadZoteroPath)
		return
	}

	// Handle download logic if -download-URL flag is provided
	if *downloadURLPath != "" {
		logger.SetupLogging(logger.Stdout, "")
		prismaid.DownloadURLList(*downloadURLPath)
		return
	}

	// Check if no valid flags are provided
	if *projectConfigPath == "" && !*initFlag && *downloadURLPath == "" && *downloadZoteroPath == "" {
		fmt.Println("Usage: ./prismAId_OS_CPU[.exe] --project <path-to-your-project-config.toml> or --init or --download-URL <path-to-url-list.txt> or --download-zotero <path-to-zotero-config.toml>")
		os.Exit(1)
	}

	// Handle project logic if -project flag is provided
	if *projectConfigPath != "" {
		// Read the file using the injected FileReader interface
		data, err := os.ReadFile(*projectConfigPath)
		if err != nil {
			fmt.Println("Error reading Review configuration:", err)
			os.Exit(1)
		}

		err = prismaid.Review(string(data))
		if err != nil {
			fmt.Println("Error running Review logic:", err)
			os.Exit(1)
		}
		return
	}

	// Handle init logic if -init flag is provided
	if *initFlag {
		terminal.RunInteractiveConfigCreation()
		return
	}
}

// handleZoteroDownload parses the Zotero config TOML file and calls the DownloadZoteroPDFs function
func handleZoteroDownload(configPath string) {
	// Set up logging
	logger.SetupLogging(logger.Stdout, "")

	// Read the TOML file
	var config ZoteroConfig
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		fmt.Printf("Error reading Zotero configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate config
	if config.User == "" || config.APIKey == "" || config.Group == "" {
		fmt.Println("Error: Zotero configuration must include user, api_key, and group")
		os.Exit(1)
	}

	// Get the directory where the TOML file is located to use as parentDir
	parentDir := filepath.Dir(configPath)

	// Call the function with provided parameters
	err = prismaid.DownloadZoteroPDFs(config.User, config.APIKey, config.Group, parentDir)
	if err != nil {
		fmt.Printf("Error downloading Zotero PDFs: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully downloaded PDFs from Zotero")
}
