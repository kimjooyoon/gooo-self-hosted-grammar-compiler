# Bootstrap contract v1

## Source authority

`meta/gooo-grammar.gooo` is authoritative for the grammar language. It must
declare `token`, `production`, `precedence`, `associativity`, `stage`,
`origin`, and `ambiguity policy`. Go is an implementation of parse, lower,
generate, execute, and verify operations; it is not a second semantic source.

## Generation boundary

1. The independent stage0 seed parser reads the `.gooo` source.
2. Stage0 lowering creates IR0 and emits the stage1 parser artifact.
3. The generated stage1 parser reads the same grammar source and lowers it to
   IR1.
4. IR1 emits the stage2 parser artifact, which is run against the grammar
   source and a sample `.gooo` program in CI.
5. Stage0, stage1, and stage2 record syntax-tree, semantic-IR, and terminal
   reason digests. Drift is `REFUTED`.

The generated parser may share the data model and digest definitions, but its
parser implementation does not call stage0. This keeps the seed trust surface
small and makes a false self-parse observable.

## Decision lattice

`REFUTED > UNKNOWN > CLOSED` is the only reduction order. Forbidden stage
origins and conflicting precedence are refutations. Unresolved grammar
ambiguity is unknown because a finite fixture cannot prove the undecidable
general property. Unknown evidence requires all six fields:

```text
stage, step, reason, unknown_class, next_operation, blocked_by
```

Byte replay additionally requires equal source bytes, generated stage1 and
stage2 bytes, and terminal-reason digest. A proposed improvement is valid only
when scenario, source digest, contract digest, toolchain identity, and runner
digest are all exactly equal before and after.

## Inventory scope

Source metrics count regular `.go` and `.gooo` files below the repository root
after excluding `.git`, `.ci`, `generated/ci`, `vendor`, `cache`, `.cache`,
`toolchain`, `.toolchain`, and `node_modules`. The root `README.md` remains a
separate explicit exclusion. Generated files are counted only when they remain
inside the source scope; caller-owned generated output is not source evidence.
