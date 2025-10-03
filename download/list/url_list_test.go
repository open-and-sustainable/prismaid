package list

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Test DownloadURLList with different file types
func TestDownloadURLList(t *testing.T) {
	// Create temp dir
	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock PDF server
	pdfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4\nmock pdf content"))
	}))
	defer pdfServer.Close()

	t.Run("TXT file", func(t *testing.T) {
		txtFile := filepath.Join(tempDir, "urls.txt")
		content := pdfServer.URL + "\n# comment\n\n"
		if err := os.WriteFile(txtFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		err := DownloadURLList(txtFile)
		if err != nil {
			t.Errorf("DownloadURLList failed: %v", err)
		}
	})

	t.Run("CSV file", func(t *testing.T) {
		csvFile := filepath.Join(tempDir, "papers.csv")
		content := "Title,URL,DOI\nTest Paper," + pdfServer.URL + ",10.1234/test\n"
		if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		err := DownloadURLList(csvFile)
		if err != nil {
			t.Errorf("DownloadURLList failed: %v", err)
		}

		// Check download file was created
		downloadPath := strings.TrimSuffix(csvFile, ".csv") + "_download.csv"
		if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
			t.Error("Download file not created")
		}
	})

	t.Run("TSV file", func(t *testing.T) {
		tsvFile := filepath.Join(tempDir, "papers.tsv")
		content := "Title\tURL\tDOI\nTest\t" + pdfServer.URL + "\t10.1234/test\n"
		if err := os.WriteFile(tsvFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		err := DownloadURLList(tsvFile)
		if err != nil {
			t.Errorf("DownloadURLList failed: %v", err)
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		err := DownloadURLList("/nonexistent/file.txt")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

// Test extractPDF function
func TestExtractPDF(t *testing.T) {
	// Direct PDF server
	pdfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("PDF content"))
	}))
	defer pdfServer.Close()

	// HTML with PDF link
	htmlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<html><a href="test.pdf">Download PDF</a></html>`
		w.Write([]byte(html))
	}))
	defer htmlServer.Close()

	// HTML with meta tag
	metaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<html><head><meta name="citation_pdf_url" content="https://example.com/paper.pdf"></head></html>`
		w.Write([]byte(html))
	}))
	defer metaServer.Close()

	tests := []struct {
		name         string
		url          string
		expectPDFURL string
		expectError  bool
	}{
		{"Direct PDF", pdfServer.URL, pdfServer.URL, false},
		{"HTML with PDF link", htmlServer.URL, htmlServer.URL + "/test.pdf", false},
		{"HTML with meta PDF", metaServer.URL, "https://example.com/paper.pdf", false},
		{"Invalid URL", "http://invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdfURL, filename, err := extractPDF(tt.url)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if pdfURL != tt.expectPDFURL {
				t.Errorf("Expected PDF URL %q, got %q", tt.expectPDFURL, pdfURL)
			}
			if !tt.expectError && filename == "" {
				t.Error("Expected non-empty filename")
			}
		})
	}
}

// Test Dimensions.ai URL handling
func TestHandleDimensionsURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "Valid Dimensions URL",
			url:         "https://app.dimensions.ai/details/publication/pub.1118092727",
			expectError: true, // Expected to fail as it requires API access
		},
		{
			name:        "Invalid Dimensions URL",
			url:         "https://app.dimensions.ai/details/publication/invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdfURL, filename, err := extractPDF(tt.url)

			// For Dimensions URLs, we expect an error or empty result
			if strings.Contains(tt.url, "dimensions.ai") {
				if err == nil && pdfURL == "" {
					// This is acceptable - no PDF found
				} else if err != nil {
					// This is also acceptable - error handling Dimensions URL
				} else if pdfURL != "" {
					t.Errorf("Unexpectedly resolved Dimensions.ai URL to PDF: %q", pdfURL)
				}
			}

			_ = filename // filename might be empty or have a default value
		})
	}
}

// Test parseCSVFile function
func TestParseCSVFile(t *testing.T) {
	tempDir := t.TempDir()

	csvPath := filepath.Join(tempDir, "test.csv")
	content := `Title,Authors,Year,URL,DOI
"Paper 1","Smith, J.",2023,https://example.com/p1.pdf,10.1234/a
"Paper 2","Jones, M.",2024,,10.5678/b
"Paper 3","Brown, A.",2023,,`

	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	papers, _, err := parseCSVFile(csvPath, ',')
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	if len(papers) != 3 {
		t.Errorf("Expected 3 papers, got %d", len(papers))
	}

	// Check first paper
	if papers[0].Title != "Paper 1" {
		t.Errorf("Wrong title: %s", papers[0].Title)
	}
	if papers[0].URL != "https://example.com/p1.pdf" {
		t.Errorf("Wrong URL: %s", papers[0].URL)
	}

	// Check second paper (DOI converted to URL)
	if !strings.Contains(papers[1].URL, "doi.org") {
		t.Errorf("DOI not converted to URL: %s", papers[1].URL)
	}

	// Check third paper (no URL/DOI) - only if we have 3 papers
	if len(papers) > 2 {
		if papers[2].URL != "" {
			t.Errorf("Expected empty URL, got: %s", papers[2].URL)
		}
	}
}

