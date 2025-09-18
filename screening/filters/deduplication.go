package filters

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// ManuscriptData represents the data structure for a manuscript
type ManuscriptData struct {
	ID           string
	OriginalData map[string]string
	Text         string
}

// DeduplicationConfig represents the configuration for deduplication
type DeduplicationConfig struct {
	Method        string
	Threshold     float64
	CompareFields []string
}

// FindDuplicates identifies duplicate manuscripts based on the configuration
// Returns a map where key is manuscript ID and value is a tuple (isDuplicate, originalID)
func FindDuplicates(records []ManuscriptData, config DeduplicationConfig) map[string][2]interface{} {
	switch config.Method {
	case "exact":
		return findExactDuplicates(records, config.CompareFields)
	case "fuzzy":
		return findFuzzyDuplicates(records, config.CompareFields, config.Threshold)
	case "semantic":
		// Semantic matching would require AI/embeddings - placeholder for now
		return findSemanticDuplicates(records, config.CompareFields, config.Threshold)
	default:
		return findExactDuplicates(records, config.CompareFields)
	}
}

// findExactDuplicates finds exact matches based on specified fields
func findExactDuplicates(manuscripts []ManuscriptData, compareFields []string) map[string][2]interface{} {
	duplicates := make(map[string][2]interface{})
	hashMap := make(map[string]string) // hash -> first occurrence ID

	for _, manuscript := range manuscripts {
		// Create a hash of the comparison fields
		hash := computeHash(manuscript, compareFields)

		if originalID, exists := hashMap[hash]; exists {
			// This is a duplicate
			duplicates[manuscript.ID] = [2]interface{}{true, originalID}
		} else {
			// First occurrence
			hashMap[hash] = manuscript.ID
			duplicates[manuscript.ID] = [2]interface{}{false, ""}
		}
	}

	return duplicates
}

// findFuzzyDuplicates finds similar matches using string similarity algorithms
func findFuzzyDuplicates(manuscripts []ManuscriptData, compareFields []string, threshold float64) map[string][2]interface{} {
	duplicates := make(map[string][2]interface{})

	// Initialize all as non-duplicates
	for _, manuscript := range manuscripts {
		duplicates[manuscript.ID] = [2]interface{}{false, ""}
	}

	// Compare each manuscript with others
	for i := 0; i < len(manuscripts); i++ {
		for j := i + 1; j < len(manuscripts); j++ {
			similarity := calculateSimilarity(manuscripts[i], manuscripts[j], compareFields)

			if similarity >= threshold {
				// Mark the second one as duplicate of the first
				if !duplicates[manuscripts[j].ID][0].(bool) {
					duplicates[manuscripts[j].ID] = [2]interface{}{true, manuscripts[i].ID}
				}
			}
		}
	}

	return duplicates
}

// findSemanticDuplicates would use embeddings for semantic similarity
// This is a placeholder implementation - actual implementation would require AI integration
func findSemanticDuplicates(manuscripts []ManuscriptData, compareFields []string, threshold float64) map[string][2]interface{} {
	// For now, fall back to fuzzy matching
	// In a full implementation, this would:
	// 1. Generate embeddings for each manuscript
	// 2. Calculate cosine similarity between embeddings
	// 3. Mark as duplicates if similarity > threshold
	return findFuzzyDuplicates(manuscripts, compareFields, threshold)
}

// computeHash creates a hash from specified fields
func computeHash(manuscript ManuscriptData, compareFields []string) string {
	h := sha256.New()

	if len(compareFields) == 0 {
		// If no fields specified, use the full text
		h.Write([]byte(normalizeText(manuscript.Text)))
	} else {
		// Use specified fields
		for _, field := range compareFields {
			if field == "text" {
				h.Write([]byte(normalizeText(manuscript.Text)))
			} else if value, exists := manuscript.OriginalData[field]; exists {
				h.Write([]byte(normalizeText(value)))
			}
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// calculateSimilarity calculates similarity between two manuscripts
func calculateSimilarity(m1, m2 ManuscriptData, compareFields []string) float64 {
	// Get comparison strings
	str1 := getComparisonString(m1, compareFields)
	str2 := getComparisonString(m2, compareFields)

	// Use Jaccard similarity for now
	return jaccardSimilarity(str1, str2)
}

// getComparisonString builds a string from specified fields
func getComparisonString(manuscript ManuscriptData, compareFields []string) string {
	var parts []string

	if len(compareFields) == 0 {
		return normalizeText(manuscript.Text)
	}

	for _, field := range compareFields {
		if field == "text" {
			parts = append(parts, normalizeText(manuscript.Text))
		} else if value, exists := manuscript.OriginalData[field]; exists {
			parts = append(parts, normalizeText(value))
		}
	}

	return strings.Join(parts, " ")
}

// normalizeText normalizes text for comparison
func normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")
	return text
}

// jaccardSimilarity calculates Jaccard similarity between two strings
func jaccardSimilarity(str1, str2 string) float64 {
	// Tokenize strings into words
	words1 := strings.Fields(str1)
	words2 := strings.Fields(str2)

	// Create sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}
	for _, word := range words2 {
		set2[word] = true
	}

	// Calculate intersection and union
	intersection := 0
	union := make(map[string]bool)

	for word := range set1 {
		union[word] = true
		if set2[word] {
			intersection++
		}
	}
	for word := range set2 {
		union[word] = true
	}

	if len(union) == 0 {
		return 0.0
	}

	return float64(intersection) / float64(len(union))
}

// LevenshteinDistance calculates the edit distance between two strings
func LevenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first column and row
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// NormalizedLevenshteinSimilarity returns similarity score between 0 and 1
func NormalizedLevenshteinSimilarity(s1, s2 string) float64 {
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 1.0
	}

	distance := LevenshteinDistance(s1, s2)
	return 1.0 - float64(distance)/float64(maxLen)
}
