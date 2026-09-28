# Read-only capability JSONL adapter

The grammar capability compiler exposes a small editor-facing adapter on stdin/stdout:

    go run ./cmd/gooo-grammar-capability-lsp

Each input line is JSON with an id, grammar text, and natural-language query:

    {"id":"one","grammar":"...authoritative .gooo grammar...","query":"what can gooo do with precedence"}

Each output line contains the original id, the capability report, and its derived provenance evidence. Every capability includes a declaration-backed description suitable for editor hover or completion labels. Those descriptions are derived only from the .gooo capability kind and name, and are included in the report digest.

The adapter is JSONL rather than a full transport implementation so an editor integration can own framing and process lifecycle. It is read-only: it does not parse a user program, generate code, execute effects, mutate grammar, or authorize changes.

UNKNOWN and DEFERRED responses retain the first mismatch and missing stage in both report and evidence. Related natural-language suggestions remain DEFERRED until a precise declared capability is confirmed.