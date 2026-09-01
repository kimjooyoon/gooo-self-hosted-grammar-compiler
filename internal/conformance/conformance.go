package conformance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/compiler"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

var FixedCaseIDs = []string{
	"base-self-parse",
	"additive-syntax-extension",
	"precedence-preserving-change",
	"ambiguous-grammar",
	"conflicting-precedence",
	"forbidden-stage-escape",
	"byte-identical-replay",
}

type CaseDefinition struct {
	ID                         string `json:"id"`
	Source                     string `json:"source"`
	ExpectedDecision           string `json:"expected_decision"`
	Mode                       string `json:"mode"`
	Reference                  string `json:"reference,omitempty"`
	RequireSemanticChange      bool   `json:"require_semantic_change,omitempty"`
	RequireSyntaxChange        bool   `json:"require_syntax_change,omitempty"`
	RequireByteIdenticalReplay bool   `json:"require_byte_identical_replay,omitempty"`
}

type CaseReport struct {
	Definition CaseDefinition  `json:"definition"`
	Actual     compiler.Result `json:"actual"`
	Pass       bool            `json:"pass"`
	Failure    string          `json:"failure,omitempty"`
}

type Report struct {
	Schema   string       `json:"schema"`
	Total    int          `json:"total"`
	Selected int          `json:"selected"`
	Executed int          `json:"executed"`
	Reused   int          `json:"reused"`
	Failed   int          `json:"failed"`
	Unknown  int          `json:"unknown"`
	Cases    []CaseReport `json:"cases"`
	AllPass  bool         `json:"all_pass"`
}

func LoadDefinition(path string) (CaseDefinition, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return CaseDefinition{}, err
	}
	var definition CaseDefinition
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&definition); err != nil {
		return CaseDefinition{}, err
	}
	if definition.ID == "" || definition.Source == "" || definition.ExpectedDecision == "" {
		return CaseDefinition{}, fmt.Errorf("%s: incomplete case definition", path)
	}
	return definition, nil
}

func RunDirectory(root, referenceRoot string) (Report, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return Report{}, err
	}
	var definitions []CaseDefinition
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		definition, err := LoadDefinition(filepath.Join(root, entry.Name(), "case.json"))
		if err != nil {
			return Report{}, err
		}
		definitions = append(definitions, definition)
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].ID < definitions[j].ID })
	if len(definitions) != len(FixedCaseIDs) {
		return Report{}, fmt.Errorf("fixed denominator requires %d cases, found %d", len(FixedCaseIDs), len(definitions))
	}
	for index, id := range FixedCaseIDs {
		found := false
		for _, definition := range definitions {
			if definition.ID == id {
				found = true
				break
			}
		}
		if !found {
			return Report{}, fmt.Errorf("fixed case %d %q is missing", index+1, id)
		}
	}

	report := Report{Schema: "gooo/self-hosted-grammar-compiler/conformance/v1", Total: len(definitions), Selected: len(definitions), Cases: []CaseReport{}}
	for _, definition := range definitions {
		source, err := os.ReadFile(filepath.Join(root, definition.ID, definition.Source))
		if err != nil {
			// Source is normally relative to the case directory; retain a useful
			// error when a fixture points elsewhere.
			source, err = os.ReadFile(filepath.Join(root, definition.Source))
		}
		if err != nil {
			return Report{}, err
		}
		actual, runErr := compiler.Bootstrap(source)
		caseReport := CaseReport{Definition: definition, Actual: actual}
		if runErr != nil {
			caseReport.Failure = runErr.Error()
		} else if actual.Decision != definition.ExpectedDecision {
			caseReport.Failure = fmt.Sprintf("expected %s, got %s", definition.ExpectedDecision, actual.Decision)
		} else if err := validateTerminal(actual); err != nil {
			caseReport.Failure = err.Error()
		} else if definition.RequireByteIdenticalReplay {
			caseReport.Failure = validateReplay(definition, source, root, referenceRoot, actual)
		} else if definition.RequireSemanticChange || definition.RequireSyntaxChange {
			caseReport.Failure = validateChange(definition, source, root, referenceRoot, actual)
		}
		caseReport.Pass = caseReport.Failure == ""
		report.Cases = append(report.Cases, caseReport)
		report.Executed++
		if actual.Decision == compiler.DecisionUnknown {
			report.Unknown++
		}
		if !caseReport.Pass {
			report.Failed++
		}
	}
	report.AllPass = report.Failed == 0
	return report, nil
}

func validateTerminal(result compiler.Result) error {
	if result.Terminal.Decision != result.Decision || result.TerminalDigest == "" {
		return "terminal record and decision digest disagree"
	}
	if result.Decision == compiler.DecisionUnknown {
		fields := []string{result.Terminal.Stage, result.Terminal.Step, result.Terminal.Reason, result.Terminal.UnknownClass, result.Terminal.NextOperation, result.Terminal.BlockedBy}
		for _, field := range fields {
			if field == "" {
				return "UNKNOWN terminal record is missing a required field"
			}
		}
	}
	return nil
}

func validateReplay(definition CaseDefinition, source []byte, root, referenceRoot string, actual compiler.Result) string {
	referencePath := filepath.Join(referenceRoot, definition.Reference)
	reference, err := os.ReadFile(referencePath)
	if err != nil {
		return "replay reference cannot be read: " + err.Error()
	}
	if model.DigestBytes(source) != model.DigestBytes(reference) {
		return "replay source bytes differ"
	}
	other, err := compiler.Bootstrap(reference)
	if err != nil {
		return "replay reference failed: " + err.Error()
	}
	if model.DigestBytes(actual.Stage1Artifact) != model.DigestBytes(other.Stage1Artifact) || model.DigestBytes(actual.Stage2Artifact) != model.DigestBytes(other.Stage2Artifact) || model.DigestJSON(actual.Generations) != model.DigestJSON(other.Generations) || actual.TerminalDigest != other.TerminalDigest {
		return "replay artifacts or terminal digest differ"
	}
	_ = root
	return ""
}

func validateChange(definition CaseDefinition, source []byte, root, referenceRoot string, actual compiler.Result) string {
	referencePath := filepath.Join(referenceRoot, definition.Reference)
	reference, err := os.ReadFile(referencePath)
	if err != nil {
		return "change reference cannot be read: " + err.Error()
	}
	other, err := compiler.Bootstrap(reference)
	if err != nil {
		return "change reference failed: " + err.Error()
	}
	if definition.RequireSyntaxChange && model.SyntaxTreeDigest(actual.Stage0Tree) == model.SyntaxTreeDigest(other.Stage0Tree) {
		return "expected syntax-tree change was not observed"
	}
	if definition.RequireSemanticChange && model.GrammarDigest(actual.SemanticIR) == model.GrammarDigest(other.SemanticIR) {
		return "expected semantic-IR change was not observed"
	}
	_ = source
	_ = root
	return ""
}
