# Concurrent State Machines scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_state_machines` in compiled and bytecode mode, refresh the
current scorecard, performance and interaction frontiers, architecture
dependency decisions, and concurrency closure ledger, and retain no compiler,
generated-runtime, bytecode VM, tree-walker, canonical-stdlib, language,
dependency, or WASM implementation change.

The application is retained as broad portable evidence. Profiling reproduced
closed generic owners and found no exact new mechanism that satisfies the
three-unlike-application admission rule.

## Promoted evidence

The chronologically first ordinary five-process cohort is the selected
scorecard snapshot. The second cohort and pooled arithmetic means remain in
the application gate. All 50 timed processes verified.

| Mode | Selected Able | Selected reference | Selected ratio | Pooled ten-process ratio |
| --- | ---: | ---: | ---: | ---: |
| compiled | 0.2640 s | Go 0.0037 s | 71.351× | 66.405× |
| bytecode | 0.3020 s | Python 0.0602 s | 5.017× | 4.923× |
| bytecode | 0.3020 s | Ruby 0.0548 s | 5.511× | 5.430× |

Both selected and pooled results are unambiguous target misses. The Go
reference cohorts moved by 10.10%, so the performance decision uses all ten
retained samples rather than selecting the more favorable cohort.

## Current product frontier

The checked selection contains 105 rows: 56 compiled and 49 bytecode. The full
status scorecard contains 112 rows across 56 applications. The performance
frontier has eight snapshot meets, 97 misses, five established guards, zero
actionable groups, and `177.42505263157895` seconds of summed target excess.

Concurrent State Machines joins the closed compiled-concurrency and
bytecode-concurrency groups. Compiled profiles reproduce the rejected
`bridge.currentGID`/`runtime.Stack` wall beneath four independently spawned
machines, two interface-selected implementations, and four captured
adjusters. Bytecode profiles record 24,586 inline-call hits and zero misses
plus 16,384 resolved-member inline hits, then split into already-reviewed
lock/cache, integer, arithmetic, nominal-field, frame/return, and dispatch
work.

The weighted interaction frontier now has minimum depth six and no depth-zero
or depth-one triples. Two interactions tie at minimum depth:

- concurrency × functions/closures × interface dispatch, adjacent to
  `46.050421` seconds of target excess;
- functions/closures × inherent methods × interface dispatch, adjacent to
  `51.668947` seconds of target excess.

The deterministic architecture dependency chain was regenerated in order.
Its decisions remain unchanged: no current local cross-engine mechanism,
native-tier prototype, portable foreign backend, shared-runtime semantic-ABI
migration, or closed-region production cutover is admitted.

## Closure reconciliation

Only `compiled-concurrency` and `bytecode-concurrency` changed membership and
evidence. Both were reviewed and advanced. The ledger contains 21 current
closures and zero invalidations. The canonical stdlib source fingerprint is
unchanged; no reusable stdlib API or correctness defect was found.

## Verification

- exact output parity across tree-walker, bytecode, compiled Able, Go, Python,
  and Ruby;
- two complete five-process cohorts per runtime lane, 50/50 verified;
- three clean compiled and three clean warmed bytecode profiles;
- catalog, selection, coverage, operation-depth, matrix, triple, scorecard,
  frontier, closure-ledger, and architecture dependency checks;
- JSON, source-identity, source-line, whitespace, and diff checks.

## Next recommendation

Add one materially different seventh portable application that covers both
minimum-depth interactions, using independently spawned graph traversals with
interface-selected visitors and captured scoring callbacks.

Why: both shallowest triples are now at depth six. The inherent-method triple
has the greater adjacent target excess at `51.668947` seconds, while the
concurrency triple has the greater semantic weight. A concurrent graph
traversal can raise both without repeating worker queues, ordered channel
pipelines, signal dispatch, callback policy batches, transform chains, or
state machines.

What it entails: build source-equivalent deterministic Able, Go, Python, and
Ruby implementations where each Future traverses an independent nominal
graph, calls inherent node/path methods, selects a visitor through a
user-defined interface, and passes a captured scorer as data. Use a
schedule-independent exact verifier, establish six-lane parity, take two
five-process cohorts per lane, and profile only after correctness is fixed.
Admit runtime or compiler work only for an exact generic owner repeated across
at least three unlike applications. Update canonical `able-stdlib` only for a
reusable API or correctness defect, and do not begin WASM work.
