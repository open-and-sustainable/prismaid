package filters

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/alembica/extraction"
	"github.com/open-and-sustainable/alembica/utils/logger"
)

// ArticleType represents the classification of an article
type ArticleType string

const (
	// Traditional publication types
	ResearchArticle  ArticleType = "research_article"
	ReviewArticle    ArticleType = "review"
	Editorial        ArticleType = "editorial"
	Letter           ArticleType = "letter"
	CaseReport       ArticleType = "case_report"
	Commentary       ArticleType = "commentary"
	Perspective      ArticleType = "perspective"
	MetaAnalysis     ArticleType = "meta_analysis"
	SystematicReview ArticleType = "systematic_review"

	// Methodological distinctions
	EmpiricalStudy   ArticleType = "empirical_study"
	TheoreticalPaper ArticleType = "theoretical_paper"
	MethodsPaper     ArticleType = "methods_paper"

	// Study scope classifications
	SingleCaseStudy ArticleType = "single_case_study"
	SampleStudy     ArticleType = "sample_study"

	Unknown ArticleType = "unknown"
)

// ArticleClassification represents multiple overlapping classifications for an article
type ArticleClassification struct {
	PrimaryType         ArticleType             `json:"primary_type"`
	AllTypes            []ArticleType           `json:"all_types"`
	TypeScores          map[ArticleType]float64 `json:"type_scores"`
	MethodologicalTypes []ArticleType           `json:"methodological_types"`
	ScopeTypes          []ArticleType           `json:"scope_types"`
}

// ArticleTypeScore represents a type with its confidence score
type ArticleTypeScore struct {
	Type       ArticleType
	Score      float64
	Confidence string // "high", "medium", "low"
}

// ClassifyArticleType determines all applicable types for an article
// Returns a JSON string representing the classification for backward compatibility
func ClassifyArticleType(text string, llmConfigs []any) (string, error) {
	if text == "" {
		return string(Unknown), fmt.Errorf("empty text provided")
	}

	// Get comprehensive classification
	classification := classifyArticleComprehensive(text)

	// If LLM configs are provided, enhance with AI classification
	if len(llmConfigs) > 0 {
		enhanceWithAI(text, llmConfigs[0], classification)
	}

	// Convert to JSON for backward compatibility
	jsonData, err := json.Marshal(classification)
	if err != nil {
		// Fallback to primary type string
		return string(classification.PrimaryType), nil
	}

	return string(jsonData), nil
}

