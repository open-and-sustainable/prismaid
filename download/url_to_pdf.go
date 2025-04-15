package download

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-and-sustainable/alembica/utils/logger"

	"github.com/PuerkitoBio/goquery"
)

func RunListDownload(path string) {
	// Extract the directory from the input file path
	dirPath := filepath.Dir(path)

	// Open the file at the given path
	file, err := os.Open(path)
	if err != nil {
		logger.Error("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read all URLs from the file
	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Error reading file:", err)
		return
	}

	// Process each URL
	for _, url := range urls {
		pdfURL, filename, err := extractPDF(url)
		if err != nil {
			logger.Error("Error processing URL", url, ":", err)
			continue
		}

		if pdfURL != "" {
			// Create full path for saving the PDF
			fullPath := filepath.Join(dirPath, filename)

			if err := downloadPDF(pdfURL, fullPath); err != nil {
				logger.Error("Download failed for", url, ":", err)
			} else {
				logger.Info("PDF downloaded successfully as", fullPath)
			}
		} else {
			logger.Info("No PDF found for", url)
		}
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
		// Get first 20 chars of pageURL for filename
		maxLen := min(len(pageURL), 20)
		urlBase := pageURL[:maxLen]
		// Sanitize for a valid filename
		filename = sanitizeFilename(urlBase) + ".pdf"
		return pageURL, filename, nil
	}

	// Parse HTML if not a direct PDF
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Extract title for naming
	title := doc.Find("title").Text()
	if title == "" {
		title = "downloaded"
	}
	// Properly sanitize the title for a valid filename
	filename = sanitizeFilename(title) + ".pdf"

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

// Add this helper function to sanitize filenames
func sanitizeFilename(name string) string {
	// Replace all invalid filename characters
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|", "#", "%", "&", "{", "}", "$", "!", "@", "+", "=", "`", "~"}
	result := name

	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Also replace whitespace characters
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, "\t", "_")
	result = strings.ReplaceAll(result, "\n", "_")
	result = strings.ReplaceAll(result, "\r", "_")

	return result
}

// Modified to accept the full path instead of just filename
func downloadPDF(pdfURL, fullPath string) error {
	resp, err := http.Get(pdfURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
