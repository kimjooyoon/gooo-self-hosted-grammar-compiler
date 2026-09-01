package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/generated"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/compiler"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/conformance"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/evidence"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/lowering"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/model"
	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/stage0"
)

func main() {
	if len(os.Args) < 2 {
		fatal(errors.New("command is required: bootstrap, generate, sample, conformance, or ci-evidence"))
	}
	var err error
	switch os.Args[1] {
	case "bootstrap":
		err = bootstrap(os.Args[2:])
	case "generate":
		err = generate(os.Args[2:])
	case "sample":
		err = sample(os.Args[2:])
	case "conformance":
		err = conformanceCommand(os.Args[2:])
	case "ci-evidence":
		err = evidence.Command(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fatal(err)
	}
}

func bootstrap(args []string) error {
	flags := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	grammarPath := flags.String("grammar", "", "authoritative .gooo grammar")
	outDir := flags.String("out-dir", "", "absolute caller-owned output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *grammarPath == "" || *outDir == "" || !filepath.IsAbs(*outDir) {
		return errors.New("bootstrap requires absolute --out-dir and --grammar")
	}
	raw, err := os.ReadFile(*grammarPath)
	if err != nil {
		return err
	}
	result, err := compiler.Bootstrap(raw)
	if err != nil {
		return err
	}
	if err := compiler.WriteArtifact(filepath.Join(*outDir, "stage1", "stage1_parser.go"), compiler.RenderParser(result.SemanticIR, 1, "stage1generated", "ParseStage1")); err != nil {
		return err
	}
	if err := compiler.WriteArtifact(filepath.Join(*outDir, "stage2", "stage2_parser.go"), compiler.RenderParser(result.SemanticIR, 2, "stage2generated", "ParseStage2")); err != nil {
		return err
	}
	return writeJSON(filepath.Join(*outDir, "bootstrap.json"), result)
}

func generate(args []string) error {
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	grammarPath := flags.String("grammar", "", "authoritative .gooo grammar")
	out := flags.String("out", "", "absolute caller-owned Go output path")
	stage := flags.Int("stage", 2, "generated parser generation")
	packageName := flags.String("package", "generatedci", "Go package name for the artifact")
	functionName := flags.String("function", "Parse", "exported parser function name")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *grammarPath == "" || *out == "" || !filepath.IsAbs(*out) {
		return errors.New("generate requires absolute --out and --grammar")
	}
	raw, err := os.ReadFile(*grammarPath)
	if err != nil {
		return err
	}
	tree, err := stage0.Parse(raw)
	if err != nil {
		return err
	}
	ir, diagnostics, err := lowering.Lower(tree)
	if err != nil {
		return err
	}
	if len(diagnostics) > 0 {
		return fmt.Errorf("grammar has diagnostics: %s", diagnostics[0].Class)
	}
	return compiler.WriteArtifact(*out, compiler.RenderParser(ir, *stage, *packageName, *functionName))
}

func sample(args []string) error {
	flags := flag.NewFlagSet("sample", flag.ContinueOnError)
	input := flags.String("input", "", "sample .gooo program")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return errors.New("sample requires --input")
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		return err
	}
	tree, err := generated.ParseStage2(raw)
	if err != nil {
		return err
	}
	return writeJSON("", struct {
		Parser       string `json:"parser"`
		SyntaxDigest string `json:"syntax_tree_digest"`
	}{Parser: generated.Stage2ParserIdentity, SyntaxDigest: model.SyntaxTreeDigest(tree)})
}

func conformanceCommand(args []string) error {
	flags := flag.NewFlagSet("conformance", flag.ContinueOnError)
	casesRoot := flags.String("cases", "", "fixed case root")
	referenceRoot := flags.String("reference-root", "", "root for case references")
	out := flags.String("out", "", "optional JSON output path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *casesRoot == "" {
		return errors.New("conformance requires --cases")
	}
	if *referenceRoot == "" {
		*referenceRoot = *casesRoot
	}
	report, err := conformance.RunDirectory(*casesRoot, *referenceRoot)
	if err != nil {
		return err
	}
	if *out == "" {
		if err := writeJSON("", report); err != nil {
			return err
		}
		if !report.AllPass {
			return errors.New("fixed conformance failed")
		}
		return nil
	}
	if err := writeJSON(*out, report); err != nil {
		return err
	}
	if !report.AllPass {
		return errors.New("fixed conformance failed")
	}
	return nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if path == "" {
		_, err = os.Stdout.Write(raw)
		return err
	}
	if !filepath.IsAbs(path) {
		return errors.New("JSON output path must be absolute")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
