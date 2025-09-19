package filters

import (
	"math"
	"testing"
)

func TestCalculateTopicRelevance(t *testing.T) {
	tests := []struct {
		name           string
		manuscriptData map[string]string
		topics         []string
		weights        ScoreWeights
		minScore       float64
		expectRelevant bool
	}{
		{
			name: "Highly relevant AI healthcare manuscript",
			manuscriptData: map[string]string{
				"title":    "Deep Learning for Medical Image Analysis: A Comprehensive Review",
				"abstract": "This paper presents a comprehensive review of deep learning applications in medical image analysis. We explore various neural network architectures for diagnostic imaging, including convolutional neural networks for radiology and pathology. Machine learning techniques have shown remarkable success in healthcare applications.",
				"keywords": "deep learning, medical imaging, artificial intelligence, healthcare, diagnosis",
				"journal":  "Journal of Medical Artificial Intelligence",
			},
			topics: []string{
				"artificial intelligence in healthcare and medicine",
				"deep learning for medical image analysis",
				"machine learning for diagnostic imaging",
			},
			weights: ScoreWeights{
				KeywordMatch:   0.4,
				ConceptMatch:   0.4,
				FieldRelevance: 0.2,
			},
			minScore:       0.5,
			expectRelevant: true,
		},
		{
			name: "Irrelevant manuscript about agriculture",
			manuscriptData: map[string]string{
				"title":    "Soil Composition Analysis in Agricultural Fields",
				"abstract": "This study examines the soil composition and nutrient levels in various agricultural fields. We analyze the impact of different fertilizers on crop yield and soil health. The research focuses on sustainable farming practices.",
				"keywords": "agriculture, soil analysis, farming, crops",
				"journal":  "Agricultural Science Quarterly",
			},
			topics: []string{
				"artificial intelligence in healthcare",
				"machine learning for medical diagnosis",
			},
			weights: ScoreWeights{
				KeywordMatch:   0.4,
				ConceptMatch:   0.4,
				FieldRelevance: 0.2,
			},
			minScore:       0.5,
			expectRelevant: false,
		},
		{
			name: "Partially relevant AI manuscript",
			manuscriptData: map[string]string{
				"title":    "Machine Learning Applications: A Survey",
				"abstract": "This survey covers various applications of machine learning across different domains including finance, healthcare, and education. We briefly discuss medical applications but focus primarily on financial modeling.",
				"keywords": "machine learning, applications, survey",
				"journal":  "Computer Science Review",
			},
			topics: []string{
				"machine learning in healthcare",
				"artificial intelligence for medical diagnosis",
			},
			weights: ScoreWeights{
				KeywordMatch:   0.4,
				ConceptMatch:   0.4,
				FieldRelevance: 0.2,
			},
			minScore:       0.6,
			expectRelevant: false, // Should be below threshold due to limited healthcare focus
		},
		{
			name: "Empty manuscript data",
			manuscriptData: map[string]string{
				"title":    "",
				"abstract": "",
			},
			topics: []string{
				"artificial intelligence in healthcare",
			},
			weights: ScoreWeights{
				KeywordMatch:   0.4,
				ConceptMatch:   0.4,
				FieldRelevance: 0.2,
			},
			minScore:       0.5,
			expectRelevant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := CalculateTopicRelevance(tt.manuscriptData, tt.topics, tt.weights)
			if err != nil {
				t.Fatalf("CalculateTopicRelevance() error = %v", err)
			}

			if score == nil {
				t.Fatal("CalculateTopicRelevance() returned nil score")
			}

			// Check if the relevance decision matches expectation
			isRelevant := score.OverallScore >= tt.minScore
			if isRelevant != tt.expectRelevant {
				t.Errorf("CalculateTopicRelevance() relevance = %v (score: %.2f), expected %v",
					isRelevant, score.OverallScore, tt.expectRelevant)
			}

			// Validate score components
			if score.ComponentScores == nil {
				t.Error("ComponentScores should not be nil")
			}

			// Check that scores are within valid range [0, 1]
			if score.OverallScore < 0 || score.OverallScore > 1 {
				t.Errorf("OverallScore %.2f is out of valid range [0, 1]", score.OverallScore)
			}

			if score.Confidence < 0 || score.Confidence > 1 {
				t.Errorf("Confidence %.2f is out of valid range [0, 1]", score.Confidence)
			}

			// For highly relevant manuscripts, expect some keyword matches
			if tt.expectRelevant && len(score.MatchedKeywords) == 0 {
				t.Error("Expected some keyword matches for relevant manuscript")
			}
		})
	}
}

