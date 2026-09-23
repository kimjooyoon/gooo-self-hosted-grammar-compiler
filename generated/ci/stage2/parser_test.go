//go:build generated_ci

package stage2generated

import (
	"os"
	"testing"
)

func TestGeneratedArtifactParsesSample(t *testing.T) {
	raw, err := os.ReadFile("../../../examples/sample.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseStage2(raw); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratedArtifactRejectsProgramWithoutName(t *testing.T) {
	if _, err := ParseStage2([]byte("program\n")); err == nil {
		t.Fatal("generated parser accepted a program declaration without a name")
	}
}
