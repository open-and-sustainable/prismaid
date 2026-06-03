package revaise

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func stringValue(item map[string]any, key string) string {
	if value, ok := item[key]; ok {
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return ""
}

func list(record map[string]any, key string) []any {
	value, ok := record[key]
	if !ok || value == nil {
		return nil
	}
	items, ok := value.([]any)
	if ok {
		return items
	}
	return nil
}

func upsertByKey(items []any, replacement map[string]any, same func(map[string]any) bool) []any {
	for index, item := range items {
		existing, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if same(existing) {
			items[index] = mergeMaps(existing, replacement)
			return items
		}
	}
	return append(items, replacement)
}

func mergeMaps(existing, replacement map[string]any) map[string]any {
	merged := make(map[string]any, len(existing)+len(replacement))
	for key, value := range existing {
		merged[key] = value
	}
	for key, value := range replacement {
		if isEmpty(value) {
			continue
		}
		merged[key] = value
	}
	return merged
}

func upsertLiteratureRecords(record Record, records []map[string]any) {
	items := list(record, "literature_records")
	for _, literatureRecord := range records {
		recordID := stringValue(literatureRecord, "record_id")
		if recordID == "" {
			continue
		}
		items = upsertByKey(items, literatureRecord, func(existing map[string]any) bool {
			return stringValue(existing, "record_id") == recordID
		})
	}
	record["literature_records"] = items
}

func upsertStage(record Record, stage map[string]any) map[string]any {
	stageType := stringValue(stage, "stage_type")
	stageLabel := stringValue(stage, "stage_label")
	items := list(record, "stages")
	for index, item := range items {
		existing, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if stringValue(existing, "stage_type") == stageType && stringValue(existing, "stage_label") == stageLabel {
			merged := mergeMaps(existing, stage)
			items[index] = merged
			record["stages"] = items
			return merged
		}
	}
	items = append(items, stage)
	record["stages"] = items
	return stage
}

func upsertStageOutputs(container map[string]any, outputs []map[string]any) {
	items := list(container, "outputs")
	for _, output := range outputs {
		uri := stringValue(output, "resource_uri")
		if uri == "" {
			continue
		}
		items = upsertByKey(items, output, func(existing map[string]any) bool {
			return stringValue(existing, "resource_uri") == uri
		})
	}
	container["outputs"] = items
}

func stageFromConfig(cfg Config, defaultType, defaultLabel string) map[string]any {
	stageType := firstNonEmpty(cfg.Stage.Type, defaultType)
	stageLabel := firstNonEmpty(cfg.Stage.Label, defaultLabel, stageType)
	stage := map[string]any{
		"stage_type":  stageType,
		"stage_label": stageLabel,
	}
	if cfg.Stage.Description != "" {
		stage["stage_description"] = cfg.Stage.Description
	}
	if cfg.Stage.StartedAt != "" {
		stage["started_at"] = cfg.Stage.StartedAt
	}
	if cfg.Stage.EndedAt != "" {
		stage["ended_at"] = cfg.Stage.EndedAt
	}
	return stage
}

func stageOutput(kind, resourceURI, format string) map[string]any {
	output := map[string]any{
		"kind":         kind,
		"resource_uri": resourceURI,
		"format":       format,
	}
	output["output_created_at"] = nowRFC3339()
	return output
}

func literatureRecordFromData(id string, data map[string]string) map[string]any {
	lookup := make(map[string]string)
	for key, value := range data {
		lookup[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	get := func(names ...string) string {
		for _, name := range names {
			if value := lookup[strings.ToLower(name)]; value != "" {
				return value
			}
		}
		return ""
	}

	recordID := firstNonEmpty(id, get("id", "record_id", "doi", "pmid"))
	if recordID == "" {
		recordID = slug(get("title"))
	}
	title := firstNonEmpty(get("title", "publication_title"), recordID)
	record := map[string]any{
		"record_id": recordID,
		"title":     title,
	}

	if authors := parseAuthors(get("authors", "author", "creators")); len(authors) > 0 {
		record["authors"] = authors
	}
	if year := parseYear(get("publication_year", "year", "date", "published")); year != nil {
		record["publication_year"] = *year
	}
	copyField(record, "abstract", get("abstract", "summary"))
	copyField(record, "doi", get("doi"))
	copyField(record, "pmid", get("pmid", "pubmed_id"))
	copyField(record, "journal", get("journal", "journal_title", "source_title", "publication"))
	copyField(record, "volume", get("volume"))
	copyField(record, "issue", get("issue"))
	copyField(record, "pages", get("pages"))
	copyField(record, "publication_language", get("language", "publication_language"))
	copyField(record, "publication_type", get("publication_type", "type", "article_type"))
	copyField(record, "full_text_url", get("full_text_url", "url", "link"))
	record["record_updated_at"] = nowRFC3339()
	return record
}

func copyField(record map[string]any, key, value string) {
	if strings.TrimSpace(value) != "" {
		record[key] = strings.TrimSpace(value)
	}
}

func parseAuthors(value string) []any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	separator := ";"
	if strings.Contains(value, " and ") {
		separator = " and "
	}
	parts := strings.Split(value, separator)
	authors := make([]any, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, " ,")
		if part != "" {
			authors = append(authors, map[string]any{"name": part})
		}
	}
	return authors
}

func parseYear(value string) *int {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if len(value) >= 4 {
		value = value[:4]
	}
	year, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &year
}

func fileURI(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "file://") {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return "file://" + abs
}
