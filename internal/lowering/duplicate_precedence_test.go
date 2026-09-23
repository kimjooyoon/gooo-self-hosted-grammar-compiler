package lowering

import (
	"testing"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

func TestValidateCollectionsRejectsDuplicatePrecedenceAtSameLevel(t *testing.T) {
	diagnostics := validateCollections(model.GrammarIR{
		Precedences: []model.PrecedenceDecl{
			{Name: "plus", Level: 10},
			{Name: "plus", Level: 10},
		},
	})
	if !HasClass(diagnostics, IssueInvalidGrammar) {
		t.Fatalf("expected duplicate precedence to be invalid: %#v", diagnostics)
	}
}
