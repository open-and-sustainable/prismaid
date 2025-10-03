package list

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-and-sustainable/alembica/utils/logger"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/semaphore"
)

// Global HTTP client with optimized connection pooling and HTTP/2 support
var httpClient *http.Client

// ConcurrentDownloader manages concurrent downloads with per-host and global limits
type ConcurrentDownloader struct {
	globalSem  *semaphore.Weighted
	hostSems   map[string]*semaphore.Weighted
	hostMutex  sync.RWMutex
	maxPerHost int64
}

// DownloadTask represents a download job
type DownloadTask struct {
	URL         string
	PDFUrl      string
	Filename    string
	FullPath    string
	Paper       *PaperMetadata
	OriginalURL string
}

// DownloadResult represents the result of a download attempt
type DownloadResult struct {
	Task    *DownloadTask
	Success bool
	Error   error
}

// UnpaywallResponse represents the response from Unpaywall API
type UnpaywallResponse struct {
	DOI            string `json:"doi"`
	IsOA           bool   `json:"is_oa"`
	BestOALocation struct {
		HostType  string `json:"host_type"`
		URLForPDF string `json:"url_for_pdf"`
	} `json:"best_oa_location"`
	OALocations []struct {
		HostType  string `json:"host_type"`
		URLForPDF string `json:"url_for_pdf"`
	} `json:"oa_locations"`
}

// RetryConfig holds retry policy configuration
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Jitter     bool
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Jitter:     true,
	}
}

func init() {
	// Configure transport with connection pooling optimizations
	transport := &http.Transport{
		MaxIdleConns:          100,              // Maximum idle connections across all hosts
		MaxIdleConnsPerHost:   10,               // Maximum idle connections per host
		IdleConnTimeout:       90 * time.Second, // How long idle connections are kept alive
		TLSHandshakeTimeout:   10 * time.Second, // TLS handshake timeout
		ResponseHeaderTimeout: 30 * time.Second, // Response header timeout
		DisableCompression:    false,            // Enable gzip compression
		ForceAttemptHTTP2:     true,             // Force HTTP/2 where supported
	}

	// Create global HTTP client with optimized settings
	httpClient = &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second, // Overall request timeout
	}

	logger.Info("HTTP client initialized with connection pooling and HTTP/2 support")
}

// NewConcurrentDownloader creates a new concurrent downloader with specified limits
func NewConcurrentDownloader(maxGlobal, maxPerHost int64) *ConcurrentDownloader {
	return &ConcurrentDownloader{
		globalSem:  semaphore.NewWeighted(maxGlobal),
		hostSems:   make(map[string]*semaphore.Weighted),
		maxPerHost: maxPerHost,
	}
}

// getHostSemaphore returns a semaphore for the given host, creating it if necessary
func (cd *ConcurrentDownloader) getHostSemaphore(host string) *semaphore.Weighted {
	cd.hostMutex.RLock()
	sem, exists := cd.hostSems[host]
	cd.hostMutex.RUnlock()

	if exists {
		return sem
	}

	cd.hostMutex.Lock()
	defer cd.hostMutex.Unlock()

	// Double-check pattern in case another goroutine created it
	if sem, exists := cd.hostSems[host]; exists {
		return sem
	}

	sem = semaphore.NewWeighted(cd.maxPerHost)
	cd.hostSems[host] = sem
	return sem
}

