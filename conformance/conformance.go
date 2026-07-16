// Package conformance checks whether a RevAIse review record satisfies a
// reporting protocol, using the SHACL shapes published by the RevAIse model.
//
// A record produced by prismAId (plain JSON following the RevAIse data model) is
// framed as JSON-LD with the RevAIse context, expanded to RDF, and validated
// against a protocol's SHACL shapes. The verdict and the per-constraint messages
// come entirely from the protocol's shapes, so conformance is decided
// symbolically rather than asserted by the tool.
//
// Protocols are selected by name and are never vendored: both the catalogue of
// available protocols and each protocol's shapes are pulled at call time from the
// latest versions RevAIse publishes on GitHub Pages, so adopting a new or revised
// protocol requires no change to prismAId.
package conformance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/tggo/goRDFlib/shacl"
)

// revaiseBaseURL is the root of the RevAIse model site on GitHub Pages. The
// protocol catalogue (index.json) and every protocol's shapes are published
// beneath it; prismAId always pulls the latest published versions from here.
const revaiseBaseURL = "https://open-and-sustainable.github.io/revaise-model/"

// revaiseVocab is the RevAIse schema namespace. RevAIse slot names equal the
// JSON keys used in records, so a minimal JSON-LD context with only this @vocab
// maps every record key and type to the absolute IRI the SHACL shapes target
// (the shapes use the same namespace under the revaise: prefix).
const revaiseVocab = revaiseBaseURL + "schema/"

// httpClient fetches the published catalogue and shapes.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// catalog mirrors the published index.json: a map of protocol identifier to its
// metadata, including the relative path of its SHACL shapes.
type catalog struct {
	Protocols map[string]protocolMeta `json:"protocols"`
}

type protocolMeta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	SHACL   string `json:"shacl"`
	Status  string `json:"status"`
}

// Report is the outcome of a conformance check. Conforms is the whole-protocol
// verdict; Summary, Passed, and Pending give a progress view so an in-progress
// review reads as partial rather than broken.
type Report struct {
	Protocol   string        `json:"protocol"`
	Conforms   bool          `json:"conforms"`
	Summary    Summary       `json:"summary"`
	Passed     []Requirement `json:"passed,omitempty"`
	Pending    []Requirement `json:"pending,omitempty"`
	Violations []Violation   `json:"violations"`
}

// Violation is a single unmet constraint, carrying the protocol's own message.
type Violation struct {
	FocusNode string `json:"focus_node,omitempty"`
	Path      string `json:"path,omitempty"`
	Message   string `json:"message"`
}

// ClassProgress summarizes how many of a record class's requirements are met.
// Present is false when the record has no instance of the class yet, in which
// case its requirements are counted as Pending (not started) rather than failed.
type ClassProgress struct {
	Present bool `json:"present"`
	Total   int  `json:"total"`
	Passed  int  `json:"passed"`
	Failed  int  `json:"failed"`
	Pending int  `json:"pending"`
}

// Summary is the progress view of a conformance check: overall counts and a
// per-record-class breakdown. Pending distinguishes requirements for stages that
// have not been started (class absent) from failed requirements, so an
// incomplete-by-design review is not mistaken for a broken one.
type Summary struct {
	Total   int                      `json:"total"`
	Passed  int                      `json:"passed"`
	Failed  int                      `json:"failed"`
	Pending int                      `json:"pending"`
	ByClass map[string]ClassProgress `json:"by_class,omitempty"`
}

// Requirement is one thing a protocol requires, with the record class it applies
// to (for example "Review" or "ScreeningStage") and the protocol's own message.
type Requirement struct {
	TargetClass string `json:"target_class,omitempty"`
	Message     string `json:"message"`
}

// Guidance is the full set of requirements a protocol imposes, together with its
// metadata. It is advisory: it helps plan a conforming review before any record
// exists and does not constrain the order in which prismAId's tools are used.
type Guidance struct {
	Protocol     string        `json:"protocol"`
	Name         string        `json:"name,omitempty"`
	Version      string        `json:"version,omitempty"`
	Status       string        `json:"status,omitempty"`
	Requirements []Requirement `json:"requirements"`
}

