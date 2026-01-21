package filters

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/alembica/extraction"
	"github.com/open-and-sustainable/alembica/utils/logger"
)

// DetectLanguage performs rule-based language detection
func DetectLanguage(text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("empty text provided")
	}

	// Normalize text
	text = strings.ToLower(strings.TrimSpace(text))

	// Count character frequencies for different scripts
	scripts := detectScripts(text)

	// Check for common language patterns
	if lang := detectByCommonWords(text); lang != "" {
		return lang, nil
	}

	// Check by script if common words didn't work
	if scripts["latin"] > 0.8 {
		// Default to English for Latin script if no specific language detected
		return "en", nil
	} else if scripts["cyrillic"] > 0.5 {
		return "ru", nil
	} else if scripts["greek"] > 0.5 {
		return "el", nil
	} else if scripts["arabic"] > 0.5 {
		return "ar", nil
	} else if scripts["hebrew"] > 0.5 {
		return "he", nil
	} else if scripts["cjk"] > 0.5 {
		// Could be Chinese, Japanese, or Korean
		return detectCJKLanguage(text), nil
	}

	return "unknown", nil
}

// LanguageDetectionConfig represents configuration for AI-based language detection
type LanguageDetectionConfig struct {
	UseAI      bool
	LLMConfigs []any // LLM configurations for AI-based detection
}

// BatchDetectLanguagesWithAI processes multiple manuscripts in a single AI call
func BatchDetectLanguagesWithAI(manuscriptsData []map[string]string, llmConfigs []any) map[string]string {
	results := make(map[string]string)

	// Initialize all results as "unknown"
	for i := range manuscriptsData {
		results[fmt.Sprintf("%d", i)] = "unknown"
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
		logger.Info("No valid AI models configured for language detection")
		return results
	}

	// Build all prompts
	var prompts []definitions.Prompt
	validIndices := []int{} // Track which manuscripts have valid data

	for idx, manuscriptData := range manuscriptsData {
		// Extract relevant fields
		title := ""
		abstract := ""
		journal := ""

		for field, value := range manuscriptData {
			fieldLower := strings.ToLower(field)
			switch fieldLower {
			case "title":
				title = value
			case "abstract":
				abstract = value
			case "journal", "journal_name", "publication":
				journal = value
			}
		}

		// Skip if no text to analyze
		if title == "" && abstract == "" && journal == "" {
			continue
		}

		// Build the data string for AI analysis
		dataStr := buildLanguageDetectionData(title, abstract, journal)

		// Create the prompt
		prompt := fmt.Sprintf(`You are a language detection expert analyzing scientific manuscripts. You need to identify the primary language of a manuscript based on the provided fields.

CONTEXT:
- You are analyzing: title, abstract, and journal information
- IMPORTANT: Many scientific databases translate abstracts to English while keeping the original title
- The title language is often more reliable than abstract language
- Journal names may indicate regional publications (e.g., "Revista Española", "Deutsche Zeitschrift")

SPECIAL CONSIDERATIONS:
- If title is in one language but abstract is in English, prioritize the title language
- Look for language-specific characters (é, ñ, ü, ø, etc.)
- Consider scientific Latin terms as part of the surrounding language context
- Mixed language content: identify the dominant/primary language

MANUSCRIPT DATA:
%s

TASK: Identify the primary language of this manuscript.
Respond with ONLY a JSON object with the ISO 639-1 language code: {"language": "en"} or {"language": "es"} or {"language": "fr"} etc.
Common codes: en (English), es (Spanish), fr (French), de (German), it (Italian), pt (Portuguese), ru (Russian), zh (Chinese), ja (Japanese), ar (Arabic)`, dataStr)

		prompts = append(prompts, definitions.Prompt{
			PromptContent:  prompt,
			SequenceID:     fmt.Sprintf("%d", idx+1),
			SequenceNumber: idx + 1,
		})
		validIndices = append(validIndices, idx)
	}

	if len(prompts) == 0 {
		logger.Info("No valid manuscripts for language detection")
		return results
	}

	logger.Info("Prepared %d manuscripts for batch language detection", len(prompts))

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
	logger.Info("Calling AI model with batch of %d language detection requests", len(prompts))
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
			var langResponse map[string]string
			if err := json.Unmarshal([]byte(response), &langResponse); err != nil {
				// Try to extract language code from plain text
				logger.Info("Failed to parse JSON response for manuscript %d, attempting text extraction", manuscriptIdx)
				response = strings.ToLower(response)
				// Look for common language codes in the response
				languageCodes := []string{"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ar", "nl", "sv", "no", "da", "fi", "pl", "cs", "hu", "ro", "el", "tr", "he", "ko"}
				for _, code := range languageCodes {
					if strings.Contains(response, `"`+code+`"`) || strings.Contains(response, "'"+code+"'") || strings.Contains(response, " "+code+" ") {
						results[fmt.Sprintf("%d", manuscriptIdx)] = code
						break
					}
				}
			} else {
				if lang, ok := langResponse["language"]; ok && lang != "" {
					results[fmt.Sprintf("%d", manuscriptIdx)] = lang
				}
			}
		}
	}

	return results
}

