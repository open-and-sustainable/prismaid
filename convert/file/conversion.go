package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-and-sustainable/alembica/utils/logger"
)

// Convert processes files from the input directory specified in the configuration and converts them into plain text files.
//
// It reads the configuration settings to identify supported formats and input directory paths. The function attempts to
// convert each file into a .txt file based on its format.
//
// Parameters:
//   - config: A pointer to a config.Config instance containing configuration details.
//
// Returns:
//   - An error if any issue occurs during reading, processing, or writing the files.
//
// Example:
//
//	> err := convert.Convert(config)
//	> if err != nil {
//	>     log.Fatalf("Conversion failed: %v", err)
//	> }
func Convert(inputDir, selectedFormats string) error {
	// Load files from the input directory
	files, err := os.ReadDir(inputDir)
	if err != nil {
		logger.Error("Error: ", err)
		return fmt.Errorf("error reading input directory: %v", err)
	}
	// formats
	formats := strings.Split(selectedFormats, ",")
	// parse files
	for format := range formats {
		for _, file := range files {
			fullPath := filepath.Join(inputDir, file.Name())

			if filepath.Ext(file.Name()) == "."+formats[format] {
				txt_content, err := readText(fullPath, formats[format])
				if err == nil {
					fileNameWithoutExt := strings.TrimSuffix(file.Name(), "."+formats[format])
					txtPath := filepath.Join(inputDir, fileNameWithoutExt+".txt")

					err = writeText(txt_content, txtPath)
					if err != nil {
						logger.Error("Error: ", err)
						return fmt.Errorf("error writing to file: %v", err)
					}
				}
			} else if filepath.Ext(file.Name()) == ".htm" { // this is to treat the special case of html files svaed with .htm extension
				txt_content, err := readText(fullPath, "html")
				if err == nil {
					fileNameWithoutExt := strings.TrimSuffix(file.Name(), ".htm")
					txtPath := filepath.Join(inputDir, fileNameWithoutExt+".txt")
					err = writeText(txt_content, txtPath)
					if err != nil {
						logger.Error("Error: ", err)
						return fmt.Errorf("error writing to file: %v", err)
					}
				}
			}
		}
	}
	return nil
}

func readText(file string, format string) (string, error) {
	var modelFunc func(string) (string, error)
	switch format {
	case "pdf":
		modelFunc = readPdf
	case "docx":
		modelFunc = readDocx
	case "html":
		modelFunc = readHtml
	default:
		logger.Error("Unsupported document type: ", format)
		return "", fmt.Errorf("unsupported document type: %s", format)
	}
	return modelFunc(file)
}

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

	logger.Info("Successfully wrote to %s\n", txtPath)
	return nil
}