// AvailableProtocols returns the protocol identifiers RevAIse publishes, sorted.
// It fetches the published catalogue, so it requires network access and returns
// an error if the catalogue cannot be retrieved.
func AvailableProtocols() ([]string, error) {
	cat, err := fetchCatalog()
	if err != nil {
		return nil, err
	}
	return protocolNames(cat), nil
}

// Check validates a RevAIse review record (JSON) against the named protocol's
// SHACL shapes and returns a Report. An unknown protocol is an error.
//
// The shapes are pulled from the latest version RevAIse publishes on GitHub
// Pages, so the check requires network access. The verdict and messages come
// entirely from the protocol's shapes.
func Check(recordJSON, protocol string) (*Report, error) {
	_, shapesGraph, err := resolveShapes(protocol)
	if err != nil {
		return nil, err
	}

	framed, err := frameRecord(recordJSON)
	if err != nil {
		return nil, fmt.Errorf("preparing record: %w", err)
	}

	dataGraph, err := shacl.LoadJsonLDString(framed, "")
	if err != nil {
		return nil, fmt.Errorf("parsing record as RDF: %w", err)
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
	addProgress(report, shapesGraph, recordJSON)
	return report, nil
}

// addProgress computes the progress view: every requirement in the shapes is
// classified as passed, failed (its message appears in a violation), or pending
// (its record class is not present yet), and summarized overall and per class.
func addProgress(report *Report, shapesGraph *shacl.Graph, recordJSON string) {
	requirements := extractRequirements(shapesGraph)
	violated := make(map[string]bool, len(report.Violations))
	for _, v := range report.Violations {
		violated[v.Message] = true
	}
	present := presentClasses(recordJSON)

	byClass := make(map[string]ClassProgress)
	for _, req := range requirements {
		classPresent := req.TargetClass == "" || present[req.TargetClass]
		cp := byClass[req.TargetClass]
		cp.Total++
		cp.Present = classPresent
		switch {
		case !classPresent:
			cp.Pending++
			report.Pending = append(report.Pending, req)
		case violated[req.Message]:
			cp.Failed++
		default:
			cp.Passed++
			report.Passed = append(report.Passed, req)
		}
		byClass[req.TargetClass] = cp
	}

	summary := Summary{ByClass: byClass}
	for _, cp := range byClass {
		summary.Total += cp.Total
		summary.Passed += cp.Passed
		summary.Failed += cp.Failed
		summary.Pending += cp.Pending
	}
	report.Summary = summary
}

// presentClasses returns the record classes that have at least one instance in
// the record, so requirements for absent classes count as pending rather than
// failed. It mirrors the typing done by frameRecord.
func presentClasses(recordJSON string) map[string]bool {
	present := map[string]bool{"Review": true}
	var record map[string]any
	if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
		return present
	}
	if stages, ok := record["stages"].([]any); ok {
		for _, s := range stages {
			stage, ok := s.(map[string]any)
			if !ok {
				continue
			}
			if class := stageClass(fmt.Sprint(stage["stage_type"])); class != "" {
				present[class] = true
			}
		}
	}
	if records, ok := record["literature_records"].([]any); ok && len(records) > 0 {
		present["LiteratureRecord"] = true
	}
	return present
}

// ProtocolGuidance returns the full requirement checklist a protocol imposes,
// extracted from the SHACL shapes RevAIse publishes, together with the protocol's
// metadata. It is advisory — it helps plan a conforming review before any record
// exists — and requires network access.
func ProtocolGuidance(protocol string) (*Guidance, error) {
	entry, shapesGraph, err := resolveShapes(protocol)
	if err != nil {
		return nil, err
	}
	return &Guidance{
		Protocol:     protocol,
		Name:         entry.Name,
		Version:      entry.Version,
		Status:       entry.Status,
		Requirements: extractRequirements(shapesGraph),
	}, nil
}

