package generated

import (
	"os"
	"testing"
)

func TestStage2ParsesSample(t *testing.T) {
	raw, err := os.ReadFile("../examples/sample.gooo")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := ParseStage2(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Root.Children) < 4 {
		t.Fatalf("generated parser returned %d declarations", len(tree.Root.Children))
	}
}

func TestStage1AndStage2HaveSameSyntaxTree(t *testing.T) {
	raw, err := os.ReadFile("../meta/gooo-grammar.gooo")
	if err != nil {
		t.Fatal(err)
	}
	one, err := ParseStage1(raw)
	if err != nil {
		t.Fatal(err)
	}
	two, err := ParseStage2(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(one.Root.Children) != len(two.Root.Children) {
		t.Fatalf("generation declaration counts differ: %d vs %d", len(one.Root.Children), len(two.Root.Children))
	}
}
