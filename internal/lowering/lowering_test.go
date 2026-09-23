package lowering

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/stage0"
)

func TestLowerRejectsDuplicateFixedDenominator(t *testing.T) {
	raw, err := os.ReadFile("../../meta/gooo-grammar.gooo")
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte("fixed_denominator cases=7\n"), []byte("fixed_denominator cases=8\nfixed_denominator cases=7\n"), 1)
	tree, err := stage0.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := Lower(tree); err == nil || !strings.Contains(err.Error(), "fixed_denominator declaration") {
		t.Fatalf("expected duplicate fixed_denominator error, got %v", err)
	}
}