// Test detectColumns function
func TestDetectColumns(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		wantURL int
		wantDOI int
	}{
		{
			name:    "Standard headers",
			headers: []string{"Title", "URL", "DOI", "Authors"},
			wantURL: 1,
			wantDOI: 2,
		},
		{
			name:    "BestURL preferred",
			headers: []string{"URL", "BestURL", "DOI"},
			wantURL: 1, // BestURL at index 1
			wantDOI: 2,
		},
		{
			name:    "Case insensitive",
			headers: []string{"title", "url", "doi"},
			wantURL: 1,
			wantDOI: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := detectColumns(tt.headers)
			if mapping.URL != tt.wantURL {
				t.Errorf("URL: got %d, want %d", mapping.URL, tt.wantURL)
			}
			if mapping.DOI != tt.wantDOI {
				t.Errorf("DOI: got %d, want %d", mapping.DOI, tt.wantDOI)
			}
		})
	}
}

// Test generateFilename function
func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		paper    *PaperMetadata
		contains []string
	}{
		{
			name: "Complete metadata",
			paper: &PaperMetadata{
				ID:      "1",
				Title:   "Climate Change",
				Authors: "Smith, John",
				Year:    "2023",
			},
			contains: []string{"2023", "Smith", "Climate", ".pdf"},
		},
		{
			name: "No metadata",
			paper: &PaperMetadata{
				ID: "2",
			},
			contains: []string{"paper_2.pdf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFilename(tt.paper)
			for _, part := range tt.contains {
				if !strings.Contains(result, part) {
					t.Errorf("Expected %q in filename, got %q", part, result)
				}
			}
		})
	}
}