// BatchClassifyArticleTypesWithAI processes multiple manuscripts in a single AI call
func BatchClassifyArticleTypesWithAI(manuscriptsData []map[string]string, llmConfigs []any) map[string]*ArticleClassification {
	results := make(map[string]*ArticleClassification)

	// Initialize all results with Unknown
	for i := range manuscriptsData {
		results[fmt.Sprintf("%d", i)] = &ArticleClassification{
			PrimaryType: Unknown,
			AllTypes:    []ArticleType{Unknown},
		}
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
		logger.Info("No valid AI models configured for article type classification")
		return results
	}

	// Build all prompts
	var prompts []definitions.Prompt
	validIndices := []int{}

	for idx, manuscriptData := range manuscriptsData {
		// Extract relevant fields
		title := ""
		abstract := ""

		for field, value := range manuscriptData {
			fieldLower := strings.ToLower(field)
			switch fieldLower {
			case "title":
				title = value
			case "abstract":
				abstract = value
			}
		}

		// Skip if no text to analyze
		if title == "" && abstract == "" {
			continue
		}

		// Create the prompt (same as in classifyWithAIComprehensive)
		prompt := fmt.Sprintf(`You are a scientific manuscript classification expert. Analyze this manuscript and provide a comprehensive type classification.

CONTEXT:
A manuscript can have MULTIPLE overlapping classifications. For example:
- A paper can be both "research_article" AND "empirical_study" AND "sample_study"
- A review can be both "review" AND "systematic_review"
- A methods paper can be both "research_article" AND "methods_paper"

CLASSIFICATION DIMENSIONS:

1. TRADITIONAL PUBLICATION TYPES (what editors would call it):
   - research_article: Original research with methods, results, and conclusions
   - review: Literature review without systematic methodology (narrative review, scoping review)
   - systematic_review: Following structured review protocol (PRISMA, etc.)
   - meta_analysis: Statistical synthesis of multiple studies
   - editorial: Opinion piece by editors
   - letter: Brief correspondence to editors
   - case_report: Single patient/case/instance report
   - commentary: Comments on published work
   - perspective: Author viewpoints and opinions

2. METHODOLOGICAL TYPES (how research is conducted):
   - empirical_study: Based on observation/experimentation with data collection
   - theoretical_paper: Conceptual work without empirical data
   - methods_paper: Presenting new methods, techniques, or protocols

3. STUDY SCOPE (for empirical studies):
   - single_case_study: In-depth analysis of ONE case/patient/organization (n=1)
   - sample_study: Multiple subjects (cohort studies, surveys, cross-sectional, etc.)

IMPORTANT RULES:
- A paper can have types from ALL three dimensions
- Single case ≠ case report (case report is a publication type, single case is a scope)
- If empirical, MUST be either single_case_study OR sample_study
- Default to "research_article" for traditional type if unclear
- Be specific - don't just say "unknown"

MANUSCRIPT:
Title: %s
Abstract: %s

Provide classification as JSON:
{
  "primary_type": "most_specific_type",
  "all_types": ["type1", "type2", ...],
  "methodological_types": ["empirical_study" | "theoretical_paper" | "methods_paper"],
  "scope_types": ["single_case_study" | "sample_study"] (only if empirical),
  "type_scores": {"type1": 0.95, "type2": 0.80, ...}
}`, title, abstract)

		prompts = append(prompts, definitions.Prompt{
			PromptContent:  prompt,
			SequenceID:     fmt.Sprintf("%d", idx+1),
			SequenceNumber: idx + 1,
		})
		validIndices = append(validIndices, idx)
	}

	if len(prompts) == 0 {
		logger.Info("No valid manuscripts for article type classification")
		return results
	}

	logger.Info("Prepared %d manuscripts for batch article type classification", len(prompts))

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
	logger.Info("Calling AI model with batch of %d article type classification requests", len(prompts))
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

			// Parse the AI response
			var aiResult struct {
				PrimaryType         string             `json:"primary_type"`
				AllTypes            []string           `json:"all_types"`
				MethodologicalTypes []string           `json:"methodological_types"`
				ScopeTypes          []string           `json:"scope_types"`
				TypeScores          map[string]float64 `json:"type_scores"`
			}

			if err := json.Unmarshal([]byte(response), &aiResult); err != nil {
				logger.Error("Failed to parse AI classification for manuscript %d: %v", manuscriptIdx, err)
				continue
			}

			// Convert string types to ArticleType enum
			classification := &ArticleClassification{
				PrimaryType:         parseArticleType(aiResult.PrimaryType),
				AllTypes:            []ArticleType{},
				MethodologicalTypes: []ArticleType{},
				ScopeTypes:          []ArticleType{},
				TypeScores:          make(map[ArticleType]float64),
			}

			// Convert all types
			for _, typeStr := range aiResult.AllTypes {
				classification.AllTypes = append(classification.AllTypes, parseArticleType(typeStr))
			}

			// Convert methodological types
			for _, typeStr := range aiResult.MethodologicalTypes {
				classification.MethodologicalTypes = append(classification.MethodologicalTypes, parseArticleType(typeStr))
			}

			// Convert scope types
			for _, typeStr := range aiResult.ScopeTypes {
				classification.ScopeTypes = append(classification.ScopeTypes, parseArticleType(typeStr))
			}

			// Convert type scores
			for typeStr, score := range aiResult.TypeScores {
				classification.TypeScores[parseArticleType(typeStr)] = score
			}

			results[fmt.Sprintf("%d", manuscriptIdx)] = classification
		}
	}

	return results
}

// ClassifyArticleTypeWithAI uses AI to classify article type
func ClassifyArticleTypeWithAI(manuscriptData map[string]string, useAI bool, llmConfigs []any) (*ArticleClassification, error) {
	// Extract relevant fields
	title := ""
	abstract := ""

	// Try to get fields with case-insensitive lookup
	for field, value := range manuscriptData {
		fieldLower := strings.ToLower(field)
		switch fieldLower {
		case "title":
			title = value
		case "abstract", "summary":
			abstract = value
		}
	}

	// Combine title and abstract for analysis
	text := title + " " + abstract
	if text == "" {
		return &ArticleClassification{PrimaryType: Unknown}, fmt.Errorf("no text fields available")
	}

	if !useAI || len(llmConfigs) == 0 {
		// Use rule-based classification
		return classifyArticleComprehensive(text), nil
	}

	// Use AI-based classification
	return classifyWithAIComprehensive(title, abstract, llmConfigs)
}

// ClassifyArticleTypes returns multiple applicable article types
func ClassifyArticleTypes(text string, llmConfigs []any) (*ArticleClassification, error) {
	if text == "" {
		return &ArticleClassification{PrimaryType: Unknown}, fmt.Errorf("empty text provided")
	}

	// Get comprehensive classification
	classification := classifyArticleComprehensive(text)

	// If LLM configs are provided, enhance with AI classification
	if len(llmConfigs) > 0 {
		enhanceWithAI(text, llmConfigs[0], classification)
	}

	return classification, nil
}

