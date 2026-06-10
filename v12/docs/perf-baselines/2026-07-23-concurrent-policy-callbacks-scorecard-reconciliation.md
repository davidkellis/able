# Concurrent Policy Callbacks scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_policy_callbacks` in compiled and bytecode mode, refresh
the current scorecard, performance and interaction frontiers, architecture
dependency decisions, and concurrency closure ledger, and retain no compiler,
generated-runtime, bytecode VM, tree-walker, canonical-stdlib, language,
dependency, or WASM implementation change.

The application is retained as broad portable evidence. Profiling found no
new exact generic mechanism that satisfies the three-unlike-application
admission rule.

## Promoted evidence

The chronologically first ordinary five-process cohort is the selected
scorecard snapshot. The second cohort and pooled arithmetic means remain in
the application gate. All 50 timed processes verified.

| Mode | Selected Able | Selected reference | Selected ratio | Pooled ten-process ratio |
| --- | ---: | ---: | ---: | ---: |
| compiled | 0.5420 s | Go 0.0049 s | 110.612× | 106.448× |
| bytecode | 0.3820 s | Python 0.0768 s | 4.974× | 5.926× |
| bytecode | 0.3820 s | Ruby 0.0574 s | 6.655× | 7.356× |

The Python reference cohorts were volatile, so the pooled result keeps every
sample. Both selected and pooled results are unambiguous target misses.

## Current product frontier

The checked selection contains 101 rows: 54 compiled and 47 bytecode. The full
status scorecard contains 108 rows across 54 applications. The performance
frontier has eight snapshot meets, 93 misses, five established guards, zero
actionable groups, and `175.81305263157896` seconds of summed target excess.
Compiled rows own `32.79778947368422` seconds and bytecode rows own
`143.01526315789474` seconds.

Concurrent Policy Callbacks joins the closed compiled-concurrency and
bytecode-concurrency groups. Compiled profiles reproduce the rejected
`bridge.currentGID`/`runtime.Stack` wall. Bytecode profiles record 28,939
inline-call hits and zero misses, plus 11,305 resolved member inline hits, then
split into already-reviewed integer, arithmetic, environment/cache-lock,
frame/return, and member-dispatch work.

The weighted interaction frontier now has minimum depth four, no depth-zero or
depth-one triples, and all 165 triples improved over the staged baseline. The
targeted concurrency × functions/closures × interface-dispatch interaction
rises from three to four applications. It remains the sole minimum-depth
triple and is adjacent to `44.438421` seconds of target excess.

The deterministic architecture dependency chain was regenerated in order.
Its decisions remain unchanged: no current local cross-engine mechanism, no
native-tier prototype, no portable foreign backend, no shared-runtime
semantic-ABI migration, and no closed-region production cutover is admitted.

## Closure reconciliation

Only `compiled-concurrency` and `bytecode-concurrency` changed membership and
evidence. Both were reviewed and advanced. The ledger contains 21 current
closures and zero invalidations. The canonical stdlib source fingerprint is
unchanged; no stdlib API or correctness defect was found.

## Verification

- exact output parity across tree-walker, bytecode, compiled Able, Go, Python,
  and Ruby;
- two complete five-process cohorts per runtime lane, 50/50 verified;
- three clean compiled and three clean warmed bytecode profiles;
- catalog, selection, coverage, operation-depth, matrix, triple, scorecard,
  frontier, closure-ledger, and architecture dependency checks;
- repository-wide Go and v12 test handoff;
- JSON, source-identity, source-line, whitespace, and diff checks.

## Next recommendation

Continue the portable concurrency × functions/closures × interface-dispatch
frontier with one materially different fifth application.

Why: this is still the sole minimum-depth interaction, now at depth four, and
its adjacent applications account for `44.438421` seconds of target excess.
The new worker/policy application closed one important shape but reproduced
known owners. A fifth application is worthwhile only if it changes the
semantic shape enough to test a different general boundary.

What it entails: build a deterministic non-file-centric application in which
concurrent stages pass first-class stateful callbacks through interface-typed
components, using different data and control structures from the policy
worker. Provide source-equivalent Able, Go, Python, and Ruby implementations
and an exact schedule-independent verifier; take two five-process cohorts per
lane; and profile only after parity. Admit runtime or compiler work only for
an exact owner repeated across at least three unlike applications. Update
canonical `able-stdlib` only for a reusable API or correctness defect, and do
not begin WASM work.