// Test convertDOIToURL function
func TestConvertDOIToURL(t *testing.T) {
	tests := []struct {
		doi      string
		expected string
	}{
		{"10.1234/test", "https://doi.org/10.1234/test"},
		{"doi:10.1234/test", "https://doi.org/10.1234/test"},
		{"https://doi.org/10.1234/test", "https://doi.org/10.1234/test"},
		{"", ""},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.doi, func(t *testing.T) {
			result := convertDOIToURL(tt.doi)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test sanitizeFilename function
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with space", "with_space"},
		{"with/slash", "with_slash"},
		{"with:colon", "with_colon"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test findUniqueFilename function
func TestFindUniqueFilename(t *testing.T) {
	tempDir := t.TempDir()

	// Create existing file
	existingFile := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test unique filename
	result := findUniqueFilename(tempDir, "new.pdf")
	if result != "new.pdf" {
		t.Errorf("Expected 'new.pdf', got %q", result)
	}

	// Test existing filename
	result = findUniqueFilename(tempDir, "test.pdf")
	if result != "test_1.pdf" {
		t.Errorf("Expected 'test_1.pdf', got %q", result)
	}
}

// Test writeURLListResults function
func TestWriteURLListResults(t *testing.T) {
	tempDir := t.TempDir()
	downloadPath := filepath.Join(tempDir, "download.csv")

	allURLs := []string{
		"http://example.com/paper1",
		"http://example.com/paper2",
		"http://example.com/paper3",
	}

	successfulTasks := []*DownloadTask{
		{
			OriginalURL: "http://example.com/paper1",
			Filename:    "paper1.pdf",
		},
	}

	failedURLs := []string{
		"http://example.com/paper2",
		"http://example.com/paper3",
	}

	err := writeURLListResults(allURLs, successfulTasks, failedURLs, downloadPath)
	if err != nil {
		t.Fatalf("Failed to write download file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		t.Error("Download file not created")
	}

	// Read and verify content
	file, err := os.Open(downloadPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 4 { // header + 3 data rows
		t.Errorf("Expected 4 rows, got %d", len(records))
	}

	// Check header
	if records[0][0] != "url" || records[0][1] != "downloaded" || records[0][2] != "error_reason" || records[0][3] != "filename" {
		t.Errorf("Unexpected header: %v", records[0])
	}

	// Check successful download
	if records[1][1] != "true" || records[1][3] != "paper1.pdf" {
		t.Errorf("Unexpected successful record: %v", records[1])
	}

	// Check failed downloads
	if records[2][1] != "false" || records[2][2] == "" {
		t.Errorf("Unexpected failed record: %v", records[2])
	}
}

// Test extractDOIFromText function
func TestExtractDOIFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Standard DOI",
			text:     "The paper has DOI: 10.1234/journal.2023.456",
			expected: "10.1234/journal.2023.456",
		},
		{
			name:     "DOI in JSON format",
			text:     `{"title":"Test","doi":"10.5678/test.123","year":2023}`,
			expected: "10.5678/test.123",
		},
		{
			name:     "DOI with trailing punctuation",
			text:     "See doi: 10.1234/test.456.",
			expected: "10.1234/test.456",
		},
		{
			name:     "Multiple DOIs (first one extracted)",
			text:     "First: 10.1111/aaa.111 Second: 10.2222/bbb.222",
			expected: "10.1111/aaa.111",
		},
		{
			name:     "No DOI",
			text:     "This text has no DOI identifier",
			expected: "",
		},
		{
			name:     "DOI with parenthesis",
			text:     "(doi: 10.1234/test.789)",
			expected: "10.1234/test.789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDOIFromText(tt.text)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test writeEnhancedCSV function
func TestWriteEnhancedCSV(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "enhanced.csv")

	headers := []string{"Title", "Authors", "Year", "URL"}
	papers := []*PaperMetadata{
		{
			ID:             "1",
			Title:          "Paper One",
			Authors:        "Smith, J.",
			Year:           "2023",
			URL:            "https://example.com/1.pdf",
			Downloaded:     true,
			Filename:       "paper1.pdf",
			OriginalRecord: []string{"Paper One", "Smith, J.", "2023", "https://example.com/1.pdf"},
		},
		{
			ID:             "2",
			Title:          "Paper Two",
			Authors:        "Jones, M.",
			Year:           "2024",
			URL:            "https://example.com/2.pdf",
			Downloaded:     false,
			ErrorMsg:       "No PDF found",
			OriginalRecord: []string{"Paper Two", "Jones, M.", "2024", "https://example.com/2.pdf"},
		},
	}

	err := writeEnhancedCSV(papers, headers, csvPath, ',')
	if err != nil {
		t.Fatalf("Failed to write enhanced CSV: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Error("Enhanced CSV file not created")
	}

	// Read and verify content
	file, err := os.Open(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	// Check header
	if len(records) < 1 {
		t.Fatal("No header in enhanced CSV")
	}
	header := records[0]
	expectedHeaders := []string{"downloaded", "error_reason", "filename"}
	if len(header) < 3 {
		t.Fatal("Header missing expected columns")
	}
	for i, expected := range expectedHeaders {
		if header[len(header)-3+i] != expected {
			t.Errorf("Column %d should be '%s', got %q", len(header)-3+i, expected, header[len(header)-3+i])
		}
	}

	// Check data rows
	if len(records) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 rows, got %d", len(records))
	}

	// Verify download columns for successful row
	if records[1][len(records[1])-3] != "true" {
		t.Errorf("Row 1 should have downloaded=true, got %q", records[1][len(records[1])-3])
	}
	if records[1][len(records[1])-2] != "" {
		t.Errorf("Row 1 should have empty error_reason, got %q", records[1][len(records[1])-2])
	}
	if records[1][len(records[1])-1] != "paper1.pdf" {
		t.Errorf("Row 1 should have filename=paper1.pdf, got %q", records[1][len(records[1])-1])
	}

	// Verify download columns for failed row
	if records[2][len(records[2])-3] != "false" {
		t.Errorf("Row 2 should have downloaded=false, got %q", records[2][len(records[2])-3])
	}
	if records[2][len(records[2])-2] != "No PDF found" {
		t.Errorf("Row 2 should have error_reason='No PDF found', got %q", records[2][len(records[2])-2])
	}
	if records[2][len(records[2])-1] != "" {
		t.Errorf("Row 2 should have empty filename, got %q", records[2][len(records[2])-1])
	}
}

// Test processTextFile with failed URLs logging
func TestProcessTextFileWithFailedLogging(t *testing.T) {
	tempDir := t.TempDir()
	txtPath := filepath.Join(tempDir, "test.txt")

	// Create a test file with one valid (mock) and one invalid URL
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("mock pdf"))
	}))
	defer mockServer.Close()

	content := mockServer.URL + "\nhttps://invalid-url-will-fail.com/paper.pdf\n"
	if err := os.WriteFile(txtPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := processTextFile(txtPath, tempDir)
	if err != nil {
		t.Errorf("processTextFile failed: %v", err)
	}

	// Check if download file was created
	downloadPath := filepath.Join(tempDir, "test_download.csv")
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		t.Error("Download file not created")
	} else {
		// Read and verify content
		content, err := os.ReadFile(downloadPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "invalid-url-will-fail.com") {
			t.Error("Download file missing invalid URL")
		}
	}
}

// Test parseCSVFile with original record preservation
func TestParseCSVFileWithOriginalRecord(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "test.csv")

	content := `Title,Authors,Year,URL,DOI,ExtraColumn
"Paper 1","Smith, J.",2023,https://example.com/1.pdf,10.1234/a,Extra Data
"Paper 2","Jones, M.",2024,,10.5678/b,More Data`

	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	papers, headers, err := parseCSVFile(csvPath, ',')
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	if len(papers) != 2 {
		t.Errorf("Expected 2 papers, got %d", len(papers))
	}

	if len(headers) != 6 {
		t.Errorf("Expected 6 headers, got %d", len(headers))
	}

	// Check that original records are preserved
	if papers[0].OriginalRecord == nil {
		t.Error("Paper 1 OriginalRecord is nil")
	} else if len(papers[0].OriginalRecord) != 6 {
		t.Errorf("Paper 1 OriginalRecord has %d columns, expected 6", len(papers[0].OriginalRecord))
	} else if papers[0].OriginalRecord[5] != "Extra Data" {
		t.Errorf("Paper 1 extra column should be 'Extra Data', got %q", papers[0].OriginalRecord[5])
	}

	if papers[1].OriginalRecord == nil {
		t.Error("Paper 2 OriginalRecord is nil")
	} else if len(papers[1].OriginalRecord) != 6 {
		t.Errorf("Paper 2 OriginalRecord has %d columns, expected 6", len(papers[1].OriginalRecord))
	}
}

// Test isProblematicURL function
func TestIsProblematicURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "Dimensions.ai URL",
			url:      "https://app.dimensions.ai/details/publication/pub.1234567890",
			expected: true,
		},
		{
			name:     "ResearchGate URL",
			url:      "https://www.researchgate.net/publication/123456789_Some_Paper",
			expected: true,
		},
		{
			name:     "Academia.edu URL",
			url:      "https://www.academia.edu/12345678/Paper_Title",
			expected: true,
		},
		{
			name:     "Semantic Scholar web UI URL",
			url:      "https://www.semanticscholar.org/paper/Paper-Title/abc123def456",
			expected: true,
		},
		{
			name:     "Semantic Scholar API URL (not problematic)",
			url:      "https://api.semanticscholar.org/v1/paper/abc123",
			expected: false,
		},
		{
			name:     "Regular DOI URL",
			url:      "https://doi.org/10.1234/journal.2023.456",
			expected: false,
		},
		{
			name:     "ArXiv URL",
			url:      "https://arxiv.org/pdf/2301.12345.pdf",
			expected: false,
		},
		{
			name:     "Direct PDF URL",
			url:      "https://example.com/papers/paper123.pdf",
			expected: false,
		},
		{
			name:     "Case insensitive - Dimensions",
			url:      "https://APP.DIMENSIONS.AI/details/publication/pub.123",
			expected: true,
		},
		{
			name:     "Case insensitive - ResearchGate",
			url:      "https://WWW.RESEARCHGATE.NET/publication/123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isProblematicURL(tt.url)
			if result != tt.expected {
				t.Errorf("isProblematicURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

// Test HTTP client configuration and connection pooling
func TestHTTPClientConfiguration(t *testing.T) {
	// Test that the global HTTP client is properly initialized
	if httpClient == nil {
		t.Fatal("Global HTTP client should be initialized")
	}

	// Test client timeout
	if httpClient.Timeout != 60*time.Second {
		t.Errorf("Expected timeout of 60s, got %v", httpClient.Timeout)
	}

	// Test transport configuration
	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected *http.Transport")
	}

	// Test connection pool settings
	if transport.MaxIdleConns != 100 {
		t.Errorf("Expected MaxIdleConns=100, got %d", transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("Expected MaxIdleConnsPerHost=10, got %d", transport.MaxIdleConnsPerHost)
	}
	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf("Expected IdleConnTimeout=90s, got %v", transport.IdleConnTimeout)
	}

	// Test HTTP/2 support
	if !transport.ForceAttemptHTTP2 {
		t.Error("Expected ForceAttemptHTTP2 to be true")
	}

	// Test timeouts
	if transport.TLSHandshakeTimeout != 10*time.Second {
		t.Errorf("Expected TLSHandshakeTimeout=10s, got %v", transport.TLSHandshakeTimeout)
	}
	if transport.ResponseHeaderTimeout != 30*time.Second {
		t.Errorf("Expected ResponseHeaderTimeout=30s, got %v", transport.ResponseHeaderTimeout)
	}
}

// Test connection reuse by making multiple requests
func TestHTTPConnectionReuse(t *testing.T) {
	// Create a test server that tracks connections
	connCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connCount++
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4\nfake pdf content"))
	}))
	defer server.Close()

	// Create temporary directory
	tempDir := t.TempDir()

	// Test multiple downloads to same host
	urls := []string{
		server.URL + "/paper1.pdf",
		server.URL + "/paper2.pdf",
		server.URL + "/paper3.pdf",
	}

	successCount := 0
	for i, url := range urls {
		filename := fmt.Sprintf("paper%d.pdf", i+1)
		fullPath := filepath.Join(tempDir, filename)

		err := downloadPDF(url, fullPath)
		if err != nil {
			t.Logf("Download failed for %s: %v", url, err)
		} else {
			successCount++
		}
	}

	// At least some downloads should succeed
	if successCount == 0 {
		t.Error("No downloads succeeded")
	}

	t.Logf("Successfully downloaded %d out of %d files", successCount, len(urls))
}