// downloadConcurrently performs concurrent downloads with proper semaphore management
func (cd *ConcurrentDownloader) downloadConcurrently(tasks []*DownloadTask) []DownloadResult {
	ctx := context.Background()
	results := make([]DownloadResult, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t *DownloadTask) {
			defer wg.Done()

			// Parse URL to get host for per-host limiting
			parsedURL, err := url.Parse(t.PDFUrl)
			if err != nil {
				results[idx] = DownloadResult{Task: t, Success: false, Error: err}
				return
			}
			host := parsedURL.Host

			// Acquire global semaphore
			if err := cd.globalSem.Acquire(ctx, 1); err != nil {
				results[idx] = DownloadResult{Task: t, Success: false, Error: err}
				return
			}
			defer cd.globalSem.Release(1)

			// Acquire host-specific semaphore
			hostSem := cd.getHostSemaphore(host)
			if err := hostSem.Acquire(ctx, 1); err != nil {
				results[idx] = DownloadResult{Task: t, Success: false, Error: err}
				return
			}
			defer hostSem.Release(1)

			// Perform the actual download with retry
			err = downloadPDFWithRetry(t.PDFUrl, t.FullPath, DefaultRetryConfig())

			// If download failed, try Unpaywall as a last resort
			if err != nil {
				logger.Info(fmt.Sprintf("Primary download failed for %s, trying Unpaywall fallback", t.OriginalURL))
				if fallbackErr := tryUnpaywallFallback(t); fallbackErr == nil {
					err = downloadPDFWithRetry(t.PDFUrl, t.FullPath, DefaultRetryConfig())
				}
			}

			results[idx] = DownloadResult{Task: t, Success: err == nil, Error: err}
		}(i, task)
	}

	wg.Wait()
	return results
}

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

	// Create concurrent downloader with reasonable limits
	downloader := NewConcurrentDownloader(25, 4) // 25 global, 4 per host

	// Track failed URLs
	var failedURLs []string
	var tasks []*DownloadTask

	// Prepare download tasks
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

			task := &DownloadTask{
				URL:         url,
				PDFUrl:      pdfURL,
				Filename:    filename,
				FullPath:    fullPath,
				OriginalURL: url,
			}
			tasks = append(tasks, task)
		} else {
			logger.Info("No PDF found for", url)
			failedURLs = append(failedURLs, url)
		}
	}

	// Perform concurrent downloads
	if len(tasks) > 0 {
		logger.Info(fmt.Sprintf("Starting concurrent download of %d PDFs (max 25 global, 4 per host)", len(tasks)))
		results := downloader.downloadConcurrently(tasks)

		// Process results
		for _, result := range results {
			if result.Success {
				logger.Info("PDF downloaded successfully as", result.Task.FullPath)
			} else {
				logger.Error("Download failed for", result.Task.OriginalURL, ":", result.Error)
				failedURLs = append(failedURLs, result.Task.OriginalURL)
			}
		}
	}

	// Generate download results file for plain text URL lists
	downloadPath := strings.TrimSuffix(path, filepath.Ext(path)) + "_download.csv"
	if err := writeURLListResults(urls, tasks, failedURLs, downloadPath); err != nil {
		logger.Error(fmt.Sprintf("Failed to write download file: %v", err))
	} else {
		logger.Info(fmt.Sprintf("Download file saved to: %s", downloadPath))
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

	// Create concurrent downloader with reasonable limits
	downloader := NewConcurrentDownloader(25, 4) // 25 global, 4 per host

	// Process each paper to prepare download tasks
	successCount := 0
	failCount := 0
	var tasks []*DownloadTask

	for i, paper := range papers {
		// Log progress every 10 papers during preparation
		if (i+1)%10 == 0 {
			logger.Info(fmt.Sprintf("Preparing paper %d of %d...", i+1, len(papers)))
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

		// Create download task
		task := &DownloadTask{
			URL:         paper.URL,
			PDFUrl:      pdfURL,
			Filename:    filename,
			FullPath:    fullPath,
			Paper:       paper,
			OriginalURL: paper.URL,
		}
		tasks = append(tasks, task)
	}

	// Perform concurrent downloads
	if len(tasks) > 0 {
		logger.Info(fmt.Sprintf("Starting concurrent download of %d PDFs (max 25 global, 4 per host)", len(tasks)))
		results := downloader.downloadConcurrently(tasks)

		// Process results
		for _, result := range results {
			if result.Success {
				logger.Info(fmt.Sprintf("Row %s: PDF downloaded successfully as %s", result.Task.Paper.ID, result.Task.Filename))
				result.Task.Paper.Downloaded = true
				successCount++
			} else {
				logger.Error(fmt.Sprintf("Row %s: Download failed for %s: %v", result.Task.Paper.ID, result.Task.Paper.URL, result.Error))
				result.Task.Paper.ErrorMsg = fmt.Sprintf("Download failed: %v", result.Error)
				failCount++
			}
		}
	}

	// Generate single enhanced CSV/TSV with original columns + download fields
	ext := filepath.Ext(path)
	enhancedPath := strings.TrimSuffix(path, ext) + "_download" + ext
	if err := writeEnhancedCSV(papers, headers, enhancedPath, delimiter); err != nil {
		logger.Error(fmt.Sprintf("Failed to write enhanced CSV: %v", err))
	} else {
		logger.Info(fmt.Sprintf("Enhanced CSV with download results saved to: %s", enhancedPath))
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

	// Read a few sample rows for content analysis
	var sampleRows [][]string
	sampleCount := 0
	for sampleCount < 5 { // Read up to 5 sample rows
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		sampleRows = append(sampleRows, record)
		sampleCount++
	}

	// Detect column mappings with content analysis
	mapping := detectColumnsWithContent(headers, sampleRows)
	if mapping.URL == -1 && mapping.DOI == -1 {
		return nil, nil, fmt.Errorf("no URL or DOI column found in CSV/TSV file")
	}

	// Log detected columns for debugging
	logDetectedColumns(headers, mapping)

	// Reset file reader for actual parsing
	file.Close()
	file, err = os.Open(filepath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to reopen file: %w", err)
	}
	defer file.Close()

	reader = csv.NewReader(file)
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Skip header row
	_, err = reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read header row on second pass: %w", err)
	}

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
// analyzeColumnContent analyzes sample data to determine if a column likely contains journal names vs database names
func analyzeColumnContent(columnData []string) (bool, string) {
	if len(columnData) == 0 {
		return false, "no_data"
	}

	journalIndicators := 0
	databaseIndicators := 0
	totalNonEmpty := 0

	for _, value := range columnData {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		totalNonEmpty++
		valueLower := strings.ToLower(value)

		// Journal indicators
		if strings.Contains(valueLower, "journal") ||
			strings.Contains(valueLower, "proceedings") ||
			strings.Contains(valueLower, "conference") ||
			strings.Contains(valueLower, "review") ||
			strings.Contains(valueLower, "nature") ||
			strings.Contains(valueLower, "science") ||
			strings.Contains(valueLower, "plos") ||
			len(strings.Fields(value)) > 2 { // Multi-word journal names are common
			journalIndicators++
		}

		// Database indicators
		if strings.Contains(valueLower, "scopus") ||
			strings.Contains(valueLower, "pubmed") ||
			strings.Contains(valueLower, "crossref") ||
			strings.Contains(valueLower, "wos") ||
			strings.Contains(valueLower, "web of science") ||
			strings.Contains(valueLower, "google scholar") ||
			strings.Contains(valueLower, "dimensions") ||
			strings.Contains(valueLower, "semantic scholar") ||
			(len(strings.Fields(value)) == 1 && len(value) < 15) { // Short single words often databases
			databaseIndicators++
		}
	}

	if totalNonEmpty == 0 {
		return false, "empty"
	}

	journalRatio := float64(journalIndicators) / float64(totalNonEmpty)
	databaseRatio := float64(databaseIndicators) / float64(totalNonEmpty)

	// If more than 30% look like journals and less than 20% look like databases, likely journal
	if journalRatio > 0.3 && databaseRatio < 0.2 {
		return true, "likely_journal"
	}
	// If more than 30% look like databases, likely database
	if databaseRatio > 0.3 {
		return false, "likely_database"
	}

	return journalRatio > databaseRatio, "unclear"
}

// detectColumnsWithContent detects columns using both headers and content analysis
func detectColumnsWithContent(headers []string, sampleRows [][]string) ColumnMapping {
	mapping := detectColumns(headers)

	// If we found a journal column, validate it with content analysis
	if mapping.Journal != -1 && len(sampleRows) > 0 {
		// Extract sample data for the detected journal column
		var journalSamples []string
		for _, row := range sampleRows {
			if mapping.Journal < len(row) {
				journalSamples = append(journalSamples, row[mapping.Journal])
			}
		}

		isJournal, reason := analyzeColumnContent(journalSamples)
		if !isJournal && reason == "likely_database" {
			logger.Info(fmt.Sprintf("Column '%s' detected as journal but content suggests database source, searching for better journal column", headers[mapping.Journal]))

			// Look for alternative journal columns
			originalJournalCol := mapping.Journal
			mapping.Journal = -1

			// Try to find a better journal column
			for i, header := range headers {
				if i == originalJournalCol {
					continue
				}
				headerLower := strings.ToLower(strings.TrimSpace(header))

				// Look for more specific journal indicators
				if strings.Contains(headerLower, "sourcetitle") ||
					strings.Contains(headerLower, "source_title") ||
					strings.Contains(headerLower, "publicationtitle") ||
					strings.Contains(headerLower, "publication_title") ||
					strings.Contains(headerLower, "journaltitle") ||
					strings.Contains(headerLower, "journal_title") ||
					(strings.Contains(headerLower, "title") &&
						(strings.Contains(headerLower, "source") ||
							strings.Contains(headerLower, "journal") ||
							strings.Contains(headerLower, "publication"))) {

					// Analyze content for this candidate
					var candidateSamples []string
					for _, row := range sampleRows {
						if i < len(row) {
							candidateSamples = append(candidateSamples, row[i])
						}
					}

					candidateIsJournal, candidateReason := analyzeColumnContent(candidateSamples)
					if candidateIsJournal || candidateReason != "likely_database" {
						mapping.Journal = i
						logger.Info(fmt.Sprintf("Found better journal column: '%s' (reason: %s)", header, candidateReason))
						break
					}
				}
			}

			// If no better column found, keep original but warn
			if mapping.Journal == -1 {
				mapping.Journal = originalJournalCol
				logger.Info(fmt.Sprintf("No better journal column found, keeping '%s' despite database-like content", headers[originalJournalCol]))
			}
		} else {
			logger.Info(fmt.Sprintf("Column '%s' confirmed as journal (reason: %s)", headers[mapping.Journal], reason))
		}
	}

	return mapping
}

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

		// Journal/Source detection - prioritize journal-specific terms
		// First check for highest priority journal columns
		if strings.Contains(headerLower, "sourcetitle") ||
			strings.Contains(headerLower, "source_title") ||
			strings.Contains(headerLower, "publication_title") ||
			strings.Contains(headerLower, "publicationtitle") ||
			strings.Contains(headerLower, "journaltitle") ||
			strings.Contains(headerLower, "journal_title") {
			mapping.Journal = i // Always override for these specific terms
		} else if mapping.Journal == -1 {
			// Second priority: general journal terms
			if strings.Contains(headerLower, "journal") ||
				strings.Contains(headerLower, "venue") {
				mapping.Journal = i
			} else if strings.Contains(headerLower, "publication") &&
				!strings.Contains(headerLower, "year") &&
				!strings.Contains(headerLower, "date") {
				mapping.Journal = i
			} else if strings.Contains(headerLower, "source") {
				// Lowest priority: generic "source" - content analysis will refine this later
				mapping.Journal = i
			}
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

// writeURLListResults creates a CSV file with download results for plain text URL lists
func writeURLListResults(allURLs []string, successfulTasks []*DownloadTask, failedURLs []string, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create results file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	headers := []string{"url", "downloaded", "error_reason", "filename"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Create lookup maps for efficiency
	successMap := make(map[string]*DownloadTask)
	for _, task := range successfulTasks {
		successMap[task.OriginalURL] = task
	}

	failedMap := make(map[string]bool)
	for _, url := range failedURLs {
		failedMap[url] = true
	}

	// Write data rows for all URLs
	for _, url := range allURLs {
		if task, exists := successMap[url]; exists {
			// Successful download
			record := []string{url, "true", "", task.Filename}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write record: %w", err)
			}
		} else if failedMap[url] {
			// Failed download
			record := []string{url, "false", "Download failed or no PDF found", ""}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write record: %w", err)
			}
		}
	}

	return nil
}

// writeEnhancedCSV writes a CSV/TSV file with all original columns plus download result columns
func writeEnhancedCSV(papers []*PaperMetadata, headers []string, outputPath string, delimiter rune) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create enhanced CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = delimiter
	defer writer.Flush()

	// Write header with additional download result columns
	enhancedHeaders := append(headers, "downloaded", "error_reason", "filename")
	if err := writer.Write(enhancedHeaders); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows with download results
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

		// Append error reason (empty if successful)
		errorReason := ""
		if paper.ErrorMsg != "" {
			errorReason = paper.ErrorMsg
		}
		record = append(record, errorReason)

		// Append filename (empty if failed)
		filename := ""
		if paper.Downloaded && paper.Filename != "" {
			filename = paper.Filename
		}
		record = append(record, filename)

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
	return extractPDFWithDepth(pageURL, 0)
}

func extractPDFWithDepth(pageURL string, depth int) (pdfURL string, filename string, err error) {
	return extractPDFWithDepthAndVisited(pageURL, depth, make(map[string]bool))
}

func extractPDFWithDepthAndVisited(pageURL string, depth int, visitedDOIs map[string]bool) (pdfURL string, filename string, err error) {
	// Prevent infinite recursion by limiting depth
	const maxDepth = 3
	if depth > maxDepth {
		return "", "", fmt.Errorf("maximum recursion depth reached for URL: %s", pageURL)
	}
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

	// Fetch the page to check content type with proper User-Agent using global client
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PrismAID/1.0; +https://github.com/open-and-sustainable/prismaid)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := httpClient.Do(req)
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
					// Check if we've already processed this DOI
					if visitedDOIs[doi] {
						logger.Info("DOI already processed, skipping:", doi)
						return
					}
					visitedDOIs[doi] = true

					logger.Info("Found DOI in meta tags:", doi, "- attempting to resolve")
					// Recursively call extractPDF with the DOI URL
					doiURL := convertDOIToURL(doi)
					resolvedURL, resolvedFilename, err := extractPDFWithDepthAndVisited(doiURL, depth+1, visitedDOIs)
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

						// Check if we've already processed this DOI
						if visitedDOIs[doi] {
							logger.Info("DOI already processed, skipping:", doi)
							return
						}
						visitedDOIs[doi] = true

						logger.Info("Found DOI in page content:", doi, "- attempting to resolve")
						doiURL := convertDOIToURL(doi)
						resolvedURL, resolvedFilename, err := extractPDFWithDepthAndVisited(doiURL, depth+1, visitedDOIs)
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
//
// isRetryableError determines if an error is worth retrying
func isRetryableError(err error, statusCode int) bool {
	// Retry on server errors (5xx)
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// Retry on timeout or connection errors
	if strings.Contains(strings.ToLower(err.Error()), "timeout") ||
		strings.Contains(strings.ToLower(err.Error()), "connection reset") ||
		strings.Contains(strings.ToLower(err.Error()), "connection refused") ||
		strings.Contains(strings.ToLower(err.Error()), "no such host") {
		return true
	}

	return false
}

// parseRetryAfter parses the Retry-After header value
func parseRetryAfter(retryAfter string) time.Duration {
	if retryAfter == "" {
		return 0
	}

	// Try parsing as seconds
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP date
	if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		return time.Until(t)
	}

	return 0
}

// calculateBackoffDelay calculates the delay for exponential backoff with jitter
func calculateBackoffDelay(attempt int, config RetryConfig) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt))) * config.BaseDelay

	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	if config.Jitter {
		// Add jitter: randomize between 50% and 100% of the calculated delay
		jitterRange := delay / 2
		jitter := time.Duration(rand.Int63n(int64(jitterRange)))
		delay = delay/2 + jitter
	}

	return delay
}

