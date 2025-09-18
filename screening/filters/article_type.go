package filters

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ArticleType represents the classification of an article
type ArticleType string

const (
	ResearchArticle  ArticleType = "research_article"
	ReviewArticle    ArticleType = "review"
	Editorial        ArticleType = "editorial"
	Letter           ArticleType = "letter"
	CaseReport       ArticleType = "case_report"
	Commentary       ArticleType = "commentary"
	Perspective      ArticleType = "perspective"
	MetaAnalysis     ArticleType = "meta_analysis"
	SystematicReview ArticleType = "systematic_review"
	Unknown          ArticleType = "unknown"
)

// ClassifyArticleType determines the type of article based on its content
func ClassifyArticleType(text string, llmConfigs []interface{}) (string, error) {
	if text == "" {
		return string(Unknown), fmt.Errorf("empty text provided")
	}

	// Try rule-based classification first
	articleType := classifyByRules(text)

	// If rule-based classification is confident, return it
	if articleType != Unknown && articleType != "" {
		return string(articleType), nil
	}

	// If LLM configs are provided, use AI classification
	if len(llmConfigs) > 0 {
		return classifyWithAI(text, llmConfigs[0])
	}

	return string(articleType), nil
}

// classifyByRules uses heuristics to classify article type
func classifyByRules(text string) ArticleType {
	textLower := strings.ToLower(text)

	// Extract first 2000 characters for abstract/introduction analysis
	sampleText := textLower
	if len(textLower) > 2000 {
		sampleText = textLower[:2000]
	}

	// Check for systematic review indicators
	if isSystematicReview(sampleText, textLower) {
		return SystematicReview
	}

	// Check for meta-analysis indicators
	if isMetaAnalysis(sampleText, textLower) {
		return MetaAnalysis
	}

	// Check for review article indicators
	if isReviewArticle(sampleText, textLower) {
		return ReviewArticle
	}

	// Check for editorial indicators
	if isEditorial(sampleText, textLower) {
		return Editorial
	}

	// Check for letter indicators
	if isLetter(sampleText, textLower) {
		return Letter
	}

	// Check for case report indicators
	if isCaseReport(sampleText, textLower) {
		return CaseReport
	}

	// Check for commentary indicators
	if isCommentary(sampleText, textLower) {
		return Commentary
	}

	// Check for perspective indicators
	if isPerspective(sampleText, textLower) {
		return Perspective
	}

	// Check for research article indicators
	if isResearchArticle(sampleText, textLower) {
		return ResearchArticle
	}

	return Unknown
}

// isSystematicReview checks for systematic review indicators
func isSystematicReview(sample, full string) bool {
	indicators := []string{
		"systematic review",
		"prisma",
		"cochrane",
		"search strategy",
		"inclusion criteria",
		"exclusion criteria",
		"quality assessment",
		"risk of bias",
		"literature search",
		"database search",
		"pooled analysis",
	}

	score := 0
	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			score += 2
		}
		if strings.Contains(full, indicator) {
			score++
		}
	}

	return score >= 6
}

// isMetaAnalysis checks for meta-analysis indicators
func isMetaAnalysis(sample, full string) bool {
	indicators := []string{
		"meta-analysis",
		"meta analysis",
		"pooled estimate",
		"forest plot",
		"funnel plot",
		"heterogeneity",
		"random effects",
		"fixed effects",
		"pooled odds ratio",
		"pooled risk ratio",
		"pooled hazard ratio",
	}

	score := 0
	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			score += 2
		}
		if strings.Contains(full, indicator) {
			score++
		}
	}

	return score >= 4
}

// isReviewArticle checks for review article indicators
func isReviewArticle(sample, full string) bool {
	indicators := []string{
		"review",
		"literature review",
		"narrative review",
		"scoping review",
		"we reviewed",
		"this review",
		"review of the literature",
		"comprehensive review",
		"critical review",
	}

	// Negative indicators (things that would indicate it's NOT a review)
	negativeIndicators := []string{
		"peer review",
		"review board",
		"ethics review",
		"review and approval",
		"reviewed the manuscript",
	}

	score := 0
	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			score += 2
		}
	}

	// Reduce score for negative indicators
	for _, indicator := range negativeIndicators {
		if strings.Contains(sample, indicator) {
			score--
		}
	}

	// Check for typical research article sections that reviews don't have
	if strings.Contains(full, "materials and methods") ||
		strings.Contains(full, "study population") ||
		strings.Contains(full, "data collection") {
		score -= 2
	}

	return score >= 3
}

// isEditorial checks for editorial indicators
func isEditorial(sample, full string) bool {
	indicators := []string{
		"editorial",
		"editor's note",
		"from the editor",
		"guest editorial",
		"this issue",
		"in this issue",
	}

	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			return true
		}
	}

	// Editorials are typically short
	if len(full) < 3000 {
		shortIndicators := []string{"we believe", "we think", "our opinion"}
		for _, indicator := range shortIndicators {
			if strings.Contains(sample, indicator) {
				return true
			}
		}
	}

	return false
}

