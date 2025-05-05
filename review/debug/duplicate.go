package debug

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/open-and-sustainable/alembica/utils/logger"
	"github.com/open-and-sustainable/prismaid/review/config"
)

const duplication_extension = "duplicate"

// DuplicateInput reads all .txt files from the configured input directory and creates copies of them with a
// specified duplication extension. Each duplicated file is named with the pattern "original_name_duplicate.txt".
// This function is useful for creating backup copies of input data or for testing purposes.
//
// Arguments:
// - config: A pointer to the application's configuration which holds the input directory details.
//
// Returns:
// - An error if the directory cannot be read or if a file operation fails, otherwise returns nil.
func DuplicateInput(config *config.Config) error {
	// Load text files from the input directory
	files, err := os.ReadDir(config.Project.Configuration.InputDirectory)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Iterate over each file in the directory
	for _, file := range files {
		// Process only .txt files
		if filepath.Ext(file.Name()) == ".txt" {
			// Construct the full file path
			filePath := filepath.Join(config.Project.Configuration.InputDirectory, file.Name())

			// Read the file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				logger.Error("Failed to read file %s: %v", file.Name(), err)
				return err
			}

			// Create the new filename with the duplication extension
			fileBaseName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			newFileName := fileBaseName + "_" + duplication_extension + ".txt"
			newFilePath := filepath.Join(config.Project.Configuration.InputDirectory, newFileName)

			// Write the duplicated content to the new file
			err = os.WriteFile(newFilePath, content, 0644)
			if err != nil {
				logger.Error("Failed to write duplicated file %s: %v", newFileName, err)
				return err
			}

			logger.Info("File %s duplicated as %s", file.Name(), newFileName)
		}
	}

	return nil
}

// RemoveDuplicateInput deletes all text files from the configured input directory that were previously
// created with the duplication extension. This function is useful for cleaning up backup copies of input
// data or for resetting after testing.
//
// Arguments:
// - config: A pointer to the application's configuration which holds the input directory details.
//
// Returns:
// - An error if the directory cannot be read or if a file operation fails, otherwise returns nil.
func RemoveDuplicateInput(config *config.Config) error {
	// Load files from the input directory
	files, err := os.ReadDir(config.Project.Configuration.InputDirectory)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Iterate over each file in the directory
	for _, file := range files {
		// Check if the file is a .txt file
		if filepath.Ext(file.Name()) == ".txt" {
			fileBaseName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			expectedSuffix := "_" + duplication_extension

			// If the filename ends with the duplication extension and .txt, it's a duplicate
			if strings.HasSuffix(fileBaseName, expectedSuffix) {
				// Construct the full file path
				filePath := filepath.Join(config.Project.Configuration.InputDirectory, file.Name())

				// Remove the file
				err := os.Remove(filePath)
				if err != nil {
					logger.Error("Failed to remove file %s: %v", file.Name(), err)
					return err
				}

				logger.Info("Removed duplicated file: %s", file.Name())
			}
		}
	}

	return nil
}