func TestProcessTopics(t *testing.T) {
	topics := []string{
		"machine learning in healthcare applications",
		"artificial intelligence for medical diagnosis",
		"deep learning models",
	}

	keywords, concepts := processTopics(topics)

	// Check that we extracted keywords
	if len(keywords) == 0 {
		t.Error("processTopics() should extract keywords")
	}

	// Check that we extracted concepts (multi-word phrases)
	if len(concepts) == 0 {
		t.Error("processTopics() should extract concepts")
	}

	// Check for specific expected keywords
	expectedKeywords := []string{"machine", "learning", "healthcare", "artificial", "intelligence"}
	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range keywords {
			if keyword == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected keyword '%s' not found", expected)
		}
	}

	// Check for duplicates in keywords
	seen := make(map[string]bool)
	for _, keyword := range keywords {
		if seen[keyword] {
			t.Errorf("Duplicate keyword found: %s", keyword)
		}
		seen[keyword] = true
	}
}

func TestCalculateKeywordScore(t *testing.T) {
	text := "This is a study about machine learning in healthcare. We use artificial intelligence for medical diagnosis."
	keywords := []string{"machine", "learning", "healthcare", "artificial", "intelligence", "medical", "diagnosis", "nonexistent"}

	score, matched := calculateKeywordScore(text, keywords)

	// Should match most keywords except "nonexistent"
	if len(matched) != 7 {
		t.Errorf("calculateKeywordScore() matched %d keywords, expected 7", len(matched))
	}

	// Score should be high but not perfect (7/8 keywords matched)
	expectedScore := 7.0 / 8.0 * 1.5 // With boost factor
	if expectedScore > 1.0 {
		expectedScore = 1.0
	}

	if math.Abs(score-expectedScore) > 0.1 {
		t.Errorf("calculateKeywordScore() score = %.2f, expected ~%.2f", score, expectedScore)
	}
}

func TestCalculateConceptScore(t *testing.T) {
	text := "We present a deep learning model for medical image analysis. Machine learning in healthcare has shown great promise."
	concepts := []string{
		"deep learning",
		"medical image",
		"machine learning",
		"learning in healthcare",
		"nonexistent phrase",
	}

	score, matched := calculateConceptScore(text, concepts)

	// Should match 4 out of 5 concepts
	if len(matched) != 4 {
		t.Errorf("calculateConceptScore() matched %d concepts, expected 4", len(matched))
	}

	// Check score calculation
	if score < 0.5 || score > 1.0 {
		t.Errorf("calculateConceptScore() score = %.2f, expected between 0.5 and 1.0", score)
	}
}

