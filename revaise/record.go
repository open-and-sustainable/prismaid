package revaise

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Record is a RevAIse Review document represented as a mutable object.
type Record map[string]any

// ReviewSeed carries minimal project metadata used when creating a new Review.
type ReviewSeed struct {
	ID       string
	Title    string
	Type     string
	Status   string
	Version  string
	Language string
	Country  string
	Authors  []string
}

func ensureRoot(record Record, cfg Config, seed ReviewSeed) {
	if isEmpty(record["review_id"]) {
		record["review_id"] = firstNonEmpty(cfg.Review.ID, seed.ID, slug(firstNonEmpty(cfg.Review.Title, seed.Title, "prismaid review")))
	}
	if isEmpty(record["review_title"]) {
		record["review_title"] = firstNonEmpty(cfg.Review.Title, seed.Title, "prismAId review")
	}
	if isEmpty(record["review_type"]) {
		record["review_type"] = firstNonEmpty(cfg.Review.Type, seed.Type, "SYSTEMATIC_REVIEW")
	}
	if isEmpty(record["review_status"]) {
		record["review_status"] = firstNonEmpty(cfg.Review.Status, seed.Status, "IN_PROGRESS")
	}
	if version := firstNonEmpty(cfg.Review.Version, seed.Version); version != "" && isEmpty(record["version"]) {
		record["version"] = version
	}
	if language := firstNonEmpty(cfg.Review.Language, seed.Language); language != "" && isEmpty(record["review_language"]) {
		record["review_language"] = language
	}
	if country := firstNonEmpty(cfg.Review.Country, seed.Country); country != "" && isEmpty(record["review_country"]) {
		record["review_country"] = country
	}
	if isEmpty(record["created_at"]) {
		record["created_at"] = nowRFC3339()
	}
	record["updated_at"] = nowRFC3339()

	authors := cfg.Review.Authors
	if len(authors) == 0 {
		authors = seed.Authors
	}
	if len(authors) == 0 {
		authors = []string{"prismAId"}
	}
	if isEmpty(record["review_authors"]) {
		record["review_authors"] = authorObjects(authors)
	}
}

func authorObjects(authors []string) []any {
	result := make([]any, 0, len(authors))
	for _, author := range authors {
		author = strings.TrimSpace(author)
		if author == "" {
			continue
		}
		result = append(result, map[string]any{"name": author})
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	switch typed := value.(type) {
	case string:
		return typed == ""
	case []any:
		return len(typed) == 0
	default:
		return false
	}
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	if value == "" {
		return "prismaid_review"
	}
	return value
}

func requireEnabled(cfg Config) error {
	if !cfg.IsEnabled() {
		return nil
	}
	if strings.TrimSpace(cfg.RecordFile) == "" {
		return fmt.Errorf("revaise record_file is required when revaise is enabled")
	}
	return nil
}
