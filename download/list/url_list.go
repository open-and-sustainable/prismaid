package list

import (
	"bufio"
	"io"
	"net/http"
	"net/url"
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

// findPDFLink searches a webpage for a link to a PDF file using multiple detection strategies
func findPDFLink(doc *goquery.Document, pageURL string) string {
	var pdfURL string

	// Strategy 1: Check for citation_pdf_url in meta tags (most reliable for academic papers)
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		property, _ := s.Attr("property")
		content, exists := s.Attr("content")

		nameLower := strings.ToLower(name)
		propertyLower := strings.ToLower(property)

		if exists && (strings.Contains(nameLower, "citation_pdf_url") ||
			strings.Contains(propertyLower, "citation_pdf_url") ||
			strings.Contains(nameLower, "fulltext_pdf_url")) {
			pdfURL = content
		}
	})

	if pdfURL != "" {
		return pdfURL
	}

	// Strategy 2: Site-specific patterns for common academic publishers
	baseURL, err := url.Parse(pageURL)
	if err == nil {
		hostname := baseURL.Hostname()

		// MDPI pattern (handles doi.org/10.3390)
		if strings.Contains(hostname, "mdpi.com") || strings.Contains(pageURL, "doi.org/10.3390") {
			doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
				href, exists := s.Attr("href")
				if exists && strings.Contains(href, "/pdf") {
					pdfURL = href
					return false
				}
				return true
			})
		}

		// ScienceDirect pattern
		if strings.Contains(hostname, "sciencedirect.com") || strings.Contains(pageURL, "doi.org/10.1016") {
			doc.Find("a.pdf-download-btn-link, a.download-link, a.article-download-pdf-link").EachWithBreak(func(i int, s *goquery.Selection) bool {
				href, exists := s.Attr("href")
				if exists {
					pdfURL = href
					return false
				}
				return true
			})
		}

		// Springer pattern
		if strings.Contains(hostname, "springer.com") || strings.Contains(pageURL, "doi.org/10.1007") {
			doc.Find("a.download-article, a[data-track-action='Download PDF']").EachWithBreak(func(i int, s *goquery.Selection) bool {
				href, exists := s.Attr("href")
				if exists {
					pdfURL = href
					return false
				}
				return true
			})
		}

		// IEEE pattern
		if strings.Contains(hostname, "ieee.org") || strings.Contains(pageURL, "doi.org/10.1109") {
			doc.Find("a.pdf-btn, a.doc-actions-link, a.pdf-file").EachWithBreak(func(i int, s *goquery.Selection) bool {
				href, exists := s.Attr("href")
				if exists {
					pdfURL = href
					return false
				}
				return true
			})
		}

		// Wiley pattern
		if strings.Contains(hostname, "wiley.com") || strings.Contains(pageURL, "doi.org/10.1002") {
			doc.Find("a.pdf-download, a[title='PDF']").EachWithBreak(func(i int, s *goquery.Selection) bool {
				href, exists := s.Attr("href")
				if exists {
					pdfURL = href
					return false
				}
				return true
			})
		}
	}

	if pdfURL != "" {
		return pdfURL
	}

	// Strategy 3: Look for links ending with .pdf (the original approach)
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if exists && strings.HasSuffix(strings.ToLower(href), ".pdf") {
			pdfURL = href
			return false // found one, stop iteration
		}
		return true
	})

	if pdfURL != "" {
		return pdfURL
	}

	// Strategy 4: Look for links with PDF-related text or attributes
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if !exists {
			return true
		}

		hrefLower := strings.ToLower(href)
		textLower := strings.ToLower(s.Text())

		// Check for download buttons/links with PDF-related text
		if strings.Contains(hrefLower, "pdf") ||
			strings.Contains(textLower, "pdf") ||
			strings.Contains(textLower, "download") && strings.Contains(textLower, "full text") {

			// Avoid false positives
			if !strings.Contains(hrefLower, "cover") && !strings.Contains(hrefLower, "sample") {
				pdfURL = href
				return false
			}
		}
		return true
	})

	if pdfURL != "" {
		return pdfURL
	}

	// Strategy 5: Look for elements with download attributes or PDF-related classes
	doc.Find("[download], .download-pdf, .pdf-download").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if exists {
			pdfURL = href
			return false
		}
		return true
	})

	return pdfURL
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

	// Use our enhanced function to find PDF links
	pdfURL = findPDFLink(doc, pageURL)

	// If found, make sure it's an absolute URL
	if pdfURL != "" && !strings.HasPrefix(pdfURL, "http") {
		// Use url.Parse to properly resolve relative URLs
		base, err := url.Parse(pageURL)
		if err == nil {
			relative, err := url.Parse(pdfURL)
			if err == nil {
				pdfURL = base.ResolveReference(relative).String()
			} else {
				// Fallback to simple joining if URL parsing fails
				if strings.HasPrefix(pdfURL, "/") {
					// Absolute path - join with scheme and host
					u, err := url.Parse(pageURL)
					if err == nil {
						pdfURL = u.Scheme + "://" + u.Host + pdfURL
					} else {
						pdfURL = pageURL + pdfURL
					}
				} else {
					// Relative path
					pdfURL = pageURL + "/" + pdfURL
				}
			}
		} else {
			// Fallback to simple joining
			pdfURL = pageURL + pdfURL
		}
	}

	// Log the found PDF URL for debugging
	if pdfURL != "" {
		logger.Info("Found PDF URL:", pdfURL)
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
