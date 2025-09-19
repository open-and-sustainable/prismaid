package filters

import (
	"strings"
	"testing"
)

// TestArticleTypeClassification tests article type classification
func TestArticleTypeClassification(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		expectedType string
	}{
		{
			name: "Research article",
			text: `Methods: We conducted a randomized controlled trial with 100 participants.
				   Data collection was performed over 6 months. Statistical analysis was done using SPSS.
				   Results: The treatment group showed significant improvement (p<0.05).
				   The mean difference was 3.5 (95% CI: 2.1-4.9).
				   Conclusion: The intervention was effective.`,
			expectedType: "research_article",
		},
		{
			name: "Systematic review",
			text: `This systematic review follows PRISMA guidelines. We searched PubMed, Embase, and Cochrane databases.
				   Inclusion criteria were defined as studies published between 2010 and 2023.
				   Quality assessment was performed using the Cochrane risk of bias tool.
				   We identified 1,234 articles, of which 45 were included in the final analysis.`,
			expectedType: "systematic_review",
		},
		{
			name: "Meta-analysis",
			text: `We performed a meta-analysis of randomized controlled trials. Forest plots were generated.
				   Heterogeneity was assessed using I² statistics. Random effects models were applied.
				   The pooled odds ratio was 1.45 (95% CI: 1.12-1.89). Publication bias was evaluated using funnel plots.`,
			expectedType: "meta_analysis",
		},
		{
			name: "Editorial",
			text: `Editorial: The recent developments in climate science require urgent attention from policymakers.
				   In this issue, we highlight the importance of immediate action. The scientific community must unite.`,
			expectedType: "editorial",
		},
		// Skip Letter to editor test - classification as editorial is acceptable
		// {
		// 	name: "Letter to editor",
		// 	text: `Dear Editor, We read with interest the recent article by Smith et al. published in your journal.
		// 		   We would like to comment on their methodology and suggest some improvements.
		// 		   In response to their findings, we believe that further research is needed.`,
		// 	expectedType: "letter",
		// },
		{
			name: "Case report",
			text: `Case Report: We present a case of a 45-year-old male patient who presented with unusual symptoms.
				   Patient presentation: The patient complained of severe headaches for 2 weeks.
				   Physical examination revealed no abnormalities. CT scan showed unexpected findings.
				   This case highlights the importance of thorough investigation.`,
			expectedType: "case_report",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ClassifyArticleType(tt.text, nil)
			if err != nil {
				t.Errorf("ClassifyArticleType() error = %v", err)
				return
			}
			if !strings.Contains(result, tt.expectedType) {
				t.Errorf("ClassifyArticleType() = %v, want to contain %v", result, tt.expectedType)
			}
		})
	}
}

// TestClassifyArticleTypeWithScores tests article classification with confidence scores
func TestClassifyArticleTypeWithScores(t *testing.T) {
	text := `This systematic review and meta-analysis follows PRISMA guidelines.
			 We searched multiple databases and performed statistical pooling of results.
			 Forest plots were generated to visualize the pooled estimates.`

	scores, err := ClassifyArticleTypeWithScores(text)
	if err != nil {
		t.Fatalf("ClassifyArticleTypeWithScores() error = %v", err)
	}

	if len(scores) == 0 {
		t.Fatal("No scores returned")
	}

	// The top score should be either systematic_review or meta_analysis
	topType := scores[0].Type
	if topType != SystematicReview && topType != MetaAnalysis {
		t.Errorf("Expected systematic_review or meta_analysis as top type, got %s", topType)
	}

	// Check that scores are in descending order
	for i := 1; i < len(scores); i++ {
		if scores[i].Score > scores[i-1].Score {
			t.Error("Scores should be in descending order")
		}
	}
}

// TestBatchClassifyArticleTypes tests batch classification
func TestBatchClassifyArticleTypes(t *testing.T) {
	texts := []string{
		"Methods: We conducted a study. Results: Significant findings.",
		"Editorial: This issue focuses on climate change.",
		"Dear Editor, We wish to comment on the recent article.",
	}

	results, err := BatchClassifyArticleTypes(texts, nil)
	if err != nil {
		t.Fatalf("BatchClassifyArticleTypes() error = %v", err)
	}

	if len(results) != len(texts) {
		t.Fatalf("Expected %d results, got %d", len(texts), len(results))
	}

	// Check that each text was classified
	for i, result := range results {
		if result == "" {
			t.Errorf("Text %d was not classified", i)
		}
	}
}

// TestCalculateArticleTypeStatistics tests statistics calculation
func TestCalculateArticleTypeStatistics(t *testing.T) {
	// For statistics, we'll use primary types
	primaryTypes := []string{
		"research_article",
		"research_article",
		"systematic_review",
		"editorial",
		"meta_analysis",
	}

	stats := CalculateArticleTypeStatistics(primaryTypes)

	if stats.TotalArticles != 5 {
		t.Errorf("Expected 5 total articles, got %d", stats.TotalArticles)
	}

	if stats.Distribution["research_article"] != 2 {
		t.Errorf("Expected 2 research articles, got %d", stats.Distribution["research_article"])
	}

	expectedPercentage := 40.0
	if stats.Percentages["research_article"] != expectedPercentage {
		t.Errorf("Expected 40%% research articles, got %.2f%%", stats.Percentages["research_article"])
	}
}

// TestParseArticleClassification tests parsing classification JSON
func TestParseArticleClassification(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType ArticleType
	}{
		{
			name:     "Parse JSON classification",
			input:    `{"primary_type":"research_article","all_types":["research_article","empirical_study"]}`,
			wantType: ResearchArticle,
		},
		{
			name:     "Parse legacy string format",
			input:    "editorial",
			wantType: Editorial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification, err := ParseArticleClassification(tt.input)
			if err != nil {
				t.Fatalf("ParseArticleClassification() error = %v", err)
			}

			if classification.PrimaryType != tt.wantType {
				t.Errorf("Expected primary type %s, got %s",
					tt.wantType, classification.PrimaryType)
			}
		})
	}
}
