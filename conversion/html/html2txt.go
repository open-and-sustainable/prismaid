package html

import (
	"io"
	"os"

	"github.com/k3a/html2text"
)

// ReadHtml converts an HTML file located at the specified path to plain text.
//
// It reads the HTML file, processes it to extract only the text content
// (excluding HTML formatting), and returns the plain text representation.
//
// Parameters:
//   - path: A string containing the file path to the HTML document to be converted.
//
// Returns:
//   - string: The plain text content extracted from the HTML document.
//   - error: An error if the file cannot be opened or if the HTML conversion fails.
func ReadHtml(path string) (string, error) {
	// Open the HTML file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the entire file content
	htmlContent, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Convert HTML to plain text
	// github.com/k3a/html2text automatically handles stripping HTML tags,
	// links, and formatting to produce clean plain text
	text := html2text.HTML2Text(string(htmlContent))

	return text, nil
}