// classifyArticleComprehensive performs comprehensive rule-based classification
func classifyArticleComprehensive(text string) *ArticleClassification {
	textLower := strings.ToLower(text)

	// Extract first 2000 characters for abstract/introduction analysis
	sampleText := textLower
	if len(textLower) > 2000 {
		sampleText = textLower[:2000]
	}

	classification := &ArticleClassification{
		AllTypes:            []ArticleType{},
		TypeScores:          make(map[ArticleType]float64),
		MethodologicalTypes: []ArticleType{},
		ScopeTypes:          []ArticleType{},
	}

	// Calculate scores for all publication types
	calculatePublicationTypeScores(sampleText, textLower, classification)

	// Calculate methodological type scores
	calculateMethodologicalScores(sampleText, textLower, classification)

	// Calculate study scope scores
	calculateStudyScopeScores(sampleText, textLower, classification)

	// Determine primary type based on highest score
	determinePrimaryType(classification)

	// Build comprehensive type list
	buildTypeList(classification)

	return classification
}

// calculatePublicationTypeScores calculates scores for traditional publication types
func calculatePublicationTypeScores(sample, full string, classification *ArticleClassification) {
	// Check for systematic review
	if score := calculateSystematicReviewScore(sample, full); score > 0 {
		classification.TypeScores[SystematicReview] = score
	}

	// Check for meta-analysis
	if score := calculateMetaAnalysisScore(sample, full); score > 0 {
		classification.TypeScores[MetaAnalysis] = score
	}

	// Check for review article
	if score := calculateReviewScore(sample, full); score > 0 {
		classification.TypeScores[ReviewArticle] = score
	}

	// Check for editorial
	if score := calculateEditorialScore(sample, full); score > 0 {
		classification.TypeScores[Editorial] = score
	}

	// Check for letter
	if score := calculateLetterScore(sample, full); score > 0 {
		classification.TypeScores[Letter] = score
	}

	// Check for case report
	if score := calculateCaseReportScore(sample, full); score > 0 {
		classification.TypeScores[CaseReport] = score
	}

	// Check for commentary
	if score := calculateCommentaryScore(sample, full); score > 0 {
		classification.TypeScores[Commentary] = score
	}

	// Check for perspective
	if score := calculatePerspectiveScore(sample, full); score > 0 {
		classification.TypeScores[Perspective] = score
	}

	// Check for research article
	if score := calculateResearchScore(sample, full); score > 0 {
		classification.TypeScores[ResearchArticle] = score
	}
}

// calculateMethodologicalScores calculates scores for methodological types
func calculateMethodologicalScores(sample, full string, classification *ArticleClassification) {
	empiricalScore := calculateEmpiricalScore(sample, full)
	theoreticalScore := calculateTheoreticalScore(sample, full)
	methodsScore := calculateMethodsScore(sample, full)

	// A paper can be both empirical and methods-focused
	if empiricalScore > 5 {
		classification.TypeScores[EmpiricalStudy] = empiricalScore
		classification.MethodologicalTypes = append(classification.MethodologicalTypes, EmpiricalStudy)
	}

	if theoreticalScore > 5 {
		classification.TypeScores[TheoreticalPaper] = theoreticalScore
		classification.MethodologicalTypes = append(classification.MethodologicalTypes, TheoreticalPaper)
	}

	if methodsScore > 5 {
		classification.TypeScores[MethodsPaper] = methodsScore
		classification.MethodologicalTypes = append(classification.MethodologicalTypes, MethodsPaper)
	}
}

// calculateStudyScopeScores calculates scores for study scope types
func calculateStudyScopeScores(sample, full string, classification *ArticleClassification) {
	// Only calculate scope if there's empirical content
	empiricalScore, hasEmpirical := classification.TypeScores[EmpiricalStudy]
	researchScore, hasResearch := classification.TypeScores[ResearchArticle]
	caseScore, hasCase := classification.TypeScores[CaseReport]

	if (hasEmpirical && empiricalScore > 0) || (hasResearch && researchScore > 0) || (hasCase && caseScore > 0) {

		singleCaseScore := calculateSingleCaseScore(sample, full)
		sampleScore := calculateSampleScore(sample, full)

		if singleCaseScore > 5 {
			classification.TypeScores[SingleCaseStudy] = singleCaseScore
			classification.ScopeTypes = append(classification.ScopeTypes, SingleCaseStudy)
		}

		if sampleScore > 5 {
			classification.TypeScores[SampleStudy] = sampleScore
			classification.ScopeTypes = append(classification.ScopeTypes, SampleStudy)
		}
	}
}

