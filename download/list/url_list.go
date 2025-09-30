package list

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/open-and-sustainable/alembica/utils/logger"

	"github.com/PuerkitoBio/goquery"
)

// PaperMetadata holds information about a paper to be downloaded
type PaperMetadata struct {
	ID             string   // Row index or identifier
	URL            string   // Best URL or resolved DOI
	DOI            string   // Digital Object Identifier
	Title          string   // Article title
	Authors        string   // Authors list
	Year           string   // Publication year
	Journal        string   // Source/Journal title
	Abstract       string   // Abstract text (for future use)
	Downloaded     bool     // Whether the file was successfully downloaded
	Filename       string   // Generated filename for the download
	ErrorMsg       string   // Error message if download failed
	OriginalRecord []string // Original CSV/TSV row data for preserving all columns
}

// ColumnMapping holds the detected column indices for relevant fields
type ColumnMapping struct {
	URL      int
	DOI      int
	Title    int
	Authors  int
	Year     int
	Journal  int
	Abstract int
}

// DownloadURLList processes a text, CSV, or TSV file containing URLs and attempts to download
// PDFs from each entry.
//
// This function supports three input formats:
// 1. Plain text file: One URL per line
// 2. CSV file: Comma-separated values with intelligent column detection
// 3. TSV file: Tab-separated values with intelligent column detection
//
// For CSV/TSV files, the function automatically detects columns for:
// - URLs (BestLink, BestURL, URL, Link, etc.)
// - DOIs (converts to URLs if no direct URL available)
// - Title, Authors, Year, Journal (used for intelligent file naming)
//
// The function will:
// - Parse the input file based on its extension (.csv, .tsv, .txt)
// - Generate meaningful filenames using paper metadata when available
// - Create a download report (_report.csv) for CSV/TSV inputs
// - Save all files to the same directory as the input file
//
// Parameters:
//   - path: The path to a file containing URLs or paper metadata
//
// Returns an error if the function fails to open or read the input file,
// but continues processing even if individual URLs fail to download.
func DownloadURLList(path string) error {
	// Extract the directory from the input file path
	dirPath := filepath.Dir(path)

	// Determine file type based on extension
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".csv", ".tsv":
		// Handle CSV/TSV files with metadata
		delimiter := ','
		if ext == ".tsv" {
			delimiter = '\t'
		}
		return processCSVFile(path, dirPath, delimiter)
	default:
		// Handle plain text files (original behavior)
		return processTextFile(path, dirPath)
	}
}

// findUniqueFilename checks if a file exists and returns a unique filename if needed
// It appends _1, _2, etc. to the base filename until a non-existing name is found
func findUniqueFilename(dirPath, filename string) string {
	fullPath := filepath.Join(dirPath, filename)

	// If file doesn't exist, return original filename
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return filename
	}

	// File exists, find a unique filename
	base := strings.TrimSuffix(filename, ".pdf")
	counter := 1
	for {
		newFilename := fmt.Sprintf("%s_%d.pdf", base, counter)
		fullPath = filepath.Join(dirPath, newFilename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			logger.Info("File already exists, using alternative name:", newFilename)
			return newFilename
		}
		counter++
		if counter > 1000 {
			// Safety limit with timestamp to guarantee uniqueness
			timestamp := time.Now().Unix()
			newFilename = fmt.Sprintf("%s_%d_%d.pdf", base, counter, timestamp)
			return newFilename
		}
	}
}

// processTextFile handles the original plain text URL list format
func processTextFile(path, dirPath string) error {
	// Open the file at the given path
	file, err := os.Open(path)
	if err != nil {
		return err
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
		return err
	}

	// Track failed URLs
	var failedURLs []string

	// Process each URL
	for _, url := range urls {
		pdfURL, filename, err := extractPDF(url)
		if err != nil {
			logger.Error("Error processing URL", url, ":", err)
			failedURLs = append(failedURLs, url)
			continue
		}

		if pdfURL != "" {
			// Ensure unique filename to prevent overwriting
			filename = findUniqueFilename(dirPath, filename)
			fullPath := filepath.Join(dirPath, filename)

			if err := downloadPDF(pdfURL, fullPath); err != nil {
				logger.Error("Download failed for", url, ":", err)
				failedURLs = append(failedURLs, url)
			} else {
				logger.Info("PDF downloaded successfully as", fullPath)
			}
		} else {
			logger.Info("No PDF found for", url)
			failedURLs = append(failedURLs, url)
		}
	}

	// Write failed URLs log if there are any failures
	if len(failedURLs) > 0 {
		failedLogPath := strings.TrimSuffix(path, filepath.Ext(path)) + "_failed.txt"
		if err := writeFailedURLsLog(failedURLs, failedLogPath); err != nil {
			logger.Error(fmt.Sprintf("Failed to write failed URLs log: %v", err))
		} else {
			logger.Info(fmt.Sprintf("Failed URLs log saved to: %s", failedLogPath))
		}
	}

	return nil
}

