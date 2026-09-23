package compiler

import (
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/generated"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/lowering"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/stage0"
)

const (
	DecisionClosed  = "CLOSED"
	DecisionUnknown = "UNKNOWN"
	DecisionRefuted = "REFUTED"
)

type GenerationRecord struct {
	Generation           string `json:"generation"`
	SyntaxTreeDigest     string `json:"syntax_tree_digest"`
	SemanticIRDigest     string `json:"semantic_ir_digest"`
	TerminalReasonDigest string `json:"terminal_reason_digest"`
	ParserIdentity       string `json:"parser_identity"`
}

type Result struct {
	Decision       string               `json:"decision"`
	Terminal       model.TerminalRecord `json:"terminal"`
	TerminalDigest string               `json:"terminal_digest"`
	Generations    []GenerationRecord   `json:"generations"`
	Diagnostics    []model.Diagnostic   `json:"diagnostics,omitempty"`
	Stage1Artifact []byte               `json:"-"`
	Stage2Artifact []byte               `json:"-"`
	SourceDigest   string               `json:"source_digest"`
	SemanticIR     model.GrammarIR      `json:"semantic_ir"`
	Stage0Tree     model.SyntaxTree     `json:"-"`
	Stage1Tree     model.SyntaxTree     `json:"-"`
	Stage2Tree     model.SyntaxTree     `json:"-"`
}

// Bootstrap performs the three-generation comparison. Stage0 is the seed,
// stage1 is the committed generated parser, and stage2 is generated from the
// stage1-lowered IR. All parser-source differences become evidence, not hidden
// side effects.
func Bootstrap(source []byte) (Result, error) {
	result := Result{SourceDigest: model.DigestBytes(source)}
	terminal := closedTerminal()

	tree0, err := stage0.Parse(source)
	if err != nil {
		return refutedResult(result, terminalFor("stage0", "seed-parse", "SEED_SYNTAX_INVALID", ""), err), nil
	}
	result.Stage0Tree = tree0
	ir0, diagnostics0, err := lowering.Lower(tree0)
	if err != nil {
		return refutedResult(result, terminalFor("stage0", "lower", "SEMANTIC_LOWERING_INVALID", ""), err), nil
	}
	result.Diagnostics = append(result.Diagnostics, diagnostics0...)
	result.SemanticIR = ir0
	result.Generations = append(result.Generations, GenerationRecord{
		Generation: "stage0", SyntaxTreeDigest: model.SyntaxTreeDigest(tree0), SemanticIRDigest: model.GrammarDigest(ir0), ParserIdentity: "independent-stage0-seed",
	})

	tree1, err := generated.ParseStage1(source)
	if err != nil {
		return unknownResult(result, terminalFor("stage1", "generated-parse", "EVIDENCE_INCOMPLETE", "generated-parser-cannot-parse-grammar-source"), err), nil
	}
	result.Stage1Tree = tree1
	ir1, diagnostics1, err := lowering.Lower(tree1)
	if err != nil {
		return unknownResult(result, terminalFor("stage1", "lower", "EVIDENCE_INCOMPLETE", "stage1-lowering-failed"), err), nil
	}
	result.Diagnostics = append(result.Diagnostics, diagnostics1...)
	result.Generations = append(result.Generations, GenerationRecord{
		Generation: "stage1", SyntaxTreeDigest: model.SyntaxTreeDigest(tree1), SemanticIRDigest: model.GrammarDigest(ir1), ParserIdentity: generated.Stage1ParserIdentity,
	})
	result.Stage1Artifact = RenderParser(ir0, 1, "generated", "ParseStage1")
	result.Stage2Artifact = RenderParser(ir1, 2, "generated", "ParseStage2")

	tree2, err := generated.ParseStage2(source)
	if err != nil {
		return unknownResult(result, terminalFor("stage2", "generated-parse", "EVIDENCE_INCOMPLETE", "stage2-parser-cannot-parse-grammar-source"), err), nil
	}
	result.Stage2Tree = tree2
	ir2, diagnostics2, err := lowering.Lower(tree2)
	if err != nil {
		return unknownResult(result, terminalFor("stage2", "lower", "EVIDENCE_INCOMPLETE", "stage2-lowering-failed"), err), nil
	}
	result.Diagnostics = append(result.Diagnostics, diagnostics2...)
	result.Generations = append(result.Generations, GenerationRecord{
		Generation: "stage2", SyntaxTreeDigest: model.SyntaxTreeDigest(tree2), SemanticIRDigest: model.GrammarDigest(ir2), ParserIdentity: generated.Stage2ParserIdentity,
	})

	decision, record := decide(result.Diagnostics, result.Generations, ir0, ir1, ir2)
	result.Decision, result.Terminal, terminal = decision, record, record
	result.TerminalDigest = model.TerminalDigest(terminal)
	stampTerminalDigest(&result)
	return result, nil
}

