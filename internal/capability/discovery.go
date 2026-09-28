package capability

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/lowering"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/stage0"
)

type Status string

const (
	StatusUnknown  Status = "UNKNOWN"
	StatusDeferred Status = "DEFERRED"
	StatusBound    Status = "BOUND"
)

type Capability struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SuggestedAction struct {
	ID            string   `json:"id"`
	Reason        string   `json:"reason"`
	CapabilityIDs []string `json:"capability_ids"`
}

type Report struct {
	Status              Status             `json:"status"`
	Query               string             `json:"query"`
	SourceDigest        string             `json:"source_digest"`
	GrammarDigest       string             `json:"grammar_digest,omitempty"`
	Capabilities        []Capability       `json:"capabilities"`
	MatchedCapabilities []string           `json:"matched_capabilities"`
	RelatedCapabilities []string           `json:"related_capabilities,omitempty"`
	SuggestedActions    []SuggestedAction  `json:"suggested_actions"`
	UnresolvedTerms     []string           `json:"unresolved_terms"`
	Diagnostics         []model.Diagnostic `json:"diagnostics,omitempty"`
	FirstMismatch       string             `json:"first_mismatch"`
	MissingStage        string             `json:"missing_stage"`
	NextQuestion        string             `json:"next_question"`
	Reason              string             `json:"reason"`
	ReadOnly            bool               `json:"read_only"`
	ReportDigest        string             `json:"report_digest"`
}

// Discover reads the authoritative grammar and exposes its declared surface.
// It never parses a user program, executes a generated parser, or mutates the
// grammar. Query matching is lexical and deliberately preserves uncertainty.
func Discover(raw []byte, query string) Report {
	report := Report{
		Status:           StatusUnknown,
		Query:            strings.TrimSpace(query),
		SourceDigest:     model.DigestBytes(raw),
		Capabilities:     []Capability{},
		SuggestedActions: []SuggestedAction{},
		ReadOnly:         true,
		FirstMismatch:    "grammar_source",
		MissingStage:     "grammar_source",
		Reason:           "authoritative grammar could not yet be bound",
	}
	tree, err := stage0.Parse(raw)
	if err != nil {
		report.FirstMismatch = "stage0_parse"
		report.MissingStage = "stage0_parse"
		report.NextQuestion = "Which authoritative .gooo grammar should be inspected?"
		report.finish()
		return report
	}
	ir, diagnostics, err := lowering.Lower(tree)
	if err != nil {
		report.FirstMismatch = "grammar_lower"
		report.MissingStage = "grammar_lower"
		report.NextQuestion = "Which grammar declaration should be repaired before capability discovery?"
		report.finish()
		return report
	}
	report.GrammarDigest = model.GrammarDigest(ir)
	report.Capabilities = capabilities(ir)
	report.Diagnostics = append([]model.Diagnostic(nil), diagnostics...)
	if len(diagnostics) > 0 {
		report.Status = StatusDeferred
		report.FirstMismatch = "grammar_diagnostics"
		report.MissingStage = "grammar_diagnostics"
		report.NextQuestion = "Which grammar diagnostic should be resolved before relying on this capability surface?"
		report.Reason = "grammar diagnostics prevent an unqualified capability binding"
		report.finish()
		return report
	}

	report.MatchedCapabilities, report.UnresolvedTerms, report.RelatedCapabilities = match(report.Query, report.Capabilities)
	report.SuggestedActions = suggestedActions(report.MatchedCapabilities, report.RelatedCapabilities)
	switch {
	case isOverviewQuery(report.Query) || strings.TrimSpace(report.Query) == "":
		report.Status = StatusBound
		report.FirstMismatch = ""
		report.MissingStage = ""
		report.NextQuestion = "Which declared grammar capability should be inspected next?"
		report.Reason = "the declared grammar capability surface is available for read-only inspection"
	case len(report.MatchedCapabilities) == 0 && len(report.RelatedCapabilities) > 0:
		report.Status = StatusDeferred
		report.FirstMismatch = "query_terms"
		report.MissingStage = "query_terms"
		report.NextQuestion = "Which related declared capability should be confirmed for this natural-language query?"
		report.Reason = "natural-language terms suggest declared capabilities, but the query is not directly bound"
	case len(report.MatchedCapabilities) == 0:
		report.Status = StatusUnknown
		report.FirstMismatch = "capability_query"
		report.MissingStage = "capability_query"
		report.NextQuestion = "Which token, production, effect, precedence, or associativity should clarify this query?"
		report.Reason = "the query did not match the declared grammar capability surface"
	case len(report.UnresolvedTerms) > 0:
		report.Status = StatusDeferred
		report.FirstMismatch = "query_terms"
		report.MissingStage = "query_terms"
		report.NextQuestion = "Which declared grammar capability should resolve the remaining query terms?"
		report.Reason = "some query terms remain unresolved against the declared grammar surface"
	default:
		report.Status = StatusBound
		report.FirstMismatch = ""
		report.MissingStage = ""
		report.NextQuestion = "Which declared grammar capability should be inspected next?"
		report.Reason = "the query matches the declared grammar capability surface"
	}
	report.finish()
	return report
}