// processCSVFile handles CSV/TSV files with metadata
func processCSVFile(path, dirPath string, delimiter rune) error {
	// Parse the CSV/TSV file
	papers, headers, err := parseCSVFile(path, delimiter)
	if err != nil {
		return fmt.Errorf("failed to parse CSV/TSV file: %w", err)
	}

	logger.Info(fmt.Sprintf("Found %d entries to process", len(papers)))

	// Process each paper
	successCount := 0
	failCount := 0

	for i, paper := range papers {
		// Log progress every 10 papers
		if (i+1)%10 == 0 {
			logger.Info(fmt.Sprintf("Processing paper %d of %d...", i+1, len(papers)))
		}

		// Skip if no URL available
		if paper.URL == "" {
			logger.Info(fmt.Sprintf("Warning: Row %s: No URL available (Title: %s)", paper.ID, paper.Title))
			paper.ErrorMsg = "No URL available"
			failCount++
			continue
		}

		// Extract PDF URL from the page
		pdfURL, defaultFilename, err := extractPDF(paper.URL)
		if err != nil {
			logger.Error(fmt.Sprintf("Row %s: Error processing URL %s: %v", paper.ID, paper.URL, err))
			paper.ErrorMsg = fmt.Sprintf("Error: %v", err)
			failCount++
			continue
		}

		if pdfURL == "" {
			logger.Info(fmt.Sprintf("Warning: Row %s: No PDF found at %s", paper.ID, paper.URL))
			paper.ErrorMsg = "No PDF found"
			failCount++
			continue
		}

		// Generate intelligent filename
		filename := generateFilename(paper)
		if filename == "" {
			filename = defaultFilename
		}
		// Ensure unique filename to prevent overwriting
		filename = findUniqueFilename(dirPath, filename)
		paper.Filename = filename
		fullPath := filepath.Join(dirPath, filename)

		// Download the PDF
		if err := downloadPDF(pdfURL, fullPath); err != nil {
			logger.Error(fmt.Sprintf("Row %s: Download failed for %s: %v", paper.ID, paper.URL, err))
			paper.ErrorMsg = fmt.Sprintf("Download failed: %v", err)
			failCount++
		} else {
			logger.Info(fmt.Sprintf("Row %s: PDF downloaded successfully as %s", paper.ID, filename))
			paper.Downloaded = true
			successCount++
		}
	}

	// Generate download report (original format for backward compatibility)
	reportPath := strings.TrimSuffix(path, filepath.Ext(path)) + "_report.csv"
	if err := writeDownloadReport(papers, reportPath); err != nil {
		logger.Error(fmt.Sprintf("Failed to write download report: %v", err))
	} else {
		logger.Info(fmt.Sprintf("Download report saved to: %s", reportPath))
	}

	// Generate enhanced CSV/TSV with original columns + downloaded status
	ext := filepath.Ext(path)
	enhancedPath := strings.TrimSuffix(path, ext) + "_with_status" + ext
	if err := writeEnhancedCSV(papers, headers, enhancedPath, delimiter); err != nil {
		logger.Error(fmt.Sprintf("Failed to write enhanced CSV: %v", err))
	} else {
		logger.Info(fmt.Sprintf("Enhanced CSV with download status saved to: %s", enhancedPath))
	}

	// Log summary
	logger.Info(fmt.Sprintf("Download complete: %d successful, %d failed out of %d total",
		successCount, failCount, len(papers)))

	return nil
}

// parseCSVFile reads a CSV/TSV file and extracts paper metadata with intelligent column detection
func parseCSVFile(filepath string, delimiter rune) ([]*PaperMetadata, []string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read header row: %w", err)
	}

	// Detect column mappings
	mapping := detectColumns(headers)
	if mapping.URL == -1 && mapping.DOI == -1 {
		return nil, nil, fmt.Errorf("no URL or DOI column found in CSV/TSV file")
	}

	// Log detected columns for debugging
	logDetectedColumns(headers, mapping)

	// Parse data rows
	var papers []*PaperMetadata
	rowIndex := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Info(fmt.Sprintf("Warning: Error reading row %d: %v", rowIndex, err))
			continue
		}

		paper := extractPaperMetadata(record, mapping, rowIndex)
		// Store the original record for later export
		paper.OriginalRecord = make([]string, len(record))
		copy(paper.OriginalRecord, record)
		// Include all papers, even those without URLs/DOIs, so we can report them
		papers = append(papers, paper)
		rowIndex++
	}

	return papers, headers, nil
}