// determinePrimaryType determines the primary type based on priority and scores
func determinePrimaryType(classification *ArticleClassification) {
	// Priority order for primary type determination
	priorityOrder := []ArticleType{
		MetaAnalysis,
		SystematicReview,
		ReviewArticle,
		MethodsPaper,
		ResearchArticle,
		CaseReport,
		Editorial,
		Letter,
		Commentary,
		Perspective,
		EmpiricalStudy,
		TheoreticalPaper,
		SingleCaseStudy,
		SampleStudy,
	}

	var maxScore float64
	classification.PrimaryType = Unknown

	for _, articleType := range priorityOrder {
		if score, exists := classification.TypeScores[articleType]; exists && score > maxScore {
			maxScore = score
			classification.PrimaryType = articleType
		}
	}

	// If still unknown, check for any type with score
	if classification.PrimaryType == Unknown {
		for articleType, score := range classification.TypeScores {
			if score > 0 {
				classification.PrimaryType = articleType
				break
			}
		}
	}
}

// buildTypeList builds the comprehensive list of all applicable types
func buildTypeList(classification *ArticleClassification) {
	// Add all types with significant scores (>5)
	for articleType, score := range classification.TypeScores {
		if score > 5 {
			// Avoid duplicates
			found := false
			for _, existing := range classification.AllTypes {
				if existing == articleType {
					found = true
					break
				}
			}
			if !found {
				classification.AllTypes = append(classification.AllTypes, articleType)
			}
		}
	}

	// Ensure primary type is in the list
	if classification.PrimaryType != Unknown {
		found := false
		for _, existing := range classification.AllTypes {
			if existing == classification.PrimaryType {
				found = true
				break
			}
		}
		if !found {
			classification.AllTypes = append([]ArticleType{classification.PrimaryType}, classification.AllTypes...)
		}
	}
}

// Scoring functions for publication types

func calculateSystematicReviewScore(sample, full string) float64 {
	var score float64

	// Strong indicators
	strongIndicators := []string{
		"systematic review",
		"prisma",
		"cochrane",
		"prospero",
		"systematic literature review",
		"systematic search",
	}

	for _, indicator := range strongIndicators {
		if strings.Contains(sample, indicator) {
			score += 10
		} else if strings.Contains(full, indicator) {
			score += 5
		}
	}

	// Methodological indicators
	methodIndicators := []string{
		"inclusion criteria",
		"exclusion criteria",
		"database search",
		"search strategy",
		"quality assessment",
		"risk of bias",
		"data extraction",
	}

	for _, indicator := range methodIndicators {
		if strings.Contains(full, indicator) {
			score += 3
		}
	}

	return score
}

func calculateMetaAnalysisScore(sample, full string) float64 {
	var score float64

	// Strong indicators
	if strings.Contains(sample, "meta-analysis") || strings.Contains(sample, "meta analysis") ||
		strings.Contains(sample, "metaanalysis") {
		score += 15
	}

	// Statistical pooling indicators
	poolingIndicators := []string{
		"pooled",
		"forest plot",
		"funnel plot",
		"heterogeneity",
		"random effects",
		"fixed effects",
		"effect size",
		"pooled estimate",
		"combined results",
		"statistical synthesis",
	}

	for _, indicator := range poolingIndicators {
		if strings.Contains(full, indicator) {
			score += 4
		}
	}

	// If it's also a systematic review, boost the score
	if strings.Contains(full, "systematic review") && score > 0 {
		score += 5
	}

	return score
}

func calculateReviewScore(sample, full string) float64 {
	var score float64

	// Direct indicators
	reviewIndicators := []string{
		"literature review",
		"narrative review",
		"scoping review",
		"integrative review",
		"critical review",
		"review of",
		"reviews the",
		"we review",
		"this review",
	}

	for _, indicator := range reviewIndicators {
		if strings.Contains(sample, indicator) {
			score += 8
		} else if strings.Contains(full, indicator) {
			score += 4
		}
	}

	// Check for review structure without empirical data
	if strings.Contains(full, "review") && !strings.Contains(full, "data collection") &&
		!strings.Contains(full, "participants") && !strings.Contains(full, "subjects") {
		score += 3
	}

	return score
}

func calculateResearchScore(sample, full string) float64 {
	var score float64

	// Research structure indicators
	methodIndicators := []string{
		"methods",
		"methodology",
		"data collection",
		"participants",
		"subjects",
		"sample",
		"procedure",
		"materials",
	}

	resultsIndicators := []string{
		"results",
		"findings",
		"analysis",
		"statistical",
		"significant",
		"p-value",
		"correlation",
		"regression",
	}

	for _, indicator := range methodIndicators {
		if strings.Contains(full, indicator) {
			score += 2
		}
	}

	for _, indicator := range resultsIndicators {
		if strings.Contains(full, indicator) {
			score += 2
		}
	}

	// Check for research article structure
	if strings.Contains(full, "introduction") && strings.Contains(full, "discussion") {
		score += 3
	}

	return score
}

