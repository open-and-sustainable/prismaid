package ocr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadWithTika tests the ReadWithTika function with a mock Tika server
func TestReadWithTika(t *testing.T) {
	// Create a mock Tika server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		if r.Header.Get("Accept") != "text/plain" {
			t.Errorf("Expected Accept header 'text/plain', got %s", r.Header.Get("Accept"))
		}

		// Return mock extracted text
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This is extracted text from the document with OCR"))
	}))
	defer mockServer.Close()

	// Create a temporary test file
	tempDir, err := os.MkdirTemp("", "tika_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "test.pdf")
	err = os.WriteFile(testFilePath, []byte("Mock PDF content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Extract address from mock server URL (remove http://)
	serverAddr := strings.TrimPrefix(mockServer.URL, "http://")

	// Test ReadWithTika
	result, err := ReadWithTika(testFilePath, serverAddr)
	if err != nil {
		t.Errorf("ReadWithTika returned error: %v", err)
	}

	expectedText := "This is extracted text from the document with OCR"
	if result != expectedText {
		t.Errorf("Expected text '%s', got '%s'", expectedText, result)
	}
}

// TestReadWithTikaServerError tests error handling when Tika server returns an error
func TestReadWithTikaServerError(t *testing.T) {
	// Create a mock Tika server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer mockServer.Close()

	// Create a temporary test file
	tempDir, err := os.MkdirTemp("", "tika_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "test.pdf")
	err = os.WriteFile(testFilePath, []byte("Mock PDF content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	serverAddr := strings.TrimPrefix(mockServer.URL, "http://")

	// Test that error is returned
	_, err = ReadWithTika(testFilePath, serverAddr)
	if err == nil {
		t.Error("Expected error when Tika server returns error status, got nil")
	}
}

// TestReadWithTikaInvalidFile tests error handling for non-existent files
func TestReadWithTikaInvalidFile(t *testing.T) {
	// Test with non-existent file
	_, err := ReadWithTika("/nonexistent/file.pdf", "localhost:9998")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestReadWithTikaServerUnavailable tests error handling when server is unreachable
func TestReadWithTikaServerUnavailable(t *testing.T) {
	// Create a temporary test file
	tempDir, err := os.MkdirTemp("", "tika_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "test.pdf")
	err = os.WriteFile(testFilePath, []byte("Mock PDF content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Use an invalid address
	_, err = ReadWithTika(testFilePath, "localhost:99999")
	if err == nil {
		t.Error("Expected error when server is unavailable, got nil")
	}
}

// TestIsTikaAvailable tests the server availability check
func TestIsTikaAvailable(t *testing.T) {
	// Create a mock Tika server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer mockServer.Close()

	serverAddr := strings.TrimPrefix(mockServer.URL, "http://")

	// Test that server is detected as available
	available := IsTikaAvailable(serverAddr)
	if !available {
		t.Error("Expected Tika server to be available")
	}

	// Test with unavailable server
	unavailable := IsTikaAvailable("localhost:99999")
	if unavailable {
		t.Error("Expected unavailable server to return false")
	}
}

// TestIsTikaAvailableWithNoContent tests server availability with 204 No Content response
func TestIsTikaAvailableWithNoContent(t *testing.T) {
	// Create a mock Tika server that returns 204 No Content
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer mockServer.Close()

	serverAddr := strings.TrimPrefix(mockServer.URL, "http://")

	// Test that server is detected as available
	available := IsTikaAvailable(serverAddr)
	if !available {
		t.Error("Expected Tika server to be available with 204 response")
	}
}

// TestReadWithTikaIntegration is a manual integration test (skipped by default)
// To run: go test -v -run TestReadWithTikaIntegration
// Requires a running Tika server at localhost:9998
func TestReadWithTikaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if Tika server is available
	if !IsTikaAvailable("localhost:9998") {
		t.Skip("Tika server not available at localhost:9998, skipping integration test")
	}

	// Create a simple test file
	tempDir, err := os.MkdirTemp("", "tika_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple text file for testing
	testFilePath := filepath.Join(tempDir, "test.txt")
	testContent := "This is a test document for Tika OCR integration testing."
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test extraction
	result, err := ReadWithTika(testFilePath, "localhost:9998")
	if err != nil {
		t.Errorf("Integration test failed: %v", err)
	}

	// The result should contain the test content
	if !strings.Contains(result, "test document") {
		t.Errorf("Expected result to contain 'test document', got: %s", result)
	}

	fmt.Printf("Integration test successful. Extracted text length: %d\n", len(result))
}
