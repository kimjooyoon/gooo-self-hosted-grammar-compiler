# Grammar capability query

`gooo-grammar-capability` reads the authoritative `.gooo` grammar and exposes its declared token, production, effect, type, precedence, and associativity surface without executing a generated parser.

Example:

    go run ./cmd/gooo-grammar-capability --grammar meta/gooo-grammar.gooo --query "What can this grammar do with precedence?"

The response includes source and grammar IR digests, matched capabilities, unresolved terms, a read-only next question, and a report digest. `UNKNOWN` and `DEFERRED` preserve the earliest missing grammar or query boundary. The command does not mutate the grammar, generate code, execute a program, or grant authorization.
