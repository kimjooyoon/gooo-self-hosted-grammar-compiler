// Code generated for the gooo grammar bootstrap. DO NOT EDIT.
// The implementation is intentionally separate from internal/stage0.
package generated

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

const Stage1ParserIdentity = "generated-stage1-from-gooo-grammar-v1"

type generatedField struct {
	Kind  string
	Value string
}

// ParseStage1 is the first generated parser. It parses the grammar source
// itself; no stage0 parser package is imported here.
func ParseStage1(raw []byte) (model.SyntaxTree, error) {
	return parseGeneratedGrammar(raw, "stage1")
}

func parseGeneratedGrammar(raw []byte, stage string) (model.SyntaxTree, error) {
	tree := model.SyntaxTree{Schema: model.SyntaxTreeSchema, Root: model.SyntaxNode{Kind: "grammar-file"}}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	line := 0
	grammarSeen := false
	for scanner.Scan() {
		line++
		fields, err := generatedFields(scanner.Text())
		if err != nil {
			return model.SyntaxTree{}, fmt.Errorf("%s line %d: %w", stage, line, err)
		}
		if len(fields) == 0 {
			continue
		}
		kind := fields[0].Value
		if err := validateGeneratedDeclaration(kind, fields, line); err != nil {
			return model.SyntaxTree{}, fmt.Errorf("%s: %w", stage, err)
		}
		if kind == "grammar" {
			if grammarSeen {
				return model.SyntaxTree{}, fmt.Errorf("%s line %d: duplicate grammar declaration", stage, line)
			}
			grammarSeen = true
		}
		node := model.SyntaxNode{Kind: kind, Line: line}
		for _, field := range fields {
			childKind := "field"
			if field.Kind == "literal" || field.Kind == "regex" {
				childKind = field.Kind
			}
			node.Children = append(node.Children, model.SyntaxNode{Kind: childKind, Value: field.Value, Line: line})
		}
		tree.Root.Children = append(tree.Root.Children, node)
	}
	if err := scanner.Err(); err != nil {
		return model.SyntaxTree{}, err
	}
	if !grammarSeen {
		return model.SyntaxTree{}, fmt.Errorf("%s: grammar declaration is missing", stage)
	}
	return tree, nil
}

func validateGeneratedDeclaration(kind string, fields []generatedField, line int) error {
	switch kind {
	case "package", "namespace":
		if len(fields) != 2 {
			return fmt.Errorf("line %d: %s expects a name", line, kind)
		}
	case "program":
		if len(fields) < 2 {
			return fmt.Errorf("line %d: program expects a name", line)
		}
	case "entity":
		if len(fields) < 2 {
			return fmt.Errorf("line %d: entity expects a name", line)
		}
	case "activity":
		if len(fields) < 2 {
			return fmt.Errorf("line %d: activity expects a signature", line)
		}
	case "grammar":
		if len(fields) != 3 {
			return fmt.Errorf("line %d: invalid grammar declaration", line)
		}
	case "stage", "ambiguity", "type", "effect", "fixed_denominator":
		if len(fields) < 2 {
			return fmt.Errorf("line %d: incomplete %s declaration", line, kind)
		}
	case "token":
		if len(fields) < 3 {
			return fmt.Errorf("line %d: incomplete token declaration", line)
		}
	case "production":
		if len(fields) < 4 || fields[2].Value != "->" {
			return fmt.Errorf("line %d: invalid production declaration", line)
		}
	case "precedence", "associativity":
		if len(fields) < 3 {
			return fmt.Errorf("line %d: incomplete %s declaration", line, kind)
		}
	default:
		return fmt.Errorf("line %d: unsupported declaration %q", line, kind)
	}
	return nil
}

func generatedFields(line string) ([]generatedField, error) {
	line = generatedCommentless(line)
	var fields []generatedField
	for i := 0; i < len(line); {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t' || line[i] == '\r') {
			i++
		}
		if i == len(line) {
			break
		}
		start := i
		switch line[i] {
		case '"':
			i++
			for i < len(line) {
				if line[i] == '\\' {
					i += 2
					continue
				}
				if line[i] == '"' {
					i++
					break
				}
				i++
			}
			if i > len(line) || i == 0 || line[i-1] != '"' {
				return nil, fmt.Errorf("unterminated string")
			}
			value, err := strconv.Unquote(line[start:i])
			if err != nil {
				return nil, fmt.Errorf("invalid string: %w", err)
			}
			fields = append(fields, generatedField{Kind: "literal", Value: value})
		case '/':
			i++
			escaped := false
			for i < len(line) {
				if !escaped && line[i] == '/' {
					i++
					break
				}
				if !escaped && line[i] == '\\' {
					escaped = true
				} else {
					escaped = false
				}
				i++
			}
			if i > len(line) || i == 0 || line[i-1] != '/' {
				return nil, fmt.Errorf("unterminated regex")
			}
			fields = append(fields, generatedField{Kind: "regex", Value: line[start+1 : i-1]})
		default:
			for i < len(line) && line[i] != ' ' && line[i] != '\t' && line[i] != '\r' {
				i++
			}
			fields = append(fields, generatedField{Kind: "bare", Value: line[start:i]})
		}
	}
	return fields, nil
}

func generatedCommentless(line string) string {
	quoted := false
	regex := false
	escaped := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && (quoted || regex) {
			escaped = true
			continue
		}
		if ch == '"' && !regex {
			quoted = !quoted
			continue
		}
		if ch == '/' && !quoted {
			regex = !regex
			continue
		}
		if ch == '#' && !quoted && !regex {
			return strings.TrimSpace(line[:i])
		}
	}
	return strings.TrimSpace(line)
}
