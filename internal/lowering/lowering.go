package lowering

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

const (
	IssueAmbiguous             = "ambiguous-grammar"
	IssueConflictingPrecedence = "conflicting-precedence"
	IssueForbiddenStageEscape  = "forbidden-stage-escape"
	IssueForbiddenEffect       = "forbidden-effect"
	IssueInvalidGrammar        = "invalid-grammar"
)

// Lower turns the stage0 or generated syntax tree into the semantic grammar
// IR. Diagnostics are deterministic and are part of the bootstrap evidence.
func Lower(tree model.SyntaxTree) (model.GrammarIR, []model.Diagnostic, error) {
	ir := model.GrammarIR{Schema: model.GrammarIRSchema, Types: []string{}, Effects: []string{}, Tokens: []model.TokenDecl{}, Productions: []model.ProductionDecl{}, Precedences: []model.PrecedenceDecl{}, Associativities: []model.AssociativityDecl{}}
	var diagnostics []model.Diagnostic
	grammarCount := 0
	stageCount := 0
	ambiguityCount := 0
	for _, node := range tree.Root.Children {
		fields := nodeValues(node)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "grammar":
			grammarCount++
			if len(fields) != 3 {
				return ir, diagnostics, fmt.Errorf("line %d: grammar expects name and version", node.Line)
			}
			ir.Name, ir.Version = fields[1], fields[2]
		case "stage":
			stageCount++
			if len(fields) < 3 {
				return ir, diagnostics, fmt.Errorf("line %d: stage expects number and origin", node.Line)
			}
			number, err := strconv.Atoi(fields[1])
			if err != nil {
				return ir, diagnostics, fmt.Errorf("line %d: invalid stage number", node.Line)
			}
			origin, ok := keyValue(fields[2:], "origin")
			if !ok {
				return ir, diagnostics, fmt.Errorf("line %d: stage origin is missing", node.Line)
			}
			ir.Stage = model.StageDecl{Number: number, Origin: origin}
			if number != 1 || origin != "generated-from-stage0" {
				diagnostics = append(diagnostics, model.Diagnostic{Class: IssueForbiddenStageEscape, Reason: "stage must be 1 with origin generated-from-stage0", Stage: "lower", Step: "stage-origin-policy"})
			}
		case "ambiguity":
			ambiguityCount++
			policy, ok := keyValue(fields[1:], "policy")
			if !ok {
				return ir, diagnostics, fmt.Errorf("line %d: ambiguity policy is missing", node.Line)
			}
			ir.AmbiguityPolicy = policy
			if policy != "reject" && policy != "prefer-first" && policy != "preserve-all" {
				return ir, diagnostics, fmt.Errorf("line %d: unsupported ambiguity policy %q", node.Line, policy)
			}
		case "type":
			if len(fields) != 2 || fields[1] == "" {
				return ir, diagnostics, fmt.Errorf("line %d: type expects a name", node.Line)
			}
			ir.Types = append(ir.Types, fields[1])
		case "effect":
			if len(fields) < 2 {
				return ir, diagnostics, fmt.Errorf("line %d: effect expects at least one capability", node.Line)
			}
			ir.Effects = append(ir.Effects, fields[1:]...)
		case "fixed_denominator":
			if len(fields) != 2 || !strings.HasPrefix(fields[1], "cases=") {
				return ir, diagnostics, fmt.Errorf("line %d: fixed_denominator expects cases=N", node.Line)
			}
			value, err := strconv.Atoi(strings.TrimPrefix(fields[1], "cases="))
			if err != nil || value < 1 {
				return ir, diagnostics, fmt.Errorf("line %d: invalid fixed denominator", node.Line)
			}
			ir.FixedDenominator = value
		case "token":
			if len(fields) < 3 {
				return ir, diagnostics, fmt.Errorf("line %d: token expects name and pattern", node.Line)
			}
			skip := false
			attributes := parseAttributes(fields[3:])
			if value, ok := attributeValue(attributes, "skip"); ok {
				parsed, err := strconv.ParseBool(value)
				if err != nil {
					return ir, diagnostics, fmt.Errorf("line %d: skip must be boolean", node.Line)
				}
				skip = parsed
			}
			ir.Tokens = append(ir.Tokens, model.TokenDecl{Name: fields[1], Pattern: fields[2], Skip: skip, Attributes: attributes})
		case "production":
			if len(fields) < 4 || fields[2] != "->" {
				return ir, diagnostics, fmt.Errorf("line %d: production expects name -> symbols", node.Line)
			}
			rhs := make([]model.Symbol, 0, len(fields)-3)
			for index, value := range fields[3:] {
				terminal := node.Children[index+3].Kind == "literal" || node.Children[index+3].Kind == "regex"
				rhs = append(rhs, model.Symbol{Value: value, Terminal: terminal})
			}
			ir.Productions = append(ir.Productions, model.ProductionDecl{Name: fields[1], RHS: rhs})
		case "precedence":
			if len(fields) < 3 {
				return ir, diagnostics, fmt.Errorf("line %d: precedence expects name and level", node.Line)
			}
			levelText, ok := keyValue(fields[2:], "level")
			if !ok {
				return ir, diagnostics, fmt.Errorf("line %d: precedence level is missing", node.Line)
			}
			level, err := strconv.Atoi(levelText)
			if err != nil || level < 0 {
				return ir, diagnostics, fmt.Errorf("line %d: invalid precedence level", node.Line)
			}
			ir.Precedences = append(ir.Precedences, model.PrecedenceDecl{Name: fields[1], Level: level})
		case "associativity":
			if len(fields) != 3 || (fields[2] != "left" && fields[2] != "right" && fields[2] != "nonassoc") {
				return ir, diagnostics, fmt.Errorf("line %d: associativity expects name and left/right/nonassoc", node.Line)
			}
			ir.Associativities = append(ir.Associativities, model.AssociativityDecl{Name: fields[1], Value: fields[2]})
		default:
			return ir, diagnostics, fmt.Errorf("line %d: unsupported declaration %q", node.Line, fields[0])
		}
	}
	if grammarCount != 1 || stageCount != 1 || ambiguityCount != 1 {
		return ir, diagnostics, fmt.Errorf("grammar requires exactly one grammar, stage, and ambiguity declaration")
	}
	if ir.Name == "" || ir.Version == "" || len(ir.Types) == 0 || len(ir.Effects) == 0 || ir.FixedDenominator != 7 || len(ir.Tokens) == 0 || len(ir.Productions) == 0 || len(ir.Precedences) == 0 || len(ir.Associativities) == 0 {
		return ir, diagnostics, fmt.Errorf("grammar declarations are incomplete")
	}
	diagnostics = append(diagnostics, validateCollections(ir)...)
	return Canonical(ir), canonicalDiagnostics(diagnostics), nil
}

