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

func TestRenderedParserValidatesProgramDeclarations(t *testing.T) {
	rendered := string(RenderParser(model.GrammarIR{}, 1, "generated", "ParseStage1"))
	if !strings.Contains(rendered, "validateGeneratedDeclaration(kind, fields, line)") {
		t.Fatal("generated parser does not validate declarations before accepting them")
	}
	if !strings.Contains(rendered, "case \"package\", \"namespace\":") {
		t.Fatal("generated parser has no strict package/namespace declaration contract")
	}
}
