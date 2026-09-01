package stage0

import (
	"os"
	"testing"
)

func TestSeedParsesAuthority(t *testing.T) {
	raw, err := os.ReadFile("../../meta/gooo-grammar.gooo")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Root.Kind != "grammar-file" || len(tree.Root.Children) == 0 {
		t.Fatal("seed parser returned an empty syntax tree")
	}
}
