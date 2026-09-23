package compiler

import (
	"os"
	"strings"
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

func TestGeneratedParserTemplatePreservesDeclarationValidation(t *testing.T) {
	for _, expected := range []string{
		"validateGeneratedDeclaration(kind, fields, line)",
		`case "package", "namespace":`,
		`case "program":`,
		"expects a name",
	} {
		if !strings.Contains(parserTemplate, expected) {
			t.Fatalf("generated parser template is missing declaration validation %q", expected)
		}
	}
}