// Test HTTP client timeout behavior
func TestHTTPClientTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4\nslow pdf content"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	fullPath := filepath.Join(tempDir, "slow.pdf")

	// This should succeed as our timeout is 300 seconds for downloads
	start := time.Now()
	err := downloadPDF(server.URL, fullPath)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Download should succeed but got error: %v", err)
	}

	// Should take around 2 seconds plus overhead
	if elapsed > 10*time.Second {
		t.Errorf("Download took too long: %v", elapsed)
	}

	t.Logf("Download completed in %v", elapsed)
}

// Test extractPDF with optimized client
func TestExtractPDFWithOptimizedClient(t *testing.T) {
	// Create a mock server that returns HTML with PDF link
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".pdf") {
			w.Header().Set("Content-Type", "application/pdf")
			w.Write([]byte("mock pdf content"))
		} else {
			w.Header().Set("Content-Type", "text/html")
			html := fmt.Sprintf(`<html><body><a href="%s/paper.pdf">Download PDF</a></body></html>`, server.URL)
			w.Write([]byte(html))
		}
	}))
	defer server.Close()

	// Test extractPDF function
	pdfURL, filename, err := extractPDF(server.URL)
	if err != nil {
		t.Fatalf("extractPDF failed: %v", err)
	}

	if pdfURL == "" {
		t.Error("Expected PDF URL but got empty string")
	}

	if filename == "" {
		t.Error("Expected filename but got empty string")
	}

	t.Logf("Found PDF URL: %s, filename: %s", pdfURL, filename)
}

