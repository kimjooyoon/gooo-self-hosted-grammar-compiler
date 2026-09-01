package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInventoryUsesSourceRepositoryScope(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{".git/objects", ".ci", "generated/ci", "vendor", "cache", ".cache", "toolchain", ".toolchain", "node_modules"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, directory, "ignored.go"), []byte("package ignored\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package source\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.gooo"), []byte("grammar source v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	values := inventory(root)
	if values[0].RegularFiles != 1 || values[0].PhysicalLines != 1 || values[1].RegularFiles != 1 || values[1].PhysicalLines != 1 {
		t.Fatalf("excluded files entered source inventory: %#v", values)
	}
	if values[0].DescendantDirs != 1 || values[1].DescendantDirs != 1 {
		t.Fatalf("excluded directories entered source inventory: %#v", values)
	}
}