// detectColumns intelligently identifies column indices based on header names
func detectColumns(headers []string) ColumnMapping {
	mapping := ColumnMapping{
		URL:      -1,
		DOI:      -1,
		Title:    -1,
		Authors:  -1,
		Year:     -1,
		Journal:  -1,
		Abstract: -1,
	}

	for i, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))

		// URL detection - prioritize "best" URLs
		if strings.Contains(headerLower, "bestlink") ||
			strings.Contains(headerLower, "besturl") ||
			strings.Contains(headerLower, "best_link") ||
			strings.Contains(headerLower, "best_url") {
			mapping.URL = i
		} else if mapping.URL == -1 && (strings.Contains(headerLower, "url") ||
			strings.Contains(headerLower, "link") ||
			strings.Contains(headerLower, "href")) {
			mapping.URL = i
		}

		// DOI detection
		if mapping.DOI == -1 && (strings.Contains(headerLower, "doi") ||
			strings.Contains(headerLower, "digital object identifier")) {
			mapping.DOI = i
		}

		// Title detection
		if mapping.Title == -1 && strings.Contains(headerLower, "title") &&
			!strings.Contains(headerLower, "source") &&
			!strings.Contains(headerLower, "journal") {
			mapping.Title = i
		}

		// Authors detection
		if mapping.Authors == -1 && (strings.Contains(headerLower, "author") ||
			strings.Contains(headerLower, "creator") ||
			strings.Contains(headerLower, "contributor")) {
			mapping.Authors = i
		}

		// Year detection (before Journal to avoid "publicationyear" matching "publication")
		if mapping.Year == -1 && (strings.Contains(headerLower, "year") ||
			strings.Contains(headerLower, "date") && strings.Contains(headerLower, "publ")) {
			mapping.Year = i
		}

		// Journal/Source detection
		if mapping.Journal == -1 && (strings.Contains(headerLower, "journal") ||
			strings.Contains(headerLower, "source") ||
			(strings.Contains(headerLower, "publication") && !strings.Contains(headerLower, "year")) ||
			strings.Contains(headerLower, "venue")) {
			mapping.Journal = i
		}

		// Abstract detection
		if mapping.Abstract == -1 && strings.Contains(headerLower, "abstract") {
			mapping.Abstract = i
		}
	}

	return mapping
}

// extractPaperMetadata creates a PaperMetadata struct from a CSV record
func extractPaperMetadata(record []string, mapping ColumnMapping, rowIndex int) *PaperMetadata {
	paper := &PaperMetadata{
		ID: fmt.Sprintf("%d", rowIndex),
	}

	// Safely extract values with bounds checking
	if mapping.URL >= 0 && mapping.URL < len(record) {
		paper.URL = strings.TrimSpace(record[mapping.URL])
	}

	if mapping.DOI >= 0 && mapping.DOI < len(record) {
		paper.DOI = strings.TrimSpace(record[mapping.DOI])
	}

	if mapping.Title >= 0 && mapping.Title < len(record) {
		paper.Title = strings.TrimSpace(record[mapping.Title])
	}

	if mapping.Authors >= 0 && mapping.Authors < len(record) {
		paper.Authors = strings.TrimSpace(record[mapping.Authors])
	}

	if mapping.Year >= 0 && mapping.Year < len(record) {
		paper.Year = strings.TrimSpace(record[mapping.Year])
	}

	if mapping.Journal >= 0 && mapping.Journal < len(record) {
		paper.Journal = strings.TrimSpace(record[mapping.Journal])
	}

	if mapping.Abstract >= 0 && mapping.Abstract < len(record) {
		paper.Abstract = strings.TrimSpace(record[mapping.Abstract])
	}

	// Check if URL is problematic (requires JavaScript/browser or API token)
	if paper.URL != "" && isProblematicURL(paper.URL) {
		logger.Info(fmt.Sprintf("Row %s: Detected problematic URL (requires browser/API): %s", paper.ID, paper.URL))

		// First, try to use the DOI if available
		if paper.DOI != "" {
			logger.Info(fmt.Sprintf("Row %s: Using DOI instead: %s", paper.ID, paper.DOI))
			paper.URL = convertDOIToURL(paper.DOI)
		} else {
			// No DOI available, try to find one via Crossref
			logger.Info(fmt.Sprintf("Row %s: No DOI available, searching Crossref...", paper.ID))
			doi := searchCrossrefForDOI(paper.Title, paper.Authors, paper.Year)
			if doi != "" {
				logger.Info(fmt.Sprintf("Row %s: Found DOI via Crossref: %s", paper.ID, doi))
				paper.DOI = doi
				paper.URL = convertDOIToURL(doi)
			} else {
				logger.Info(fmt.Sprintf("Row %s: Could not find DOI via Crossref, will attempt original URL", paper.ID))
			}
		}
	}

	// If no URL but DOI exists, convert DOI to URL
	if paper.URL == "" && paper.DOI != "" {
		paper.URL = convertDOIToURL(paper.DOI)
	}

	return paper
}

