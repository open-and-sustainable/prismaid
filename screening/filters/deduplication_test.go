package filters

import (
	"strings"
	"testing"
)

// TestDeduplication tests the deduplication functionality
func TestDeduplication(t *testing.T) {
	manuscripts := []ManuscriptData{
		{
			ID: "1",
			OriginalData: map[string]string{
				"title":    "Climate Change and Global Warming",
				"abstract": "This study examines the effects of climate change on global temperatures.",
			},
			Text: "This study examines the effects of climate change on global temperatures.",
		},
		{
			ID: "2",
			OriginalData: map[string]string{
				"title":    "Climate Change and Global Warming",
				"abstract": "This study examines the effects of climate change on global temperatures.",
			},
			Text: "This study examines the effects of climate change on global temperatures.",
		},
		{
			ID: "3",
			OriginalData: map[string]string{
				"title":    "Ocean Acidification Research",
				"abstract": "This research focuses on ocean pH levels and marine ecosystems.",
			},
			Text: "This research focuses on ocean pH levels and marine ecosystems.",
		},
	}

	// Test exact matching
	t.Run("Exact matching", func(t *testing.T) {
		config := DeduplicationConfig{
			UseAI:         false,
			CompareFields: []string{"title", "abstract"},
		}

		duplicates := FindDuplicates(manuscripts, config)

		// Check that manuscript 2 is marked as duplicate of manuscript 1
		if dup, exists := duplicates["2"]; exists {
			isDuplicate := dup[0].(bool)
			originalID := dup[1].(string)

			if !isDuplicate {
				t.Error("Manuscript 2 should be marked as duplicate")
			}
			if originalID != "1" {
				t.Errorf("Manuscript 2 should be duplicate of 1, got %s", originalID)
			}
		} else {
			t.Error("Manuscript 2 should exist in duplicates map")
		}

		// Check that manuscript 3 is not a duplicate
		if dup, exists := duplicates["3"]; exists {
			isDuplicate := dup[0].(bool)
			if isDuplicate {
				t.Error("Manuscript 3 should not be marked as duplicate")
			}
		}
	})

	// Test simple matching with single character tolerance
	t.Run("Simple matching", func(t *testing.T) {
		manuscriptsSimple := []ManuscriptData{
			{
				ID: "1",
				OriginalData: map[string]string{
					"title": "Climate Change and Global Warming",
				},
				Text: "Study on climate",
			},
			{
				ID: "2",
				OriginalData: map[string]string{
					"title": "Climate Change and Global Warming", // Exact match
				},
				Text: "Study on climate",
			},
			{
				ID: "3",
				OriginalData: map[string]string{
					"title": "Climate Changes and Global Warming", // Single char difference (Change -> Changes)
				},
				Text: "Study on climate",
			},
			{
				ID: "4",
				OriginalData: map[string]string{
					"title": "Completely Different Study",
				},
				Text: "Ocean research",
			},
		}

		config := DeduplicationConfig{
			UseAI:         false,
			CompareFields: []string{"title"},
		}

		duplicates := FindDuplicates(manuscriptsSimple, config)

		// Check that manuscript 2 is marked as duplicate of manuscript 1 (exact match)
		if dup, exists := duplicates["2"]; exists {
			isDuplicate := dup[0].(bool)
			if !isDuplicate {
				t.Error("Exact matching manuscripts should be detected")
			}
		}

		// Check that manuscript 3 is marked as duplicate (single char difference)
		if dup, exists := duplicates["3"]; exists {
			isDuplicate := dup[0].(bool)
			if !isDuplicate {
				t.Error("Manuscripts with single character difference should be detected")
			}
		}

		// Check that manuscript 4 is not similar
		if dup, exists := duplicates["4"]; exists {
			isDuplicate := dup[0].(bool)
			if isDuplicate {
				t.Error("Dissimilar manuscript should not be marked as duplicate")
			}
		}
	})
}

// TestBuildComparisonData tests the buildComparisonData function with improved formatting
func TestBuildComparisonData(t *testing.T) {
	manuscript := ManuscriptData{
		ID: "test1",
		OriginalData: map[string]string{
			"title":   "Climate Change Study",
			"authors": "Smith, J.; Johnson, K.",
			"year":    "2023",
			"doi":     "10.1234/example.doi",
		},
		LowerFieldMap: map[string]string{
			"title":   "title",
			"authors": "authors",
			"year":    "year",
			"doi":     "doi",
		},
	}

	compareFields := []string{"title", "authors", "year", "doi"}
	result := buildComparisonData(manuscript, compareFields)

	// Check that fields are formatted correctly (values are normalized to lowercase)
	if !strings.Contains(result, "TITLE: climate change study") {
		t.Errorf("Expected formatted TITLE field, got: %s", result)
	}
	if !strings.Contains(result, "AUTHORS: smith, j.; johnson, k.") {
		t.Errorf("Expected formatted AUTHORS field, got: %s", result)
	}
	if !strings.Contains(result, "YEAR: 2023") {
		t.Errorf("Expected formatted YEAR field, got: %s", result)
	}
	if !strings.Contains(result, "DOI: 10.1234/example.doi") {
		t.Errorf("Expected formatted DOI field, got: %s", result)
	}

	// Test with empty fields
	emptyManuscript := ManuscriptData{
		ID:            "test2",
		OriginalData:  map[string]string{},
		LowerFieldMap: map[string]string{},
	}
	emptyResult := buildComparisonData(emptyManuscript, compareFields)
	if emptyResult != "[No data available for comparison fields]" {
		t.Errorf("Expected empty data message, got: %s", emptyResult)
	}
}