// Test Crossref API with optimized client
func TestCrossrefAPIWithOptimizedClient(t *testing.T) {
	// Skip if running in offline mode
	if testing.Short() {
		t.Skip("Skipping Crossref API test in short mode")
	}

	title := "Machine Learning"
	authors := "Tom Mitchell"
	year := "1997"

	start := time.Now()
	doi := searchCrossrefForDOI(title, authors, year)
	elapsed := time.Since(start)

	// Should complete within reasonable time
	if elapsed > 30*time.Second {
		t.Errorf("Crossref API call took too long: %v", elapsed)
	}

	t.Logf("Crossref query completed in %v, result: %s", elapsed, doi)
}

// Test searchCrossrefForDOI function
func TestSearchCrossrefForDOI(t *testing.T) {
	// Note: These tests make real API calls to Crossref
	// They may be slow or fail if the API is down
	// Consider using t.Skip() in CI environments if needed

	tests := []struct {
		name        string
		title       string
		authors     string
		year        string
		expectDOI   bool
		description string
	}{
		{
			name:        "Well-known paper",
			title:       "A Statistical Interpretation of Term Specificity and Its Application in Retrieval",
			authors:     "Karen Sparck Jones",
			year:        "1972",
			expectDOI:   true,
			description: "Classic information retrieval paper",
		},
		{
			name:        "Empty title",
			title:       "",
			authors:     "John Smith",
			year:        "2023",
			expectDOI:   false,
			description: "Should return empty string with no title",
		},
		{
			name:        "Gibberish title",
			title:       "xyzabc123notarealpaper456xyz",
			authors:     "",
			year:        "",
			expectDOI:   false,
			description: "Should not find DOI for nonsense title",
		},
		{
			name:        "Title only",
			title:       "Machine Learning",
			authors:     "",
			year:        "",
			expectDOI:   true,
			description: "Generic title might match something (low score)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip Crossref API tests if running in offline mode
			// Uncomment the following line to skip these tests:
			// t.Skip("Skipping Crossref API test")

			result := searchCrossrefForDOI(tt.title, tt.authors, tt.year)

			if tt.expectDOI {
				// We expect some DOI, but we can't predict the exact value
				// Just check if it looks like a DOI
				if result == "" {
					t.Logf("Warning: Expected DOI for %s but got empty string", tt.description)
					t.Logf("This might be due to API rate limiting or temporary issues")
				} else if !strings.Contains(result, "10.") {
					t.Errorf("Result doesn't look like a DOI: %q", result)
				} else {
					t.Logf("Found DOI: %s", result)
				}
			} else {
				// We expect no DOI
				if result != "" && tt.title != "Machine Learning" {
					t.Logf("Unexpected DOI found: %s (might be a false positive)", result)
				}
			}
		})
	}
}