// validatePDFResponse performs early validation of the response
func validatePDFResponse(resp *http.Response) error {
	// Check Content-Type header
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		// Accept application/pdf, application/octet-stream, or missing content-type
		if !strings.Contains(contentType, "pdf") &&
			!strings.Contains(contentType, "octet-stream") &&
			!strings.Contains(contentType, "binary") {
			return fmt.Errorf("unexpected content type: %s (expected PDF)", contentType)
		}
	}

	// Read first few bytes to check for PDF signature
	buffer := make([]byte, 4)
	n, err := resp.Body.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read response body for validation: %w", err)
	}

	if n >= 4 && string(buffer) != "%PDF" {
		return fmt.Errorf("response does not start with PDF signature, got: %q", string(buffer[:n]))
	}

	return nil
}

// downloadPDFWithRetry downloads a PDF with retry logic and early validation
func downloadPDFWithRetry(pdfURL, fullPath string, config RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			logger.Info(fmt.Sprintf("Retrying download of %s (attempt %d/%d) after %v", pdfURL, attempt+1, config.MaxRetries+1, delay))
			time.Sleep(delay)
		}

		err := downloadPDF(pdfURL, fullPath)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if this is a retryable error
		statusCode := 0
		if strings.Contains(err.Error(), "bad status:") {
			// Try to extract status code from error message
			parts := strings.Split(err.Error(), " ")
			if len(parts) > 2 {
				if code, parseErr := strconv.Atoi(parts[2]); parseErr == nil {
					statusCode = code
				}
			}
		}

		if !isRetryableError(err, statusCode) {
			logger.Info(fmt.Sprintf("Non-retryable error for %s: %v", pdfURL, err))
			break
		}

		if attempt < config.MaxRetries {
			logger.Info(fmt.Sprintf("Retryable error for %s: %v", pdfURL, err))
		}
	}

	return fmt.Errorf("download failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// tryUnpaywallFallback attempts to find an open access version using Unpaywall API
func tryUnpaywallFallback(task *DownloadTask) error {
	// Extract or find DOI from the original URL/task
	doi := extractDOIFromURL(task.OriginalURL)
	if doi == "" && task.Paper != nil {
		doi = task.Paper.DOI
	}

	if doi == "" {
		return fmt.Errorf("no DOI available for Unpaywall lookup")
	}

	// Clean the DOI for API usage (remove URL prefixes if present)
	cleanDOI := strings.TrimSpace(doi)
	if strings.HasPrefix(strings.ToLower(cleanDOI), "https://doi.org/") {
		cleanDOI = cleanDOI[16:]
	} else if strings.HasPrefix(strings.ToLower(cleanDOI), "http://doi.org/") {
		cleanDOI = cleanDOI[15:]
	} else if strings.HasPrefix(strings.ToLower(cleanDOI), "https://dx.doi.org/") {
		cleanDOI = cleanDOI[19:]
	} else if strings.HasPrefix(strings.ToLower(cleanDOI), "http://dx.doi.org/") {
		cleanDOI = cleanDOI[18:]
	} else if strings.HasPrefix(strings.ToLower(cleanDOI), "doi:") {
		cleanDOI = strings.TrimSpace(cleanDOI[4:])
	}

	// Query Unpaywall API
	unpaywallURL := fmt.Sprintf("https://api.unpaywall.org/v2/%s?email=prismaid@ourresearch.org", cleanDOI)

	resp, err := httpClient.Get(unpaywallURL)
	if err != nil {
		return fmt.Errorf("Unpaywall API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unpaywall API returned status %d", resp.StatusCode)
	}

	var unpaywallResp UnpaywallResponse
	if err := json.NewDecoder(resp.Body).Decode(&unpaywallResp); err != nil {
		return fmt.Errorf("failed to decode Unpaywall response: %w", err)
	}

	if !unpaywallResp.IsOA {
		return fmt.Errorf("no open access version available according to Unpaywall")
	}

	// Try the best OA location first
	if unpaywallResp.BestOALocation.URLForPDF != "" {
		logger.Info(fmt.Sprintf("Found open access PDF via Unpaywall: %s", unpaywallResp.BestOALocation.URLForPDF))
		task.PDFUrl = unpaywallResp.BestOALocation.URLForPDF
		return nil
	}

	// Try other OA locations
	for _, location := range unpaywallResp.OALocations {
		if location.URLForPDF != "" {
			logger.Info(fmt.Sprintf("Found alternative open access PDF via Unpaywall: %s", location.URLForPDF))
			task.PDFUrl = location.URLForPDF
			return nil
		}
	}

	return fmt.Errorf("Unpaywall found open access record but no PDF URLs available")
}

// extractDOIFromURL attempts to extract a DOI from a URL
func extractDOIFromURL(url string) string {
	// Try to extract DOI from common DOI URL patterns
	doiPatterns := []string{
		`^https?://doi\.org/(.+)`,
		`^https?://dx\.doi\.org/(.+)`,
		`^doi:(.+)`,
		`^DOI:(.+)`,
		`.*/doi/(.+)`,
	}

	for _, pattern := range doiPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			doi := strings.TrimSpace(matches[1])
			// Clean up common suffixes that aren't part of the DOI
			doi = strings.Split(doi, "?")[0]
			doi = strings.Split(doi, "#")[0]
			doi = strings.Split(doi, "&")[0]
			return doi
		}
	}

	return ""
}

