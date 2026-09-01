# Parser bootstrapping and ambiguity notes

This implementation treats stage0 as a trust anchor rather than pretending
that self-application proves the seed correct. The following primary or
official sources informed the boundary:

| Source | Evidence used here |
| --- | --- |
| Parr and Fisher, *LL(*): The Foundation of the ANTLR Parser Generator*, PLDI 2011, [official PDF](https://www.antlr.org/papers/LL-star-PLDI11.pdf) | Production order can choose among viable alternatives; ambiguity policy is therefore observable compiler semantics. |
| Tree-sitter, [Grammar DSL](https://tree-sitter.github.io/tree-sitter/creating-parsers/2-the-grammar-dsl.html) and [Writing the Grammar](https://tree-sitter.github.io/tree-sitter/creating-parsers/3-writing-the-grammar.html) | Precedence and associativity resolve LR conflicts; intentional conflicts can preserve multiple parses through GLR. |
| Štěpán and Meduna, *Parglare: A LR/GLR parser for Python*, 2021, [publisher record](https://doi.org/10.1016/j.scico.2021.102734) | Grammar bootstrapping is practical, but general ambiguity is undecidable; a bounded corpus can only support `UNKNOWN`, not universal closure. |
| Sjölund, Fritzson, and Pop, *Bootstrapping a Compiler for an Equation-Based Object-Oriented Language*, [paper record](https://www.mic-journal.no/ABS/MIC-2014-1-1.asp/) | Successive self-compiling generations motivate the stage0 → stage1 → stage2 comparison. |

## Trust consequence

The seed parser is a hand-written implementation in Go and does not import the
generated parser. Stage1 and stage2 parse the same source with generated Go
artifacts. The verifier compares syntax-tree, semantic-IR, and terminal-reason
digests per generation. A generated parser that cannot parse its own source is
not silently accepted: the evidence is incomplete or refuted according to the
decision lattice.

The `ambiguous-grammar` fixture deliberately produces `UNKNOWN`. A duplicate
production is a concrete witness of more than one derivation, while absence of
such a witness is not a proof that no longer input is ambiguous. Conflicting
precedence and forbidden stage/origin declarations are finite, explicit
contract violations and therefore produce `REFUTED`.
