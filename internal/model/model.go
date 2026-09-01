package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	SyntaxTreeSchema = "gooo/self-hosted-grammar-compiler/syntax-tree/v1"
	GrammarIRSchema  = "gooo/self-hosted-grammar-compiler/grammar-ir/v1"
	EvidenceSchema   = "gooo/self-hosted-grammar-compiler/ci-evidence/v1"
)

type SyntaxTree struct {
	Schema string     `json:"schema"`
	Root   SyntaxNode `json:"root"`
}

type SyntaxNode struct {
	Kind     string       `json:"kind"`
	Value    string       `json:"value,omitempty"`
	Line     int          `json:"line,omitempty"`
	Children []SyntaxNode `json:"children,omitempty"`
}

type StageDecl struct {
	Number int    `json:"number"`
	Origin string `json:"origin"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TokenDecl struct {
	Name       string      `json:"name"`
	Pattern    string      `json:"pattern"`
	Skip       bool        `json:"skip"`
	Attributes []Attribute `json:"attributes,omitempty"`
}

type Symbol struct {
	Value    string `json:"value"`
	Terminal bool   `json:"terminal"`
}

type ProductionDecl struct {
	Name string   `json:"name"`
	RHS  []Symbol `json:"rhs"`
}

type PrecedenceDecl struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type AssociativityDecl struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type GrammarIR struct {
	Schema           string              `json:"schema"`
	Name             string              `json:"name"`
	Version          string              `json:"version"`
	Stage            StageDecl           `json:"stage"`
	AmbiguityPolicy  string              `json:"ambiguity_policy"`
	Types            []string            `json:"types"`
	Effects          []string            `json:"effects"`
	FixedDenominator int                 `json:"fixed_denominator"`
	Tokens           []TokenDecl         `json:"tokens"`
	Productions      []ProductionDecl    `json:"productions"`
	Precedences      []PrecedenceDecl    `json:"precedences"`
	Associativities  []AssociativityDecl `json:"associativities"`
}

type Diagnostic struct {
	Class  string `json:"class"`
	Reason string `json:"reason"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
}

type TerminalRecord struct {
	Decision      string `json:"decision"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	UnknownClass  string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
	BlockedBy     string `json:"blocked_by"`
}

func DigestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func DigestJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "json-error"
	}
	return DigestBytes(raw)
}

func SyntaxTreeDigest(tree SyntaxTree) string {
	// Source line numbers are evidence, not syntax. Excluding them makes a
	// formatting-only change compare by structure and terminal values.
	type canonicalNode struct {
		Kind     string          `json:"kind"`
		Value    string          `json:"value,omitempty"`
		Children []canonicalNode `json:"children,omitempty"`
	}
	var canon func(SyntaxNode) canonicalNode
	canon = func(node SyntaxNode) canonicalNode {
		children := make([]canonicalNode, 0, len(node.Children))
		for _, child := range node.Children {
			children = append(children, canon(child))
		}
		return canonicalNode{Kind: node.Kind, Value: node.Value, Children: children}
	}
	return DigestJSON(struct {
		Schema string        `json:"schema"`
		Root   canonicalNode `json:"root"`
	}{Schema: tree.Schema, Root: canon(tree.Root)})
}

func GrammarDigest(grammar GrammarIR) string { return DigestJSON(grammar) }

func TerminalDigest(record TerminalRecord) string { return DigestJSON(record) }
