package conversion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-and-sustainable/alembica/utils/logger"

	"github.com/open-and-sustainable/prismaid/conversion/doc"
	"github.com/open-and-sustainable/prismaid/conversion/html"
	"github.com/open-and-sustainable/prismaid/conversion/pdf"
)

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
	for _, format := range formats { // FIXED: use value, not index
		for _, file := range files {
			fullPath := filepath.Join(inputDir, file.Name())

			if filepath.Ext(file.Name()) == "."+format {
				txt_content, err := readText(fullPath, format)
				if err == nil {
					fileNameWithoutExt := strings.TrimSuffix(file.Name(), "."+format)
					txtPath := filepath.Join(inputDir, fileNameWithoutExt+".txt")

					err = writeText(txt_content, txtPath)
					if err != nil {
						logger.Error("Error: ", err)
						return fmt.Errorf("error writing to file: %v", err)
					}
				}
			} else if filepath.Ext(file.Name()) == ".htm" && format == "html" { // FIXED: only process .htm when html is selected
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
