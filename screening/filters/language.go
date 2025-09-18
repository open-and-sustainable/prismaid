package filters

import (
	"fmt"
	"strings"
	"unicode"
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

// DetectLanguageWithAI uses an LLM to detect language
func DetectLanguageWithAI(text string, llmConfig interface{}) (string, error) {
	// This would integrate with the alembica package for AI calls
	// For now, returning a placeholder implementation

	// Extract first 500 characters for efficiency
	sampleText := text
	if len(text) > 500 {
		sampleText = text[:500]
	}

	// In actual implementation, this would:
	// 1. Create a prompt asking the LLM to identify the language
	// 2. Call the appropriate LLM API through alembica
	// 3. Parse the response to extract the language code

	// Placeholder: fall back to rule-based detection
	return DetectLanguage(sampleText)
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
