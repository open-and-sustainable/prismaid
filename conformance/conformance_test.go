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

// skipIfOffline skips network-dependent tests in short mode and when the
// published RevAIse catalogue cannot be reached, so they never fail offline.
func skipIfOffline(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping network-dependent conformance test in short mode")
	}
	if _, err := fetchCatalog(); err != nil {
		t.Skipf("skipping: RevAIse catalogue unreachable: %v", err)
	}
}

func TestCheckReportsPrismaGaps(t *testing.T) {
	skipIfOffline(t)

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
	skipIfOffline(t)

	if _, err := Check(sampleRecord, "made-up"); err == nil {
		t.Fatal("expected an error for an unknown protocol")
	}
}

func TestAvailableProtocols(t *testing.T) {
	skipIfOffline(t)

	got, err := AvailableProtocols()
	if err != nil {
		t.Fatalf("AvailableProtocols failed: %v", err)
	}
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

func TestProtocolGuidance(t *testing.T) {
	skipIfOffline(t)

	guidance, err := ProtocolGuidance("prisma-2020")
	if err != nil {
		t.Fatalf("ProtocolGuidance failed: %v", err)
	}
	if guidance.Name == "" {
		t.Error("expected protocol metadata (name) to be populated")
	}
	if len(guidance.Requirements) == 0 {
		t.Fatal("expected at least one requirement")
	}

	// Requirements should carry the protocol's own messages and be grouped by
	// the record class they apply to.
	foundReview := false
	for _, r := range guidance.Requirements {
		if r.Message == "" {
			t.Error("expected every requirement to carry a message")
		}
		if r.TargetClass == "Review" {
			foundReview = true
		}
	}
	if !foundReview {
		t.Error("expected at least one Review-class requirement")
	}
}

func TestProtocolGuidanceUnknown(t *testing.T) {
	skipIfOffline(t)

	if _, err := ProtocolGuidance("made-up"); err == nil {
		t.Fatal("expected an error for an unknown protocol")
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
