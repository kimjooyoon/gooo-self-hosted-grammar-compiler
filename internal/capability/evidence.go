package capability

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

type Evidence struct {
	Status           Status `json:"status"`
	FirstMismatch    string `json:"first_mismatch"`
	MissingStage     string `json:"missing_stage"`
	QueryDigest      string `json:"query_digest"`
	SourceDigest     string `json:"source_digest"`
	GrammarDigest    string `json:"grammar_digest"`
	CapabilityDigest string `json:"capability_digest"`
	ReportDigest     string `json:"report_digest"`
	EvidenceDigest   string `json:"evidence_digest"`
}

func (report Report) Evidence() Evidence {
	evidence := Evidence{
		Status:           report.Status,
		FirstMismatch:    report.FirstMismatch,
		MissingStage:     report.MissingStage,
		QueryDigest:      model.DigestBytes([]byte(report.Query)),
		SourceDigest:     report.SourceDigest,
		GrammarDigest:    report.GrammarDigest,
		CapabilityDigest: model.DigestJSON(report.Capabilities),
		ReportDigest:     report.ReportDigest,
	}
	evidence.EvidenceDigest = evidence.digest()
	return evidence
}

func (evidence Evidence) Validate() error {
	for name, value := range map[string]string{
		"query":    evidence.QueryDigest,
		"source":   evidence.SourceDigest,
		"report":   evidence.ReportDigest,
		"evidence": evidence.EvidenceDigest,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("grammar capability evidence %s digest is missing", name)
		}
	}
	switch evidence.Status {
	case StatusUnknown, StatusDeferred:
		if strings.TrimSpace(evidence.FirstMismatch) == "" || strings.TrimSpace(evidence.MissingStage) == "" {
			return fmt.Errorf("unresolved grammar capability evidence lost its first boundary")
		}
	case StatusBound:
		if evidence.FirstMismatch != "" || evidence.MissingStage != "" {
			return fmt.Errorf("bound grammar capability evidence retained an unresolved boundary")
		}
		for name, value := range map[string]string{
			"grammar":    evidence.GrammarDigest,
			"capability": evidence.CapabilityDigest,
		} {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("bound grammar capability evidence %s digest is missing", name)
			}
		}
	default:
		return fmt.Errorf("grammar capability evidence status %q is invalid", evidence.Status)
	}
	if evidence.EvidenceDigest != evidence.digest() {
		return fmt.Errorf("grammar capability evidence digest does not match")
	}
	return nil
}

func (evidence Evidence) digest() string {
	return model.DigestJSON(struct {
		Status           Status
		FirstMismatch    string
		MissingStage     string
		QueryDigest      string
		SourceDigest     string
		GrammarDigest    string
		CapabilityDigest string
		ReportDigest     string
	}{
		Status:           evidence.Status,
		FirstMismatch:    evidence.FirstMismatch,
		MissingStage:     evidence.MissingStage,
		QueryDigest:      evidence.QueryDigest,
		SourceDigest:     evidence.SourceDigest,
		GrammarDigest:    evidence.GrammarDigest,
		CapabilityDigest: evidence.CapabilityDigest,
		ReportDigest:     evidence.ReportDigest,
	})
}