// buildLanguageDetectionData builds a formatted string with manuscript data for language detection
func buildLanguageDetectionData(title, abstract, journal string) string {
	var parts []string

	if title != "" {
		parts = append(parts, fmt.Sprintf("TITLE: %s", title))
	}
	if abstract != "" {
		// Limit abstract to first 500 characters for efficiency
		abstractSample := abstract
		if len(abstract) > 500 {
			abstractSample = abstract[:500] + "..."
		}
		parts = append(parts, fmt.Sprintf("ABSTRACT: %s", abstractSample))
	}
	if journal != "" {
		parts = append(parts, fmt.Sprintf("JOURNAL: %s", journal))
	}

	if len(parts) == 0 {
		return "[No data available]"
	}

	return strings.Join(parts, "\n")
}

// detectScripts analyzes the script composition of the text
func detectScripts(text string) map[string]float64 {
	scripts := map[string]int{
		"latin":    0,
		"cyrillic": 0,
		"greek":    0,
		"arabic":   0,
		"hebrew":   0,
		"cjk":      0,
		"other":    0,
	}

	totalChars := 0

	for _, r := range text {
		if unicode.IsLetter(r) {
			totalChars++

			if isLatinScript(r) {
				scripts["latin"]++
			} else if isCyrillicScript(r) {
				scripts["cyrillic"]++
			} else if isGreekScript(r) {
				scripts["greek"]++
			} else if isArabicScript(r) {
				scripts["arabic"]++
			} else if isHebrewScript(r) {
				scripts["hebrew"]++
			} else if isCJKScript(r) {
				scripts["cjk"]++
			} else {
				scripts["other"]++
			}
		}
	}

	// Convert to percentages
	percentages := make(map[string]float64)
	if totalChars > 0 {
		for script, count := range scripts {
			percentages[script] = float64(count) / float64(totalChars)
		}
	}

	return percentages
}

