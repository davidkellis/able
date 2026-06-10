# Parser residual-span caller-pair gate

## Decision

Keep one final general source-parser cleanup. The parser no longer writes the
same span twice in the remaining direct-identifier, pipe-wrapper, identifier
pattern, and literal-pattern paths. The direct-identifier expression route
also now uses one grouped Go switch case instead of falling through to the
recursive identifier fallback.

The changes preserve the existing span. They add no cache, allocation,
Tree-sitter call, parser state, or retained node identity. No grammar, AST
shape, diagnostic contract, language syntax, benchmark source, canonical
standard library, compiler lowering, VM/runtime path, nominal-type rule, or
WASM code changed.

## Temporary caller-pair audit

Environment-gated, per-parse instrumentation associated each span request with
its AST object identity, Tree-sitter `Node.Id()`, and first/repeat Go call
sites. Each application ran once because the counts are structural for a fixed
source graph. The instrumentation and its stack-walk cost were removed before
building either timing binary.

| Application | Modules | Requests | Duplicate requests | Same AST/node | Different AST |
| --- | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 13 | 101,901 | 10,059 | 9,659 | 400 |
| Dependency Plan | 9 | 49,080 | 4,696 | 4,479 | 217 |
| Future Pipeline | 2 | 28,330 | 2,584 | 2,462 | 122 |
| Configuration Validation/Extraction | 27 | 255,375 | 27,387 | 26,310 | 1,077 |

Five exact adjacent caller pairs account for every same-AST/same-node repeat
in all four applications:

| First write | Redundant repeat | Document | Dependency | Future | Config | Total |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| identifier parser | expression identifier fallback | 2,852 | 1,368 | 677 | 8,667 | 13,564 |
| pipe parser | `pipe_expression` dispatcher | 2,716 | 1,196 | 668 | 7,487 | 12,067 |
| pipe parser | low-precedence pipe dispatcher | 2,716 | 1,196 | 668 | 7,487 | 12,067 |
| identifier parser | identifier-pattern dispatcher | 1,311 | 714 | 447 | 2,538 | 5,010 |
| literal-pattern parser | literal-pattern dispatcher | 64 | 5 | 2 | 131 | 202 |

This clears the advance gate by a wide margin: each exact pair occurs in all
four unlike applications, and the four material paths are common grammar and
dispatcher machinery rather than benchmark-specific constructs.

## Candidate mechanism

`parseExpressionInternal` had consecutive `case "identifier":` and
`case "keyword_identifier":` clauses. Go does not implicitly fall through an
empty switch clause, so ordinary identifiers missed the direct route and were
rediscovered by the recursive `findIdentifier` fallback. Grouping the labels
routes both identifier kinds through `parseIdentifier` directly.

The pipe parser already annotates its result from the same pipe syntax node.
Likewise, `parseIdentifier` and `parseLiteralPattern` already annotate the
objects returned to their pattern dispatchers. Those dispatchers now return
the annotated values directly. Intentional annotations from wider wrapper
nodes remain unchanged.

## Frozen-binary timing gate

Separate clean baseline and candidate test binaries ran twenty independent,
order-balanced process pairs per application. Means include every sample.
`program_load` is the application-facing metric and `ast_mapping` is the
directly affected mechanism metric.

| Application | Baseline load | Candidate load | Load change | Baseline mapping | Candidate mapping | Mapping change | Candidate load wins |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 121.472 ms | 119.038 ms | -2.00% | 49.020 ms | 47.153 ms | -3.81% | 12/20 |
| Dependency Plan | 77.344 ms | 77.123 ms | -0.29% | 27.896 ms | 26.922 ms | -3.49% | 8/20 |
| Future Pipeline | 49.323 ms | 48.389 ms | -1.89% | 15.265 ms | 14.540 ms | -4.75% | 11/20 |
| Configuration Validation/Extraction | 294.660 ms | 280.741 ms | -4.72% | 127.949 ms | 116.798 ms | -8.71% | 14/20 |

The directly affected phase improves in all four applications. Total loading
improves in three and is neutral in Dependency Plan, where unrelated native
parsing moved against the candidate. This is a broad positive gate on a busy
workstation, not a single-sample selection.

## Normal verifier controls

The ordinary external-comparison harness ran five independent candidate
bytecode processes per application. All 20 processes completed and every
stdout capture passed its external verifier.

| Application | Mean complete process | Verification |
| --- | ---: | ---: |
| Document Audit | 0.254 s | 5/5 |
| Dependency Plan | 0.462 s | 5/5 |
| Future Pipeline | 0.474 s | 5/5 |
| Configuration Validation/Extraction | 1.306 s | 5/5 |

## Verification

- `go test ./pkg/parser ./pkg/ast ./pkg/driver -count=1 -timeout 55s`
- Focused identifier, typed-pattern, destructuring, match, and loop-pattern
  interpreter tests
- Compiler diagnostic-origin registration and execution-harness tests
- Two representative compiler diagnostic-parity fixtures, serialized
- Four applications x twenty clean order-balanced timing pairs
- Four applications x five normal verifier-backed bytecode processes
- `git diff --check`

One broadened interpreter parity control still exposes the existing
`alias_reexport_impl_ambiguity` expected-error mismatch: the runtime includes
the qualified interface name while the fixture expects the older unqualified
text. The full compiler diagnostic-parity umbrella also exceeds the bounded
55-second package limit when it launches many compiled fixtures concurrently
on this workstation. Neither failure depends on parser spans; bounded affected
controls pass.

## Next selection

Close parser span micro-optimization work. The caller audit exhausted every
same-AST/same-node residual repeat shared by these four source graphs; pursuing
the remaining different-object node reuse would require semantic analysis or
a cache for a much smaller ceiling.

Next, first reconcile the qualified-name fixture expectation so the broad
parity control is green, then refresh bounded compiler and bytecode profiles
for the current portable suite and choose the next larger shared wall. This
entails a small correctness-only fixture/runtime provenance check, followed by
fresh one-process profiles and a candidate only when the same generic owner is
material across multiple unlike programs. The reason is to return effort to
the compiler/VM performance targets while preserving a trustworthy test gate;
do not reopen parser span work, add benchmark or named-nominal fast paths, or
begin WASM work.

This follow-up is complete. The qualified diagnostic fixture was reconciled,
and fresh bounded compiled/bytecode profiles admitted no new shared leaf; see
`2026-07-20-qualified-impl-diagnostic-and-profile-refresh.md`.