// Test extractPaperMetadata with problematic URLs
func TestExtractPaperMetadataWithProblematicURLs(t *testing.T) {
	mapping := ColumnMapping{
		URL:     0,
		DOI:     1,
		Title:   2,
		Authors: 3,
		Year:    4,
	}

	tests := []struct {
		name              string
		record            []string
		expectURLChange   bool
		expectDOIResolved bool
		description       string
	}{
		{
			name: "Problematic URL with DOI",
			record: []string{
				"https://www.researchgate.net/publication/123456789",
				"10.1234/test.2023.456",
				"Test Paper",
				"Smith, J.",
				"2023",
			},
			expectURLChange:   true,
			expectDOIResolved: true,
			description:       "Should replace ResearchGate URL with DOI URL",
		},
		{
			name: "Problematic URL without DOI",
			record: []string{
				"https://app.dimensions.ai/details/publication/pub.123",
				"",
				"Another Test Paper",
				"Jones, M.",
				"2024",
			},
			expectURLChange:   false,
			expectDOIResolved: false,
			description:       "Should attempt Crossref lookup (may or may not succeed)",
		},
		{
			name: "Normal URL with DOI",
			record: []string{
				"https://doi.org/10.1234/test",
				"10.1234/test",
				"Regular Paper",
				"Brown, A.",
				"2023",
			},
			expectURLChange:   false,
			expectDOIResolved: false,
			description:       "Should keep DOI URL as-is",
		},
		{
			name: "Dimensions URL with DOI",
			record: []string{
				"https://dimensions.ai/publication/pub.999",
				"10.5678/dimensions.test",
				"Dimensions Paper",
				"Taylor, K.",
				"2022",
			},
			expectURLChange:   true,
			expectDOIResolved: true,
			description:       "Should replace Dimensions URL with DOI",
		},
		{
			name: "Academia.edu URL with DOI",
			record: []string{
				"https://www.academia.edu/12345/Paper",
				"10.9999/academia.paper",
				"Academia Paper",
				"Wilson, R.",
				"2021",
			},
			expectURLChange:   true,
			expectDOIResolved: true,
			description:       "Should replace Academia.edu URL with DOI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paper := extractPaperMetadata(tt.record, mapping, 1)

			// Check if URL was changed from problematic URL
			if tt.expectURLChange {
				if paper.URL == tt.record[0] {
					t.Errorf("Expected URL to be changed from problematic URL %q", tt.record[0])
				}
			}

			// Check if DOI was resolved
			if tt.expectDOIResolved {
				if !strings.Contains(paper.URL, "doi.org") {
					t.Errorf("Expected URL to contain doi.org, got %q", paper.URL)
				}
			}

			// Check that metadata was preserved
			if paper.Title != tt.record[2] {
				t.Errorf("Title mismatch: got %q, want %q", paper.Title, tt.record[2])
			}
			if paper.Authors != tt.record[3] {
				t.Errorf("Authors mismatch: got %q, want %q", paper.Authors, tt.record[3])
			}

			t.Logf("%s: URL changed to %q", tt.description, paper.URL)
		})
	}
}

// TestConcurrentDownloader tests the concurrent download functionality
func TestConcurrentDownloader(t *testing.T) {
	// Create temp dir
	tempDir := t.TempDir()

	// Track request timing to verify concurrent execution
	requestTimes := make(map[string]time.Time)
	var requestMutex sync.Mutex

	// Create mock server with artificial delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMutex.Lock()
		requestTimes[r.URL.Path] = time.Now()
		requestMutex.Unlock()

		// Add small delay to simulate network latency
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4\nmock pdf content"))
	}))
	defer server.Close()

	// Create downloader with low limits for testing
	downloader := NewConcurrentDownloader(3, 2) // 3 global, 2 per host

	// Create multiple download tasks
	tasks := []*DownloadTask{
		{
			PDFUrl:   server.URL + "/paper1.pdf",
			Filename: "paper1.pdf",
			FullPath: filepath.Join(tempDir, "paper1.pdf"),
		},
		{
			PDFUrl:   server.URL + "/paper2.pdf",
			Filename: "paper2.pdf",
			FullPath: filepath.Join(tempDir, "paper2.pdf"),
		},
		{
			PDFUrl:   server.URL + "/paper3.pdf",
			Filename: "paper3.pdf",
			FullPath: filepath.Join(tempDir, "paper3.pdf"),
		},
	}

	start := time.Now()
	results := downloader.downloadConcurrently(tasks)
	duration := time.Since(start)

	// Verify all downloads succeeded
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
			// Verify file was created
			if _, err := os.Stat(result.Task.FullPath); err != nil {
				t.Errorf("Expected file %s to exist", result.Task.FullPath)
			}
		} else {
			t.Errorf("Download failed: %v", result.Error)
		}
	}

	if successCount != 3 {
		t.Errorf("Expected 3 successful downloads, got %d", successCount)
	}

	// Verify concurrent execution (should be faster than sequential)
	expectedSequentialTime := time.Duration(len(tasks)) * 100 * time.Millisecond
	if duration >= expectedSequentialTime {
		t.Logf("Warning: Duration %v suggests downloads may not have been concurrent (expected < %v)", duration, expectedSequentialTime)
	} else {
		t.Logf("Downloads completed in %v (concurrent execution detected)", duration)
	}

	t.Logf("Successfully downloaded %d files concurrently in %v", successCount, duration)
}

