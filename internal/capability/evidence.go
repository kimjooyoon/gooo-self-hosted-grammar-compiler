package capability

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

type Evidence struct {
	QueryDigest      string `json:"query_digest"`
	SourceDigest     string `json:"source_digest"`
	GrammarDigest    string `json:"grammar_digest"`
	CapabilityDigest string `json:"capability_digest"`
	ReportDigest     string `json:"report_digest"`
	EvidenceDigest   string `json:"evidence_digest"`
}

func (report Report) Evidence() Evidence {
	evidence := Evidence{
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
		"query":      evidence.QueryDigest,
		"source":     evidence.SourceDigest,
		"grammar":    evidence.GrammarDigest,
		"capability": evidence.CapabilityDigest,
		"report":     evidence.ReportDigest,
		"evidence":   evidence.EvidenceDigest,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("grammar capability evidence %s digest is missing", name)
		}
	}
	if evidence.EvidenceDigest != evidence.digest() {
		return fmt.Errorf("grammar capability evidence digest does not match")
	}
	return nil
}

func (evidence Evidence) digest() string {
	return model.DigestJSON(struct {
		QueryDigest      string
		SourceDigest     string
		GrammarDigest    string
		CapabilityDigest string
		ReportDigest     string
	}{
		QueryDigest:      evidence.QueryDigest,
		SourceDigest:     evidence.SourceDigest,
		GrammarDigest:    evidence.GrammarDigest,
		CapabilityDigest: evidence.CapabilityDigest,
		ReportDigest:     evidence.ReportDigest,
	})
}
