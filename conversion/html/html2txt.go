package html

import (
	"os"

	html "jaytaylor.com/html2text"
)

func ReadHtml(path string) (string, error) {
	// Open the HTML file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Set options with TextOnly flag set to true
	options := html.Options{
		TextOnly: true,
	}

	// Convert HTML to plain text
	text, err := html.FromReader(file, options)
	if err != nil {
		return "", err
	}

	return text, nil
}