// resolveShapes looks a protocol up in the published catalogue and returns its
// metadata together with its shapes loaded as an RDF graph.
func resolveShapes(protocol string) (protocolMeta, *shacl.Graph, error) {
	cat, err := fetchCatalog()
	if err != nil {
		return protocolMeta{}, nil, err
	}
	entry, ok := cat.Protocols[protocol]
	if !ok {
		return protocolMeta{}, nil, fmt.Errorf("unknown protocol %q: available protocols are %s", protocol, strings.Join(protocolNames(cat), ", "))
	}
	if entry.SHACL == "" {
		return entry, nil, fmt.Errorf("protocol %q publishes no SHACL shapes", protocol)
	}
	shapesTTL, err := fetch(revaiseBaseURL + entry.SHACL)
	if err != nil {
		return entry, nil, fmt.Errorf("fetching %s shapes: %w", protocol, err)
	}
	g, err := shacl.LoadTurtleString(string(shapesTTL), "")
	if err != nil {
		return entry, nil, fmt.Errorf("parsing %s shapes: %w", protocol, err)
	}
	return entry, g, nil
}

// extractRequirements collects every constraint message in the shapes graph,
// grouped by the target class of the shape that carries it, deduplicated and
// sorted. The messages are the protocol's own checklist wording.
func extractRequirements(g *shacl.Graph) []Requirement {
	msgPred := shacl.IRI(shacl.SH + "message")
	propPred := shacl.IRI(shacl.SH + "property")
	tcPred := shacl.IRI(shacl.SH + "targetClass")

	var reqs []Requirement
	seen := make(map[string]bool)
	for _, tr := range g.All(nil, &msgPred, nil) {
		message := tr.Object.Value()
		if message == "" {
			continue
		}
		class := owningClass(g, tr.Subject, propPred, tcPred)
		key := class + "\x00" + message
		if seen[key] {
			continue
		}
		seen[key] = true
		reqs = append(reqs, Requirement{TargetClass: class, Message: message})
	}
	sort.Slice(reqs, func(i, j int) bool {
		if reqs[i].TargetClass != reqs[j].TargetClass {
			return reqs[i].TargetClass < reqs[j].TargetClass
		}
		return reqs[i].Message < reqs[j].Message
	})
	return reqs
}

// owningClass finds the target class a constraint message applies to: the node
// shape that references the constraint via sh:property, or the node itself when
// the message sits directly on a targeted shape.
func owningClass(g *shacl.Graph, node, propPred, tcPred shacl.Term) string {
	for _, owner := range g.Subjects(propPred, node) {
		for _, tc := range g.Objects(owner, tcPred) {
			return localName(tc.Value())
		}
	}
	for _, tc := range g.Objects(node, tcPred) {
		return localName(tc.Value())
	}
	return ""
}

// localName returns the final segment of an IRI (after the last '#' or '/').
func localName(iri string) string {
	if i := strings.LastIndexAny(iri, "#/"); i >= 0 {
		return iri[i+1:]
	}
	return iri
}

// fetchCatalog retrieves and parses the published protocol catalogue.
func fetchCatalog() (catalog, error) {
	var cat catalog
	body, err := fetch(revaiseBaseURL + "index.json")
	if err != nil {
		return cat, fmt.Errorf("fetching protocol catalogue: %w", err)
	}
	if err := json.Unmarshal(body, &cat); err != nil {
		return cat, fmt.Errorf("parsing protocol catalogue: %w", err)
	}
	return cat, nil
}

// fetch performs a GET and returns the response body, or an error for any
// non-200 status.
func fetch(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// protocolNames returns the catalogue's protocol identifiers, sorted.
func protocolNames(cat catalog) []string {
	names := make([]string, 0, len(cat.Protocols))
	for name := range cat.Protocols {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