func calculateEditorialScore(sample, full string) float64 {
	var score float64

	editorialIndicators := []string{
		"editorial",
		"editor's",
		"from the editor",
		"guest editorial",
		"this issue",
		"in this issue",
		"special issue",
	}

	for _, indicator := range editorialIndicators {
		if strings.Contains(sample, indicator) {
			score += 10
		} else if strings.Contains(full, indicator) {
			score += 5
		}
	}

	// Editorial characteristics
	if score > 0 && len(full) < 3000 {
		score += 3
	}

	return score
}

func calculateLetterScore(sample, full string) float64 {
	var score float64

	letterIndicators := []string{
		"letter to",
		"dear editor",
		"to the editor",
		"correspondence",
		"we read with interest",
		"response to",
		"reply to",
		"comment on",
	}

	for _, indicator := range letterIndicators {
		if strings.Contains(sample, indicator) {
			score += 10
		} else if strings.Contains(full, indicator) {
			score += 5
		}
	}

	// Letter characteristics
	if score > 0 && len(full) < 2000 {
		score += 3
	}

	return score
}

func calculateCaseReportScore(sample, full string) float64 {
	var score float64

	caseIndicators := []string{
		"case report",
		"case presentation",
		"case study",
		"patient presentation",
		"clinical case",
		"case description",
		"we report",
		"we present a case",
		"year-old",
		"presented with",
		"chief complaint",
		"medical history",
		"clinical findings",
		"diagnosis",
		"treatment",
		"follow-up",
	}

	for _, indicator := range caseIndicators {
		if strings.Contains(sample, indicator) {
			score += 4
		} else if strings.Contains(full, indicator) {
			score += 2
		}
	}

	// Medical case indicators
	if strings.Contains(full, "patient") && strings.Contains(full, "diagnosis") {
		score += 3
	}

	return score
}

func calculateCommentaryScore(sample, full string) float64 {
	var score float64

	commentaryIndicators := []string{
		"commentary",
		"comment",
		"viewpoint",
		"opinion",
		"we comment",
		"authors comment",
		"invited commentary",
	}

	for _, indicator := range commentaryIndicators {
		if strings.Contains(sample, indicator) {
			score += 8
		} else if strings.Contains(full, indicator) {
			score += 4
		}
	}

	return score
}

func calculatePerspectiveScore(sample, full string) float64 {
	var score float64

	perspectiveIndicators := []string{
		"perspective",
		"point of view",
		"personal view",
		"author's perspective",
		"our perspective",
	}

	for _, indicator := range perspectiveIndicators {
		if strings.Contains(sample, indicator) {
			score += 8
		} else if strings.Contains(full, indicator) {
			score += 4
		}
	}

	return score
}

// Scoring functions for methodological types

func calculateEmpiricalScore(sample, full string) float64 {
	var score float64

	// Data collection indicators
	dataIndicators := []string{
		"data collection",
		"data were collected",
		"collected data",
		"gathered data",
		"survey",
		"experiment",
		"observation",
		"measurement",
		"empirical",
		"fieldwork",
		"interviews",
		"questionnaire",
	}

	for _, indicator := range dataIndicators {
		if strings.Contains(sample, indicator) {
			score += 4
		} else if strings.Contains(full, indicator) {
			score += 2
		}
	}

	// Analysis indicators
	analysisIndicators := []string{
		"statistical analysis",
		"data analysis",
		"analyzed",
		"tested",
		"measured",
		"calculated",
		"regression",
		"correlation",
		"anova",
		"t-test",
	}

	for _, indicator := range analysisIndicators {
		if strings.Contains(full, indicator) {
			score += 2
		}
	}

	// Results from data
	if strings.Contains(full, "results") && (strings.Contains(full, "data") ||
		strings.Contains(full, "participants") || strings.Contains(full, "sample")) {
		score += 3
	}

	return score
}

func calculateTheoreticalScore(sample, full string) float64 {
	var score float64

	// Theoretical indicators
	theoryIndicators := []string{
		"theoretical",
		"conceptual",
		"framework",
		"model",
		"theory",
		"proposition",
		"hypothesis",
		"conceptualize",
		"theorize",
		"theoretical framework",
		"conceptual model",
		"theoretical model",
	}

	for _, indicator := range theoryIndicators {
		if strings.Contains(sample, indicator) {
			score += 4
		} else if strings.Contains(full, indicator) {
			score += 2
		}
	}

	// Abstract concepts without empirical data
	abstractIndicators := []string{
		"we propose",
		"we argue",
		"we posit",
		"we theorize",
		"this paper argues",
		"we conceptualize",
		"philosophical",
		"epistemological",
		"ontological",
	}

	for _, indicator := range abstractIndicators {
		if strings.Contains(full, indicator) {
			score += 3
		}
	}

	// Penalty for empirical indicators
	if strings.Contains(full, "data collection") || strings.Contains(full, "empirical") {
		score -= 5
	}

	return score
}

