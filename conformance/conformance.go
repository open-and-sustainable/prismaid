// Package conformance checks whether a RevAIse review record satisfies a
// reporting protocol, using the SHACL shapes published by the RevAIse model.
//
// A record produced by prismAId (plain JSON following the RevAIse data model) is
// framed as JSON-LD with the RevAIse context, expanded to RDF, and validated
// against a protocol's SHACL shapes. The verdict and the per-constraint messages
// come entirely from the protocol's shapes, so conformance is decided
// symbolically rather than asserted by the tool. Protocols are selected by name
// and are pluggable: adding one is a matter of embedding its shapes and
// registering them in the protocols map.
package conformance

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tggo/goRDFlib/shacl"
)

//go:embed protocols/prisma-2020.shacl.ttl
var prisma2020Shapes string

// revaiseVocab is the RevAIse schema namespace. RevAIse slot names equal the
// JSON keys used in records, so a minimal JSON-LD context with only this @vocab
// maps every record key and type to the absolute IRI the SHACL shapes target
// (the shapes use the same namespace under the revaise: prefix).
const revaiseVocab = "https://open-and-sustainable.github.io/revaise-model/schema/"

// protocols maps a protocol identifier to its SHACL shapes (Turtle). To support
// a new protocol, embed its shapes above and add an entry here.
var protocols = map[string]string{
	"prisma-2020": prisma2020Shapes,
}

// Report is the outcome of a conformance check.
type Report struct {
	Protocol   string      `json:"protocol"`
	Conforms   bool        `json:"conforms"`
	Violations []Violation `json:"violations"`
}

// Violation is a single unmet constraint, carrying the protocol's own message.
type Violation struct {
	FocusNode string `json:"focus_node,omitempty"`
	Path      string `json:"path,omitempty"`
	Message   string `json:"message"`
}

// AvailableProtocols returns the registered protocol identifiers, sorted.
func AvailableProtocols() []string {
	names := make([]string, 0, len(protocols))
	for name := range protocols {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Check validates a RevAIse review record (JSON) against the named protocol's
// SHACL shapes and returns a Report. An unknown protocol is an error.
//
// The check is read-only and offline: it uses the vendored RevAIse context and
// protocol shapes, and performs no network access.
func Check(recordJSON, protocol string) (*Report, error) {
	shapes, ok := protocols[protocol]
	if !ok {
		return nil, fmt.Errorf("unknown protocol %q: available protocols are %s", protocol, strings.Join(AvailableProtocols(), ", "))
	}

	framed, err := frameRecord(recordJSON)
	if err != nil {
		return nil, fmt.Errorf("preparing record: %w", err)
	}

	dataGraph, err := shacl.LoadJsonLDString(framed, "")
	if err != nil {
		return nil, fmt.Errorf("parsing record as RDF: %w", err)
	}
	shapesGraph, err := shacl.LoadTurtleString(shapes, "")
	if err != nil {
		return nil, fmt.Errorf("parsing %s shapes: %w", protocol, err)
	}

	result := shacl.Validate(dataGraph, shapesGraph)
	report := &Report{Protocol: protocol, Conforms: result.Conforms}
	for _, r := range result.Results {
		report.Violations = append(report.Violations, Violation{
			FocusNode: r.FocusNode.Value(),
			Path:      r.ResultPath.Value(),
			Message:   messageText(r.ResultMessages),
		})
	}
	return report, nil
}

// frameRecord turns a plain RevAIse record into a JSON-LD document: it attaches
// the RevAIse context and adds the @type declarations the SHACL shapes target
// (the root Review, each stage by its stage_type, and literature records) so
// their targetClass constraints bind during validation.
func frameRecord(recordJSON string) (string, error) {
	var record map[string]any
	if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
		return "", err
	}

	record["@context"] = map[string]any{"@vocab": revaiseVocab}
	record["@type"] = "Review"

	if stages, ok := record["stages"].([]any); ok {
		for _, s := range stages {
			stage, ok := s.(map[string]any)
			if !ok {
				continue
			}
			if class := stageClass(fmt.Sprint(stage["stage_type"])); class != "" {
				stage["@type"] = class
			}
		}
	}

	if records, ok := record["literature_records"].([]any); ok {
		for _, l := range records {
			if lit, ok := l.(map[string]any); ok {
				lit["@type"] = "LiteratureRecord"
			}
		}
	}

	framed, err := json.Marshal(record)
	return string(framed), err
}

// stageClass maps a RevAIse stage_type to the RevAIse class the SHACL shapes
// target. Stage types without a targeted shape return an empty string and are
// left untyped (their absence is instead caught by the review-level shapes).
func stageClass(stageType string) string {
	switch stageType {
	case "search":
		return "SearchStage"
	case "screening_title_abstract", "screening_fulltext":
		return "ScreeningStage"
	case "data_extraction":
		return "ExtractionStage"
	case "risk_of_bias":
		return "RiskOfBiasAssessmentStage"
	case "synthesis_meta_analysis", "synthesis_narrative":
		return "SynthesisStage"
	default:
		return ""
	}
}

func messageText(messages []shacl.Term) string {
	parts := make([]string, 0, len(messages))
	for _, m := range messages {
		parts = append(parts, m.Value())
	}
	return strings.Join(parts, "; ")
}
