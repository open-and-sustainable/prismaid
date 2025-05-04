package pdf

import (
	"os"
	"regexp"

	pdf "github.com/ledongthuc/pdf"

	api "github.com/pdfcpu/pdfcpu/pkg/api"
	model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	pdfTypes "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/open-and-sustainable/alembica/utils/logger"
)

// ReadPdf extracts text content from a PDF file using the ledongthuc/pdf library.
//
// This function reads a PDF file from the specified path and attempts to extract
// all text content page by page. It processes each page by retrieving text organized
// in rows and concatenating it with newlines. If a page contains no extractable text
// or encounters an error, appropriate warnings are logged and processing continues
// with the next page. If no text is successfully extracted from any page, the function
// falls back to an alternative extraction method using pdfcpu.
//
// Parameters:
//   - path: The file path to the PDF document to be processed
//
// Returns:
//   - A string containing all extracted text with newlines between text rows
//   - An error if the PDF file cannot be opened or processed
func ReadPdf(path string) (string, error) {
	text := ""

	// Open the PDF file
	f, r, err := pdf.Open(path)
	if err != nil {
		logger.Error("Failed to open PDF: %v", err)
		return "", err
	}
	defer f.Close()

	totalPage := r.NumPage()
	if totalPage == 0 {
		logger.Error("The PDF contains no pages")
		return "", nil
	}

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			logger.Error("Page %d is null or not available", pageIndex)
			continue
		}

		rows, err := p.GetTextByRow()
		if err != nil {
			logger.Error("Error retrieving text for page %d: %v", pageIndex, err)
			continue
		}
		if len(rows) == 0 {
			logger.Error("No text rows found on page %d", pageIndex)
			continue
		}

		for _, row := range rows {
			if len(row.Content) == 0 {
				logger.Error("Empty content on page %d", pageIndex)
				continue
			}
			line := textsToString(row.Content)
			if line == "" {
				logger.Error("Converted text is empty on page %d", pageIndex)
			}
			text += line + "\n"
		}
	}

	// Fallback if no text was extracted
	if text == "" {
		logger.Info("No text extracted from any pages of the PDF, attempting alternative method.")
		return extractTextWithPdfCpu(path)
	}
	return text, nil
}

// textsToString converts a slice of PDF text objects into a single string.
//
// This function iterates through a slice of Text structs from the ledongthuc/pdf
// package and concatenates their string values (S field) into a single result string.
// It preserves the exact text content without adding any spacing or formatting.
//
// Parameters:
//   - texts: A slice of pdf.Text objects containing text content from a PDF
//
// Returns:
//   - A string containing the concatenated text values
func textsToString(texts []pdf.Text) string {
	result := ""
	for _, text := range texts {
		result += text.S
	}
	return result
}

// extractTextWithPdfCpu extracts text from a PDF using the pdfcpu library as a fallback method.
//
// This function serves as an alternative text extraction method when the primary extraction
// fails. It processes each page of the PDF, handling various content stream structures
// (direct, indirect references, and arrays), decodes them, and extracts text using regex.
// The function uses relaxed validation to handle potentially problematic PDFs.
//
// Parameters:
//   - filePath: The path to the PDF file to extract text from
//
// Returns:
//   - A string containing all extracted text with newlines between pages
//   - An error if file opening, context creation, or text extraction fails
func extractTextWithPdfCpu(filePath string) (string, error) {
	// Open the PDF file
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Create a pdfcpu configuration with relaxed validation
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	// Create a pdfcpu configuration and context
	ctx, err := api.ReadContext(f, conf)
	if err != nil {
		return "", err
	}

	// Optimize the PDF context to fix minor issues
	if err := api.OptimizeContext(ctx); err != nil {
		logger.Error("Optimization failed: %v", err)
	}

	var extractedText string
	// Process each page
	for i := 1; i <= ctx.PageCount; i++ {
		pageDict, _, _, err := ctx.PageDict(i, false)
		if err != nil {
			logger.Error("Error extracting page %d: %v", i, err)
			continue
		}

		contentsEntry, ok := pageDict.Find("Contents")
		if !ok {
			logger.Error("No content stream found for page %d", i)
			continue
		}

		var contentData []byte

		switch obj := contentsEntry.(type) {
		case pdfTypes.IndirectRef:
			// Single content stream
			streamDict, found, err := ctx.DereferenceStreamDict(obj)
			if err != nil || !found {
				logger.Error("Failed to dereference single content stream for page %d: %v", i, err)
				continue
			}

			err = streamDict.Decode()
			if err != nil {
				logger.Error("Failed to decode single content stream for page %d: %v", i, err)
				continue
			}

			contentData = append(contentData, streamDict.Content...)

		case pdfTypes.Array:
			// Array of content streams
			for _, element := range obj {
				// Check if the element is an indirect reference
				indirectRef, ok := element.(pdfTypes.IndirectRef)
				if !ok {
					logger.Error("Invalid content stream reference (not IndirectRef) for page %d: %T", i, element)
					continue
				}

				// Dereference the indirect reference
				obj, err := ctx.Dereference(indirectRef)
				if err != nil {
					logger.Error("Failed to dereference object in array for page %d: %v", i, err)
					continue
				}
				if obj == nil {
					logger.Error("Dereferenced object is nil for page %d", i)
					continue
				}

				// Check if the object is a StreamDict
				streamDict, ok := obj.(pdfTypes.StreamDict)
				if !ok {
					logger.Error("Dereferenced object is not a StreamDict for page %d: %T", i, obj)
					continue
				}

				err = streamDict.Decode()
				if err != nil {
					logger.Error("Failed to decode stream for page %d: %v", i, err)
					continue
				}

				contentData = append(contentData, streamDict.Content...)

				// Decode the stream content
				err = streamDict.Decode()
				if err != nil {
					logger.Error("Failed to decode stream in array for page %d: %v", i, err)
					continue
				}

				contentData = append(contentData, streamDict.Content...)
			}

		case *pdfTypes.StreamDict:
			// Direct content stream
			err := obj.Decode()
			if err != nil {
				logger.Error("Failed to decode direct content stream for page %d: %v", i, err)
				continue
			}

			contentData = append(contentData, obj.Content...)

		default:
			logger.Error("Unexpected type for 'Contents' entry on page %d: %T", i, obj)
			continue
		}

		// Now use the collected content data
		content := contentData
		text := parseText(content)
		extractedText += text + "\n"
	}

	return extractedText, nil
}

// parseText extracts text from PDF content streams using regular expressions.
//
// This function searches for text drawing operators (Tj) in the PDF content stream
// and extracts the text within parentheses that precedes these operators.
// It concatenates all found text segments with spaces between them.
//
// Parameters:
//   - content: A byte slice containing the raw PDF content stream data
//
// Returns:
//   - A string containing all extracted text with spaces between segments
func parseText(content []byte) string {
	re := regexp.MustCompile(`\((.*?)\)Tj`)
	matches := re.FindAllSubmatch(content, -1)

	var text string
	for _, match := range matches {
		if len(match) > 1 {
			text += string(match[1]) + " "
		}
	}
	return text
}