// detectByCommonWords checks for common words in various languages
func detectByCommonWords(text string) string {
	// Common words in different languages
	commonWords := map[string][]string{
		"en": {"the", "and", "of", "to", "in", "is", "that", "for", "with", "as", "on", "by", "at", "from"},
		"es": {"el", "la", "de", "que", "y", "en", "un", "por", "con", "para", "los", "las", "del"},
		"fr": {"le", "de", "et", "la", "les", "des", "est", "un", "une", "dans", "que", "pour", "sur"},
		"de": {"der", "die", "und", "das", "den", "dem", "des", "ist", "mit", "von", "für", "auf", "ein"},
		"it": {"il", "di", "e", "la", "che", "in", "un", "per", "con", "del", "della", "dei", "delle"},
		"pt": {"o", "de", "e", "que", "do", "da", "em", "um", "para", "com", "na", "os", "dos"},
		"nl": {"de", "het", "van", "een", "in", "en", "is", "op", "aan", "met", "voor", "zijn", "dat"},
		"ru": {"и", "в", "на", "с", "что", "это", "не", "как", "для", "по", "из", "у", "от"},
		"zh": {"的", "一", "是", "了", "我", "不", "在", "人", "有", "他", "这", "为", "之"},
		"ja": {"の", "に", "は", "を", "が", "と", "で", "て", "も", "な", "い", "か", "ある"},
		"ar": {"في", "من", "على", "إلى", "أن", "هذا", "ذلك", "التي", "الذي", "كان", "هو", "هي"},
	}

	// Count occurrences for each language
	scores := make(map[string]int)
	words := strings.Fields(text)

	for lang, wordList := range commonWords {
		for _, word := range words {
			for _, commonWord := range wordList {
				if strings.EqualFold(word, commonWord) {
					scores[lang]++
				}
			}
		}
	}

	// Find language with highest score
	maxScore := 0
	detectedLang := ""

	for lang, score := range scores {
		if score > maxScore {
			maxScore = score
			detectedLang = lang
		}
	}

	// Require at least 3 common word matches
	if maxScore >= 3 {
		return detectedLang
	}

	return ""
}

// Script detection helper functions
func isLatinScript(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= 0x00C0 && r <= 0x00FF) || // Latin-1 Supplement
		(r >= 0x0100 && r <= 0x017F) || // Latin Extended-A
		(r >= 0x0180 && r <= 0x024F) // Latin Extended-B
}

func isCyrillicScript(r rune) bool {
	return (r >= 0x0400 && r <= 0x04FF) || // Cyrillic
		(r >= 0x0500 && r <= 0x052F) // Cyrillic Supplement
}

func isGreekScript(r rune) bool {
	return (r >= 0x0370 && r <= 0x03FF) || // Greek and Coptic
		(r >= 0x1F00 && r <= 0x1FFF) // Greek Extended
}

func isArabicScript(r rune) bool {
	return (r >= 0x0600 && r <= 0x06FF) || // Arabic
		(r >= 0x0750 && r <= 0x077F) || // Arabic Supplement
		(r >= 0xFB50 && r <= 0xFDFF) || // Arabic Presentation Forms-A
		(r >= 0xFE70 && r <= 0xFEFF) // Arabic Presentation Forms-B
}

func isHebrewScript(r rune) bool {
	return r >= 0x0590 && r <= 0x05FF
}

func isCJKScript(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana
		(r >= 0xAC00 && r <= 0xD7AF) // Hangul Syllables
}

// detectCJKLanguage attempts to distinguish between Chinese, Japanese, and Korean
func detectCJKLanguage(text string) string {
	hasHiragana := false
	hasKatakana := false
	hasHangul := false
	hasChinese := false

	for _, r := range text {
		if r >= 0x3040 && r <= 0x309F {
			hasHiragana = true
		} else if r >= 0x30A0 && r <= 0x30FF {
			hasKatakana = true
		} else if r >= 0xAC00 && r <= 0xD7AF {
			hasHangul = true
		} else if r >= 0x4E00 && r <= 0x9FFF {
			hasChinese = true
		}
	}

	// Japanese has hiragana or katakana
	if hasHiragana || hasKatakana {
		return "ja"
	}

	// Korean has hangul
	if hasHangul {
		return "ko"
	}

	// Chinese has only Chinese characters
	if hasChinese {
		// Could distinguish between simplified (zh-CN) and traditional (zh-TW)
		// but would need more sophisticated detection
		return "zh"
	}

	return "unknown"
}

// GetLanguageName returns the full name of a language from its ISO code
func GetLanguageName(code string) string {
	languages := map[string]string{
		"en": "English",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"nl": "Dutch",
		"ru": "Russian",
		"zh": "Chinese",
		"ja": "Japanese",
		"ko": "Korean",
		"ar": "Arabic",
		"he": "Hebrew",
		"el": "Greek",
	}

	if name, exists := languages[strings.ToLower(code)]; exists {
		return name
	}

	return "Unknown"
}
