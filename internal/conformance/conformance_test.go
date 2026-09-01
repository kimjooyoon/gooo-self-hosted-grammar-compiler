package conformance

import "testing"

func TestFixedSevenCaseDenominator(t *testing.T) {
	report, err := RunDirectory("../../fixtures/cases", "../..")
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 7 || report.Selected != 7 || report.Executed != 7 || report.Failed != 0 || !report.AllPass {
		for _, result := range report.Cases {
			if !result.Pass {
				t.Logf("case %s expected=%s actual=%s failure=%s diagnostics=%v", result.Definition.ID, result.Definition.ExpectedDecision, result.Actual.Decision, result.Failure, result.Actual.Diagnostics)
			}
		}
		t.Fatalf("fixed denominator failed: total=%d selected=%d executed=%d failed=%d all_pass=%v", report.Total, report.Selected, report.Executed, report.Failed, report.AllPass)
	}
}
