# Grammar capability evidence

The grammar capability command can emit a provenance-only projection without parsing or executing a user program:

    go run ./cmd/gooo-grammar-capability --grammar path/to/grammar.gooo --query "what can gooo do with precedence" --evidence-only

The projection binds the query digest, source digest, lowered grammar digest, capability-surface digest, report digest, and derived evidence digest. UNKNOWN and DEFERRED responses preserve first_mismatch and missing_stage instead of inventing later-stage evidence. BOUND is evidence binding for the declared grammar surface, not a claim of language completeness.

The declaration-level contract is examples/grammar-capability-evidence.gooo. It is intentionally read-only and does not authorize parsing, generation, execution, mutation, or scope expansion.
