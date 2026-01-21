package filters

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/alembica/extraction"
	"github.com/open-and-sustainable/alembica/utils/logger"
)

// TopicRelevanceConfig represents configuration for topic relevance filtering
type TopicRelevanceConfig struct {
	Enabled      bool         `toml:"enabled"`
	UseAI        bool         `toml:"use_ai"`
	Topics       []string     `toml:"topics"`        // List of topic descriptions
	MinScore     float64      `toml:"min_score"`     // Minimum score (0-1) to include
	ScoreWeights ScoreWeights `toml:"score_weights"` // Weights for different scoring components
}

// ScoreWeights defines the weights for different scoring components
type ScoreWeights struct {
	KeywordMatch   float64 `toml:"keyword_match"`   // Weight for keyword matching
	ConceptMatch   float64 `toml:"concept_match"`   // Weight for concept matching
	FieldRelevance float64 `toml:"field_relevance"` // Weight for field/domain relevance
}

// TopicRelevanceScore represents the relevance score for a manuscript
type TopicRelevanceScore struct {
	OverallScore    float64            `json:"overall_score"`    // Combined score (0-1)
	ComponentScores map[string]float64 `json:"component_scores"` // Individual component scores
	MatchedKeywords []string           `json:"matched_keywords"` // Keywords that matched
	MatchedConcepts []string           `json:"matched_concepts"` // Concepts that matched
	Confidence      float64            `json:"confidence"`       // Confidence in the score
	IsRelevant      bool               `json:"is_relevant"`      // Whether manuscript is relevant
}

// CalculateTopicRelevance calculates relevance score without AI
func CalculateTopicRelevance(manuscriptData map[string]string, topics []string, weights ScoreWeights) (*TopicRelevanceScore, error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("no topics provided for relevance calculation")
	}

	// Normalize weights if not provided
	if weights.KeywordMatch == 0 && weights.ConceptMatch == 0 && weights.FieldRelevance == 0 {
		weights = ScoreWeights{
			KeywordMatch:   0.4,
			ConceptMatch:   0.4,
			FieldRelevance: 0.2,
		}
	}

	// Extract text from manuscript
	text := extractRelevantText(manuscriptData)
	if text == "" {
		return &TopicRelevanceScore{
			OverallScore: 0,
			ComponentScores: map[string]float64{
				"keyword_match":   0,
				"concept_match":   0,
				"field_relevance": 0,
			},
			MatchedKeywords: []string{},
			MatchedConcepts: []string{},
			IsRelevant:      false,
			Confidence:      0.5,
		}, nil
	}

	// Process topics to extract keywords and concepts
	topicKeywords, topicConcepts := processTopics(topics)

	// Calculate component scores
	keywordScore, matchedKeywords := calculateKeywordScore(text, topicKeywords)
	conceptScore, matchedConcepts := calculateConceptScore(text, topicConcepts)
	fieldScore := calculateFieldRelevanceScore(manuscriptData, topics)

	// Calculate weighted overall score
	totalWeight := weights.KeywordMatch + weights.ConceptMatch + weights.FieldRelevance
	overallScore := (keywordScore*weights.KeywordMatch +
		conceptScore*weights.ConceptMatch +
		fieldScore*weights.FieldRelevance) / totalWeight

	// Calculate confidence based on amount of text and matches
	confidence := calculateConfidence(text, len(matchedKeywords), len(matchedConcepts))

	return &TopicRelevanceScore{
		OverallScore: overallScore,
		ComponentScores: map[string]float64{
			"keyword_match":   keywordScore,
			"concept_match":   conceptScore,
			"field_relevance": fieldScore,
		},
		MatchedKeywords: matchedKeywords,
		MatchedConcepts: matchedConcepts,
		Confidence:      confidence,
		IsRelevant:      overallScore >= 0.5, // Default threshold
	}, nil
}