// convertDOIToURL converts a DOI to a resolvable URL
func convertDOIToURL(doi string) string {
	// Clean the DOI
	doi = strings.TrimSpace(doi)

	// Remove common prefixes if present
	if strings.HasPrefix(strings.ToLower(doi), "doi:") {
		doi = strings.TrimSpace(doi[4:])
	}
	if strings.HasPrefix(strings.ToLower(doi), "http://dx.doi.org/") {
		return doi // Already a URL
	}
	if strings.HasPrefix(strings.ToLower(doi), "https://dx.doi.org/") {
		return doi // Already a URL
	}
	if strings.HasPrefix(strings.ToLower(doi), "http://doi.org/") {
		return doi // Already a URL
	}
	if strings.HasPrefix(strings.ToLower(doi), "https://doi.org/") {
		return doi // Already a URL
	}

	// Convert to URL if it's a plain DOI
	if doi != "" && strings.Contains(doi, "10.") {
		return "https://doi.org/" + doi
	}

	return ""
}

// generateFilename creates an intelligent filename based on paper metadata
func generateFilename(paper *PaperMetadata) string {
	var parts []string

	// Add year if available
	if paper.Year != "" {
		// Extract just the year if it's a full date
		year := paper.Year
		if len(year) > 4 {
			// Try to extract 4-digit year
			for i := 0; i <= len(year)-4; i++ {
				substr := year[i : i+4]
				if isYear(substr) {
					year = substr
					break
				}
			}
		}
		parts = append(parts, year)
	}

	// Add first author's last name if available
	if paper.Authors != "" {
		firstAuthor := extractFirstAuthor(paper.Authors)
		if firstAuthor != "" {
			parts = append(parts, firstAuthor)
		}
	}

	// Add truncated title if available
	if paper.Title != "" {
		titlePart := truncateTitle(paper.Title, 50)
		if titlePart != "" {
			parts = append(parts, titlePart)
		}
	}

	// If we have parts, join them
	if len(parts) > 0 {
		filename := strings.Join(parts, "_")
		return sanitizeFilename(filename) + ".pdf"
	}

	// Fallback to row ID
	return fmt.Sprintf("paper_%s.pdf", paper.ID)
}

// isYear checks if a string is a valid year (1900-2099)
func isYear(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	// Check if it's a reasonable year (1900-2099)
	if s >= "1900" && s <= "2099" {
		return true
	}
	return false
}

// extractFirstAuthor attempts to extract the first author's last name
func extractFirstAuthor(authors string) string {
	// Handle various separators - semicolon takes precedence
	separators := []string{";", " and ", " & ", "|"}

	firstAuthor := authors
	for _, sep := range separators {
		if idx := strings.Index(authors, sep); idx > 0 {
			firstAuthor = authors[:idx]
			break
		}
	}

	firstAuthor = strings.TrimSpace(firstAuthor)

	// Try to extract last name (assume "Last, First" or "First Last" format)
	if strings.Contains(firstAuthor, ",") {
		// "Last, First" format
		parts := strings.Split(firstAuthor, ",")
		return sanitizeForFilename(strings.TrimSpace(parts[0]))
	} else {
		// "First Last" format - take the last word
		words := strings.Fields(firstAuthor)
		if len(words) > 0 {
			return sanitizeForFilename(words[len(words)-1])
		}
	}

	return sanitizeForFilename(firstAuthor)
}

// truncateTitle creates a filename-friendly truncated version of the title
func truncateTitle(title string, maxLen int) string {
	// Remove special characters and normalize spaces
	title = strings.TrimSpace(title)

	// Remove common stop words at the beginning
	stopWords := []string{"The ", "A ", "An "}
	for _, word := range stopWords {
		if strings.HasPrefix(title, word) {
			title = title[len(word):]
			break
		}
	}

	// Take first few words up to maxLen characters
	words := strings.Fields(title)
	result := ""
	for _, word := range words {
		word = sanitizeForFilename(word)
		if len(result)+len(word)+1 > maxLen {
			break
		}
		if result != "" {
			result += "_"
		}
		result += word
	}

	return result
}

// sanitizeForFilename removes characters that are problematic in filenames
func sanitizeForFilename(s string) string {
	// Remove or replace problematic characters
	replacements := map[string]string{
		"/":  "-",
		"\\": "-",
		":":  "",
		"*":  "",
		"?":  "",
		"\"": "",
		"<":  "",
		">":  "",
		"|":  "",
		".":  "",
		",":  "",
		";":  "",
		"'":  "",
		"&":  "and",
		"#":  "",
		"%":  "",
		"{":  "",
		"}":  "",
		"[":  "",
		"]":  "",
		"(":  "",
		")":  "",
		"!":  "",
		"@":  "",
		"$":  "",
		"^":  "",
		"~":  "",
		"`":  "",
		"+":  "",
		"=":  "",
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Replace multiple spaces/underscores with single underscore
	result = strings.ReplaceAll(result, "  ", " ")
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, "__", "_")

	// Trim underscores from ends
	result = strings.Trim(result, "_")

	return result
}

