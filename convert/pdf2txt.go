package convert

import (
    "log"
    "os"
    "regexp"
    pdf "github.com/ledongthuc/pdf"

	api "github.com/pdfcpu/pdfcpu/pkg/api"
    model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
    pdfTypes "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Primary text extraction function using github.com/ledongthuc/pdf
func readPdf(path string) (string, error) {
    text := ""

    // Open the PDF file
    f, r, err := pdf.Open(path)
    if err != nil {
        log.Printf("Failed to open PDF: %v", err)
        return "", err
    }
    defer f.Close()

    totalPage := r.NumPage()
    if totalPage == 0 {
        log.Println("The PDF contains no pages")
        return "", nil
    }

    for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
        p := r.Page(pageIndex)
        if p.V.IsNull() {
            log.Printf("Page %d is null or not available", pageIndex)
            continue
        }

        rows, err := p.GetTextByRow()
        if err != nil {
            log.Printf("Error retrieving text for page %d: %v", pageIndex, err)
            continue
        }
        if len(rows) == 0 {
            log.Printf("No text rows found on page %d", pageIndex)
            continue
        }

        for _, row := range rows {
            if len(row.Content) == 0 {
                log.Printf("Empty content on page %d", pageIndex)
                continue
            }
            line := textsToString(row.Content)
            if line == "" {
                log.Printf("Converted text is empty on page %d", pageIndex)
            }
            text += line + "\n"
        }
    }

    // Fallback if no text was extracted
    if text == "" {
        log.Println("No text extracted from any pages of the PDF, attempting alternative method.")
        return extractTextWithPdfCpu(path)
    }
    return text, nil
}

// Convert a []Text to a single string by concatenating the Value fields
func textsToString(texts []pdf.Text) string {
    result := ""
    for _, text := range texts {
        result += text.S
    }
    return result
}

// extractTextFromPDF reads a PDF and extracts text from each page's content stream.
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
		log.Printf("Optimization failed: %v", err)
	}

	var extractedText string
	// Process each page
	for i := 1; i <= ctx.PageCount; i++ {
		pageDict, _, _, err := ctx.PageDict(i, false)
		if err != nil {
			log.Printf("Error extracting page %d: %v", i, err)
			continue
		}

		contentsEntry, ok := pageDict.Find("Contents")
        if !ok {
            log.Printf("No content stream found for page %d", i)
            continue
        }

        var contentData []byte

        switch obj := contentsEntry.(type) {
        case pdfTypes.IndirectRef:
            // Single content stream
            streamDict, found, err := ctx.DereferenceStreamDict(obj)
            if err != nil || !found {
                log.Printf("Failed to dereference single content stream for page %d: %v", i, err)
                continue
            }

            err = streamDict.Decode()
            if err != nil {
                log.Printf("Failed to decode single content stream for page %d: %v", i, err)
                continue
            }

            contentData = append(contentData, streamDict.Content...)

        case pdfTypes.Array:
            // Array of content streams
            for _, element := range obj {
                // Check if the element is an indirect reference
                indirectRef, ok := element.(pdfTypes.IndirectRef)
                if !ok {
                    log.Printf("Invalid content stream reference (not IndirectRef) for page %d: %T", i, element)
                    continue
                }
        
                // Dereference the indirect reference
                obj, err := ctx.Dereference(indirectRef)
                if err != nil {
                    log.Printf("Failed to dereference object in array for page %d: %v", i, err)
                    continue
                }
                if obj == nil {
                    log.Printf("Dereferenced object is nil for page %d", i)
                    continue
                }

                // Check if the object is a StreamDict
                streamDict, ok := obj.(pdfTypes.StreamDict)
                if !ok {
                    log.Printf("Dereferenced object is not a StreamDict for page %d: %T", i, obj)
                    continue
                }

                err = streamDict.Decode()
                if err != nil {
                    log.Printf("Failed to decode stream for page %d: %v", i, err)
                    continue
                }

                contentData = append(contentData, streamDict.Content...)

        
                // Decode the stream content
                err = streamDict.Decode()
                if err != nil {
                    log.Printf("Failed to decode stream in array for page %d: %v", i, err)
                    continue
                }
        
                contentData = append(contentData, streamDict.Content...)
            }
                

        case *pdfTypes.StreamDict:
            // Direct content stream
            err := obj.Decode()
            if err != nil {
                log.Printf("Failed to decode direct content stream for page %d: %v", i, err)
                continue
            }
        
            contentData = append(contentData, obj.Content...)

        default:
            log.Printf("Unexpected type for 'Contents' entry on page %d: %T", i, obj)
            continue
        }

        // Now use the collected content data
        content := contentData
        text := parseText(content)
        extractedText += text + "\n"
	}

	return extractedText, nil
}

// parseText extracts text from PDF content using regular expressions.
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