func (report Report) Validate() error {
	if !report.ReadOnly || strings.TrimSpace(report.SourceDigest) == "" || strings.TrimSpace(report.ReportDigest) == "" {
		return fmt.Errorf("grammar capability report is incomplete or not read-only")
	}
	if report.Status != StatusUnknown && report.Status != StatusDeferred && report.Status != StatusBound {
		return fmt.Errorf("grammar capability report status %q is invalid", report.Status)
	}
	if report.Status != StatusBound && (report.FirstMismatch == "" || report.MissingStage == "") {
		return fmt.Errorf("unresolved grammar capability report lost its first boundary")
	}
	if report.Status == StatusBound && (report.FirstMismatch != "" || report.MissingStage != "") {
		return fmt.Errorf("bound grammar capability report retained an unresolved boundary")
	}
	for index, action := range report.SuggestedActions {
		if strings.TrimSpace(action.ID) == "" || strings.TrimSpace(action.Reason) == "" || len(action.CapabilityIDs) == 0 {
			return fmt.Errorf("suggested action %d is incomplete", index)
		}
	}
	if report.ReportDigest != report.digest() {
		return fmt.Errorf("grammar capability report digest mismatch")
	}
	return nil
}

func (report *Report) finish() {
	report.ReportDigest = report.digest()
}

func (report Report) digest() string {
	return model.DigestJSON(struct {
		Status              Status
		Query               string
		SourceDigest        string
		GrammarDigest       string
		Capabilities        []Capability
		MatchedCapabilities []string
		RelatedCapabilities []string
		SuggestedActions    []SuggestedAction
		UnresolvedTerms     []string
		Diagnostics         []model.Diagnostic
		FirstMismatch       string
		MissingStage        string
		NextQuestion        string
		Reason              string
		ReadOnly            bool
	}{
		Status: report.Status, Query: report.Query, SourceDigest: report.SourceDigest,
		GrammarDigest: report.GrammarDigest, Capabilities: report.Capabilities,
		MatchedCapabilities: report.MatchedCapabilities, RelatedCapabilities: report.RelatedCapabilities,
		SuggestedActions: report.SuggestedActions, UnresolvedTerms: report.UnresolvedTerms,
		Diagnostics: report.Diagnostics, FirstMismatch: report.FirstMismatch,
		MissingStage: report.MissingStage, NextQuestion: report.NextQuestion,
		Reason: report.Reason, ReadOnly: report.ReadOnly,
	})
}

