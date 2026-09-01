# gooo-self-hosted-grammar-compiler

`gooo-self-hosted-grammar-compiler` is a small, auditable bootstrap slice for
Gooo. The authority is the `.gooo` grammar in
[`meta/gooo-grammar.gooo`](meta/gooo-grammar.gooo). Go code only parses,
lowers, generates, executes, and verifies that authority.

The bootstrap chain is:

```text
.gooo authority
    │
    ▼
independent stage0 seed parser ──lower──▶ IR0 ──generate──▶ stage1 Go parser
                                                               │
                                                               ▼
                                                     parse/lower source again
                                                               │
                                                               ▼
                                                          IR1 ──generate──▶ stage2 Go parser
```

Stage0 lives in `internal/stage0` and shares only the model/digest data types
with generated code. `generated/stage1_parser.go` and
`generated/stage2_parser.go` are real Go parser artifacts. The CI workflow
also generates fresh stage1/stage2 sources, builds them, and runs the fresh
stage2 parser against [`examples/sample.gooo`](examples/sample.gooo).

## Authority and comparison

The `.gooo` language declares tokens, productions, precedence,
associativity, stage/origin, and ambiguity policy. Lowering produces a
canonical semantic IR. Every generation records a syntax-tree digest, a
semantic-IR digest, and a terminal-reason digest.

Decision precedence is fixed: `REFUTED > UNKNOWN > CLOSED`. An `UNKNOWN`
record always carries `stage`, `step`, `reason`, `unknown_class`,
`next_operation`, and `blocked_by`. Improvement is accepted only for an exact
scenario/source/contract/toolchain/runner digest before/after pair; aggregate
scores and estimated percentages are not used.

The fixed conformance denominator has seven cases:

| Case | Expected |
| --- | --- |
| base self-parse | `CLOSED` |
| additive syntax extension | `CLOSED` |
| precedence-preserving change | `CLOSED` |
| ambiguous grammar | `UNKNOWN` |
| conflicting precedence | `REFUTED` |
| forbidden stage escape | `REFUTED` |
| byte-identical replay | `CLOSED` |

## CI-only verification

The development contract intentionally records local compile, build, test,
vet, conformance, and integration executions as zero. GitHub Actions is the
validation authority. It runs Go 1.27, formatting, vet, build, tests, fresh
generated-artifact build/run, fixed conformance, and the exact integer evidence
writer. The evidence contract is
[`contracts/ci-evidence-v1.schema.json`](contracts/ci-evidence-v1.schema.json):
all required metrics are integers, including physical lines, descendant
directories, regular files, generated files/bytes, phase wall time and peak
RSS, and test total/selected/executed/reused/failed/unknown counts. The root
README is explicitly excluded from the `.go`/`.gooo` inventory.
The source-repository inventory also skips `.git`, `.ci`, `generated/ci`,
`vendor`, cache directories, toolchain directories, and `node_modules`; these
are caller-owned outputs or non-source internals.

Useful CI commands are:

```text
go run ./cmd/gooo-grammar-compiler bootstrap --grammar meta/gooo-grammar.gooo --out-dir /absolute/caller-owned/output
go run ./cmd/gooo-grammar-compiler sample --input examples/sample.gooo
go run ./cmd/gooo-grammar-compiler conformance --cases fixtures/cases --reference-root .
```

No command in this repository commits, pushes, merges, overwrites public tags,
or releases. Release verification is documented separately and uses only the
public release API, never an administration-settings endpoint.

The CI receipt records `input_repository_writes=0`, caller-owned temporary or
output paths only, and automatic commit/push/merge/release authority as exact
zeroes.

## Research basis

The boundary is intentionally conservative about parser trust and ambiguity:

- Parr and Fisher, *LL(*): The Foundation of the ANTLR Parser Generator*,
  PLDI 2011, [official paper](https://www.antlr.org/papers/LL-star-PLDI11.pdf).
  It describes production-order disambiguation and why a parser generator's
  ambiguity policy must be explicit.
- The Tree-sitter team documents parse precedence, associativity, lexical
  precedence, and intentional GLR conflicts in the
  [official grammar DSL guide](https://tree-sitter.github.io/tree-sitter/creating-parsers/2-the-grammar-dsl.html)
  and [grammar-writing guide](https://tree-sitter.github.io/tree-sitter/creating-parsers/3-writing-the-grammar.html).
- V. Štěpán and M. Meduna, *Parglare: A LR/GLR parser for Python*,
  Software: Practice and Experience 51 (2021),
  [publisher record](https://doi.org/10.1016/j.scico.2021.102734). It reports
  grammar bootstrapping and states the practical limit: general grammar
  ambiguity is undecidable, so bounded evidence cannot prove all inputs
  unambiguous.
- Sjölund, Fritzson, and Pop, *Bootstrapping a Compiler for an
  Equation-Based Object-Oriented Language*, Modelica 2014,
  [paper record](https://www.mic-journal.no/ABS/MIC-2014-1-1.asp/), is the
  self-hosting case study behind treating stage0 as a separately trusted seed
  and checking successive compiler generations.