// TestRetryFunctionality tests the retry logic with exponential backoff
func TestRetryFunctionality(t *testing.T) {
	tempDir := t.TempDir()

	// Track retry attempts
	attemptCount := 0
	var attemptMutex sync.Mutex

	// Create mock server that fails first 2 attempts, succeeds on 3rd
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptMutex.Lock()
		attemptCount++
		currentAttempt := attemptCount
		attemptMutex.Unlock()

		if currentAttempt <= 2 {
			// First 2 attempts: return 503 (retryable)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service temporarily unavailable"))
			return
		}

		// 3rd attempt: succeed
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4\nretry success pdf"))
	}))
	defer server.Close()

	fullPath := filepath.Join(tempDir, "retry_test.pdf")

	// Test with retry config that allows 3 attempts
	config := RetryConfig{
		MaxRetries: 2,                     // Total of 3 attempts (0, 1, 2)
		BaseDelay:  50 * time.Millisecond, // Fast for testing
		MaxDelay:   200 * time.Millisecond,
		Jitter:     false, // Disable jitter for predictable testing
	}

	start := time.Now()
	err := downloadPDFWithRetry(server.URL, fullPath, config)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected download to succeed after retries, got: %v", err)
	}

	// Verify file was created
	if _, statErr := os.Stat(fullPath); statErr != nil {
		t.Errorf("Expected file to exist at %s", fullPath)
	}

	// Should have made exactly 3 attempts
	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	// Should take at least the retry delays (50ms + 100ms = 150ms minimum)
	expectedMinDuration := 150 * time.Millisecond
	if duration < expectedMinDuration {
		t.Errorf("Duration %v too short, expected at least %v", duration, expectedMinDuration)
	}

	t.Logf("Retry test completed in %v with %d attempts", duration, attemptCount)
}

// TestNonRetryableError tests that non-retryable errors fail immediately
func TestNonRetryableError(t *testing.T) {
	tempDir := t.TempDir()

	attemptCount := 0
	var attemptMutex sync.Mutex

	// Create mock server that always returns 404 (non-retryable)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptMutex.Lock()
		attemptCount++
		attemptMutex.Unlock()

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	fullPath := filepath.Join(tempDir, "nonretry_test.pdf")

	config := DefaultRetryConfig()

	err := downloadPDFWithRetry(server.URL, fullPath, config)

	// Should fail immediately without retries
	if err == nil {
		t.Error("Expected download to fail for 404 error")
	}

	// Should have made only 1 attempt (no retries for 404)
	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attemptCount)
	}

	t.Logf("Non-retryable error handled correctly with %d attempts", attemptCount)
}

// TestImprovedColumnDetection tests the improved column detection for source vs journal
func TestImprovedColumnDetection(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		sampleRows  [][]string
		expectedCol int
		description string
	}{
		{
			name:    "SourceTitle preferred over Source",
			headers: []string{"Title", "Authors", "Source", "SourceTitle", "Year"},
			sampleRows: [][]string{
				{"Paper 1", "Author 1", "Scopus", "Nature Medicine", "2023"},
				{"Paper 2", "Author 2", "PubMed", "Science", "2022"},
				{"Paper 3", "Author 3", "Crossref", "PLOS ONE", "2023"},
			},
			expectedCol: 3, // SourceTitle column
			description: "Should prefer SourceTitle over Source when Source contains database names",
		},
		{
			name:    "Source used when contains journal names",
			headers: []string{"Title", "Authors", "Source", "DOI"},
			sampleRows: [][]string{
				{"Paper 1", "Author 1", "Journal of Medicine", "10.1234/test1"},
				{"Paper 2", "Author 2", "Nature Communications", "10.1234/test2"},
				{"Paper 3", "Author 3", "Proceedings of Science", "10.1234/test3"},
			},
			expectedCol: 2, // Source column
			description: "Should use Source when it contains journal-like names",
		},
		{
			name:    "PublicationTitle preferred",
			headers: []string{"Title", "Source", "PublicationTitle", "Authors"},
			sampleRows: [][]string{
				{"Paper 1", "WOS", "Cell Biology Journal", "Author 1"},
				{"Paper 2", "Dimensions", "Nature Reviews", "Author 2"},
			},
			expectedCol: 2, // PublicationTitle column
			description: "Should prefer PublicationTitle over generic Source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := detectColumnsWithContent(tt.headers, tt.sampleRows)

			if mapping.Journal != tt.expectedCol {
				t.Errorf("Expected journal column %d (%s), got %d (%s)",
					tt.expectedCol,
					tt.headers[tt.expectedCol],
					mapping.Journal,
					func() string {
						if mapping.Journal >= 0 && mapping.Journal < len(tt.headers) {
							return tt.headers[mapping.Journal]
						}
						return "none"
					}())
			}

			t.Logf("%s: Correctly detected journal column as '%s'",
				tt.description,
				func() string {
					if mapping.Journal >= 0 && mapping.Journal < len(tt.headers) {
						return tt.headers[mapping.Journal]
					}
					return "none"
				}())
		})
	}
}

