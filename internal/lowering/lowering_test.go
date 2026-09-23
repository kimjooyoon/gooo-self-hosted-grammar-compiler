package lowering

import (
	"testing"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

func TestValidateCollectionsRejectsDuplicateTokenAttribute(t *testing.T) {
	diagnostics := validateCollections(model.GrammarIR{
		Tokens: []model.TokenDecl{{
			Name: "IDENT",
			Attributes: []model.Attribute{
				{Key: "skip", Value: "true"},
				{Key: "skip", Value: "false"},
			},
		}},
	})
	for _, diagnostic := range diagnostics {
		if diagnostic.Class == IssueInvalidGrammar && diagnostic.Step == "token-attribute-uniqueness" {
			return
		}
	}
	t.Fatal("duplicate token attribute key was not refuted")
}
