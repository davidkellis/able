# Concurrency × lexical-pattern depth reconciliation

Date: 2026-07-21

## Decision

Recognize Dependency Wave Validation as an existing portable guard for
`concurrency × lexical_blocks_bindings_patterns`. Keep the coverage annotation
and the matrix generator's exact family-membership baseline support. Add no new
benchmark application and keep no compiler, bytecode-VM, runtime, canonical-
stdlib, language, or WASM performance change.

The application already executes the missing interaction deeply. Adding a
second program would duplicate a current source-equivalent, verifier-backed
contract rather than broaden the semantic corpus.

## Reconciliation

The original Dependency Wave application gate deliberately credited only its
six target control-flow interactions. That was conservative tranche
bookkeeping, not a statement that the program lacked lexical or destructuring
patterns. The v12 specification defines match-clause struct and typed patterns
through the same pattern grammar used by destructuring bindings.

Dependency Wave's concurrent hot path contains both forms:

- every worker selects `case task: WaveTask` or `case nil` after a Channel
  receive;
- every completed node calls `WaveResult.accepted()`, `value()`, and
  `checksum()`, whose `Accepted { value }` / `Recovered { value }` clauses
  destructure payload-bearing union variants;
- dependency construction calls the same payload-selecting methods before
  enqueuing later waves.

The default contract processes 4,096 nodes. Even without counting dependency
lookups, it executes 12,288 union-pattern selections while aggregating results
and 4,100 typed-or-nil Channel receive selections. That is a conservative
lower bound of 16,388 relevant pattern selections per run. The patterns are
therefore material application behavior, not dead syntax or setup-only code.

## Source and execution evidence

The canonical and sibling Able sources remain byte-identical with SHA-256
`7d200dbe3ee867e95668ae8c74a2e45893f4d369db6bf00962363348dc632a50`.
The foreign suite preserves the same graph, worker count, validation,
aggregation, and deterministic verifier contract:

```text
4096:32:4096:4096:4054:42:2053151855:383298
```

The promoted scorecard already supplies current source-exact evidence, so no
timing rerun or scorecard mutation is needed for a coverage-only
reclassification:

| Mode/reference | Able processes | Able mean | Reference mean | Ratio |
| --- | ---: | ---: | ---: | ---: |
| compiled / Go | 5 verified | 1.2440 s | 0.0042 s | 296.19x |
| bytecode / Python | 5 verified | 0.4200 s | 0.0480 s | 8.75x |
| bytecode / Ruby | 5 verified | 0.4200 s | 0.0514 s | 8.17x |

All ten Able processes and all applicable reference processes had zero
failures and timeouts. The Able source, verifier, and foreign-source hashes in
those rows match the files inspected for this reconciliation.

## Coverage-depth result

The matrix baseline removes only Dependency Wave's lexical-family membership;
it does not pretend the existing application was newly added. Across the 11
discriminating portable/mixed families and 55 pairs:

| Measure | Before | Current |
| --- | ---: | ---: |
| zero-depth pairs | 0 | 0 |
| minimum pair depth | 1 | 2 |
| depth-one pairs | 1 | 0 |
| strengthened pairs | — | 8 |

`concurrency × lexical_blocks_bindings_patterns` rises from one guard to two:
Concurrent Event Routing and Dependency Wave Validation. The other seven
improvements are the lexical family's intersections with control flow,
functions, inherent methods, interfaces, Option/Result, stdlib protocols, and
nominal/union types. Machine-readable and rendered evidence:
`2026-07-21-concurrency-lexical-depth-matrix.{json,md}`.

## Performance disposition

This tranche changes metadata and reporting only. Dependency Wave's exact
compiled profile is still dominated by `bridge.currentGID`; its exact bytecode
profile still resolves to call/member, executor, return, atomic environment,
and type-match owners. Those generic families have already completed broad
cross-application gates. Reclassifying already-executed patterns supplies no
new profile leaf and invalidates none of those closures, so no implementation
candidate is admissible.

The selection manifest remains 46 compiled and 39 bytecode applications. The
promoted scoreboard remains 92 full-status rows, and the frontier remains 77
misses, 143.927 seconds above target budget, and zero actionable groups.

## Verification

- canonical/sibling Able byte identity and current verifier/source hashes;
- feature-coverage manifest validation;
- exact family-membership baseline and interaction-matrix unit tests;
- current selection-manifest, scoreboard, and performance-frontier replay;
- full bytecode test family;
- JSON syntax, file-size, whitespace, and temporary-artifact checks.

## Next recommendation

Audit three-way feature interactions, weighted by semantic importance and
current target excess, before adding another application or reopening an
implementation family.

Why: every pair now has at least two independent portable guards, so another
pairwise application would mostly add redundant breadth. Pairwise completeness
can still hide a missing realistic combination of three features, while the
implementation frontier currently has no evidence-backed open candidate.

What it entails: extend the report-only matrix tooling to enumerate portable
three-family intersections, rank weak combinations using current scorecard
excess and language significance, and audit existing applications first. Add
at most one bounded, source-equivalent, verifier-backed program only when a
material combination is genuinely absent. Profile implementation code only if
that work exposes one new concrete generic leaf in at least three unlike
applications. Continue to defer WASM.