// BatchCalculateTopicRelevanceWithAI processes multiple manuscripts in a single AI call
func BatchCalculateTopicRelevanceWithAI(manuscriptsData []map[string]string, topics []string, minScore float64, llmConfigs []any) map[string]*TopicRelevanceScore {
	results := make(map[string]*TopicRelevanceScore)

	// Initialize all results with zero scores
	for i := range manuscriptsData {
		results[fmt.Sprintf("%d", i)] = &TopicRelevanceScore{
			OverallScore:    0.0,
			ComponentScores: make(map[string]float64),
			MatchedKeywords: []string{},
			MatchedConcepts: []string{},
			Confidence:      0.0,
			IsRelevant:      false,
		}
	}

	if len(topics) == 0 {
		logger.Info("No topics provided for relevance calculation")
		return results
	}

	// Prepare AI model configurations
	var models []definitions.Model
	for _, llmConfig := range llmConfigs {
		if llm, ok := llmConfig.(map[string]any); ok {
			model := definitions.Model{
				Provider:     getStringValue(llm, "provider"),
				APIKey:       getStringValue(llm, "api_key"),
				Model:        getStringValue(llm, "model"),
				Temperature:  getFloatValue(llm, "temperature"),
				TPMLimit:     getIntValue(llm, "tpm_limit"),
				RPMLimit:     getIntValue(llm, "rpm_limit"),
				BaseURL:      getStringValue(llm, "base_url"),
				EndpointType: getStringValue(llm, "endpoint_type"),
				Region:       getStringValue(llm, "region"),
				ProjectID:    getStringValue(llm, "project_id"),
				Location:     getStringValue(llm, "location"),
				APIVersion:   getStringValue(llm, "api_version"),
			}
			models = append(models, model)
		}
	}

	if len(models) == 0 {
		logger.Info("No valid AI models configured for topic relevance")
		return results
	}

	// Build all prompts
	var prompts []definitions.Prompt
	validIndices := []int{}
	topicsStr := strings.Join(topics, "\n- ")

	for idx, manuscriptData := range manuscriptsData {
		// Build the data string for AI analysis
		dataStr := buildTopicRelevanceData(manuscriptData)

		// Skip if no data to analyze
		if dataStr == "" || dataStr == "[No relevant data available]" {
			continue
		}

		// Create the prompt
		prompt := fmt.Sprintf(`You are an expert in academic manuscript screening. Your task is to evaluate whether a manuscript is relevant to specific research topics.

TOPICS OF INTEREST:
- %s

MANUSCRIPT DATA:
%s

TASK: Evaluate the relevance of this manuscript to the specified topics.

Analyze the manuscript considering:
1. Direct keyword matches with the topics
2. Conceptual alignment with the research areas
3. Field/domain relevance
4. Methodological relevance
5. Research questions and objectives alignment

Respond with a JSON object containing:
{
  "overall_score": 0.75,  // Score from 0.0 to 1.0
  "component_scores": {
    "keyword_match": 0.8,
    "concept_match": 0.7,
    "field_relevance": 0.75
  },
  "matched_keywords": ["keyword1", "keyword2"],
  "matched_concepts": ["concept1", "concept2"],
  "confidence": 0.85,  // Confidence in the assessment (0-1)
  "is_relevant": true,  // Boolean decision
  "reasoning": "Brief explanation of the relevance assessment"
}`, topicsStr, dataStr)

		prompts = append(prompts, definitions.Prompt{
			PromptContent:  prompt,
			SequenceID:     fmt.Sprintf("%d", idx+1),
			SequenceNumber: idx + 1,
		})
		validIndices = append(validIndices, idx)
	}

	if len(prompts) == 0 {
		logger.Info("No valid manuscripts for topic relevance assessment")
		return results
	}

	logger.Info("Prepared %d manuscripts for batch topic relevance assessment", len(prompts))

	// Prepare the input for alembica with all prompts
	input := definitions.Input{
		Metadata: definitions.InputMetadata{
			Version:       "1.0",
			SchemaVersion: "1.0",
		},
		Models:  models,
		Prompts: prompts,
	}

	// Convert to JSON
	jsonInput, err := json.Marshal(input)
	if err != nil {
		logger.Error("Failed to marshal input for AI: %v", err)
		return results
	}

	// Call alembica once with all prompts
	logger.Info("Calling AI model with batch of %d topic relevance requests", len(prompts))
	result, err := extraction.Extract(string(jsonInput))
	if err != nil {
		logger.Error("AI extraction failed: %v", err)
		return results
	}

	// Parse the response
	var output definitions.Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		logger.Error("Failed to parse AI response: %v", err)
		return results
	}

	// Process responses
	for respIdx, manuscriptIdx := range validIndices {
		if respIdx < len(output.Responses) && len(output.Responses[respIdx].ModelResponses) > 0 {
			response := output.Responses[respIdx].ModelResponses[0]

			// Try to parse JSON response
			var relevanceResponse TopicRelevanceScore
			if err := json.Unmarshal([]byte(response), &relevanceResponse); err != nil {
				logger.Error("Failed to parse relevance response for manuscript %d: %v", manuscriptIdx, err)
				continue
			}

			// Update IsRelevant based on minScore
			relevanceResponse.IsRelevant = relevanceResponse.OverallScore >= minScore

			results[fmt.Sprintf("%d", manuscriptIdx)] = &relevanceResponse
		}
	}

	return results
}

