package convert

import (
    "bytes"
    "log"
    "os"
    pdf "github.com/ledongthuc/pdf"
	api "github.com/pdfcpu/pdfcpu/pkg/api"
    model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
    rsc "rsc.io/pdf"
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
        return extractWithPdfcpuAndRscio(path)
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

// Fallback function using pdfcpu for preprocessing and rsc.io/pdf for extraction
func extractWithPdfcpuAndRscio(path string) (string, error) {
    var buf bytes.Buffer
    file, err := os.Open(path)
    if err != nil {
        log.Printf("Failed to open PDF for pdfcpu: %v", err)
        return "", err
    }
    defer file.Close()

    conf := model.NewDefaultConfiguration()
    ctx, err := api.ReadValidateAndOptimize(file, conf)
    if err != nil {
        log.Printf("pdfcpu optimization failed: %v", err)
        return "", err
    }

    // Write the optimized PDF to a buffer
    err = api.Write(ctx, &buf, conf)
    if err != nil {
        log.Printf("Failed to write optimized PDF: %v", err)
        return "", err
    }

    // Use rsc.io/pdf to extract text from the optimized PDF
    reader, err := rsc.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
    if err != nil {
        log.Printf("Failed to read optimized PDF with rsc.io/pdf: %v", err)
        return "", err
    }

    var extractedText bytes.Buffer
    for i := 1; i <= reader.NumPage(); i++ {
        page := reader.Page(i)
        content := page.Content()
        for _, text := range content.Text {
            extractedText.WriteString(text.S)
            extractedText.WriteString(" ")
        }
    }

    result := extractedText.String()
    if result == "" {
        log.Println("Text extraction from optimized PDF returned empty result.")
        return "", nil
    }

    return result, nil
}