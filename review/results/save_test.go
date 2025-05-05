package results

import (
	"testing"
)

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