func calculateMethodsScore(sample, full string) float64 {
	var score float64

	methodsIndicators := []string{
		"novel method",
		"new method",
		"method for",
		"technique for",
		"algorithm",
		"protocol",
		"procedure",
		"methodology",
		"methodological",
		"we present a method",
		"we introduce",
		"we develop",
	}

	for _, indicator := range methodsIndicators {
		if strings.Contains(sample, indicator) {
			score += 5
		} else if strings.Contains(full, indicator) {
			score += 2
		}
	}

	return score
}

// Scoring functions for study scope types

func calculateSingleCaseScore(sample, full string) float64 {
	var score float64

	singleIndicators := []string{
		"single case",
		"one case",
		"individual case",
		"one patient",
		"single patient",
		"one company",
		"single company",
		"one organization",
		"single organization",
		"n=1",
		"n = 1",
		"single subject",
		"individual subject",
	}

	for _, indicator := range singleIndicators {
		if strings.Contains(sample, indicator) {
			score += 10
		} else if strings.Contains(full, indicator) {
			score += 5
		}
	}

	// Check for case study without multiple
	if (strings.Contains(full, "case study") || strings.Contains(full, "case analysis")) &&
		!strings.Contains(full, "multiple") && !strings.Contains(full, "cases") &&
		!strings.Contains(full, "comparative") {
		score += 5
	}

	return score
}

func calculateSampleScore(sample, full string) float64 {
	var score float64

	// Multiple subjects indicators
	multipleIndicators := []string{
		"participants",
		"subjects",
		"respondents",
		"patients",
		"sample",
		"cohort",
		"population",
		"cases",
		"companies",
		"organizations",
		"individuals",
	}

	for _, indicator := range multipleIndicators {
		if strings.Contains(sample, indicator) {
			score += 3
		} else if strings.Contains(full, indicator) {
			score += 1.5
		}
	}

	// Study design indicators
	designIndicators := []string{
		"cross-sectional",
		"longitudinal",
		"cohort study",
		"case-control",
		"randomized",
		"controlled trial",
		"survey",
		"questionnaire",
		"recruited",
		"enrolled",
		"sampled",
	}

	for _, indicator := range designIndicators {
		if strings.Contains(full, indicator) {
			score += 4
		}
	}

	// Sample size indicators
	if strings.Contains(full, "n=") || strings.Contains(full, "n =") {
		// Check if it's not n=1
		if !strings.Contains(full, "n=1") && !strings.Contains(full, "n = 1") {
			score += 5
		}
	}

	return score
}

// AI enhancement functions

func enhanceWithAI(text string, llmConfig interface{}, classification *ArticleClassification) error {
	// This would integrate with the AI model to enhance classification
	// For now, it's a placeholder that can be implemented with actual LLM calls
	return nil
}

func classifyWithAI(text string, llmConfig interface{}) (string, error) {
	// Placeholder for AI-based classification
	// This would make actual LLM API calls
	return string(Unknown), nil
}

// parseArticleType converts string to ArticleType enum
func parseArticleType(typeStr string) ArticleType {
	switch strings.ToLower(strings.TrimSpace(typeStr)) {
	case "research_article":
		return ResearchArticle
	case "review":
		return ReviewArticle
	case "editorial":
		return Editorial
	case "letter":
		return Letter
	case "case_report":
		return CaseReport
	case "commentary":
		return Commentary
	case "perspective":
		return Perspective
	case "meta_analysis":
		return MetaAnalysis
	case "systematic_review":
		return SystematicReview
	case "empirical_study":
		return EmpiricalStudy
	case "theoretical_paper":
		return TheoreticalPaper
	case "methods_paper":
		return MethodsPaper
	case "single_case_study":
		return SingleCaseStudy
	case "sample_study":
		return SampleStudy
	default:
		return Unknown
	}
}

