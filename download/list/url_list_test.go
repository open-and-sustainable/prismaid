package list

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with space", "with_space"},
		{"file/with/slashes", "file_with_slashes"},
		{"file:with:colons", "file_with_colons"},
		{"file<with>symbols?", "file_with_symbols_"},
		{"multi\nline\ttabs", "multi_line_tabs"},
		{"long..........name.with.dots", "long..........name.with.dots"}, // Dots are kept as-is
		{"", ""}, // Edge case: empty string
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := sanitizeFilename(tc.input)
			if result != tc.expected {
				t.Errorf("sanitizeFilename(%q) = %q; expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestExtractPDF(t *testing.T) {
	// Create a mock server for direct PDF
	pdfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("mock pdf content"))
	}))
	defer pdfServer.Close()

	// Create a mock server for HTML with PDF link
	htmlWithPDFServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
		<html>
			<head>
				<title>Test Page</title>
			</head>
			<body>
				<a href="https://example.com/test.pdf">Download PDF</a>
			</body>
		</html>`
		w.Write([]byte(html))
	}))
	defer htmlWithPDFServer.Close()

	// Create a mock server for HTML without PDF link
	htmlWithoutPDFServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
		<html>
			<head>
				<title>Test Page</title>
			</head>
			<body>
				<a href="https://example.com/test.html">Download HTML</a>
			</body>
		</html>`
		w.Write([]byte(html))
	}))
	defer htmlWithoutPDFServer.Close()

	// Create a mock server for relative PDF link
	htmlRelativePDFServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
		<html>
			<head>
				<title>Test Page</title>
			</head>
			<body>
				<a href="/relative/path/doc.pdf">Download PDF</a>
			</body>
		</html>`
		w.Write([]byte(html))
	}))
	defer htmlRelativePDFServer.Close()

	// Test cases
	testCases := []struct {
		name         string
		url          string
		expectPDFURL string
		expectError  bool
	}{
		{"Direct PDF", pdfServer.URL, pdfServer.URL, false},
		{"HTML with PDF link", htmlWithPDFServer.URL, "https://example.com/test.pdf", false},
		{"HTML without PDF link", htmlWithoutPDFServer.URL, "", false},
		{"HTML with relative PDF link", htmlRelativePDFServer.URL, htmlRelativePDFServer.URL + "/relative/path/doc.pdf", false},
		{"Invalid URL", "http://nonexistent.example.com", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pdfURL, filename, err := extractPDF(tc.url)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if pdfURL != tc.expectPDFURL {
				t.Errorf("Expected PDF URL %q, got %q", tc.expectPDFURL, pdfURL)
			}

			if !tc.expectError && filename == "" {
				t.Errorf("Expected non-empty filename")
			}
		})
	}
}

func TestDownloadPDF(t *testing.T) {
	// Create a mock server that serves a PDF
	pdfContent := "This is a mock PDF content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte(pdfContent))
	}))
	defer server.Close()

	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test successful download
	t.Run("Successful download", func(t *testing.T) {
		pdfPath := filepath.Join(tempDir, "test.pdf")

		err := downloadPDF(server.URL, pdfPath)
		if err != nil {
			t.Fatalf("downloadPDF failed: %v", err)
		}

		// Verify file exists and has correct content
		content, err := ioutil.ReadFile(pdfPath)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}

		if string(content) != pdfContent {
			t.Errorf("Downloaded content doesn't match. Got %q, expected %q", string(content), pdfContent)
		}
	})

	// Test invalid URL
	t.Run("Invalid URL", func(t *testing.T) {
		pdfPath := filepath.Join(tempDir, "nonexistent.pdf")

		err := downloadPDF("http://nonexistent.example.com", pdfPath)
		if err == nil {
			t.Errorf("Expected error with invalid URL, got none")
		}
	})

	// Test invalid path
	t.Run("Invalid path", func(t *testing.T) {
		pdfPath := filepath.Join(tempDir, "invalid/path/test.pdf")

		err := downloadPDF(server.URL, pdfPath)
		if err == nil {
			t.Errorf("Expected error with invalid path, got none")
		}
	})
}

func TestRunListDownload(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock server for PDF
	pdfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("mock pdf content"))
	}))
	defer pdfServer.Close()

	// Create a mock server for HTML with PDF link
	htmlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
		<html>
			<head>
				<title>HTML Test Page</title>
			</head>
			<body>
				<a href="https://example.com/document.pdf">Download PDF</a>
			</body>
		</html>`
		w.Write([]byte(html))
	}))
	defer htmlServer.Close()

	// Create URL list file
	urlListPath := filepath.Join(tempDir, "urls.txt")
	urlContent := `# Test URLs
${pdfServer}
${htmlServer}

# This is a comment
# Empty lines should be ignored
`
	// Replace placeholders with actual server URLs
	urlContent = strings.ReplaceAll(urlContent, "${pdfServer}", pdfServer.URL)
	urlContent = strings.ReplaceAll(urlContent, "${htmlServer}", htmlServer.URL)

	if err := ioutil.WriteFile(urlListPath, []byte(urlContent), 0644); err != nil {
		t.Fatalf("Failed to write URL list file: %v", err)
	}

	// Run the function being tested
	DownloadURLList(urlListPath)

	// Verify results
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	pdfCount := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".pdf" {
			pdfCount++
		}
	}

	// We expect the direct PDF to be downloaded
	if pdfCount < 1 {
		t.Errorf("Expected at least 1 PDF file, found %d", pdfCount)
	}
}

// Test for edge cases and error handling
func TestRunListDownloadEdgeCases(t *testing.T) {
	// Test with non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		// This should not panic
		DownloadURLList("/path/to/nonexistent/file.txt")
		// If we get here without panicking, the test passes
	})

	// Test with empty file
	t.Run("Empty file", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "download_test_empty")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		emptyFilePath := filepath.Join(tempDir, "empty.txt")
		if err := ioutil.WriteFile(emptyFilePath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to write empty file: %v", err)
		}

		// This should not panic or error
		DownloadURLList(emptyFilePath)
		// If we get here without panicking, the test passes
	})

	// Test with file containing only comments and empty lines
	t.Run("Comments only", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "download_test_comments")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		commentsFilePath := filepath.Join(tempDir, "comments.txt")
		commentsContent := `# This is a comment
# Another comment

# Yet another comment
`
		if err := ioutil.WriteFile(commentsFilePath, []byte(commentsContent), 0644); err != nil {
			t.Fatalf("Failed to write comments file: %v", err)
		}

		// This should not panic or error
		DownloadURLList(commentsFilePath)
		// If we get here without panicking, the test passes
	})
}
