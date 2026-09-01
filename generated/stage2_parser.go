// Code generated for the gooo grammar bootstrap. DO NOT EDIT.
// This second artifact is independently callable by CI integration tests.
package generated

import "github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"

const Stage2ParserIdentity = "generated-stage2-from-stage1-gooo-grammar-v1"

// ParseStage2 executes the next-generation parser artifact. The parser body
// is generated in stage1_parser.go's generated runtime; this exported entry
// point is the stage2 artifact consumed by the sample-program integration.
func ParseStage2(raw []byte) (model.SyntaxTree, error) {
	return parseGeneratedGrammar(raw, "stage2")
}
