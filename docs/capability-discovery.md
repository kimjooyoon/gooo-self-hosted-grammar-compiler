# Capability discovery from .gooo declarations

Capability discovery answers a narrow, useful question without pretending to infer the whole language:

> What can this Gooo grammar do, based on declarations that are present here?

The answer is declaration-backed. It does not execute a capability, grant authority, or claim that the language is complete.

## CLI

Run a bounded discovery query against the grammar authority:

```text
go run ./cmd/gooo-grammar-capability \
  --grammar meta/gooo-grammar.gooo \
  --query "What can gooo do?"
```

A Korean query is supported as the same bounded form:

```text
go run ./cmd/gooo-grammar-capability \
  --grammar meta/gooo-grammar.gooo \
  --query "gooo 언어에서 무엇을 할 수 있는지?"
```

The report is useful as a starting point for a human or an editor:

1. Inspect the declarations that matched the query.
2. Follow the first suggested read-only action.
3. Use the declaration and evidence digests when deciding whether to write a plan.
4. Keep execution, authorization, and semantic completeness as separate later contracts.

## Result states

| State | Meaning |
| --- | --- |
| `BOUND` | The query has a direct declaration-backed capability match. |
| `RELATED` | A bounded alias or related phrase was found, but the exact declaration still needs inspection. |
| `UNKNOWN` | No declaration-backed evidence was found, or an earlier stage is unresolved. |

Related language is deliberately not promoted to a direct capability. The report preserves the first unresolved stage and emits only bounded suggestions such as inspecting declarations or reviewing a declared effect.

## Provenance

A discovery result is tied to:

- the grammar source digest,
- the normalized query,
- the matched and related capability digests,
- the first unresolved stage when present,
- the suggested-action digest,
- the evidence prefix used by the report.

This makes “what can this language do?” a reproducible inspection operation rather than an opaque model guess. The JSONL adapter exposes the same read-only contract for editor and LSP integrations.
