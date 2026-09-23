package compiler

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

func TestGeneratedParserTemplateValidatesProgramDeclarations(t *testing.T) {
	rendered := string(RenderParser(model.GrammarIR{}, 1, "generated", "ParseStage1"))
	for _, expected := range []string{"package expects a name", "program expects a name"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("generated parser template is missing %q", expected)
		}
	}
}

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
