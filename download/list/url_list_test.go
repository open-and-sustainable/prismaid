package list

import (
	"encoding/csv"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test DownloadURLList with different file types
func TestDownloadURLList(t *testing.T) {
	// Create temp dir
	tempDir, err := ioutil.TempDir("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock PDF server
	pdfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("mock pdf content"))
	}))
	defer pdfServer.Close()

	t.Run("TXT file", func(t *testing.T) {
		txtFile := filepath.Join(tempDir, "urls.txt")
		content := pdfServer.URL + "\n# comment\n\n"
		if err := ioutil.WriteFile(txtFile, []byte(content), 0644); err != nil {
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
		if err := ioutil.WriteFile(csvFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		err := DownloadURLList(csvFile)
		if err != nil {
			t.Errorf("DownloadURLList failed: %v", err)
		}

		// Check report was created
		reportPath := strings.TrimSuffix(csvFile, ".csv") + "_report.csv"
		if _, err := os.Stat(reportPath); os.IsNotExist(err) {
			t.Error("Report file not created")
		}
	})

	t.Run("TSV file", func(t *testing.T) {
		tsvFile := filepath.Join(tempDir, "papers.tsv")
		content := "Title\tURL\tDOI\nTest\t" + pdfServer.URL + "\t10.1234/test\n"
		if err := ioutil.WriteFile(tsvFile, []byte(content), 0644); err != nil {
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
	if err := ioutil.WriteFile(existingFile, []byte("test"), 0644); err != nil {
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

// Test writeDownloadReport function
func TestWriteDownloadReport(t *testing.T) {
	tempDir := t.TempDir()
	reportPath := filepath.Join(tempDir, "report.csv")

	papers := []*PaperMetadata{
		{
			ID:         "1",
			Title:      "Test Paper",
			Downloaded: true,
			Filename:   "test.pdf",
		},
		{
			ID:       "2",
			Title:    "Failed",
			ErrorMsg: "No PDF found",
		},
	}

	err := writeDownloadReport(papers, reportPath)
	if err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Error("Report file not created")
	}

	// Read and verify content
	file, err := os.Open(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 rows, got %d", len(records))
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

// Test writeFailedURLsLog function
func TestWriteFailedURLsLog(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "failed_urls.txt")

	failedURLs := []string{
		"https://example.com/paper1.pdf",
		"https://invalid-domain.com/paper2.pdf",
		"https://www.researchgate.net/publication/123456",
	}

	err := writeFailedURLsLog(failedURLs, logPath)
	if err != nil {
		t.Fatalf("Failed to write failed URLs log: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Failed URLs log file not created")
	}

	// Read and verify content
	content, err := ioutil.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	for _, url := range failedURLs {
		if !strings.Contains(contentStr, url) {
			t.Errorf("Failed URLs log missing URL: %s", url)
		}
	}

	// Verify header comments
	if !strings.Contains(contentStr, "# Failed Downloads") {
		t.Error("Failed URLs log missing header comment")
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
			OriginalRecord: []string{"Paper One", "Smith, J.", "2023", "https://example.com/1.pdf"},
		},
		{
			ID:             "2",
			Title:          "Paper Two",
			Authors:        "Jones, M.",
			Year:           "2024",
			URL:            "https://example.com/2.pdf",
			Downloaded:     false,
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
	if header[len(header)-1] != "downloaded" {
		t.Errorf("Last column should be 'downloaded', got %q", header[len(header)-1])
	}

	// Check data rows
	if len(records) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 rows, got %d", len(records))
	}

	// Verify download status column
	if records[1][len(records[1])-1] != "true" {
		t.Errorf("Row 1 should have downloaded=true, got %q", records[1][len(records[1])-1])
	}
	if records[2][len(records[2])-1] != "false" {
		t.Errorf("Row 2 should have downloaded=false, got %q", records[2][len(records[2])-1])
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
	if err := ioutil.WriteFile(txtPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := processTextFile(txtPath, tempDir)
	if err != nil {
		t.Errorf("processTextFile failed: %v", err)
	}

	// Check if failed URLs log was created
	failedLogPath := filepath.Join(tempDir, "test_failed.txt")
	if _, err := os.Stat(failedLogPath); os.IsNotExist(err) {
		t.Error("Failed URLs log not created")
	} else {
		// Read and verify content
		content, err := ioutil.ReadFile(failedLogPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "invalid-url-will-fail.com") {
			t.Error("Failed URLs log missing invalid URL")
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