func downloadPDF(pdfURL, fullPath string) error {
	// Use global client with optimized transport and extended timeout for large downloads
	downloadClient := &http.Client{
		Transport: httpClient.Transport, // Reuse the optimized transport with connection pooling
		Timeout:   300 * time.Second,    // 5 minutes for large PDF files
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Create request with proper headers
	req, err := http.NewRequest("GET", pdfURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PrismAID/1.0; +https://github.com/open-and-sustainable/prismaid)")
	req.Header.Set("Accept", "application/pdf,application/octet-stream,*/*;q=0.8")

	resp, err := downloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if we got a successful response
	if resp.StatusCode != http.StatusOK {
		// Check for Retry-After header on rate limiting
		if resp.StatusCode == 429 || resp.StatusCode == 503 {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				delay := parseRetryAfter(retryAfter)
				if delay > 0 && delay < 5*time.Minute { // Cap at 5 minutes
					logger.Info(fmt.Sprintf("Server requested retry after %v for %s", delay, pdfURL))
					time.Sleep(delay)
				}
			}
		}
		return fmt.Errorf("bad status: %d %s", resp.StatusCode, resp.Status)
	}

	// Early validation of response
	if err := validatePDFResponse(resp); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy the response body (note: first 4 bytes were already read during validation)
	// We need to create a multi-reader that includes those bytes
	buffer := []byte("%PDF") // We know this is what we read during validation
	combinedReader := io.MultiReader(strings.NewReader(string(buffer)), resp.Body)

	_, err = io.Copy(out, combinedReader)
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

	// Make the request with proper User-Agent using global client
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		logger.Info(fmt.Sprintf("Error creating Crossref request: %v", err))
		return ""
	}

	// Crossref requests a polite User-Agent
	req.Header.Set("User-Agent", "PrismAID/1.0 (https://github.com/open-and-sustainable/prismaid; mailto:info@example.com)")

	resp, err := httpClient.Do(req)
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
				doiURL := convertDOIToURL(doi)
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
				doiURL := convertDOIToURL(doi)
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

	// Use global client but with extended timeout for large PDF downloads
	downloadClient := &http.Client{
		Transport: httpClient.Transport, // Reuse the optimized transport
		Timeout:   120 * time.Second,    // 2 minutes for large files
	}

	resp, err := downloadClient.Do(req)
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
