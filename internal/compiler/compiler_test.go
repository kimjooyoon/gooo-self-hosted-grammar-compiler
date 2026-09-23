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

func TestRenderParserValidatesProgramDeclarationShape(t *testing.T) {
	source := string(RenderParser(model.GrammarIR{}, 1, "generated", "ParseStage1"))
	if !strings.Contains(source, `kind == "package" || kind == "namespace"`) {
		t.Fatal("generated parser must validate package and namespace declarations")
	}
	if strings.Contains(source, `kind == "package" || kind == "program"`) {
		t.Fatal("generated parser must not bypass package validation")
	}
	if !strings.Contains(source, `invalid program declaration`) {
		t.Fatal("generated parser must validate program declarations")
	}
}