func nodeValues(node model.SyntaxNode) []string {
	values := make([]string, 0, len(node.Children))
	for _, child := range node.Children {
		values = append(values, child.Value)
	}
	return values
}

func keyValue(fields []string, key string) (string, bool) {
	needle := key + "="
	found := false
	value := ""
	for _, field := range fields {
		if strings.HasPrefix(field, needle) && len(field) > len(needle) {
			if found {
				return "", false
			}
			value = strings.TrimPrefix(field, needle)
			found = true
		}
	}
	return value, found
}

func parseAttributes(fields []string) []model.Attribute {
	attributes := make([]model.Attribute, 0, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			attributes = append(attributes, model.Attribute{Key: parts[0], Value: parts[1]})
		}
	}
	sort.Slice(attributes, func(i, j int) bool {
		if attributes[i].Key == attributes[j].Key {
			return attributes[i].Value < attributes[j].Value
		}
		return attributes[i].Key < attributes[j].Key
	})
	return attributes
}

func attributeValue(attributes []model.Attribute, key string) (string, bool) {
	for _, attribute := range attributes {
		if attribute.Key == key {
			return attribute.Value, true
		}
	}
	return "", false
}

func validateCollections(ir model.GrammarIR) []model.Diagnostic {
	var diagnostics []model.Diagnostic
	allowedEffects := map[string]bool{"parse": true, "lower": true, "generate": true, "execute": true, "verify": true}
	for _, effect := range ir.Effects {
		if !allowedEffects[effect] {
			diagnostics = append(diagnostics, model.Diagnostic{Class: IssueForbiddenEffect, Reason: "effect is outside parse/lower/generate/execute/verify", Stage: "lower", Step: "effect-policy"})
		}
	}
	tokens := map[string]bool{}
	for _, token := range ir.Tokens {
		if token.Name == "" || tokens[token.Name] {
			diagnostics = append(diagnostics, model.Diagnostic{Class: IssueInvalidGrammar, Reason: "duplicate or empty token name", Stage: "lower", Step: "token-uniqueness"})
		}
		tokens[token.Name] = true
		attributes := map[string]bool{}
		for _, attribute := range token.Attributes {
			if attributes[attribute.Key] {
				diagnostics = append(diagnostics, model.Diagnostic{Class: IssueInvalidGrammar, Reason: "duplicate token attribute key", Stage: "lower", Step: "token-attribute-uniqueness"})
			}
			attributes[attribute.Key] = true
		}
	}
	productions := map[string]bool{}
	productionBodies := map[string]bool{}
	for _, production := range ir.Productions {
		productions[production.Name] = true
		body := make([]string, 0, len(production.RHS))
		for _, symbol := range production.RHS {
			body = append(body, fmt.Sprintf("%t:%s", symbol.Terminal, symbol.Value))
		}
		key := production.Name + "->" + strings.Join(body, " ")
		if productionBodies[key] {
			diagnostics = append(diagnostics, model.Diagnostic{Class: IssueAmbiguous, Reason: "duplicate production admits more than one derivation", Stage: "lower", Step: "ambiguity-analysis"})
		}
		productionBodies[key] = true
	}
	precedenceLevels := map[string]int{}
	for _, precedence := range ir.Precedences {
		if previous, exists := precedenceLevels[precedence.Name]; exists && previous != precedence.Level {
			diagnostics = append(diagnostics, model.Diagnostic{Class: IssueConflictingPrecedence, Reason: "one operator has multiple precedence levels", Stage: "lower", Step: "precedence-consistency"})
		}
		precedenceLevels[precedence.Name] = precedence.Level
	}
	associativities := map[string]string{}
	for _, associativity := range ir.Associativities {
		if previous, exists := associativities[associativity.Name]; exists && previous != associativity.Value {
			diagnostics = append(diagnostics, model.Diagnostic{Class: IssueConflictingPrecedence, Reason: "one operator has multiple associativities", Stage: "lower", Step: "associativity-consistency"})
		}
		associativities[associativity.Name] = associativity.Value
	}
	for _, production := range ir.Productions {
		if !productions[production.Name] {
			diagnostics = append(diagnostics, model.Diagnostic{Class: IssueInvalidGrammar, Reason: "production name is not bound", Stage: "lower", Step: "production-binding"})
		}
		for _, symbol := range production.RHS {
			if symbol.Terminal {
				continue
			}
			if !tokens[symbol.Value] && !productions[symbol.Value] {
				diagnostics = append(diagnostics, model.Diagnostic{Class: IssueInvalidGrammar, Reason: "production references unknown symbol", Stage: "lower", Step: "symbol-binding"})
			}
		}
	}
	for _, associativity := range ir.Associativities {
		if !containsPrecedence(ir.Precedences, associativity.Name) {
			diagnostics = append(diagnostics, model.Diagnostic{Class: IssueInvalidGrammar, Reason: "associativity has no precedence declaration", Stage: "lower", Step: "precedence-binding"})
		}
	}
	return diagnostics
}