// classifyWithAIComprehensive performs AI-based article type classification
func classifyWithAIComprehensive(title, abstract string, llmConfigs []any) (*ArticleClassification, error) {
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
		logger.Info("No valid AI models configured for article type classification, falling back to rule-based")
		// Fall back to rule-based classification
		text := title + " " + abstract
		return classifyArticleComprehensive(text), nil
	}

	// Build the manuscript data string
	dataStr := fmt.Sprintf("Title: %s\nAbstract: %s", title, abstract)

	// Create the comprehensive prompt
	prompt := fmt.Sprintf(`You are an expert in scientific literature classification. Analyze the following manuscript and classify it into ALL applicable categories.

IMPORTANT: A manuscript can belong to MULTIPLE overlapping categories. For example:
- A paper can be both "research_article" AND "empirical_study" AND "sample_study"
- A paper can be both "systematic_review" AND "meta_analysis"
- A paper can be both "methods_paper" AND "empirical_study"

MANUSCRIPT DATA:
%s

CLASSIFICATION CATEGORIES:

1. TRADITIONAL PUBLICATION TYPES (select all that apply):
- research_article: Original research with methods and results
- review: Literature review or narrative review
- systematic_review: Following structured protocols (e.g., PRISMA)
- meta_analysis: Statistical synthesis of multiple studies
- editorial: Opinion piece by editors
- letter: Correspondence to editors
- case_report: Report of individual case(s)
- commentary: Comment on published work
- perspective: Author viewpoint/opinion

2. METHODOLOGICAL TYPES (select all that apply):
- empirical_study: Based on observation/experimentation with data collection
- theoretical_paper: Conceptual work without empirical data
- methods_paper: Presenting new methods/techniques/protocols

3. STUDY SCOPE (for empirical studies, select if applicable):
- single_case_study: In-depth analysis of single case/patient/organization (n=1)
- sample_study: Multiple participants/subjects (cohort, cross-sectional, survey, etc.)

RESPONSE FORMAT:
Provide a JSON object with:
{
  "primary_type": "most_specific_type",
  "all_types": ["type1", "type2", "type3"],
  "methodological_types": ["empirical_study", "theoretical_paper", or "methods_paper"],
  "scope_types": ["single_case_study" or "sample_study"] if applicable
}

Example response for a research article with empirical data from multiple participants:
{
  "primary_type": "research_article",
  "all_types": ["research_article", "empirical_study", "sample_study"],
  "methodological_types": ["empirical_study"],
  "scope_types": ["sample_study"]
}`, dataStr)

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
		// Fall back to rule-based
		text := title + " " + abstract
		return classifyArticleComprehensive(text), nil
	}

	// Call alembica
	logger.Info("Calling AI model for article type classification")
	result, err := extraction.Extract(string(jsonInput))
	if err != nil {
		logger.Error("AI extraction failed: %v", err)
		// Fall back to rule-based
		text := title + " " + abstract
		return classifyArticleComprehensive(text), nil
	}

	// Parse the response
	var output definitions.Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		logger.Error("Failed to parse AI response: %v", err)
		// Fall back to rule-based
		text := title + " " + abstract
		return classifyArticleComprehensive(text), nil
	}

	// Extract classification from the response
	if len(output.Responses) > 0 && len(output.Responses[0].ModelResponses) > 0 {
		response := output.Responses[0].ModelResponses[0]

		// Try to parse JSON response
		var aiClassification struct {
			PrimaryType         string   `json:"primary_type"`
			AllTypes            []string `json:"all_types"`
			MethodologicalTypes []string `json:"methodological_types"`
			ScopeTypes          []string `json:"scope_types"`
		}

		if err := json.Unmarshal([]byte(response), &aiClassification); err != nil {
			logger.Error("Failed to parse AI classification response: %v", err)
			// Fall back to rule-based
			text := title + " " + abstract
			return classifyArticleComprehensive(text), nil
		}

		// Convert to ArticleClassification
		classification := &ArticleClassification{
			PrimaryType:         ArticleType(aiClassification.PrimaryType),
			AllTypes:            []ArticleType{},
			TypeScores:          make(map[ArticleType]float64),
			MethodologicalTypes: []ArticleType{},
			ScopeTypes:          []ArticleType{},
		}

		// Convert all types
		for _, t := range aiClassification.AllTypes {
			articleType := ArticleType(t)
			classification.AllTypes = append(classification.AllTypes, articleType)
			// Assign high score for AI-identified types
			classification.TypeScores[articleType] = 20.0
		}

		// Convert methodological types
		for _, t := range aiClassification.MethodologicalTypes {
			classification.MethodologicalTypes = append(classification.MethodologicalTypes, ArticleType(t))
		}

		// Convert scope types
		for _, t := range aiClassification.ScopeTypes {
			classification.ScopeTypes = append(classification.ScopeTypes, ArticleType(t))
		}

		// Ensure primary type is in AllTypes
		if classification.PrimaryType != Unknown {
			found := false
			for _, t := range classification.AllTypes {
				if t == classification.PrimaryType {
					found = true
					break
				}
			}
			if !found {
				classification.AllTypes = append([]ArticleType{classification.PrimaryType}, classification.AllTypes...)
			}
		}

		logger.Info("AI classification successful: primary=%s, all=%v", classification.PrimaryType, classification.AllTypes)
		return classification, nil
	}

	// If we couldn't get a valid response, fall back to rule-based
	logger.Info("Could not extract valid classification from AI response, falling back to rule-based")
	text := title + " " + abstract
	return classifyArticleComprehensive(text), nil
}

