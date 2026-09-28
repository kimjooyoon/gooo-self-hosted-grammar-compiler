package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/capability"
)

func main() {
	flags := flag.NewFlagSet("gooo-grammar-capability", flag.ExitOnError)
	grammarPath := flags.String("grammar", "", "authoritative .gooo grammar")
	query := flags.String("query", "", "read-only natural-language capability query")
	evidenceOnly := flags.Bool("evidence-only", false, "emit only the provenance evidence projection")
	flags.Parse(os.Args[1:])
	if *grammarPath == "" {
		fatal(errors.New("--grammar is required"))
	}
	raw, err := os.ReadFile(*grammarPath)
	if err != nil {
		fatal(err)
	}
	report := capability.Discover(raw, *query)
	if err := report.Validate(); err != nil {
		fatal(err)
	}
	if *evidenceOnly {
		evidence := report.Evidence()
		if err := evidence.Validate(); err != nil {
			fatal(err)
		}
		encode(evidence)
		return
	}
	encode(report)
}

func encode(value any) {
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