// isLetter checks for letter indicators
func isLetter(sample, full string) bool {
	indicators := []string{
		"letter to the editor",
		"letter to editor",
		"correspondence",
		"in response to",
		"we read with interest",
		"dear editor",
		"to the editor",
	}

	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			return true
		}
	}

	// Letters are typically very short
	if len(full) < 2000 {
		if strings.Contains(sample, "dear ") || strings.Contains(sample, "sincerely,") {
			return true
		}
	}

	return false
}

// isCaseReport checks for case report indicators
func isCaseReport(sample, full string) bool {
	indicators := []string{
		"case report",
		"case presentation",
		"case study",
		"patient presentation",
		"we report a case",
		"we present a case",
		"year-old patient",
		"year-old male",
		"year-old female",
		"year-old woman",
		"year-old man",
		"chief complaint",
		"physical examination revealed",
	}

	score := 0
	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			score += 2
		}
		if strings.Contains(full, indicator) {
			score++
		}
	}

	return score >= 4
}

// isCommentary checks for commentary indicators
func isCommentary(sample, full string) bool {
	indicators := []string{
		"commentary",
		"comment on",
		"viewpoint",
		"opinion piece",
		"we comment",
		"authors comment",
	}

	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			return true
		}
	}

	return false
}

// isPerspective checks for perspective indicators
func isPerspective(sample, full string) bool {
	indicators := []string{
		"perspective",
		"perspectives on",
		"personal view",
		"point of view",
	}

	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			return true
		}
	}

	return false
}

// isResearchArticle checks for research article indicators
func isResearchArticle(sample, full string) bool {
	indicators := []string{
		"methods",
		"methodology",
		"participants",
		"results",
		"data collection",
		"statistical analysis",
		"study design",
		"inclusion criteria",
		"sample size",
		"primary outcome",
		"secondary outcome",
		"p value",
		"confidence interval",
		"standard deviation",
		"we conducted",
		"we performed",
		"we analyzed",
		"we investigated",
		"this study",
		"our study",
		"the present study",
	}

	score := 0
	for _, indicator := range indicators {
		if strings.Contains(sample, indicator) {
			score++
		}
		if strings.Contains(full, indicator) {
			score++
		}
	}

	// Check for typical research article structure
	hasMethodsSection := strings.Contains(full, "method") || strings.Contains(full, "materials")
	hasResultsSection := strings.Contains(full, "results") || strings.Contains(full, "findings")
	hasDiscussionSection := strings.Contains(full, "discussion") || strings.Contains(full, "conclusions")

	if hasMethodsSection && hasResultsSection {
		score += 5
	}

	if hasDiscussionSection {
		score += 2
	}

	return score >= 8
}

// classifyWithAI uses LLM to classify article type
func classifyWithAI(text string, llmConfig interface{}) (string, error) {
	// This would integrate with the alembica package for AI calls
	// For now, returning a placeholder implementation

	// Extract sample for efficiency
	sampleText := text
	if len(text) > 3000 {
		sampleText = text[:3000]
	}

	// In actual implementation, this would:
	// 1. Create a structured prompt asking the LLM to classify the article
	// 2. Call the appropriate LLM API through alembica
	// 3. Parse the JSON response to extract the article type

	// Placeholder: fall back to rule-based classification
	result := classifyByRules(sampleText)
	return string(result), nil
}

// GetArticleTypeDescription returns a description of the article type
func GetArticleTypeDescription(articleType string) string {
	descriptions := map[string]string{
		"research_article":  "Original research presenting new empirical findings",
		"review":            "Review of existing literature on a topic",
		"systematic_review": "Systematic review following structured methodology",
		"meta_analysis":     "Statistical analysis of multiple studies",
		"editorial":         "Opinion piece by journal editors",
		"letter":            "Letter to the editor or correspondence",
		"case_report":       "Report of individual patient case(s)",
		"commentary":        "Commentary on published work",
		"perspective":       "Author's perspective on a topic",
		"unknown":           "Article type could not be determined",
	}

	if desc, exists := descriptions[articleType]; exists {
		return desc
	}

	return "Unknown article type"
}

// ArticleTypeScore represents confidence scores for each article type
type ArticleTypeScore struct {
	Type       string  `json:"type"`
	Score      float64 `json:"score"`
	Confidence string  `json:"confidence"`
}

// ClassifyArticleTypeWithScores returns classification with confidence scores
func ClassifyArticleTypeWithScores(text string) ([]ArticleTypeScore, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text provided")
	}

	scores := []ArticleTypeScore{}
	textLower := strings.ToLower(text)

	// Calculate scores for each type
	types := map[ArticleType]float64{
		SystematicReview: calculateSystematicReviewScore(textLower),
		MetaAnalysis:     calculateMetaAnalysisScore(textLower),
		ReviewArticle:    calculateReviewScore(textLower),
		ResearchArticle:  calculateResearchScore(textLower),
		Editorial:        calculateEditorialScore(textLower),
		Letter:           calculateLetterScore(textLower),
		CaseReport:       calculateCaseReportScore(textLower),
		Commentary:       calculateCommentaryScore(textLower),
		Perspective:      calculatePerspectiveScore(textLower),
	}

	// Convert to ArticleTypeScore and determine confidence
	for articleType, score := range types {
		confidence := "low"
		if score > 0.7 {
			confidence = "high"
		} else if score > 0.4 {
			confidence = "medium"
		}

		scores = append(scores, ArticleTypeScore{
			Type:       string(articleType),
			Score:      score,
			Confidence: confidence,
		})
	}

	// Sort by score (highest first)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].Score > scores[i].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	return scores, nil
}

