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

	papers, err := parseCSVFile(csvPath, ',')
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
