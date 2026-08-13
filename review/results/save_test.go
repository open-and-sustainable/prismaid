package results

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/open-and-sustainable/alembica/definitions"
)

func TestSaveJSONMapsFilenamesBySequenceID(t *testing.T) {
	tmp := t.TempDir()
	outputPath := filepath.Join(tmp, "results.json")
	results := extractionAndJustificationResults(t)

	if err := saveJSON(outputPath, results, []string{"doc_a", "doc_b"}); err != nil {
		t.Fatalf("saveJSON failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var written []map[string]any
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(written) != 4 {
		t.Fatalf("expected 4 output objects, got %d", len(written))
	}

	expected := []struct {
		filename string
		field    string
		value    string
	}{
		{"doc_a", "extraction", "doc_a_extraction"},
		{"doc_a", "justification", "doc_a_justification"},
		{"doc_b", "extraction", "doc_b_extraction"},
		{"doc_b", "justification", "doc_b_justification"},
	}
	for i, want := range expected {
		if got := written[i]["filename"]; got != want.filename {
			t.Fatalf("object %d filename = %v, expected %s", i, got, want.filename)
		}
		if got := written[i][want.field]; got != want.value {
			t.Fatalf("object %d %s = %v, expected %s", i, want.field, got, want.value)
		}
	}
}

func TestSaveCSVMapsFilenamesBySequenceID(t *testing.T) {
	tmp := t.TempDir()
	outputPath := filepath.Join(tmp, "results.csv")
	results := extractionAndJustificationResults(t)

	if err := saveCSV(outputPath, results, []string{"doc_a", "doc_b"}, []string{"extraction"}); err != nil {
		t.Fatalf("saveCSV failed: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer file.Close()

	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	expected := [][]string{
		{"Provider", "Model", "File Name", "extraction"},
		{"provider", "model", "doc_a", "doc_a_extraction"},
		{"provider", "model", "doc_b", "doc_b_extraction"},
	}
	if len(rows) != len(expected) {
		t.Fatalf("expected %d rows, got %d: %#v", len(expected), len(rows), rows)
	}
	for i := range expected {
		for j := range expected[i] {
			if rows[i][j] != expected[i][j] {
				t.Fatalf("row %d col %d = %q, expected %q", i, j, rows[i][j], expected[i][j])
			}
		}
	}
}

func extractionAndJustificationResults(t *testing.T) string {
	t.Helper()
	output := definitions.Output{
		Responses: []definitions.Response{
			{
				Provider:       "provider",
				Model:          "model",
				SequenceID:     "1",
				SequenceNumber: 1,
				ModelResponses: []string{`{"extraction":"doc_a_extraction"}`},
			},
			{
				Provider:       "provider",
				Model:          "model",
				SequenceID:     "1",
				SequenceNumber: 2,
				ModelResponses: []string{`{"justification":"doc_a_justification"}`},
			},
			{
				Provider:       "provider",
				Model:          "model",
				SequenceID:     "2",
				SequenceNumber: 1,
				ModelResponses: []string{`{"extraction":"doc_b_extraction"}`},
			},
			{
				Provider:       "provider",
				Model:          "model",
				SequenceID:     "2",
				SequenceNumber: 2,
				ModelResponses: []string{`{"justification":"doc_b_justification"}`},
			},
		},
	}
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshal test output: %v", err)
	}
	return string(data)
}

// TestGetDirectoryPath tests the directory path extraction logic
func TestGetDirectoryPath(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Current directory",
			input:    "results.json",
			expected: "",
		},
		{
			name:     "Subdirectory",
			input:    "output/results.json",
			expected: "output",
		},
		{
			name:     "Nested subdirectory",
			input:    "data/output/results.json",
			expected: "data/output",
		},
		{
			name:     "Just filename with dot prefix",
			input:    "./results.json",
			expected: "",
		},
		{
			name:     "Absolute path",
			input:    "/tmp/results.json",
			expected: "/tmp",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetDirectoryPath(tc.input)
			if result != tc.expected {
				t.Errorf("GetDirectoryPath(%s) = %s, expected %s",
					tc.input, result, tc.expected)
			}
		})
	}
}