func closedTerminal() model.TerminalRecord {
	return model.TerminalRecord{Decision: DecisionClosed, Stage: "compare", Step: "generation-equivalence", Reason: "SEMANTIC_IR_CANONICAL", UnknownClass: "", NextOperation: "", BlockedBy: ""}
}

func terminalFor(stage, step, reason, blockedBy string) model.TerminalRecord {
	record := model.TerminalRecord{Decision: DecisionUnknown, Stage: stage, Step: step, Reason: reason, UnknownClass: "evidence-incomplete", NextOperation: "inspect generation evidence", BlockedBy: blockedBy}
	if reason == "SEED_SYNTAX_INVALID" || reason == "SEMANTIC_LOWERING_INVALID" {
		record.Decision = DecisionRefuted
		record.UnknownClass, record.NextOperation = "", ""
	}
	return record
}

func refutedResult(result Result, terminal model.TerminalRecord, err error) Result {
	result.Decision, result.Terminal = DecisionRefuted, terminal
	result.TerminalDigest = model.TerminalDigest(terminal)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, model.Diagnostic{Class: lowering.IssueInvalidGrammar, Reason: err.Error(), Stage: terminal.Stage, Step: terminal.Step})
	}
	stampTerminalDigest(&result)
	return result
}

func unknownResult(result Result, terminal model.TerminalRecord, err error) Result {
	result.Decision, result.Terminal = DecisionUnknown, terminal
	result.TerminalDigest = model.TerminalDigest(terminal)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, model.Diagnostic{Class: "evidence-incomplete", Reason: err.Error(), Stage: terminal.Stage, Step: terminal.Step})
	}
	stampTerminalDigest(&result)
	return result
}

func stampTerminalDigest(result *Result) {
	for index := range result.Generations {
		result.Generations[index].TerminalReasonDigest = result.TerminalDigest
	}
}

func decide(diagnostics []model.Diagnostic, generations []GenerationRecord, ir0, ir1, ir2 model.GrammarIR) (string, model.TerminalRecord) {
	if lowering.HasClass(diagnostics, lowering.IssueConflictingPrecedence) || lowering.HasClass(diagnostics, lowering.IssueForbiddenStageEscape) || lowering.HasClass(diagnostics, lowering.IssueForbiddenEffect) || lowering.HasClass(diagnostics, lowering.IssueInvalidGrammar) {
		return DecisionRefuted, model.TerminalRecord{Decision: DecisionRefuted, Stage: "compare", Step: "contract-policy", Reason: highestDiagnosticReason(diagnostics, lowering.IssueConflictingPrecedence, lowering.IssueForbiddenStageEscape, lowering.IssueForbiddenEffect, lowering.IssueInvalidGrammar), UnknownClass: "", NextOperation: "", BlockedBy: ""}
	}
	if lowering.HasClass(diagnostics, lowering.IssueAmbiguous) {
		return DecisionUnknown, model.TerminalRecord{Decision: DecisionUnknown, Stage: "lower", Step: "ambiguity-analysis", Reason: "AMBIGUITY_NOT_DECIDABLE", UnknownClass: "grammar-ambiguity", NextOperation: "supply explicit disambiguation or bounded witness", BlockedBy: "duplicate-productions-have-multiple-derivations"}
	}
	if len(generations) != 3 || generations[0].SyntaxTreeDigest != generations[1].SyntaxTreeDigest || generations[1].SyntaxTreeDigest != generations[2].SyntaxTreeDigest || model.GrammarDigest(ir0) != model.GrammarDigest(ir1) || model.GrammarDigest(ir1) != model.GrammarDigest(ir2) {
		return DecisionRefuted, model.TerminalRecord{Decision: DecisionRefuted, Stage: "compare", Step: "generation-equivalence", Reason: "GENERATION_DRIFT", UnknownClass: "", NextOperation: "", BlockedBy: "stage-digests-differ"}
	}
	return DecisionClosed, closedTerminal()
}

