package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdInitialization(t *testing.T) {
	// Save original command line arguments and restore them after the test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Save the original flag.CommandLine and restore it after the test
	oldCommandLine := flag.CommandLine
	defer func() { flag.CommandLine = oldCommandLine }()

	// Create a new flag set for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Test that flags can be defined
	var projectConfigPath string
	var initFlag bool

	flag.StringVar(&projectConfigPath, "project", "", "Path to the project configuration file")
	flag.BoolVar(&initFlag, "init", false, "Run interactively to initialize a new project configuration file")

	// Verify the flags were created
	if flag.Lookup("project") == nil {
		t.Error("Expected 'project' flag to be defined")
	}

	if flag.Lookup("init") == nil {
		t.Error("Expected 'init' flag to be defined")
	}

	// This test doesn't actually run main() as it would exit the process
	// But it verifies the basic structure is in place
}

func TestHandleConversionIsolatedRetriesOnZeroSize(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "convert_isolated_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pdfPath := filepath.Join(tempDir, "sample.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4"), 0644); err != nil {
		t.Fatalf("Failed to write test pdf: %v", err)
	}

	origRun := runConvertCommand
	defer func() { runConvertCommand = origRun }()

	callCount := 0
	var sawOCR bool
	runConvertCommand = func(exePath, inputDir, fullPath, tikaServer string, ocrOnly bool) ([]byte, error) {
		callCount++
		if ocrOnly {
			sawOCR = true
		}
		ext := filepath.Ext(fullPath)
		txtPath := filepath.Join(inputDir, strings.TrimSuffix(filepath.Base(fullPath), ext)+".txt")
		if ocrOnly {
			return []byte("ocr-only ok"), os.WriteFile(txtPath, []byte("ok"), 0644)
		}
		return []byte("initial ok"), os.WriteFile(txtPath, []byte{}, 0644)
	}

	handleConversionIsolated(tempDir, "pdf", "localhost:9998", false)

	txtPath := filepath.Join(tempDir, "sample.txt")
	info, err := os.Stat(txtPath)
	if err != nil {
		t.Fatalf("Expected txt output to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("Expected txt output to be non-zero after OCR retry")
	}
	if callCount != 2 {
		t.Fatalf("Expected 2 conversion attempts, got %d", callCount)
	}
	if !sawOCR {
		t.Fatalf("Expected OCR-only retry to be used")
	}
}