// CalculateTopicRelevanceWithAI uses AI to calculate topic relevance
func CalculateTopicRelevanceWithAI(manuscriptData map[string]string, topics []string, llmConfigs []any) (*TopicRelevanceScore, error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("no topics provided for relevance calculation")
	}

	// Prepare AI model configurations
	var models []definitions.Model
	for _, llmConfig := range llmConfigs {
		if llm, ok := llmConfig.(map[string]interface{}); ok {
			model := definitions.Model{
				Provider:     getStringValue(llm, "provider"),
				APIKey:       getStringValue(llm, "api_key"),
				Model:        getStringValue(llm, "model"),
				Temperature:  getFloatValue(llm, "temperature"),
				TPMLimit:     getIntValue(llm, "tpm_limit"),
				RPMLimit:     getIntValue(llm, "rpm_limit"),
				BaseURL:      getStringValue(llm, "base_url"),
				EndpointType: getStringValue(llm, "endpoint_type"),
				Region:       getStringValue(llm, "region"),
				ProjectID:    getStringValue(llm, "project_id"),
				Location:     getStringValue(llm, "location"),
				APIVersion:   getStringValue(llm, "api_version"),
			}
			models = append(models, model)
		}
	}

	if len(models) == 0 {
		logger.Info("No valid AI models configured, falling back to non-AI relevance calculation")
		return CalculateTopicRelevance(manuscriptData, topics, ScoreWeights{
			KeywordMatch:   0.4,
			ConceptMatch:   0.4,
			FieldRelevance: 0.2,
		})
	}

	// Build the data string for AI analysis
	dataStr := buildTopicRelevanceData(manuscriptData)
	topicsStr := strings.Join(topics, "\n- ")

	// Create the prompt
	prompt := fmt.Sprintf(`You are an expert in academic manuscript screening. Your task is to evaluate whether a manuscript is relevant to specific research topics.

TOPICS OF INTEREST:
- %s

MANUSCRIPT DATA:
%s

TASK: Evaluate the relevance of this manuscript to the specified topics.

Analyze the manuscript considering:
1. Direct keyword matches with the topics
2. Conceptual alignment with the research areas
3. Field/domain relevance
4. Methodological relevance
5. Research questions and objectives alignment

Respond with a JSON object containing:
{
  "overall_score": 0.75,  // Score from 0.0 to 1.0
  "component_scores": {
    "keyword_match": 0.8,
    "concept_match": 0.7,
    "field_relevance": 0.75
  },
  "matched_keywords": ["keyword1", "keyword2"],
  "matched_concepts": ["concept1", "concept2"],
  "confidence": 0.85,  // Confidence in the assessment (0-1)
  "is_relevant": true,  // Boolean decision
  "reasoning": "Brief explanation of the relevance assessment"
}`, topicsStr, dataStr)

	// Prepare the input for alembica
	input := definitions.Input{
		Metadata: definitions.InputMetadata{
			Version:       "1.0",
			SchemaVersion: "1.0",
		},
		Models: models,
		Prompts: []definitions.Prompt{
			{
				PromptContent:  prompt,
				SequenceID:     "1",
				SequenceNumber: 1,
			},
		},
	}

	// Convert to JSON
	jsonInput, err := json.Marshal(input)
	if err != nil {
		logger.Error("Failed to marshal input for AI: %v", err)
		return nil, err
	}

	// Call alembica
	logger.Info("Calling AI model for topic relevance assessment")
	result, err := extraction.Extract(string(jsonInput))
	if err != nil {
		logger.Error("AI extraction failed: %v", err)
		// Fall back to non-AI method
		return CalculateTopicRelevance(manuscriptData, topics, ScoreWeights{
			KeywordMatch:   0.4,
			ConceptMatch:   0.4,
			FieldRelevance: 0.2,
		})
	}

	// Parse the response
	var output definitions.Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		logger.Error("Failed to parse AI response: %v", err)
		// Fall back to non-AI method
		return CalculateTopicRelevance(manuscriptData, topics, ScoreWeights{
			KeywordMatch:   0.4,
			ConceptMatch:   0.4,
			FieldRelevance: 0.2,
		})
	}

	// Extract relevance score from the response
	if len(output.Responses) > 0 && len(output.Responses[0].ModelResponses) > 0 {
		response := output.Responses[0].ModelResponses[0]

		// Try to parse JSON response
		var relevanceResponse TopicRelevanceScore
		if err := json.Unmarshal([]byte(response), &relevanceResponse); err != nil {
			logger.Error("Failed to parse relevance response: %v", err)
			// Fall back to non-AI method
			return CalculateTopicRelevance(manuscriptData, topics, ScoreWeights{
				KeywordMatch:   0.4,
				ConceptMatch:   0.4,
				FieldRelevance: 0.2,
			})
		}

		return &relevanceResponse, nil
	}

	// If we couldn't get a valid response, fall back to non-AI method
	logger.Info("Could not extract valid relevance score from AI response, falling back to non-AI method")
	return CalculateTopicRelevance(manuscriptData, topics, ScoreWeights{
		KeywordMatch:   0.4,
		ConceptMatch:   0.4,
		FieldRelevance: 0.2,
	})
}

