# Parser redundant-span dispatcher gate

## Decision

Keep one general source-parser improvement. Expression and statement
dispatchers no longer annotate an AST object a second time when the delegated
parser has already annotated that same object from the same Tree-sitter node.
Intentional re-annotation from a wider wrapper node remains unchanged.

This removes Tree-sitter position calls rather than caching their results. It
therefore needs no parse-scoped map, adds no per-node allocation or lookup, and
cannot retain a node ID after its syntax tree closes.

No grammar, AST shape, diagnostic contract, language syntax, compiler
lowering, VM/runtime path, benchmark, canonical standard library, nominal-type
rule, or WASM code changed.

## Measurement-only duplicate census

Temporary per-parse instrumentation counted `spanFromNode` requests and unique
`Node.Id()` values. `Node.Id()` performs no C call and is unique within the
active syntax tree. Each application ran once because these structural counts
are deterministic for a fixed source graph.

| Application | Modules | Span requests | Unique nodes | Duplicate requests | Duplicate share |
| --- | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 13 | 147,283 | 92,078 | 55,205 | 37.48% |
| Dependency Plan | 9 | 69,296 | 44,509 | 24,787 | 35.77% |
| Future Await Race | 2 | 37,645 | 24,645 | 13,000 | 34.53% |
| Configuration Validation/Extraction | 27 | 379,582 | 228,581 | 151,001 | 39.78% |

A second temporary identity census associated each annotation with its AST
object and Tree-sitter node. Between 99.09% and 99.29% of duplicate requests
were the same AST object being annotated again from the same syntax node. That
evidence selected dispatcher cleanup over a speculative general cache.

All instrumentation was removed before building the frozen timing candidate.

## Candidate mechanism

Specialized expression parsers already return annotated expressions. The
generic expression dispatcher nevertheless called `annotateExpression` again
for 36 delegated routes. The statement dispatcher did the same for 11
definition/host-statement routes whose specialized parsers already annotate
their results. Those second calls produced the same final span and made two
additional C calls through `StartPosition()` and `EndPosition()`.

The candidate returns the delegated AST value directly on those routes.
Dispatcher-owned nodes and wider wrapper constructs still annotate normally.
Temporary post-candidate counters confirmed the intended reduction:

| Application | Baseline requests | Candidate requests | Reduction | Remaining duplicates |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 147,283 | 101,901 | -30.81% | 10,059 (9.87%) |
| Dependency Plan | 69,296 | 49,080 | -29.17% | 4,696 (9.57%) |
| Future Await Race | 37,645 | 27,006 | -28.26% | 2,431 (9.00%) |
| Configuration Validation/Extraction | 379,582 | 255,375 | -32.72% | 27,387 (10.72%) |

## Frozen-binary phase gate

Separate clean baseline and candidate test binaries ran twenty independent,
order-balanced process pairs per application. Means include every sample.
`program_load` is the application-facing selection metric; `ast_mapping` is
the directly affected mechanism metric.

| Application | Baseline load | Candidate load | Load change | AST mapping change | Candidate wins |
| --- | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 132.963 ms | 128.566 ms | -3.31% | -9.46% | 13/20 |
| Dependency Plan, first cohort | 70.808 ms | 74.938 ms | +5.83% | +0.20% | 6/20 |
| Future Await Race | 51.403 ms | 51.218 ms | -0.36% | -7.36% | 11/20 |
| Configuration Validation/Extraction | 291.978 ms | 278.200 ms | -4.72% | -10.77% | 14/20 |

Dependency Plan was volatile, so two independent twenty-pair cohorts used
different balanced order patterns. The second changed total loading -0.28%
and AST mapping -7.59%; the third changed total loading -3.88% and AST mapping
-10.95%. Across all 60 runs per side, total loading is neutral at 74.458 ms
baseline versus 74.740 ms candidate (+0.38%), while AST mapping improves from
28.516 ms to 26.713 ms (-6.32%). Native parsing and origin annotation—both
untouched—moved against the candidate in the pooled mean and explain the small
total difference.

Thus the directly affected phase improves in every repeated aggregate, total
loading improves in two applications, remains neutral in Future and in the
60-run Dependency control, and shows no broad regression.

## Normal verifier controls

The ordinary external-comparison harness ran five independent bytecode
processes per application. All 20 processes completed and all outputs passed
their external verifier.

| Application | Mean complete process | Verification |
| --- | ---: | ---: |
| Document Audit | 0.252 s | 5/5 |
| Dependency Plan | 0.448 s | 5/5 |
| Future Await Race | 0.124 s | 5/5 |
| Configuration Validation/Extraction | 1.244 s | 5/5 |

Machine-readable and rendered controls:

- `2026-07-20-parser-redundant-span-candidate.json`
- `2026-07-20-parser-redundant-span-candidate-scorecard.md`

## Verification

- Exact parser span, pattern/type span, error-fixture, and error-handling tests
- `go test ./pkg/parser -count=1 -timeout 60s`
- `go test ./pkg/ast ./pkg/driver -count=1 -timeout 60s`
- Focused dynamic-import and regex fixture parity tests
- Compiler diagnostic-origin, diagnostic-parity, and execution-harness tests
- Four applications x five normal bytecode processes: 20/20 successful and
  externally verified
- `git diff --check`

Every focused test completed in under 13 seconds. No audit instrumentation
remains in production code.

## Next selection

Next perform one bounded caller-pair audit of the remaining 9.00%-10.72%
duplicate span requests. The potential gain is now much smaller, but the
residual repeats in all four unlike source graphs and may expose one more
zero-overhead wrapper/dispatcher cleanup.

This entails temporary per-parse attribution of first and repeated annotation
call sites, followed by complete removal of the counters. Advance a candidate
only if the same exact redundant caller pair occurs materially in at least
three applications; otherwise close parser span micro-work and return to
larger shared compiler/VM profile walls. Do not introduce a general node-ID
cache for this residual or begin benchmark-, nominal-container-, stdlib-, or
WASM-specific work.

This follow-up is complete. The caller-pair audit found and removed the final
shared same-AST/same-node repeats without a cache; see
`2026-07-20-parser-residual-span-caller-pair-gate.md`. Parser span
micro-optimization is now closed.
