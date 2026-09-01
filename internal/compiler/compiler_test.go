package compiler

import (
	"os"
	"testing"
)

func TestBootstrapAuthorityIsClosed(t *testing.T) {
	raw, err := os.ReadFile("../../meta/gooo-grammar.gooo")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Bootstrap(raw)
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != DecisionClosed {
		t.Fatalf("expected CLOSED, got %s", result.Decision)
	}
	if len(result.Generations) != 3 {
		t.Fatalf("expected three generation records, got %d", len(result.Generations))
	}
}
