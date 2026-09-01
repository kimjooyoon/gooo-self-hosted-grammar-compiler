// Package stage0 is the independent seed parser. It intentionally does not
// call generated parser code or share parser routines with stage1.
package stage0

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
)

type seedField struct {
	Kind  string
	Value string
}

// Parse is a small, deliberately independent line parser for the seed
// bootstrap. Its only shared surface with generated parsers is the data model.
func Parse(raw []byte) (model.SyntaxTree, error) {
	tree := model.SyntaxTree{
		Schema: model.SyntaxTreeSchema,
		Root:   model.SyntaxNode{Kind: "grammar-file"},
	}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	line := 0
	grammarSeen := false
	for scanner.Scan() {
		line++
		fields, err := seedFields(scanner.Text())
		if err != nil {
			return model.SyntaxTree{}, fmt.Errorf("stage0 line %d: %w", line, err)
		}
		if len(fields) == 0 {
			continue
		}
		kind := fields[0].Value
		switch kind {
		case "grammar":
			if grammarSeen || len(fields) != 3 {
				return model.SyntaxTree{}, fmt.Errorf("stage0 line %d: invalid grammar declaration", line)
			}
			grammarSeen = true
		case "stage", "ambiguity", "type":
			if len(fields) < 2 {
				return model.SyntaxTree{}, fmt.Errorf("stage0 line %d: incomplete %s declaration", line, kind)
			}
		case "effect", "fixed_denominator":
			if len(fields) < 2 {
				return model.SyntaxTree{}, fmt.Errorf("stage0 line %d: incomplete %s declaration", line, kind)
			}
		case "token":
			if len(fields) < 3 {
				return model.SyntaxTree{}, fmt.Errorf("stage0 line %d: incomplete token declaration", line)
			}
		case "production":
			if len(fields) < 4 || fields[2].Value != "->" {
				return model.SyntaxTree{}, fmt.Errorf("stage0 line %d: invalid production declaration", line)
			}
		case "precedence", "associativity":
			if len(fields) < 3 {
				return model.SyntaxTree{}, fmt.Errorf("stage0 line %d: incomplete %s declaration", line, kind)
			}
		default:
			return model.SyntaxTree{}, fmt.Errorf("stage0 line %d: unsupported declaration %q", line, kind)
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
		return model.SyntaxTree{}, fmt.Errorf("stage0: grammar declaration is missing")
	}
	return tree, nil
}

func seedFields(line string) ([]seedField, error) {
	line = seedCommentless(line)
	var fields []seedField
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
			if i > len(line) || line[i-1] != '"' {
				return nil, fmt.Errorf("unterminated string")
			}
			value, err := strconv.Unquote(line[start:i])
			if err != nil {
				return nil, fmt.Errorf("invalid string: %w", err)
			}
			fields = append(fields, seedField{Kind: "literal", Value: value})
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
			if i > len(line) || line[i-1] != '/' {
				return nil, fmt.Errorf("unterminated regex")
			}
			fields = append(fields, seedField{Kind: "regex", Value: line[start+1 : i-1]})
		default:
			for i < len(line) && line[i] != ' ' && line[i] != '\t' && line[i] != '\r' {
				i++
			}
			if start == i {
				return nil, fmt.Errorf("empty field")
			}
			fields = append(fields, seedField{Kind: "bare", Value: line[start:i]})
		}
	}
	return fields, nil
}

func seedCommentless(line string) string {
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