func capabilities(ir model.GrammarIR) []Capability {
	result := []Capability{{ID: "grammar:" + ir.Name, Kind: "grammar", Name: ir.Name, Description: capabilityDescription("grammar", ir.Name)}}
	for _, value := range ir.Types {
		result = append(result, Capability{ID: "type:" + value, Kind: "type", Name: value, Description: capabilityDescription("type", value)})
	}
	for _, value := range ir.Effects {
		result = append(result, Capability{ID: "effect:" + value, Kind: "effect", Name: value, Description: capabilityDescription("effect", value)})
	}
	for _, value := range ir.Tokens {
		result = append(result, Capability{ID: "token:" + value.Name, Kind: "token", Name: value.Name, Description: capabilityDescription("token", value.Name)})
	}
	for _, value := range ir.Productions {
		result = append(result, Capability{ID: "production:" + value.Name, Kind: "production", Name: value.Name, Description: capabilityDescription("production", value.Name)})
	}
	for _, value := range ir.Precedences {
		result = append(result, Capability{ID: "precedence:" + value.Name, Kind: "precedence", Name: value.Name, Description: capabilityDescription("precedence", value.Name)})
	}
	for _, value := range ir.Associativities {
		result = append(result, Capability{ID: "associativity:" + value.Name, Kind: "associativity", Name: value.Name, Description: capabilityDescription("associativity", value.Name)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func capabilityDescription(kind, name string) string {
	switch kind {
	case "grammar":
		return "Authoritative .gooo grammar declaration: " + name
	case "type":
		return "Declared .gooo grammar type: " + name
	case "effect":
		return "Declared .gooo grammar effect: " + name
	case "token":
		return "Declared .gooo lexical token: " + name
	case "production":
		return "Declared .gooo syntax production: " + name
	case "precedence":
		return "Declared .gooo precedence entry: " + name
	case "associativity":
		return "Declared .gooo associativity entry: " + name
	default:
		return "Declared .gooo grammar capability: " + name
	}
}

func suggestedActions(matched, related []string) []SuggestedAction {
	ids := make([]string, 0, len(matched)+len(related))
	seen := make(map[string]struct{})
	for _, value := range append(append([]string{}, matched...), related...) {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ids = append(ids, value)
	}
	if len(ids) == 0 {
		return []SuggestedAction{}
	}
	sort.Strings(ids)
	actions := []SuggestedAction{{
		ID:            "inspect-declarations",
		Reason:        "Inspect declaration-backed capability evidence before attempting any operation; this hint is read-only.",
		CapabilityIDs: ids,
	}}
	for _, value := range matched {
		if !strings.HasPrefix(value, "effect:") {
			continue
		}
		effect := strings.TrimPrefix(value, "effect:")
		switch effect {
		case "execute", "generate", "lower", "parse", "verify":
			actions = append(actions, SuggestedAction{
				ID:            effect,
				Reason:        "A declared .gooo effect names this bounded next operation; the suggestion does not execute it.",
				CapabilityIDs: []string{value},
			})
		}
	}
	return actions
}

func match(query string, all []Capability) ([]string, []string, []string) {
	if isOverviewQuery(query) || strings.TrimSpace(query) == "" {
		matched := make([]string, 0, len(all))
		for _, value := range all {
			matched = append(matched, value.ID)
		}
		return matched, []string{}, []string{}
	}
	terms := queryTerms(query)
	matchedSet := make(map[string]struct{})
	relatedSet := make(map[string]struct{})
	unresolved := make([]string, 0)
	for _, term := range terms {
		found := false
		for _, value := range all {
			haystack := strings.ToLower(value.ID + " " + value.Kind + " " + value.Name)
			if strings.Contains(haystack, term) {
				matchedSet[value.ID] = struct{}{}
				found = true
			}
		}
		if found {
			continue
		}
		for _, kind := range relatedKinds(term) {
			for _, value := range all {
				if value.Kind == kind {
					relatedSet[value.ID] = struct{}{}
				}
			}
		}
		unresolved = append(unresolved, term)
	}
	matched := make([]string, 0, len(matchedSet))
	for value := range matchedSet {
		matched = append(matched, value)
	}
	related := make([]string, 0, len(relatedSet))
	for value := range relatedSet {
		if _, ok := matchedSet[value]; !ok {
			related = append(related, value)
		}
	}
	sort.Strings(matched)
	sort.Strings(related)
	return matched, unresolved, related
}

var queryCapabilityAliases = map[string][]string{
	"analyze":    {"parse", "lower"},
	"build":      {"generate"},
	"check":      {"verify"},
	"compile":    {"generate"},
	"explore":    {"grammar", "production", "token"},
	"generate":   {"generate"},
	"inspect":    {"grammar", "production", "token"},
	"parse":      {"parse"},
	"read":       {"parse"},
	"structure":  {"production", "token"},
	"syntax":     {"production", "token"},
	"transform":  {"lower"},
	"understand": {"parse", "lower"},
	"validate":   {"verify"},
	"verify":     {"verify"},
	"what":       {"grammar"},
	"operator":   {"precedence", "associativity"},
	"operators":  {"precedence", "associativity"},
}

func relatedKinds(term string) []string {
	return queryCapabilityAliases[term]
}

func queryTerms(query string) []string {
	stop := map[string]bool{"a": true, "an": true, "and": true, "can": true, "capabilities": true, "do": true, "does": true, "gooo": true, "grammar": true, "how": true, "i": true, "is": true, "language": true, "me": true, "of": true, "on": true, "show": true, "support": true, "tell": true, "the": true, "this": true, "to": true, "what": true, "with": true, "would": true}
	fields := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(query)), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	seen := make(map[string]struct{})
	terms := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" || stop[field] {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		terms = append(terms, field)
	}
	sort.Strings(terms)
	return terms
}

func isOverviewQuery(query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	for _, phrase := range []string{"what can", "what can i do", "what can it do", "what is possible", "what is supported", "show examples", "what does", "무엇을", "무엇인지", "무엇을 할 수", "할 수 있는", "어떤 것", "가능한", "가능"} {
		if strings.Contains(query, phrase) {
			return true
		}
	}
	return false
}
