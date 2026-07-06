package conformance

import (
	"strings"
	"testing"
)

// sampleRecord is a RevAIse record in the shape prismAId produces: it has a
// title and included literature records, but omits several PRISMA-required
// reporting fields (abstract, funding, competing interests, data availability).
const sampleRecord = `{
  "review_id": "r1",
  "review_title": "Test review",
  "review_type": "SYSTEMATIC_REVIEW",
  "review_status": "IN_PROGRESS",
  "review_authors": [{"name": "Tester"}],
  "literature_records": [{"record_id": "rec1", "title": "A study"}],
  "stages": [
    {"stage_type": "data_extraction", "stage_label": "Extraction"}
  ]
}`

func TestCheckReportsPrismaGaps(t *testing.T) {
	report, err := Check(sampleRecord, "prisma-2020")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if report.Conforms {
		t.Fatalf("expected the record not to conform (it omits required reporting fields)")
	}
	if len(report.Violations) == 0 {
		t.Fatal("expected at least one violation")
	}

	// The verdict and messages must come from the protocol shapes.
	want := []string{"review abstract", "funding sources"}
	for _, w := range want {
		if !hasViolationContaining(report, w) {
			t.Errorf("expected a violation mentioning %q; got: %s", w, allMessages(report))
		}
	}
}

func TestCheckUnknownProtocol(t *testing.T) {
	if _, err := Check(sampleRecord, "made-up"); err == nil {
		t.Fatal("expected an error for an unknown protocol")
	}
}

func TestAvailableProtocols(t *testing.T) {
	got := AvailableProtocols()
	found := false
	for _, p := range got {
		if p == "prisma-2020" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected prisma-2020 among available protocols, got %v", got)
	}
}

func hasViolationContaining(report *Report, substr string) bool {
	for _, v := range report.Violations {
		if strings.Contains(strings.ToLower(v.Message), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

func allMessages(report *Report) string {
	msgs := make([]string, 0, len(report.Violations))
	for _, v := range report.Violations {
		msgs = append(msgs, v.Message)
	}
	return strings.Join(msgs, " | ")
}
