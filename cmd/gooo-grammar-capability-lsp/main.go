package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-self-hosted-grammar-compiler/internal/capability"
)

func main() {
	if err := capability.ServeCapabilityJSONL(os.Stdin, os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
