package html

import (
	"os"

	html "jaytaylor.com/html2text"
)

// ReadHtml converts an HTML file located at the specified path to plain text.
//
// It reads the HTML file, processes it with specific options to extract only the text content
// (excluding links and tables), and returns the plain text representation.
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

	// Set options with TextOnly flag set to true
	options := html.Options{
		TextOnly:     true,
		PrettyTables: false,
		OmitLinks:    true,
	}

	// Convert HTML to plain text
	text, err := html.FromReader(file, options)
	if err != nil {
		return "", err
	}

	return text, nil
}