// ClassifyArticleTypeWithScores returns classification with confidence scores
func ClassifyArticleTypeWithScores(text string) ([]ArticleTypeScore, error) {
	classification, err := ClassifyArticleTypes(text, nil)
	if err != nil {
		return nil, err
	}

	var scores []ArticleTypeScore
	for articleType, score := range classification.TypeScores {
		confidence := "low"
		if score > 15 {
			confidence = "high"
		} else if score > 8 {
			confidence = "medium"
		}

		scores = append(scores, ArticleTypeScore{
			Type:       articleType,
			Score:      score,
			Confidence: confidence,
		})
	}

	// Sort scores in descending order
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].Score > scores[i].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	return scores, nil
}

// BatchClassifyArticleTypes classifies multiple articles
func BatchClassifyArticleTypes(texts []string, llmConfigs []any) ([]string, error) {
	results := make([]string, len(texts))

	for i, text := range texts {
		result, err := ClassifyArticleType(text, llmConfigs)
		if err != nil {
			results[i] = string(Unknown)
		} else {
			results[i] = result
		}
	}

	return results, nil
}

// ArticleTypeStatistics represents statistics about article type distribution
type ArticleTypeStatistics struct {
	TotalArticles int                `json:"total_articles"`
	Distribution  map[string]int     `json:"distribution"`
	Percentages   map[string]float64 `json:"percentages"`
}

// CalculateArticleTypeStatistics calculates statistics for article types
func CalculateArticleTypeStatistics(articleTypes []string) ArticleTypeStatistics {
	stats := ArticleTypeStatistics{
		TotalArticles: len(articleTypes),
		Distribution:  make(map[string]int),
		Percentages:   make(map[string]float64),
	}

	for _, articleType := range articleTypes {
		stats.Distribution[articleType]++
	}

	for articleType, count := range stats.Distribution {
		stats.Percentages[articleType] = (float64(count) / float64(stats.TotalArticles)) * 100
	}

	return stats
}

// ExportClassificationResults exports classification results to various formats
func ExportClassificationResults(classifications []*ArticleClassification, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(classifications, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GetArticleTypeDescription returns a description of the article type
func GetArticleTypeDescription(articleType string) string {
	descriptions := map[string]string{
		"research_article":  "Original research presenting new empirical findings",
		"review":            "Review of existing literature on a topic",
		"systematic_review": "Systematic review following structured methodology",
		"meta_analysis":     "Statistical analysis of multiple studies",
		"editorial":         "Editorial or opinion piece by journal editors",
		"letter":            "Letter to the editor or correspondence",
		"case_report":       "Report of a specific case or patient",
		"commentary":        "Commentary on published work",
		"perspective":       "Author's perspective on a topic",
		"empirical_study":   "Study based on observation or experimentation with data",
		"theoretical_paper": "Conceptual or theoretical work without empirical data",
		"methods_paper":     "Paper presenting new methods or techniques",
		"single_case_study": "In-depth analysis of a single case",
		"sample_study":      "Study involving multiple participants or subjects",
		"unknown":           "Article type could not be determined",
	}

	if desc, exists := descriptions[articleType]; exists {
		return desc
	}
	return "No description available"
}

// Helper functions for checking article types in exclusion logic

// HasArticleType checks if a classification contains a specific type
func HasArticleType(classification *ArticleClassification, articleType ArticleType) bool {
	for _, t := range classification.AllTypes {
		if t == articleType {
			return true
		}
	}
	return false
}

// HasAnyArticleType checks if a classification contains any of the specified types
func HasAnyArticleType(classification *ArticleClassification, types ...ArticleType) bool {
	for _, t := range types {
		if HasArticleType(classification, t) {
			return true
		}
	}
	return false
}

// ParseArticleClassification parses the JSON classification string back to struct
func ParseArticleClassification(classificationJSON string) (*ArticleClassification, error) {
	var classification ArticleClassification

	// First try to parse as JSON
	err := json.Unmarshal([]byte(classificationJSON), &classification)
	if err != nil {
		// Fallback: treat as simple string type for backward compatibility
		classification = ArticleClassification{
			PrimaryType: ArticleType(classificationJSON),
			AllTypes:    []ArticleType{ArticleType(classificationJSON)},
		}
	}

	return &classification, nil
}
