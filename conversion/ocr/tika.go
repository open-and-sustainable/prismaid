package ocr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/open-and-sustainable/alembica/utils/logger"
)

// ReadWithTika extracts text from a file using Apache Tika server with OCR support.
//
// This function sends a file to a running Apache Tika server for text extraction.
// Tika supports OCR via Tesseract for images and scanned PDFs when no other text
// extraction method works. The server must be running and accessible at the specified
// address and port.
//
// Parameters:
//   - path: The file path to the document to be processed
//   - tikaAddress: The Tika server address (e.g., "localhost:9998" or "0.0.0.0:9998")
//
// Returns:
//   - A string containing all extracted text (including OCR results)
//   - An error if the file cannot be read, the server is unreachable, or extraction fails
func ReadWithTika(path string, tikaAddress string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		logger.Error("Failed to open file for Tika OCR: %v", err)
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read file content for Tika OCR: %v", err)
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Prepare HTTP request to Tika server
	tikaURL := fmt.Sprintf("http://%s/tika", tikaAddress)
	req, err := http.NewRequest("PUT", tikaURL, bytes.NewReader(fileContent))
	if err != nil {
		logger.Error("Failed to create Tika request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to get plain text output
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", "application/octet-stream")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute, // OCR can take time for large documents
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to connect to Tika server at %s: %v", tikaAddress, err)
		return "", fmt.Errorf("failed to connect to Tika server: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Error("Tika server returned error: %d - %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("tika server returned status %d", resp.StatusCode)
	}

	// Read response body
	textContent, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read Tika response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	logger.Info("Successfully extracted text from %s using Tika OCR", path)
	return string(textContent), nil
}

// IsTikaAvailable checks if a Tika server is running and accessible at the specified address.
//
// This function sends a GET request to the Tika server's root endpoint to verify
// that the server is running and responding. It can be used before attempting text
// extraction to provide better error messages to users.
//
// Parameters:
//   - tikaAddress: The Tika server address (e.g., "localhost:9998" or "0.0.0.0:9998")
//
// Returns:
//   - true if the server is accessible and responding, false otherwise
func IsTikaAvailable(tikaAddress string) bool {
	tikaURL := fmt.Sprintf("http://%s/tika", tikaAddress)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(tikaURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent
}