// Helper functions for calculating scores (0.0 to 1.0)
func calculateSystematicReviewScore(text string) float64 {
	score := 0.0
	indicators := []string{
		"systematic review", "prisma", "cochrane", "search strategy",
		"inclusion criteria", "exclusion criteria", "quality assessment",
	}

	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			score += 0.15
		}
	}

	if score > 1.0 {
		return 1.0
	}
	return score
}

func calculateMetaAnalysisScore(text string) float64 {
	score := 0.0
	indicators := []string{
		"meta-analysis", "pooled estimate", "forest plot", "heterogeneity",
		"random effects", "fixed effects",
	}

	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			score += 0.17
		}
	}

	if score > 1.0 {
		return 1.0
	}
	return score
}

func calculateReviewScore(text string) float64 {
	score := 0.0
	indicators := []string{
		"review", "literature review", "narrative review", "we reviewed",
	}

	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			score += 0.25
		}
	}

	// Reduce score if it has research article characteristics
	if strings.Contains(text, "methods") && strings.Contains(text, "results") {
		score -= 0.3
	}

	if score < 0 {
		return 0
	}
	if score > 1.0 {
		return 1.0
	}
	return score
}

func calculateResearchScore(text string) float64 {
	score := 0.0
	indicators := []string{
		"methods", "results", "participants", "data collection",
		"statistical analysis", "p value", "confidence interval",
	}

	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			score += 0.14
		}
	}

	if score > 1.0 {
		return 1.0
	}
	return score
}

func calculateEditorialScore(text string) float64 {
	if strings.Contains(text, "editorial") || strings.Contains(text, "from the editor") {
		return 0.9
	}
	if len(text) < 3000 && strings.Contains(text, "this issue") {
		return 0.6
	}
	return 0.0
}

func calculateLetterScore(text string) float64 {
	if strings.Contains(text, "letter to the editor") || strings.Contains(text, "dear editor") {
		return 0.95
	}
	if strings.Contains(text, "correspondence") || strings.Contains(text, "in response to") {
		return 0.7
	}
	return 0.0
}

func calculateCaseReportScore(text string) float64 {
	score := 0.0
	indicators := []string{
		"case report", "case presentation", "patient presentation",
		"year-old patient", "chief complaint",
	}

	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			score += 0.3
		}
	}

	if score > 1.0 {
		return 1.0
	}
	return score
}

func calculateCommentaryScore(text string) float64 {
	if strings.Contains(text, "commentary") || strings.Contains(text, "comment on") {
		return 0.8
	}
	if strings.Contains(text, "viewpoint") || strings.Contains(text, "opinion piece") {
		return 0.6
	}
	return 0.0
}

func calculatePerspectiveScore(text string) float64 {
	if strings.Contains(text, "perspective") || strings.Contains(text, "personal view") {
		return 0.8
	}
	if strings.Contains(text, "point of view") {
		return 0.6
	}
	return 0.0
}

// BatchClassifyArticleTypes classifies multiple articles efficiently
func BatchClassifyArticleTypes(texts []string, llmConfigs []interface{}) ([]string, error) {
	results := make([]string, len(texts))

	for i, text := range texts {
		articleType, err := ClassifyArticleType(text, llmConfigs)
		if err != nil {
			results[i] = string(Unknown)
		} else {
			results[i] = articleType
		}
	}

	return results, nil
}

// ArticleTypeStatistics represents statistics about article type classification
type ArticleTypeStatistics struct {
	TotalArticles int                `json:"total_articles"`
	Distribution  map[string]int     `json:"distribution"`
	Percentages   map[string]float64 `json:"percentages"`
}

// CalculateArticleTypeStatistics computes statistics for a batch of classified articles
func CalculateArticleTypeStatistics(articleTypes []string) ArticleTypeStatistics {
	stats := ArticleTypeStatistics{
		TotalArticles: len(articleTypes),
		Distribution:  make(map[string]int),
		Percentages:   make(map[string]float64),
	}

	// Count occurrences
	for _, articleType := range articleTypes {
		stats.Distribution[articleType]++
	}

	// Calculate percentages
	for articleType, count := range stats.Distribution {
		stats.Percentages[articleType] = float64(count) / float64(stats.TotalArticles) * 100
	}

	return stats
}

// ExportClassificationResults exports classification results to JSON
func ExportClassificationResults(results []ArticleTypeScore) (string, error) {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling results: %v", err)
	}
	return string(jsonData), nil
}
