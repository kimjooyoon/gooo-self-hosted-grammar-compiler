package lowering

import "testing"

func TestKeyValueRejectsDuplicateKey(t *testing.T) {
	if _, ok := keyValue([]string{"level=5", "level=10"}, "level"); ok {
		t.Fatal("keyValue accepted duplicate precedence levels")
	}
}