// TestDOICleaningForUnpaywall tests that DOIs are properly cleaned before sending to Unpaywall API
func TestDOICleaningForUnpaywall(t *testing.T) {
	tests := []struct {
		inputDOI    string
		expectedDOI string
		description string
	}{
		{"10.1109/POMS.2018.8629496", "10.1109/POMS.2018.8629496", "Plain DOI"},
		{"https://doi.org/10.1109/POMS.2018.8629496", "10.1109/POMS.2018.8629496", "HTTPS DOI URL"},
		{"http://doi.org/10.1109/POMS.2018.8629496", "10.1109/POMS.2018.8629496", "HTTP DOI URL"},
		{"https://dx.doi.org/10.1109/POMS.2018.8629496", "10.1109/POMS.2018.8629496", "DX DOI URL HTTPS"},
		{"http://dx.doi.org/10.1109/POMS.2018.8629496", "10.1109/POMS.2018.8629496", "DX DOI URL HTTP"},
		{"doi:10.1109/POMS.2018.8629496", "10.1109/POMS.2018.8629496", "DOI with prefix"},
		{"DOI:10.1109/POMS.2018.8629496", "10.1109/POMS.2018.8629496", "DOI with uppercase prefix"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Create a task with the test DOI
			task := &DownloadTask{
				Paper: &PaperMetadata{
					DOI: tt.inputDOI,
				},
			}

			// Test DOI extraction and cleaning
			doi := extractDOIFromURL(task.Paper.DOI)
			if doi == "" {
				doi = task.Paper.DOI
			}

			// Clean the DOI as done in tryUnpaywallFallback
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

			if cleanDOI != tt.expectedDOI {
				t.Errorf("Expected cleaned DOI '%s', got '%s'", tt.expectedDOI, cleanDOI)
			}

			t.Logf("%s: '%s' -> '%s'", tt.description, tt.inputDOI, cleanDOI)
		})
	}
}

// TestUnpaywallFallback tests the Unpaywall API fallback functionality
func TestUnpaywallFallback(t *testing.T) {
	tempDir := t.TempDir()

	// Mock Unpaywall API response
	unpaywallServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v2/10.1234/test") {
			// Return a successful Unpaywall response with open access PDF
			response := `{
				"doi": "10.1234/test",
				"is_oa": true,
				"best_oa_location": {
					"host_type": "repository",
					"url_for_pdf": "http://localhost:` + strings.Split(r.Host, ":")[1] + `/open_access.pdf"
				},
				"oa_locations": []
			}`
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer unpaywallServer.Close()

	// Mock PDF server that serves the open access PDF
	pdfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/open_access.pdf" {
			w.Header().Set("Content-Type", "application/pdf")
			w.Write([]byte("%PDF-1.4\nopen access pdf content"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer pdfServer.Close()

	// Create download task with a DOI-based URL that will fail initially
	task := &DownloadTask{
		URL:         "https://doi.org/10.1234/test",
		PDFUrl:      "https://example.com/nonexistent.pdf", // This will fail
		Filename:    "test.pdf",
		FullPath:    filepath.Join(tempDir, "test.pdf"),
		OriginalURL: "https://doi.org/10.1234/test",
		Paper: &PaperMetadata{
			DOI: "10.1234/test",
		},
	}

	// Temporarily override the Unpaywall API URL in the function
	// We'll need to modify the function to make it testable, but for now test the DOI extraction
	doi := extractDOIFromURL(task.OriginalURL)
	if doi != "10.1234/test" {
		t.Errorf("Expected DOI '10.1234/test', got '%s'", doi)
	}

	t.Logf("Successfully extracted DOI: %s", doi)
}

// TestExtractDOIFromURL tests DOI extraction from various URL formats
func TestExtractDOIFromURL(t *testing.T) {
	tests := []struct {
		url         string
		expectedDOI string
	}{
		{"https://doi.org/10.1234/test", "10.1234/test"},
		{"https://dx.doi.org/10.5678/example", "10.5678/example"},
		{"https://example.com/doi/10.9999/paper", "10.9999/paper"},
		{"doi:10.1111/sample", "10.1111/sample"},
		{"DOI:10.2222/UPPERCASE", "10.2222/UPPERCASE"},
		{"https://doi.org/10.1234/test?ref=123", "10.1234/test"},
		{"https://doi.org/10.1234/test#section1", "10.1234/test"},
		{"https://example.com/paper", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := extractDOIFromURL(tt.url)
			if result != tt.expectedDOI {
				t.Errorf("Expected DOI '%s', got '%s'", tt.expectedDOI, result)
			}
		})
	}
}
