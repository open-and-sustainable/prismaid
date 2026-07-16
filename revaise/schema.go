package revaise

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Released RevAIse data-model artifacts, produced by RevAIse's build cycle and
// published on its documentation site. The /stable/ channel is the released,
// verified version; nothing is vendored — these are fetched live.
const (
	dataModelSchemaURL  = "https://revaise-model.readthedocs.io/stable/api/revaise.schema.json"
	dataModelContextURL = "https://revaise-model.readthedocs.io/stable/api/context.jsonld"
)

var schemaHTTPClient = &http.Client{Timeout: 30 * time.Second}

// SchemaSlot describes one property of a data-model class: its name, whether it
// is required, its type (a JSON type, a referenced class, or "array of X"), and,
// when the property is enum-valued, the allowed values.
type SchemaSlot struct {
	Name       string   `json:"name"`
	Type       string   `json:"type,omitempty"`
	Required   bool     `json:"required,omitempty"`
	EnumValues []string `json:"enum_values,omitempty"`
}

// SchemaDescription describes the RevAIse data model or one of its types. When
// no type is requested, Classes and Enums list the available names. When a class
// is requested, Required and Properties are populated; when an enum is requested,
// EnumValues is populated.
type SchemaDescription struct {
	Version    string       `json:"version,omitempty"`
	Classes    []string     `json:"classes,omitempty"`
	Enums      []string     `json:"enums,omitempty"`
	Type       string       `json:"type,omitempty"`
	Kind       string       `json:"kind,omitempty"`
	Required   []string     `json:"required,omitempty"`
	Properties []SchemaSlot `json:"properties,omitempty"`
	EnumValues []string     `json:"enum_values,omitempty"`
}

// FetchSchema returns the released RevAIse data-model JSON Schema as a string.
func FetchSchema() (string, error) {
	body, err := fetchArtifact(dataModelSchemaURL)
	return string(body), err
}

// FetchContext returns the released RevAIse JSON-LD context as a string.
func FetchContext() (string, error) {
	body, err := fetchArtifact(dataModelContextURL)
	return string(body), err
}

// DescribeSchema fetches the released data-model JSON Schema and describes it.
// With an empty typeName it lists the available classes and enums; with a class
// or enum name it describes that type (required slots, properties with types and
// inlined enum values, or the enum's allowed values). It requires network access.
func DescribeSchema(typeName string) (*SchemaDescription, error) {
	raw, err := fetchArtifact(dataModelSchemaURL)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Version string                     `json:"version"`
		Defs    map[string]json.RawMessage `json:"$defs"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parsing RevAIse schema: %w", err)
	}

	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		classes := make([]string, 0)
		enums := make([]string, 0)
		for name, def := range doc.Defs {
			if len(enumValues(def)) > 0 {
				enums = append(enums, name)
			} else {
				classes = append(classes, name)
			}
		}
		sort.Strings(classes)
		sort.Strings(enums)
		return &SchemaDescription{Version: doc.Version, Classes: classes, Enums: enums}, nil
	}

	def, ok := doc.Defs[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown type %q in the RevAIse data model (call without a type to list available classes and enums)", typeName)
	}

	desc := &SchemaDescription{Version: doc.Version, Type: typeName}
	if values := enumValues(def); len(values) > 0 {
		desc.Kind = "enum"
		desc.EnumValues = values
		return desc, nil
	}

	desc.Kind = "class"
	var parsed struct {
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(def, &parsed); err != nil {
		return nil, fmt.Errorf("parsing type %q: %w", typeName, err)
	}
	desc.Required = parsed.Required
	requiredSet := make(map[string]bool, len(parsed.Required))
	for _, r := range parsed.Required {
		requiredSet[r] = true
	}
	names := make([]string, 0, len(parsed.Properties))
	for name := range parsed.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		slotType, slotEnum := slotTypeAndEnum(parsed.Properties[name], doc.Defs)
		desc.Properties = append(desc.Properties, SchemaSlot{
			Name:       name,
			Type:       slotType,
			Required:   requiredSet[name],
			EnumValues: slotEnum,
		})
	}
	return desc, nil
}

// fetchArtifact performs a GET and returns the body, or an error for any non-200.
func fetchArtifact(url string) ([]byte, error) {
	resp, err := schemaHTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// enumValues returns the enum values declared directly on a JSON Schema node.
func enumValues(def json.RawMessage) []string {
	var d struct {
		Enum []string `json:"enum"`
	}
	_ = json.Unmarshal(def, &d)
	return d.Enum
}

// slotTypeAndEnum resolves a property node into a human-readable type and, when
// the property is enum-valued (directly or via a $ref to an enum), its values.
func slotTypeAndEnum(raw json.RawMessage, defs map[string]json.RawMessage) (string, []string) {
	var p struct {
		Type  any      `json:"type"`
		Ref   string   `json:"$ref"`
		Enum  []string `json:"enum"`
		Items struct {
			Type any    `json:"type"`
			Ref  string `json:"$ref"`
		} `json:"items"`
	}
	_ = json.Unmarshal(raw, &p)

	if len(p.Enum) > 0 {
		return "string", p.Enum
	}
	if p.Ref != "" {
		name := refName(p.Ref)
		return name, enumValues(defs[name])
	}
	if typeString(p.Type) == "array" {
		if p.Items.Ref != "" {
			name := refName(p.Items.Ref)
			return "array of " + name, enumValues(defs[name])
		}
		if it := typeString(p.Items.Type); it != "" {
			return "array of " + it, nil
		}
		return "array", nil
	}
	return typeString(p.Type), nil
}

func refName(ref string) string {
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		return ref[i+1:]
	}
	return ref
}

// typeString normalizes a JSON Schema "type" (a string, or an array like
// ["string","null"]) to a single type name, ignoring "null".
func typeString(t any) string {
	switch v := t.(type) {
	case string:
		return v
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "null" {
				return s
			}
		}
	}
	return ""
}
