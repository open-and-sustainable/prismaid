package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid"
	"github.com/open-and-sustainable/prismaid/download"
	terminal "github.com/open-and-sustainable/prismaid/init"
)

// Main function
func main() {
	// Define flags for the project configuration file, the init option, and the download option
	projectConfigPath := flag.String("project", "", "Path to the project configuration file")
	initFlag := flag.Bool("init", false, "Run interactively to initialize a new project configuration file")
	downloadPath := flag.String("download", "", "Path to a text file containing URLs to download")

	// Parse the flags
	flag.Parse()

	// Check if the user requested help
	if flag.Arg(0) == "-help" || flag.Arg(0) == "--help" {
		flag.Usage()
		return
	}

	// Handle download logic if -download flag is provided
	if *downloadPath != "" {
		logger.SetupLogging(logger.Stdout, "")
		download.RunListDownload(*downloadPath)
		return
	}

	// Check if no valid flags are provided
	if *projectConfigPath == "" && !*initFlag && *downloadPath == "" {
		fmt.Println("Usage: ./prismAId_OS_CPU[.exe] --project <path-to-your-project-config.toml> or --init or --download <path-to-url-list.txt>")
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

		err = prismaid.RunReview(string(data))
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
