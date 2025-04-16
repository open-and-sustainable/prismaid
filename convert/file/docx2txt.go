package file

import (
	"os"
	"strings"

	docx "github.com/fumiama/go-docx"
)

func readDocx(path string) (string, error) {
	// Create a strings.Builder to collect the content
	var textBuilder strings.Builder
	readFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer readFile.Close() // Ensure the file is closed after reading
	fileinfo, err := readFile.Stat()
	if err != nil {
		return "", err
	}
	size := fileinfo.Size()
	doc, err := docx.Parse(readFile, size)
	if err != nil {
		return "", err
	}
	for _, it := range doc.Document.Body.Items {
		switch it.(type) {
		case *docx.Paragraph, *docx.Table:
			// Append the content of Paragraph or Table to the text builder
			textBuilder.WriteString(it.(interface{ String() string }).String())
			textBuilder.WriteString("\n") // Add a newline for formatting
		}
	}
	// Return the accumulated text content
	return textBuilder.String(), nil
}