// extractRelevantText extracts and combines relevant text fields from manuscript data
func extractRelevantText(manuscriptData map[string]string) string {
	var textParts []string

	// Priority fields for relevance assessment
	relevantFields := []string{
		"title", "abstract", "keywords", "subject", "research_area",
		"methodology", "objectives", "summary", "introduction",
	}

	for _, field := range relevantFields {
		// Case-insensitive field lookup
		for key, value := range manuscriptData {
			if strings.EqualFold(key, field) && value != "" {
				textParts = append(textParts, value)
				break
			}
		}
	}

	// Also check for any field containing "keyword" or "subject"
	for key, value := range manuscriptData {
		keyLower := strings.ToLower(key)
		if (strings.Contains(keyLower, "keyword") ||
			strings.Contains(keyLower, "subject") ||
			strings.Contains(keyLower, "topic")) && value != "" {
			// Avoid duplicates
			isDuplicate := false
			for _, part := range textParts {
				if part == value {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				textParts = append(textParts, value)
			}
		}
	}

	return strings.Join(textParts, " ")
}

// processTopics extracts keywords and concepts from topic descriptions
func processTopics(topics []string) ([]string, []string) {
	var keywords []string
	var concepts []string

	for _, topic := range topics {
		// Normalize the topic
		topic = strings.ToLower(strings.TrimSpace(topic))

		// Extract individual words as keywords (excluding common stop words)
		words := extractKeywordsFromText(topic)
		keywords = append(keywords, words...)

		// Extract phrases as concepts (2-4 word combinations)
		phrases := extractConceptsFromText(topic)
		concepts = append(concepts, phrases...)
	}

	// Remove duplicates
	keywords = removeDuplicates(keywords)
	concepts = removeDuplicates(concepts)

	return keywords, concepts
}

// extractKeywordsFromText extracts meaningful keywords from text
func extractKeywordsFromText(text string) []string {
	// Common stop words to exclude
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
		"are": true, "were": true, "been": true, "be": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true, "must": true,
		"can": true, "this": true, "that": true, "these": true, "those": true,
		"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
	}

	// Clean and split text
	re := regexp.MustCompile(`[^a-z0-9\s-]`)
	cleanText := re.ReplaceAllString(strings.ToLower(text), "")
	words := strings.Fields(cleanText)

	var keywords []string
	for _, word := range words {
		if len(word) > 2 && !stopWords[word] { // Skip short words and stop words
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// extractConceptsFromText extracts multi-word concepts from text
func extractConceptsFromText(text string) []string {
	var concepts []string

	// Clean text
	re := regexp.MustCompile(`[^a-z0-9\s-]`)
	cleanText := re.ReplaceAllString(strings.ToLower(text), "")
	words := strings.Fields(cleanText)

	// Extract 2-word phrases
	for i := 0; i < len(words)-1; i++ {
		phrase := words[i] + " " + words[i+1]
		concepts = append(concepts, phrase)
	}

	// Extract 3-word phrases
	for i := 0; i < len(words)-2; i++ {
		phrase := words[i] + " " + words[i+1] + " " + words[i+2]
		concepts = append(concepts, phrase)
	}

	return concepts
}

// calculateKeywordScore calculates score based on keyword matches
func calculateKeywordScore(text string, keywords []string) (float64, []string) {
	if len(keywords) == 0 {
		return 0, []string{}
	}

	textLower := strings.ToLower(text)
	var matchedKeywords []string
	matchCount := 0

	for _, keyword := range keywords {
		// Check for whole word match
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		if matched, _ := regexp.MatchString(pattern, textLower); matched {
			matchCount++
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}

	// Calculate score with diminishing returns for many matches
	score := float64(matchCount) / float64(len(keywords))
	score = math.Min(1.0, score*1.5) // Boost score slightly but cap at 1.0

	return score, matchedKeywords
}

// calculateConceptScore calculates score based on concept matches
func calculateConceptScore(text string, concepts []string) (float64, []string) {
	if len(concepts) == 0 {
		return 0, []string{}
	}

	textLower := strings.ToLower(text)
	var matchedConcepts []string
	matchCount := 0

	for _, concept := range concepts {
		if strings.Contains(textLower, concept) {
			matchCount++
			matchedConcepts = append(matchedConcepts, concept)
		}
	}

	// Calculate score
	score := float64(matchCount) / float64(len(concepts))
	score = math.Min(1.0, score*1.5) // Boost score slightly but cap at 1.0

	return score, matchedConcepts
}

// calculateFieldRelevanceScore calculates relevance based on journal/field information
func calculateFieldRelevanceScore(manuscriptData map[string]string, topics []string) float64 {
	// Extract field-related information
	fieldIndicators := []string{"journal", "field", "discipline", "category", "subject_area"}
	var fieldText string

	for _, indicator := range fieldIndicators {
		for key, value := range manuscriptData {
			if strings.Contains(strings.ToLower(key), indicator) && value != "" {
				fieldText += " " + value
			}
		}
	}

	if fieldText == "" {
		return 0.5 // Neutral score if no field information
	}

	// Check for topic-related terms in field information
	fieldTextLower := strings.ToLower(fieldText)
	matchCount := 0

	for _, topic := range topics {
		topicKeywords := extractKeywordsFromText(topic)
		for _, keyword := range topicKeywords {
			if strings.Contains(fieldTextLower, keyword) {
				matchCount++
			}
		}
	}

	// Calculate score based on matches
	if matchCount == 0 {
		return 0.2 // Low score for no field matches
	} else if matchCount <= 2 {
		return 0.5
	} else if matchCount <= 5 {
		return 0.7
	} else {
		return 0.9
	}
}

// calculateConfidence calculates confidence in the relevance assessment
func calculateConfidence(text string, keywordMatches int, conceptMatches int) float64 {
	// Base confidence on amount of text available
	textLength := len(strings.Fields(text))
	textConfidence := math.Min(1.0, float64(textLength)/500.0) // 500 words = max confidence

	// Boost confidence based on number of matches
	matchConfidence := math.Min(1.0, float64(keywordMatches+conceptMatches)/10.0)

	// Weighted average
	return textConfidence*0.6 + matchConfidence*0.4
}

// buildTopicRelevanceData builds a formatted string with manuscript data for AI analysis
func buildTopicRelevanceData(manuscriptData map[string]string) string {
	var parts []string

	// Priority fields for AI analysis
	priorityFields := []string{"title", "abstract", "keywords", "journal", "subject_area", "methodology"}

	for _, field := range priorityFields {
		for key, value := range manuscriptData {
			if strings.EqualFold(key, field) && value != "" {
				// Limit very long fields
				displayValue := value
				if field == "abstract" && len(value) > 1000 {
					displayValue = value[:1000] + "..."
				}
				parts = append(parts, fmt.Sprintf("%s: %s", strings.ToUpper(field), displayValue))
				break
			}
		}
	}

	// Add any other potentially relevant fields
	for key, value := range manuscriptData {
		keyLower := strings.ToLower(key)
		if value != "" && !isInPriorityFields(key, priorityFields) {
			if strings.Contains(keyLower, "topic") ||
				strings.Contains(keyLower, "subject") ||
				strings.Contains(keyLower, "category") ||
				strings.Contains(keyLower, "research") {
				parts = append(parts, fmt.Sprintf("%s: %s", strings.ToUpper(key), value))
			}
		}
	}

	if len(parts) == 0 {
		return "[No relevant data available]"
	}

	return strings.Join(parts, "\n")
}

// Helper function to check if a field is in priority fields
func isInPriorityFields(field string, priorityFields []string) bool {
	for _, pf := range priorityFields {
		if strings.EqualFold(field, pf) {
			return true
		}
	}
	return false
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// BatchCalculateTopicRelevance processes multiple manuscripts for topic relevance
func BatchCalculateTopicRelevance(manuscripts []map[string]string, config TopicRelevanceConfig, llmConfigs []any) ([]TopicRelevanceScore, error) {
	var scores []TopicRelevanceScore

	for _, manuscript := range manuscripts {
		var score *TopicRelevanceScore
		var err error

		if config.UseAI && len(llmConfigs) > 0 {
			score, err = CalculateTopicRelevanceWithAI(manuscript, config.Topics, llmConfigs)
		} else {
			score, err = CalculateTopicRelevance(manuscript, config.Topics, config.ScoreWeights)
		}

		if err != nil {
			logger.Error("Failed to calculate relevance for manuscript: %v", err)
			// Add a zero score for failed manuscripts
			scores = append(scores, TopicRelevanceScore{
				OverallScore: 0,
				IsRelevant:   false,
				Confidence:   0,
			})
		} else {
			// Apply minimum score threshold
			score.IsRelevant = score.OverallScore >= config.MinScore
			scores = append(scores, *score)
		}
	}

	return scores, nil
}
