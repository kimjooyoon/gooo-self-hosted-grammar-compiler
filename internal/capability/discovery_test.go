package capability

import "testing"

const validGrammar = `grammar gooo v1
stage 1 origin=generated-from-stage0
ambiguity policy=reject
type GrammarSource
effect parse lower generate execute verify
fixed_denominator cases=7
token IDENT /[A-Za-z_][A-Za-z0-9_]*/ role=identifier
production file -> IDENT
precedence IDENT level=10
associativity IDENT left
`

func TestDiscoverReturnsBoundGrammarCapabilities(t *testing.T) {
	report := Discover([]byte(validGrammar), "what can gooo do with precedence")
	if report.Status != StatusBound || len(report.MatchedCapabilities) == 0 || report.GrammarDigest == "" {
		t.Fatalf("unexpected bound report: %+v", report)
	}
	if err := report.Validate(); err != nil {
		t.Fatalf("bound report should validate: %v", err)
	}
}

func TestDiscoverPreservesUnresolvedQueryTerms(t *testing.T) {
	report := Discover([]byte(validGrammar), "precedence quantum")
	if report.Status != StatusDeferred || report.MissingStage != "query_terms" || len(report.UnresolvedTerms) != 1 {
		t.Fatalf("unexpected deferred report: %+v", report)
	}
	if err := report.Validate(); err != nil {
		t.Fatalf("deferred report should validate: %v", err)
	}
}

func TestDiscoverPreservesInvalidGrammarBoundary(t *testing.T) {
	report := Discover([]byte("grammar incomplete"), "what can gooo do?")
	if report.Status != StatusUnknown || report.MissingStage != "stage0_parse" {
		t.Fatalf("unexpected unknown report: %+v", report)
	}
	if err := report.Validate(); err != nil {
		t.Fatalf("unknown report should validate: %v", err)
	}
}