func highestDiagnosticReason(diagnostics []model.Diagnostic, classes ...string) string {
	priority := map[string]int{}
	for index, class := range classes {
		priority[class] = len(classes) - index
	}
	best := "CONTRACT_POLICY_VIOLATION"
	bestPriority := 0
	for _, diagnostic := range diagnostics {
		if value := priority[diagnostic.Class]; value > bestPriority {
			best, bestPriority = strings.ToUpper(strings.ReplaceAll(diagnostic.Class, "-", "_")), value
		}
	}
	return best
}

func RenderParser(ir model.GrammarIR, stage int, packageName, functionName string) []byte {
	if packageName == "" {
		packageName = "generated"
	}
	if functionName == "" {
		functionName = fmt.Sprintf("ParseStage%d", stage)
	}
	tokenNames := make([]string, 0, len(ir.Tokens))
	for _, token := range ir.Tokens {
		tokenNames = append(tokenNames, token.Name)
	}
	sort.Strings(tokenNames)
	tokenJSON, _ := json.Marshal(tokenNames)
	tokenLiteral := "{" + strings.TrimPrefix(strings.TrimSuffix(string(tokenJSON), "]"), "[") + "}"
	source := parserTemplate
	replacements := map[string]string{
		"__PACKAGE__":        packageName,
		"__FUNCTION__":       functionName,
		"__STAGE__":          fmt.Sprint(stage),
		"__GRAMMAR_DIGEST__": model.GrammarDigest(ir),
		"__TOKEN_NAMES__":    tokenLiteral,
	}
	for key, value := range replacements {
		source = strings.ReplaceAll(source, key, value)
	}
	if formatted, err := format.Source([]byte(source)); err == nil {
		return formatted
	}
	return []byte(source)
}

