package capability

import "testing"

func TestReportEvidenceBindsQueryAndCapabilitySurface(t *testing.T) {
	report := Discover([]byte(validGrammar), "what can gooo do with precedence")
	evidence := report.Evidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("report evidence should validate: %v", err)
	}
	if evidence.QueryDigest == "" || evidence.CapabilityDigest == "" || evidence.SuggestedActionDigest == "" || evidence.ReportDigest == "" {
		t.Fatalf("report evidence is incomplete: %+v", evidence)
	}
	other := Discover([]byte(validGrammar), "what can gooo do with token")
	if evidence.EvidenceDigest == other.Evidence().EvidenceDigest {
		t.Fatal("different queries unexpectedly share the same evidence digest")
	}
}

func TestReportEvidencePreservesUnknownReportIdentity(t *testing.T) {
	report := Discover([]byte("grammar incomplete"), "what can gooo do?")
	evidence := report.Evidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("unknown report evidence should validate: %v", err)
	}
	if evidence.ReportDigest == "" || evidence.SuggestedActionDigest == "" || evidence.EvidenceDigest == "" {
		t.Fatalf("unknown report evidence is incomplete: %+v", evidence)
	}
}
