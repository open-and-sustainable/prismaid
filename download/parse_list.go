package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	url := "https://example.com/journal-article"
	pdfURL, filename, err := extractPDF(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if pdfURL != "" {
		if err := downloadPDF(pdfURL, filename); err != nil {
			fmt.Println("Download failed:", err)
		} else {
			fmt.Println("PDF downloaded successfully as", filename)
		}
	} else {
		fmt.Println("No PDF found.")
	}
}

func extractPDF(pageURL string) (pdfURL string, filename string, err error) {
	// Fetch the page
	resp, err := http.Get(pageURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// Direct PDF check
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/pdf") {
		// Derive a default filename, or use metadata if available
		filename = "downloaded.pdf"
		return pageURL, filename, nil
	}

	// Parse HTML if not a direct PDF
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Optionally, extract title for naming
	title := doc.Find("title").Text()
	if title == "" {
		title = "downloaded"
	} else {
		// Replace spaces and non-alphanumeric characters to create a valid filename
		title = strings.ReplaceAll(title, " ", "_")
	}
	filename = title + ".pdf"

	// Look for PDF link
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if href, exists := s.Attr("href"); exists && strings.HasSuffix(strings.ToLower(href), ".pdf") {
			pdfURL = href
			return false // found one, stop iteration
		}
		return true
	})

	// If found, make sure it's an absolute URL
	if pdfURL != "" && !strings.HasPrefix(pdfURL, "http") {
		pdfURL = pageURL + pdfURL // This is a simplistic approach; consider using URL parsing for robustness.
	}
	return pdfURL, filename, nil
}

func downloadPDF(pdfURL, filename string) error {
	resp, err := http.Get(pdfURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
