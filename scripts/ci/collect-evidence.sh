#!/usr/bin/env bash
set -euo pipefail

# This script is CI-only. The development contract records all local
# compile/build/test/vet/conformance/integration executions as zero.
go run ./cmd/gooo-grammar-compiler ci-evidence \
  --grammar "$PWD/meta/gooo-grammar.gooo" \
  --cases "$PWD/fixtures/cases" \
  --output "$PWD/.ci/evidence.json"
