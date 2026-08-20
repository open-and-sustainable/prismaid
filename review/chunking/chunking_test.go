package chunking

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/open-and-sustainable/alembica/definitions"
	"github.com/open-and-sustainable/prismaid/review/config"
)

func TestSplitPreservesTextAndPrefersParagraphBoundaries(t *testing.T) {
	prefix := "instructions\n\n"
	text := "First paragraph has enough words to require a boundary.\n\nSecond paragraph also has enough words.\n\nThird paragraph completes the document."
	chunks, err := Split(text, prefix, 80, 0, byteCounter{})
	if err != nil {
		t.Fatalf("Split returned an error: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	var rebuilt strings.Builder
	for _, chunk := range chunks {
		if got := len(prefix + chunk.Text); got > 80 {
			t.Fatalf("chunk exceeds input limit: got %d, limit 80", got)
		}
		rebuilt.WriteString(chunk.CoreText)
	}
	if got := rebuilt.String(); got != text {
		t.Fatalf("chunk cores did not preserve the source text\nwant: %q\n got: %q", text, got)
	}
	if !strings.HasSuffix(chunks[0].CoreText, "\n\n") {
		t.Fatalf("expected the first chunk to end at a paragraph boundary, got %q", chunks[0].CoreText)
	}
}

func TestSplitUsesSentenceBoundaryBeforeHardSplit(t *testing.T) {
	prefix := "instructions\n\n"
	text := "The first sentence is deliberately long. The second sentence is also deliberately long. The third sentence completes the paragraph.\n\nA final paragraph follows."
	chunks, err := Split(text, prefix, 67, 0, byteCounter{})
	if err != nil {
		t.Fatalf("Split returned an error: %v", err)
	}
	if len(chunks) < 3 {
		t.Fatalf("expected several chunks, got %d", len(chunks))
	}
	if !strings.HasSuffix(strings.TrimSpace(chunks[0].CoreText), ".") {
		t.Fatalf("expected oversized paragraph to split at a sentence boundary, got %q", chunks[0].CoreText)
	}
}

func TestSplitAddsConfiguredOverlap(t *testing.T) {
	prefix := "instructions\n\n"
	text := "A first paragraph with several words.\n\nA second paragraph with several words.\n\nA third paragraph with several words."
	chunks, err := Split(text, prefix, 58, 10, byteCounter{})
	if err != nil {
		t.Fatalf("Split returned an error: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	overlap := strings.TrimSuffix(chunks[1].Text, chunks[1].CoreText)
	if overlap == "" || !strings.HasSuffix(chunks[0].CoreText, overlap) {
		t.Fatalf("expected second chunk to begin with text from the previous core, got overlap %q", overlap)
	}
}

func TestMergeAppliesExplicitRules(t *testing.T) {
	config := config.ChunkingConfig{
		Enabled: true,
		Merge: map[string]config.MergeRule{
			"endpoints": {Rule: "union", Sentinels: []string{"none", "not_specified"}},
			"status":    {Rule: "ordinal", Order: []string{"none", "implicit", "explicit"}},
			"scope":     {Rule: "categorical", Defaults: []string{"not_specified", "other"}, TieBreak: "first"},
			"detail":    {Rule: "unique_text", Separator: " | ", MaxLength: 100},
			"score":     {Rule: "numeric", Operation: "max"},
			"doi":       {Rule: "metadata", OnMismatch: "warn"},
		},
	}
	output := definitions.Output{
		Metadata: definitions.OutputMetadata{SchemaVersion: "1.0"},
		Responses: []definitions.Response{
			{
				Provider: "SelfHosted", Model: "model", SequenceID: "chunk-1", SequenceNumber: 1,
				ModelResponses: []string{`{"endpoints":["none","method_a"],"status":"implicit","scope":"not_specified","detail":"First detail.","score":0.4,"doi":"10.1/example"}`},
			},
			{
				Provider: "SelfHosted", Model: "model", SequenceID: "chunk-2", SequenceNumber: 1,
				ModelResponses: []string{`{"endpoints":["method_b","method_a"],"status":"explicit","scope":"regional","detail":"Second detail.","score":0.9,"doi":"10.1/example"}`},
			},
		},
	}
	raw, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	bindings := map[string]Binding{
		"chunk-1": {DocumentIndex: 0, Filename: "paper", ChunkIndex: 0, ChunkCount: 2},
		"chunk-2": {DocumentIndex: 0, Filename: "paper", ChunkIndex: 1, ChunkCount: 2},
	}

	merged, report, err := Merge(string(raw), bindings, config)
	if err != nil {
		t.Fatalf("Merge returned an error: %v", err)
	}
	var parsed definitions.Output
	if err := json.Unmarshal([]byte(merged), &parsed); err != nil {
		t.Fatalf("unmarshal merged output: %v", err)
	}
	if len(parsed.Responses) != 1 {
		t.Fatalf("expected one merged response, got %d", len(parsed.Responses))
	}
	var object map[string]interface{}
	if err := json.Unmarshal([]byte(parsed.Responses[0].ModelResponses[0]), &object); err != nil {
		t.Fatalf("unmarshal merged object: %v", err)
	}
	if got := object["status"]; got != "explicit" {
		t.Fatalf("expected strongest ordinal status, got %v", got)
	}
	if got := object["scope"]; got != "regional" {
		t.Fatalf("expected non-default categorical scope, got %v", got)
	}
	if got := object["score"]; got != 0.9 {
		t.Fatalf("expected max score, got %v", got)
	}
	if got := object["detail"]; got != "First detail. | Second detail." {
		t.Fatalf("expected concatenated detail, got %v", got)
	}
	if got := object["doi"]; got != "10.1/example" {
		t.Fatalf("expected preserved metadata, got %v", got)
	}
	if got := object["endpoints"]; !equalStrings(got, []string{"method_a", "method_b"}) {
		t.Fatalf("expected sentinel-free union, got %v", got)
	}
	if len(report.Conflicts) == 0 {
		t.Fatal("expected conflicts to be reported")
	}
}

func TestMergeNumericOperations(t *testing.T) {
	values := []interface{}{0.2, 0.6, 0.4}
	for _, test := range []struct {
		operation string
		expected  float64
	}{
		{"max", 0.6},
		{"mean", 0.4},
		{"min", 0.2},
	} {
		t.Run(test.operation, func(t *testing.T) {
			value, _, err := mergeField("score", values, config.MergeRule{Rule: "numeric", Operation: test.operation})
			got, ok := value.(float64)
			if err != nil || !ok || math.Abs(got-test.expected) > 1e-9 {
				t.Fatalf("expected %v for %s, got value=%v err=%v", test.expected, test.operation, value, err)
			}
		})
	}
}

func TestMergeRecordsConfiguredMetadataMismatch(t *testing.T) {
	raw, err := json.Marshal(definitions.Output{
		Metadata: definitions.OutputMetadata{SchemaVersion: "1.0"},
		Responses: []definitions.Response{
			{Provider: "x", Model: "y", SequenceID: "one", SequenceNumber: 1, ModelResponses: []string{`{"doi":"10.1/first"}`}},
			{Provider: "x", Model: "y", SequenceID: "two", SequenceNumber: 1, ModelResponses: []string{`{"doi":"10.1/second"}`}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	merged, report, err := Merge(string(raw), map[string]Binding{
		"one": {DocumentIndex: 0, Filename: "paper", ChunkIndex: 0, ChunkCount: 2},
		"two": {DocumentIndex: 0, Filename: "paper", ChunkIndex: 1, ChunkCount: 2},
	}, config.ChunkingConfig{Enabled: true, Merge: map[string]config.MergeRule{
		"doi": {Rule: "metadata", OnMismatch: "error"},
	}})
	if err != nil {
		t.Fatalf("Merge returned an unexpected global error: %v", err)
	}
	if len(report.Failures) != 1 || !strings.Contains(report.Failures[0].Message, "metadata values differ") {
		t.Fatalf("expected one metadata failure, got %#v", report.Failures)
	}
	var parsed definitions.Output
	if err := json.Unmarshal([]byte(merged), &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Responses) != 0 {
		t.Fatalf("expected failed document to be omitted, got %#v", parsed.Responses)
	}
}

func TestMergeNormalizesHeterogeneousModelValues(t *testing.T) {
	cfg := config.ChunkingConfig{Enabled: true, Merge: map[string]config.MergeRule{
		"union":       {Rule: "union", Sentinels: []string{"none", "not_specified"}},
		"ordinal":     {Rule: "ordinal", Order: []string{"none", "implicit", "explicit"}},
		"categorical": {Rule: "categorical", Defaults: []string{"not_specified", "other"}, TieBreak: "first"},
		"text":        {Rule: "unique_text", Separator: " | ", MaxLength: 100},
		"numeric":     {Rule: "numeric", Operation: "max"},
		"metadata":    {Rule: "metadata", OnMismatch: "warn"},
	}}
	output := `{"metadata":{"schemaVersion":"v1"},"responses":[
		{"provider":"x","model":"y","sequenceId":"one","sequenceNumber":1,"modelResponses":["{\"union\":\"method_a\",\"ordinal\":[\"unknown\",\"implicit\"],\"categorical\":[\"not_specified\",\"regional\"],\"text\":[\"First detail\",4],\"numeric\":\"0.4\",\"metadata\":null}" ]},
		{"provider":"x","model":"y","sequenceId":"two","sequenceNumber":1,"modelResponses":["{\"union\":[true,\"none\"],\"ordinal\":null,\"text\":\"Second detail\",\"numeric\":[\"bad\",0.9],\"metadata\":\"10.1/example\"}" ]}
	]}`
	merged, report, err := Merge(output, map[string]Binding{
		"one": {DocumentIndex: 0, Filename: "paper", ChunkIndex: 0, ChunkCount: 2},
		"two": {DocumentIndex: 0, Filename: "paper", ChunkIndex: 1, ChunkCount: 2},
	}, cfg)
	if err != nil {
		t.Fatalf("Merge returned an error for heterogeneous values: %v", err)
	}
	if len(report.Failures) != 0 || len(report.Coercions) == 0 {
		t.Fatalf("expected coercions and no failures, got report %#v", report)
	}
	var parsed definitions.Output
	if err := json.Unmarshal([]byte(merged), &parsed); err != nil {
		t.Fatal(err)
	}
	var object map[string]interface{}
	if err := json.Unmarshal([]byte(parsed.Responses[0].ModelResponses[0]), &object); err != nil {
		t.Fatal(err)
	}
	if !equalStrings(object["union"], []string{"method_a", "true"}) || object["ordinal"] != "implicit" || object["categorical"] != "regional" || object["text"] != "First detail | Second detail" || object["numeric"] != 0.9 || object["metadata"] != "10.1/example" {
		t.Fatalf("unexpected normalized result: %#v", object)
	}
}

func TestMergeRetainsOtherDocumentsWhenOneFails(t *testing.T) {
	output := `{"metadata":{"schemaVersion":"v1"},"responses":[
		{"provider":"x","model":"y","sequenceId":"good","sequenceNumber":1,"modelResponses":["{\"status\":\"yes\"}"]},
		{"provider":"x","model":"y","sequenceId":"bad","sequenceNumber":1,"modelResponses":["not json"]}
	]}`
	bindings := map[string]Binding{
		"good": {DocumentIndex: 0, Filename: "good", ChunkIndex: 0, ChunkCount: 1},
		"bad":  {DocumentIndex: 1, Filename: "bad", ChunkIndex: 0, ChunkCount: 1},
	}
	expected := []ExpectedGroup{
		{DocumentIndex: 0, Filename: "good", Provider: "x", Model: "y", SequenceNumber: 1, ChunkCount: 1},
		{DocumentIndex: 1, Filename: "bad", Provider: "x", Model: "y", SequenceNumber: 1, ChunkCount: 1},
	}
	merged, report, err := MergeWithExpected(output, bindings, expected, config.ChunkingConfig{Enabled: true, Merge: map[string]config.MergeRule{"status": {Rule: "ordinal", Order: []string{"no", "yes"}}}})
	if err != nil || len(report.Failures) != 1 || report.Failures[0].Filename != "bad" {
		t.Fatalf("expected one isolated bad-document failure, got merged=%s report=%#v err=%v", merged, report, err)
	}
	var parsed definitions.Output
	if err := json.Unmarshal([]byte(merged), &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Responses) != 1 || parsed.Responses[0].SequenceID != "1" {
		t.Fatalf("expected good document only, got %#v", parsed.Responses)
	}
}

func TestMergeOrdinalPreservesBooleanType(t *testing.T) {
	value, conflict, err := mergeField("present", []interface{}{false, true}, config.MergeRule{
		Rule:  "ordinal",
		Order: []string{"false", "true"},
	})
	if err != nil {
		t.Fatalf("mergeField returned an error: %v", err)
	}
	if !conflict || value != true {
		t.Fatalf("expected strongest boolean value and conflict, got value=%v conflict=%t", value, conflict)
	}
}

func equalStrings(value interface{}, expected []string) bool {
	items, ok := value.([]interface{})
	if !ok || len(items) != len(expected) {
		return false
	}
	for index, expectedItem := range expected {
		if items[index] != expectedItem {
			return false
		}
	}
	return true
}