// logDetectedColumns logs the detected column mappings for debugging
func logDetectedColumns(headers []string, mapping ColumnMapping) {
	logger.Info("Detected CSV/TSV columns:")

	if mapping.URL >= 0 {
		logger.Info(fmt.Sprintf("  URL column: %d (%s)", mapping.URL, headers[mapping.URL]))
	}
	if mapping.DOI >= 0 {
		logger.Info(fmt.Sprintf("  DOI column: %d (%s)", mapping.DOI, headers[mapping.DOI]))
	}
	if mapping.Title >= 0 {
		logger.Info(fmt.Sprintf("  Title column: %d (%s)", mapping.Title, headers[mapping.Title]))
	}
	if mapping.Authors >= 0 {
		logger.Info(fmt.Sprintf("  Authors column: %d (%s)", mapping.Authors, headers[mapping.Authors]))
	}
	if mapping.Year >= 0 {
		logger.Info(fmt.Sprintf("  Year column: %d (%s)", mapping.Year, headers[mapping.Year]))
	}
	if mapping.Journal >= 0 {
		logger.Info(fmt.Sprintf("  Journal column: %d (%s)", mapping.Journal, headers[mapping.Journal]))
	}
}

// writeDownloadReport generates a CSV report of download results
func writeDownloadReport(papers []*PaperMetadata, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	headers := []string{"ID", "Title", "Authors", "Year", "Journal", "URL", "DOI", "Downloaded", "Filename", "Error"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for _, paper := range papers {
		record := []string{
			paper.ID,
			paper.Title,
			paper.Authors,
			paper.Year,
			paper.Journal,
			paper.URL,
			paper.DOI,
			fmt.Sprintf("%t", paper.Downloaded),
			paper.Filename,
			paper.ErrorMsg,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// writeFailedURLsLog writes a list of failed URLs to a text file
func writeFailedURLsLog(failedURLs []string, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create failed URLs log: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header comment
	if _, err := writer.WriteString("# Failed Downloads - URLs that could not be retrieved\n"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := writer.WriteString("# One URL per line\n\n"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write each failed URL
	for _, url := range failedURLs {
		if _, err := writer.WriteString(url + "\n"); err != nil {
			return fmt.Errorf("failed to write URL: %w", err)
		}
	}

	return nil
}

// writeEnhancedCSV writes a CSV/TSV file with all original columns plus a 'downloaded' column
func writeEnhancedCSV(papers []*PaperMetadata, headers []string, outputPath string, delimiter rune) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create enhanced CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = delimiter
	defer writer.Flush()

	// Write header with additional 'downloaded' column
	enhancedHeaders := append(headers, "downloaded")
	if err := writer.Write(enhancedHeaders); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows with download status
	for _, paper := range papers {
		// Start with the original record
		record := make([]string, len(paper.OriginalRecord))
		copy(record, paper.OriginalRecord)

		// Append download status
		downloadedStatus := "false"
		if paper.Downloaded {
			downloadedStatus = "true"
		}
		record = append(record, downloadedStatus)

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// findPDFLink searches a webpage for a link to a PDF file using multiple detection strategies.
//
// This function employs a hierarchical approach with 5 different strategies to locate PDF links on
// academic publisher websites. The strategies are tried in order of reliability:
//
//  1. Metadata examination: Searches for citation_pdf_url or fulltext_pdf_url in meta tags, which is
//     the most reliable method for academic papers.
//
// 2. Publisher-specific patterns: Applies specialized selectors for common academic publishers:
//   - MDPI (doi.org/10.3390)
//   - ScienceDirect (doi.org/10.1016)
//   - Springer (doi.org/10.1007)
//   - IEEE (doi.org/10.1109)
//   - Wiley (doi.org/10.1002)
//
// 3. File extension detection: Looks for links that end with ".pdf"
//
//  4. Text content analysis: Identifies links containing PDF-related text while filtering out
//     false positives like "cover" or "sample"
//
// 5. CSS attributes examination: Searches for elements with download attributes or PDF-related classes
//
// Parameters:
//   - doc: A goquery Document containing the parsed HTML of the page
//   - pageURL: The URL of the webpage being analyzed, used for publisher detection
//
// Returns:
//   - A string containing the URL to the PDF if found, or an empty string if no PDF link is detected
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

// extractPDF attempts to find a PDF link on a webpage and determine an appropriate filename.
//
// Given a URL to a webpage, this function:
// 1. Checks if the URL itself is a direct PDF
// 2. If not, parses the HTML and searches for PDF links using various strategies
// 3. Generates a filename based on either the page title or URL
// 4. Converts relative PDF URLs to absolute URLs
//
// Parameters:
//   - pageURL: The URL of the webpage to check for PDF links
//
// Returns:
//   - pdfURL: The URL of the found PDF, or an empty string if none was found
//   - filename: A sanitized filename to use when saving the PDF
//   - err: An error if the HTTP request or HTML parsing fails, nil otherwise
func extractPDF(pageURL string) (pdfURL string, filename string, err error) {
	// Check if URL looks like a direct PDF link
	urlLower := strings.ToLower(pageURL)
	if strings.Contains(urlLower, ".pdf") ||
		strings.Contains(urlLower, "/pdf") ||
		strings.Contains(urlLower, "pdfdirect") ||
		strings.Contains(urlLower, "download/pdf") {
		// This looks like a direct PDF link, use it directly
		// Extract filename from URL if possible
		parts := strings.Split(pageURL, "/")
		lastPart := parts[len(parts)-1]
		if strings.HasSuffix(strings.ToLower(lastPart), ".pdf") {
			filename = sanitizeFilename(lastPart)
		} else {
			// Generate filename from URL
			filename = sanitizeFilename(lastPart)
			if !strings.HasSuffix(filename, ".pdf") {
				filename = filename + ".pdf"
			}
		}
		return pageURL, filename, nil
	}

	// Fetch the page to check content type with proper User-Agent
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PrismAID/1.0; +https://github.com/open-and-sustainable/prismaid)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/pdf") ||
		strings.Contains(contentType, "application/octet-stream") {
		// Direct PDF or binary stream (often used for PDFs)
		// Get filename from Content-Disposition header if available
		contentDisp := resp.Header.Get("Content-Disposition")
		if contentDisp != "" && strings.Contains(contentDisp, "filename=") {
			start := strings.Index(contentDisp, "filename=") + 9
			end := strings.IndexByte(contentDisp[start:], ';')
			if end == -1 {
				filename = sanitizeFilename(strings.Trim(contentDisp[start:], `"`))
			} else {
				filename = sanitizeFilename(strings.Trim(contentDisp[start:start+end], `"`))
			}
		} else {
			// Generate filename from URL
			parts := strings.Split(pageURL, "/")
			lastPart := parts[len(parts)-1]
			filename = sanitizeFilename(lastPart)
			if !strings.HasSuffix(filename, ".pdf") {
				filename = filename + ".pdf"
			}
		}
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

	// If no PDF link found, try to extract DOI from the page and resolve it
	if pdfURL == "" {
		// Look for DOI in meta tags
		doc.Find("meta").Each(func(i int, s *goquery.Selection) {
			if pdfURL != "" {
				return // Already found
			}

			name, _ := s.Attr("name")
			property, _ := s.Attr("property")
			content, _ := s.Attr("content")

			// Check for DOI in meta tags
			if (strings.Contains(strings.ToLower(name), "citation_doi") ||
				strings.Contains(strings.ToLower(property), "citation_doi") ||
				strings.Contains(strings.ToLower(name), "dc.identifier") ||
				strings.Contains(strings.ToLower(property), "dc.identifier")) && content != "" {
				// Found a DOI, try to resolve it
				if strings.Contains(content, "10.") {
					doi := content
					if strings.HasPrefix(strings.ToLower(doi), "doi:") {
						doi = strings.TrimSpace(doi[4:])
					}
					logger.Info("Found DOI in meta tags:", doi, "- attempting to resolve")
					// Recursively call extractPDF with the DOI URL
					doiURL := "https://doi.org/" + doi
					resolvedURL, resolvedFilename, err := extractPDF(doiURL)
					if err == nil && resolvedURL != "" {
						pdfURL = resolvedURL
						if resolvedFilename != "" {
							filename = resolvedFilename
						}
					}
				}
			}
		})

		// If still no PDF, look for DOI in the page body
		if pdfURL == "" {
			// Look for DOI patterns in various elements
			doc.Find("div, span, p, a").Each(func(i int, s *goquery.Selection) {
				if pdfURL != "" {
					return // Already found
				}

				text := s.Text()
				// Look for DOI patterns
				if strings.Contains(text, "10.") && (strings.Contains(strings.ToLower(text), "doi") || strings.Contains(text, "/")) {
					// Extract DOI using regex
					doiPattern := regexp.MustCompile(`(10\.\d{4,}(?:\.\d+)*\/[-._;()\/:a-zA-Z0-9]+)`)
					matches := doiPattern.FindStringSubmatch(text)
					if len(matches) > 0 {
						doi := matches[0]
						// Clean up the DOI
						doi = strings.TrimSuffix(doi, ".")
						doi = strings.TrimSuffix(doi, ",")
						doi = strings.TrimSuffix(doi, ";")
						doi = strings.TrimSuffix(doi, ")")

						logger.Info("Found DOI in page content:", doi, "- attempting to resolve")
						doiURL := "https://doi.org/" + doi
						resolvedURL, resolvedFilename, err := extractPDF(doiURL)
						if err == nil && resolvedURL != "" {
							pdfURL = resolvedURL
							if resolvedFilename != "" {
								filename = resolvedFilename
							}
						}
					}
				}
			})
		}
	}

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

// sanitizeFilename converts a string into a valid filename by replacing invalid
// characters with underscores. This includes replacing path separators, special
// characters, and whitespace that might cause issues in various filesystems.
//
// Parameters:
//   - name: The original string to be sanitized
//
// Returns:
//   - A sanitized string that can be safely used as a filename
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

// downloadPDF downloads a PDF file from the given URL and saves it to the specified path.
// It handles the HTTP request, creates the output file, and copies the content.
//
// Parameters:
//   - pdfURL: The URL of the PDF file to download
//   - fullPath: The full filesystem path where the PDF should be saved
//
// Returns:
//   - error: nil if successful, otherwise an error describing what went wrong
func downloadPDF(pdfURL, fullPath string) error {
	// Create a client that follows redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Get(pdfURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if we got a successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Verify we got a PDF or binary content
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" &&
		!strings.Contains(contentType, "pdf") &&
		!strings.Contains(contentType, "octet-stream") &&
		!strings.Contains(contentType, "binary") {
		// Log warning but still try to download
		logger.Info(fmt.Sprintf("Warning: Unexpected content type %s for %s", contentType, pdfURL))
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isProblematicURL checks if a URL is from a platform that requires JavaScript/browser rendering
// or API tokens to access content (Dimensions.app, ResearchGate, Academia.edu, Semantic Scholar web UI)
func isProblematicURL(urlStr string) bool {
	urlLower := strings.ToLower(urlStr)

	problematicDomains := []string{
		"dimensions.ai",
		"app.dimensions.ai",
		"researchgate.net",
		"www.researchgate.net",
		"academia.edu",
		"www.academia.edu",
		"semanticscholar.org/paper/",
		"www.semanticscholar.org/paper/",
	}

	for _, domain := range problematicDomains {
		if strings.Contains(urlLower, domain) {
			return true
		}
	}

	return false
}

// searchCrossrefForDOI queries the Crossref API to find a DOI based on paper metadata
// Returns empty string if no DOI is found
func searchCrossrefForDOI(title, authors, year string) string {
	// Need at least a title to search
	if title == "" {
		return ""
	}

	// Build the query URL for Crossref API
	// API documentation: https://github.com/CrossRef/rest-api-doc
	baseURL := "https://api.crossref.org/works"

	// Construct query parameters
	params := url.Values{}

	// Use bibliographic query which searches across multiple fields
	queryParts := []string{}
	if title != "" {
		queryParts = append(queryParts, title)
	}
	if authors != "" {
		queryParts = append(queryParts, authors)
	}

	query := strings.Join(queryParts, " ")
	params.Add("query.bibliographic", query)

	// Add rows limit (we only need the top result)
	params.Add("rows", "1")

	// Construct full URL
	searchURL := baseURL + "?" + params.Encode()

	// Make the request with proper User-Agent
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		logger.Info(fmt.Sprintf("Error creating Crossref request: %v", err))
		return ""
	}

	// Crossref requests a polite User-Agent
	req.Header.Set("User-Agent", "PrismAID/1.0 (https://github.com/open-and-sustainable/prismaid; mailto:info@example.com)")

	resp, err := client.Do(req)
	if err != nil {
		logger.Info(fmt.Sprintf("Error querying Crossref API: %v", err))
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Info(fmt.Sprintf("Crossref API returned status: %d", resp.StatusCode))
		return ""
	}

	// Parse the JSON response
	var result struct {
		Message struct {
			Items []struct {
				DOI   string   `json:"DOI"`
				Title []string `json:"title"`
				Score float64  `json:"score"`
			} `json:"items"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Info(fmt.Sprintf("Error parsing Crossref response: %v", err))
		return ""
	}

	// Check if we got any results
	if len(result.Message.Items) == 0 {
		logger.Info("No results found in Crossref")
		return ""
	}

	// Get the first (best) result
	item := result.Message.Items[0]

	// Log the match score for debugging
	logger.Info(fmt.Sprintf("Crossref match score: %.2f, Title: %v", item.Score, item.Title))

	// Only accept results with a reasonable score (Crossref uses 0-100 scale typically)
	// A score above 50 usually indicates a good match
	if item.Score < 50 {
		logger.Info(fmt.Sprintf("Crossref match score too low (%.2f), rejecting", item.Score))
		return ""
	}

	return item.DOI
}

// handleDimensionsURL attempts to handle Dimensions.ai publication URLs
// Since Dimensions.ai is a JavaScript-heavy SPA, we try alternative approaches:
// 1. Extract publication ID and try to find the DOI through alternative APIs
// 2. Try to construct a direct export URL
// 3. Return an informative error if we can't resolve it
func handleDimensionsURL(pageURL string) (pdfURL string, filename string, err error) {
	// Extract publication ID from URL
	// URL format: https://app.dimensions.ai/details/publication/pub.XXXXXXXXX
	parts := strings.Split(pageURL, "/")
	if len(parts) == 0 {
		return "", "", fmt.Errorf("invalid Dimensions.ai URL format")
	}

	pubID := parts[len(parts)-1]
	if !strings.HasPrefix(pubID, "pub.") {
		return "", "", fmt.Errorf("invalid Dimensions.ai publication ID: %s", pubID)
	}

	// Try different approaches to get the actual paper

	// Approach 1: Try Dimensions badge API which sometimes has metadata
	badgeURL := fmt.Sprintf("https://badge.dimensions.ai/details/id/%s.json", pubID)
	resp, err := http.Get(badgeURL)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			// Look for DOI in the response
			bodyStr := string(body)
			if doi := extractDOIFromText(bodyStr); doi != "" {
				// Found a DOI, resolve it
				doiURL := "https://doi.org/" + doi
				logger.Info(fmt.Sprintf("Found DOI %s for Dimensions publication %s, resolving...", doi, pubID))
				return extractPDF(doiURL)
			}
		}
	}

	// Approach 2: Try the Dimensions details API
	detailsURL := fmt.Sprintf("https://app.dimensions.ai/api/publications/%s", pubID)
	resp2, err := http.Get(detailsURL)
	if err == nil && resp2.StatusCode == 200 {
		defer resp2.Body.Close()
		body, err := io.ReadAll(resp2.Body)
		if err == nil {
			bodyStr := string(body)
			if doi := extractDOIFromText(bodyStr); doi != "" {
				doiURL := "https://doi.org/" + doi
				logger.Info(fmt.Sprintf("Found DOI %s for Dimensions publication %s, resolving...", doi, pubID))
				return extractPDF(doiURL)
			}
		}
	}

	// If we can't resolve it, return an informative error
	logger.Info(fmt.Sprintf("Unable to resolve Dimensions.ai publication %s. This may require manual download or institutional access.", pubID))
	return "", "", fmt.Errorf("Dimensions.ai URLs require JavaScript rendering or API access. Please provide the DOI or direct publisher URL instead")
}

// extractDOIFromText attempts to extract a DOI from a text string
func extractDOIFromText(text string) string {
	// Common DOI pattern
	doiPattern := regexp.MustCompile(`10\.\d{4,}(?:\.\d+)*\/[-._;()\/:a-zA-Z0-9]+`)
	matches := doiPattern.FindStringSubmatch(text)
	if len(matches) > 0 {
		doi := matches[0]
		// Clean up the DOI
		doi = strings.TrimSuffix(doi, ".")
		doi = strings.TrimSuffix(doi, ",")
		doi = strings.TrimSuffix(doi, ";")
		doi = strings.TrimSuffix(doi, ")")
		return doi
	}

	// Also check for DOI with prefix
	if strings.Contains(text, "\"doi\":\"") {
		start := strings.Index(text, "\"doi\":\"") + 7
		end := strings.Index(text[start:], "\"")
		if end > 0 {
			return text[start : start+end]
		}
	}

	return ""
}

// extractPDFDirectly is a simplified version of extractPDF that doesn't use JavaScript rendering
// to avoid infinite recursion when resolving DOIs found in JavaScript pages
func extractPDFDirectly(pageURL string) (pdfURL string, filename string, err error) {
	// Check if URL looks like a direct PDF link
	urlLower := strings.ToLower(pageURL)
	if strings.Contains(urlLower, ".pdf") ||
		strings.Contains(urlLower, "/pdf") ||
		strings.Contains(urlLower, "pdfdirect") ||
		strings.Contains(urlLower, "download/pdf") {
		parts := strings.Split(pageURL, "/")
		lastPart := parts[len(parts)-1]
		if strings.HasSuffix(strings.ToLower(lastPart), ".pdf") {
			filename = sanitizeFilename(lastPart)
		} else {
			filename = sanitizeFilename(lastPart)
			if !strings.HasSuffix(filename, ".pdf") {
				filename = filename + ".pdf"
			}
		}
		return pageURL, filename, nil
	}

	// Fetch the page
	// Fetch the page to check content type
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PrismAID/1.0; +https://github.com/open-and-sustainable/prismaid)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/pdf") ||
		strings.Contains(contentType, "application/octet-stream") {
		contentDisp := resp.Header.Get("Content-Disposition")
		if contentDisp != "" && strings.Contains(contentDisp, "filename=") {
			start := strings.Index(contentDisp, "filename=") + 9
			end := strings.IndexByte(contentDisp[start:], ';')
			if end == -1 {
				filename = sanitizeFilename(strings.Trim(contentDisp[start:], `"`))
			} else {
				filename = sanitizeFilename(strings.Trim(contentDisp[start:start+end], `"`))
			}
		} else {
			parts := strings.Split(pageURL, "/")
			lastPart := parts[len(parts)-1]
			filename = sanitizeFilename(lastPart)
			if !strings.HasSuffix(filename, ".pdf") {
				filename = filename + ".pdf"
			}
		}
		return pageURL, filename, nil
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Extract title for naming
	title := doc.Find("title").Text()
	if title == "" {
		title = "downloaded"
	}
	filename = sanitizeFilename(title) + ".pdf"

	// Use our enhanced function to find PDF links
	pdfURL = findPDFLink(doc, pageURL)

	// Make sure it's an absolute URL
	if pdfURL != "" && !strings.HasPrefix(pdfURL, "http") {
		base, err := url.Parse(pageURL)
		if err == nil {
			relative, err := url.Parse(pdfURL)
			if err == nil {
				pdfURL = base.ResolveReference(relative).String()
			}
		}
	}

	return pdfURL, filename, nil
}
