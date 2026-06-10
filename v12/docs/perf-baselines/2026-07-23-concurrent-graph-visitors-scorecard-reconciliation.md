# Concurrent Graph Visitors scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_graph_visitors` in compiled and bytecode mode, refresh the
current scorecard, performance and interaction frontiers, closure ledger, and
architecture dependency decisions, and retain the generic 4,096-entry bound
on the bytecode member-method map cache.

Retain no compiler, generated-runtime, tree-walker, canonical-stdlib,
language, dependency, named-container, non-primitive nominal, or WASM change.
The cache bound is keyed only by generic VM cache state. It continues to
refresh existing map entries and always updates the existing hot and
instruction-indexed direct caches.

## Promoted evidence

The chronologically first ordinary five-process cohort is the selected
scorecard snapshot. The second cohort and pooled arithmetic means remain in
the application gate. All 50 timed processes verified.

| Mode | Selected Able | Selected limiting ratio | Pooled Able | Pooled limiting ratio |
| --- | ---: | ---: | ---: | ---: |
| compiled | 0.258 s | 64.500× Go | 0.280 s | 69.295× Go |
| bytecode | 0.954 s | 17.699× Ruby | 0.957 s | 16.919× Python / 17.662× Ruby |

Both selected and pooled results are unambiguous target misses. The compiled
cohort means differ by 15.7%, so the performance conclusion uses all ten
workstation samples rather than selecting the more favorable cohort.

## Retained generic VM change

The pre-fix graph profile found a one-shot activation-environment key pattern
in the generic member-method map cache: one main call recorded 191,088 misses
and 180,729 hits while the unbounded map retained entries and repeatedly
rehashed. Bounding retained map entries at 4,096 reduced the graph's pooled
bytecode time from 1.110 seconds to 0.957 seconds, a 13.8% improvement.

Repeated candidate/control guards remained within workstation noise:
iterator collect was neutral; state machines moved -0.8%; policy callbacks
moved +0.8%; and the stateful pipeline moved +1.5%. A separate 64-entry
direct-cache experiment was removed because it increased every VM's fixed
footprint and produced mixed guard results.

Three final bytecode profiles average 755.63 ms/op, 94.09 MB/op, and 1.205
million allocations/op. Three compiled profiles continue to put
`bridge.currentGID`/`runtime.Stack` at 93.33% cumulative, preserving the
closed compiled-concurrency owner.

## Current product frontier

The checked selection contains 107 rows: 57 compiled and 50 bytecode. The full
status scorecard contains 114 rows across 57 applications. The performance
frontier has eight snapshot meets, 99 misses, five established guards, zero
actionable groups, and `178.5761052631579` seconds of summed target excess.

The weighted interaction frontier has no zero-depth or depth-one triple and
minimum depth seven. Its two shallowest interactions are:

- concurrency × functions/closures × interface dispatch, semantic weight
  nine and `47.201474` seconds of adjacent target excess;
- functions/closures × inherent methods × interface dispatch, semantic
  weight eight and `52.820000` seconds of adjacent target excess.

## Closure and architecture reconciliation

The retained VM change alters the shared bytecode-production fingerprint, so
ten closures that reference that scope were selected together with the
changed compiled-concurrency definition. The ledger correctly rejected
partial advancement because it would have masked shared scope drift. After
the application profile, iterator guard, and three unlike concurrency guards
were complete, the selected set received one coordinated baseline refresh.
The ledger now contains 21 current closures and zero invalidations.

Reconciliation also corrected the ledger executable's default frontier path:
it now names the cross-mode performance frontier required by its schema
validator, so direct checks and dependent architecture generators use the
same input without command-line overrides.

The deterministic dependency chain was regenerated in order. Its decisions
remain unchanged: no current local cross-engine mechanism, semantic-region
tier, native-tier prototype, portable foreign backend, shared-runtime
semantic-ABI migration, or closed-region production cutover is admitted.

## Verification

- six-lane exact output parity;
- two five-process cohorts per compiled, bytecode, and reference lane;
- three compiled, three pre-fix bytecode, and three final bytecode profiles;
- focused member-cache, call-member, reset, and return tests;
- catalog, selection, coverage, operation-depth, matrix, triple, scorecard,
  frontier, closure-ledger, and architecture dependency checks;
- JSON, source-identity, source-line, formatting, and whitespace checks.

## Next recommendation

Add one materially different eighth portable application that raises both
minimum-depth interactions with independently spawned fork/join tree folds.

Why: both shallowest triples remain at depth seven. A nominal tree/fold
topology exercises the same high-value language interactions without
repeating graph breadth-first search, worker queues, ordered pipelines,
signals, callback batches, transform chains, or state machines.

What it entails: build source-equivalent deterministic Able, Go, Python, and
Ruby implementations where each Future folds an independent nominal tree,
calls inherent node/fold-state methods, chooses a fold algebra through a
user-defined interface, and receives captured weight or pruning callbacks as
data. Establish six-lane parity, run two verifier-backed five-process cohorts
per lane, and profile only after correctness is fixed. Admit another
implementation change only for an exact generic owner repeated across unlike
programs. Update canonical `able-stdlib` only for a reusable API or
correctness defect, and do not begin WASM work.
