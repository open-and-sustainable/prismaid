package revaise

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// MergeStage merges a stage into an existing RevAIse record and returns the
// updated record as indented JSON. The stage is matched to an existing stage by
// stage_type and stage_label (filling a seed stub) or appended when none matches;
// only non-empty fields overwrite existing values. It bumps updated_at.
func MergeStage(recordJSON, stageJSON string) (string, error) {
	var record Record
	if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
		return "", fmt.Errorf("parsing record: %w", err)
	}
	var stage map[string]any
	if err := json.Unmarshal([]byte(stageJSON), &stage); err != nil {
		return "", fmt.Errorf("parsing stage: %w", err)
	}
	if stringValue(stage, "stage_type") == "" {
		return "", fmt.Errorf("stage must have a stage_type")
	}

	upsertStage(record, stage)
	record["updated_at"] = nowRFC3339()

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RecordValidation is the outcome of validating a record against the RevAIse
// data-model JSON Schema: whether it is valid, and any validation messages.
type RecordValidation struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// ValidateRecord validates a RevAIse record (JSON) against the released
// data-model JSON Schema, fetched live. A well-formed but incomplete record (for
// example a freshly seeded one) reports the missing required fields as errors. It
// requires network access; an operational failure (schema unreachable, unparsable
// schema) is returned as the error, distinct from a validation failure.
func ValidateRecord(recordJSON string) (*RecordValidation, error) {
	schemaStr, err := FetchSchema()
	if err != nil {
		return nil, err
	}
	schemaDoc, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaStr))
	if err != nil {
		return nil, fmt.Errorf("parsing RevAIse schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(dataModelSchemaURL, schemaDoc); err != nil {
		return nil, fmt.Errorf("loading RevAIse schema: %w", err)
	}
	schema, err := compiler.Compile(dataModelSchemaURL)
	if err != nil {
		return nil, fmt.Errorf("compiling RevAIse schema: %w", err)
	}

	instance, err := jsonschema.UnmarshalJSON(strings.NewReader(recordJSON))
	if err != nil {
		return &RecordValidation{Valid: false, Errors: []string{fmt.Sprintf("invalid JSON: %v", err)}}, nil
	}
	if err := schema.Validate(instance); err != nil {
		return &RecordValidation{Valid: false, Errors: validationMessages(err)}, nil
	}
	return &RecordValidation{Valid: true}, nil
}

// validationMessages flattens a schema validation error into readable lines.
func validationMessages(err error) []string {
	lines := strings.Split(err.Error(), "\n")
	messages := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "jsonschema validation failed") {
			continue
		}
		messages = append(messages, strings.TrimPrefix(trimmed, "- "))
	}
	return messages
}