func WriteArtifact(path string, raw []byte) error {
	if path == "" || !filepath.IsAbs(path) {
		return fmt.Errorf("generated artifact path must be absolute")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

const parserTemplate = `// Code generated by gooo-self-hosted-grammar-compiler; DO NOT EDIT.
package __PACKAGE__

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

const GeneratedStage = __STAGE__
const GeneratedGrammarDigest = "__GRAMMAR_DIGEST__"
var GeneratedTokenNames = []string__TOKEN_NAMES__

type generatedField struct { Kind string; Value string }

func __FUNCTION__(raw []byte) (model.SyntaxTree, error) {
	tree := model.SyntaxTree{Schema: model.SyntaxTreeSchema, Root: model.SyntaxNode{Kind: "grammar-file"}}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	line := 0
	grammarSeen := false
	programSeen := false
	for scanner.Scan() {
		line++
		fields, err := generatedFields(scanner.Text())
		if err != nil { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: %w", GeneratedStage, line, err) }
		if len(fields) == 0 { continue }
		kind := fields[0].Value
		if kind == "grammar" {
			if grammarSeen || len(fields) != 3 { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: invalid grammar declaration", GeneratedStage, line) }
			grammarSeen = true
		} else if kind == "package" || kind == "namespace" {
			if len(fields) != 2 { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: invalid %s declaration", GeneratedStage, line, kind) }
			programSeen = true
		} else if kind == "program" {
			if len(fields) < 2 { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: incomplete %s declaration", GeneratedStage, line, kind) }
			programSeen = true
		} else if kind == "entity" || kind == "activity" {
			if len(fields) < 2 { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: incomplete %s declaration", GeneratedStage, line, kind) }
		} else if kind == "stage" || kind == "ambiguity" || kind == "type" || kind == "effect" || kind == "fixed_denominator" {
			if len(fields) < 2 { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: incomplete declaration", GeneratedStage, line) }
		} else if kind == "token" {
			if len(fields) < 3 { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: incomplete token declaration", GeneratedStage, line) }
		} else if kind == "production" {
			if len(fields) < 4 || fields[2].Value != "->" { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: invalid production declaration", GeneratedStage, line) }
		} else if kind == "precedence" || kind == "associativity" {
			if len(fields) < 3 { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: incomplete declaration", GeneratedStage, line) }
		} else { return model.SyntaxTree{}, fmt.Errorf("generated stage %d line %d: unsupported declaration %q", GeneratedStage, line, kind) }
		node := model.SyntaxNode{Kind: kind, Line: line}
		for _, field := range fields { childKind := "field"; if field.Kind == "literal" || field.Kind == "regex" { childKind = field.Kind }; node.Children = append(node.Children, model.SyntaxNode{Kind: childKind, Value: field.Value, Line: line}) }
		tree.Root.Children = append(tree.Root.Children, node)
	}
	if err := scanner.Err(); err != nil { return model.SyntaxTree{}, err }
	if !grammarSeen && !programSeen { return model.SyntaxTree{}, fmt.Errorf("generated stage %d: grammar or Gooo program declaration is missing", GeneratedStage) }
	return tree, nil
}

func generatedFields(line string) ([]generatedField, error) {
	line = generatedCommentless(line)
	var fields []generatedField
	for i := 0; i < len(line); {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t' || line[i] == '\r') { i++ }
		if i == len(line) { break }
		start := i
		if line[i] == '"' {
			i++
			for i < len(line) { if line[i] == '\\' { i += 2; continue }; if line[i] == '"' { i++; break }; i++ }
			if i > len(line) || line[i-1] != '"' { return nil, fmt.Errorf("unterminated string") }
			value, err := strconv.Unquote(line[start:i]); if err != nil { return nil, err }; fields = append(fields, generatedField{Kind: "literal", Value: value}); continue
		}
		if line[i] == '/' {
			i++; escaped := false
			for i < len(line) { if !escaped && line[i] == '/' { i++; break }; if !escaped && line[i] == '\\' { escaped = true } else { escaped = false }; i++ }
			if i > len(line) || line[i-1] != '/' { return nil, fmt.Errorf("unterminated regex") }; fields = append(fields, generatedField{Kind: "regex", Value: line[start+1:i-1]}); continue
		}
		for i < len(line) && line[i] != ' ' && line[i] != '\t' && line[i] != '\r' { i++ }
		fields = append(fields, generatedField{Kind: "bare", Value: line[start:i]})
	}
	return fields, nil
}

func generatedCommentless(line string) string {
	quoted, regex, escaped := false, false, false
	for i := 0; i < len(line); i++ { ch := line[i]; if escaped { escaped = false; continue }; if ch == '\\' && (quoted || regex) { escaped = true; continue }; if ch == '"' && !regex { quoted = !quoted; continue }; if ch == '/' && !quoted { regex = !regex; continue }; if ch == '#' && !quoted && !regex { return strings.TrimSpace(line[:i]) } }
	return strings.TrimSpace(line)
}
`
