package filters

import (
	"testing"
)

// TestLanguageDetection tests the language detection functionality
func TestLanguageDetection(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "English text",
			text:     "This is an English text about scientific research and methodology. The study examines the effects of climate change.",
			expected: "en",
		},
		{
			name:     "Spanish text",
			text:     "Este es un texto en español sobre investigación científica. El estudio examina los efectos del cambio climático.",
			expected: "es",
		},
		{
			name:     "French text",
			text:     "Ceci est un texte en français sur la recherche scientifique. L'étude examine les effets du changement climatique.",
			expected: "fr",
		},
		{
			name:     "German text",
			text:     "Dies ist ein deutscher Text über wissenschaftliche Forschung. Die Studie untersucht die Auswirkungen des Klimawandels.",
			expected: "de",
		},
		{
			name:     "Italian text",
			text:     "Questo è un testo italiano sulla ricerca scientifica. Il lavoro esamina gli effetti del cambiamento climatico con il metodo della analisi.",
			expected: "it",
		},
		{
			name:     "Portuguese text",
			text:     "Este é um texto em português sobre pesquisa científica. O estudo examina os efeitos das mudanças climáticas.",
			expected: "pt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DetectLanguage(tt.text)
			if err != nil {
				t.Errorf("DetectLanguage() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("DetectLanguage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetLanguageName tests language name lookup
func TestGetLanguageName(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"en", "English"},
		{"es", "Spanish"},
		{"fr", "French"},
		{"de", "German"},
		{"zh", "Chinese"},
		{"ja", "Japanese"},
		{"xx", "Unknown"},
		{"", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := GetLanguageName(tt.code)
			if result != tt.expected {
				t.Errorf("GetLanguageName(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}
