package compiler

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
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

func TestRenderParserKeepsGeneratedDeclarationBoundaryChecks(t *testing.T) {
	source := string(RenderParser(model.GrammarIR{}, 2, "generated", "Parse"))
	for _, want := range []string{
		`else if kind == "program" {`,
		`incomplete program declaration`,
		`else if kind == "package" || kind == "namespace" {`,
		`else if kind == "entity" || kind == "activity" {`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated parser template lost declaration boundary check %q", want)
		}
	}
}