func TestCalculateFieldRelevanceScore(t *testing.T) {
	tests := []struct {
		name           string
		manuscriptData map[string]string
		topics         []string
		minScore       float64
		maxScore       float64
	}{
		{
			name: "Highly relevant journal",
			manuscriptData: map[string]string{
				"journal":      "Journal of Medical Artificial Intelligence",
				"field":        "Medical AI",
				"subject_area": "Healthcare Technology",
			},
			topics: []string{
				"artificial intelligence in healthcare",
				"medical AI applications",
			},
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			name: "Irrelevant journal",
			manuscriptData: map[string]string{
				"journal": "Agricultural Science Quarterly",
				"field":   "Agriculture",
			},
			topics: []string{
				"artificial intelligence in healthcare",
				"medical diagnosis",
			},
			minScore: 0.0,
			maxScore: 0.3,
		},
		{
			name:           "No field information",
			manuscriptData: map[string]string{},
			topics: []string{
				"machine learning",
			},
			minScore: 0.4,
			maxScore: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateFieldRelevanceScore(tt.manuscriptData, tt.topics)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateFieldRelevanceScore() = %.2f, expected between %.2f and %.2f",
					score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestExtractKeywordsFromText(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog in machine learning applications"
	keywords := extractKeywordsFromText(text)

	// Should exclude common stop words
	stopWords := []string{"the", "in"}
	for _, stopWord := range stopWords {
		for _, keyword := range keywords {
			if keyword == stopWord {
				t.Errorf("Stop word '%s' should not be included in keywords", stopWord)
			}
		}
	}

	// Should include content words
	expectedWords := []string{"quick", "brown", "fox", "jumps", "lazy", "dog", "machine", "learning", "applications"}
	for _, expected := range expectedWords {
		found := false
		for _, keyword := range keywords {
			if keyword == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected keyword '%s' not found", expected)
		}
	}
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		keywordMatches int
		conceptMatches int
		minConfidence  float64
		maxConfidence  float64
	}{
		{
			name:           "High confidence - long text with many matches",
			text:           generateLongText(500),
			keywordMatches: 8,
			conceptMatches: 5,
			minConfidence:  0.8,
			maxConfidence:  1.0,
		},
		{
			name:           "Low confidence - short text with few matches",
			text:           "This is a very short text.",
			keywordMatches: 1,
			conceptMatches: 0,
			minConfidence:  0.0,
			maxConfidence:  0.3,
		},
		{
			name:           "Medium confidence - moderate text and matches",
			text:           generateLongText(250),
			keywordMatches: 3,
			conceptMatches: 2,
			minConfidence:  0.4,
			maxConfidence:  0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := calculateConfidence(tt.text, tt.keywordMatches, tt.conceptMatches)

			if confidence < tt.minConfidence || confidence > tt.maxConfidence {
				t.Errorf("calculateConfidence() = %.2f, expected between %.2f and %.2f",
					confidence, tt.minConfidence, tt.maxConfidence)
			}

			if confidence < 0 || confidence > 1 {
				t.Errorf("Confidence %.2f is out of valid range [0, 1]", confidence)
			}
		})
	}
}

func TestBatchCalculateTopicRelevance(t *testing.T) {
	manuscripts := []map[string]string{
		{
			"title":    "AI in Healthcare",
			"abstract": "Machine learning for medical diagnosis using artificial intelligence",
		},
		{
			"title":    "Agriculture Study",
			"abstract": "Soil analysis and farming techniques",
		},
		{
			"title":    "Deep Learning for Radiology",
			"abstract": "Neural networks for medical imaging in healthcare applications using machine learning",
		},
	}

	config := TopicRelevanceConfig{
		Enabled: true,
		UseAI:   false,
		Topics: []string{
			"artificial intelligence in healthcare",
			"machine learning for medical applications",
		},
		MinScore: 0.4,
		ScoreWeights: ScoreWeights{
			KeywordMatch:   0.4,
			ConceptMatch:   0.4,
			FieldRelevance: 0.2,
		},
	}

	scores, err := BatchCalculateTopicRelevance(manuscripts, config, nil)
	if err != nil {
		t.Fatalf("BatchCalculateTopicRelevance() error = %v", err)
	}

	if len(scores) != len(manuscripts) {
		t.Errorf("BatchCalculateTopicRelevance() returned %d scores, expected %d",
			len(scores), len(manuscripts))
	}

	// First manuscript should be relevant
	if !scores[0].IsRelevant {
		t.Error("First manuscript should be relevant")
	}

	// Second manuscript should not be relevant
	if scores[1].IsRelevant {
		t.Error("Second manuscript should not be relevant")
	}

	// Third manuscript should be relevant
	if !scores[2].IsRelevant {
		t.Error("Third manuscript should be relevant")
	}
}

func TestEmptyTopics(t *testing.T) {
	manuscriptData := map[string]string{
		"title":    "Test Title",
		"abstract": "Test abstract",
	}

	_, err := CalculateTopicRelevance(manuscriptData, []string{}, ScoreWeights{})
	if err == nil {
		t.Error("CalculateTopicRelevance() should error with empty topics")
	}
}

// Helper function to generate text with specified word count
func generateLongText(wordCount int) string {
	words := []string{"machine", "learning", "artificial", "intelligence", "healthcare",
		"medical", "diagnosis", "treatment", "patient", "clinical", "research", "study",
		"analysis", "data", "model", "algorithm", "neural", "network", "deep", "system"}

	result := ""
	for i := 0; i < wordCount; i++ {
		result += words[i%len(words)] + " "
	}
	return result
}
