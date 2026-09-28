package capability

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func TestServeCapabilityJSONLReturnsReportAndEvidence(t *testing.T) {
	input := `{"id":"one","grammar":` + strconv.Quote(validGrammar) + `,"query":"what can gooo do with precedence"}
`
	var output bytes.Buffer
	if err := ServeCapabilityJSONL(strings.NewReader(input), &output); err != nil {
		t.Fatalf("capability JSONL service failed: %v", err)
	}
	var response CapabilityJSONLResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("capability JSONL response is invalid: %v", err)
	}
	if string(response.ID) != `"one"` || response.Report.Status != StatusBound || !response.ReadOnly {
		t.Fatalf("unexpected capability response: %+v", response)
	}
	if err := response.Evidence.Validate(); err != nil {
		t.Fatalf("capability response evidence should validate: %v", err)
	}
}

func TestServeCapabilityJSONLPreservesUnknownBoundary(t *testing.T) {
	input := `{"id":2,"grammar":"grammar incomplete","query":"what can gooo do?"}
`
	var output bytes.Buffer
	if err := ServeCapabilityJSONL(strings.NewReader(input), &output); err != nil {
		t.Fatalf("capability JSONL service failed: %v", err)
	}
	var response CapabilityJSONLResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("capability JSONL response is invalid: %v", err)
	}
	if response.Report.Status != StatusUnknown || response.Evidence.MissingStage != "stage0_parse" {
		t.Fatalf("unknown boundary was not preserved: %+v", response)
	}
}