func containsPrecedence(values []model.PrecedenceDecl, name string) bool {
	for _, value := range values {
		if value.Name == name {
			return true
		}
	}
	return false
}

func Canonical(ir model.GrammarIR) model.GrammarIR {
	canonical := ir
	canonical.Tokens = append([]model.TokenDecl(nil), ir.Tokens...)
	canonical.Productions = append([]model.ProductionDecl(nil), ir.Productions...)
	canonical.Precedences = append([]model.PrecedenceDecl(nil), ir.Precedences...)
	canonical.Associativities = append([]model.AssociativityDecl(nil), ir.Associativities...)
	canonical.Types = append([]string(nil), ir.Types...)
	canonical.Effects = append([]string(nil), ir.Effects...)
	sort.Strings(canonical.Types)
	sort.Strings(canonical.Effects)
	sort.Slice(canonical.Tokens, func(i, j int) bool { return canonical.Tokens[i].Name < canonical.Tokens[j].Name })
	sort.Slice(canonical.Precedences, func(i, j int) bool {
		if canonical.Precedences[i].Name == canonical.Precedences[j].Name {
			return canonical.Precedences[i].Level < canonical.Precedences[j].Level
		}
		return canonical.Precedences[i].Name < canonical.Precedences[j].Name
	})
	sort.Slice(canonical.Associativities, func(i, j int) bool { return canonical.Associativities[i].Name < canonical.Associativities[j].Name })
	for index := range canonical.Tokens {
		canonical.Tokens[index].Attributes = append([]model.Attribute(nil), canonical.Tokens[index].Attributes...)
	}
	return canonical
}

func canonicalDiagnostics(values []model.Diagnostic) []model.Diagnostic {
	copyValues := append([]model.Diagnostic(nil), values...)
	sort.Slice(copyValues, func(i, j int) bool {
		if copyValues[i].Class != copyValues[j].Class {
			return copyValues[i].Class < copyValues[j].Class
		}
		if copyValues[i].Reason != copyValues[j].Reason {
			return copyValues[i].Reason < copyValues[j].Reason
		}
		return copyValues[i].Step < copyValues[j].Step
	})
	return copyValues
}

func HasClass(diagnostics []model.Diagnostic, class string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Class == class {
			return true
		}
	}
	return false
}
