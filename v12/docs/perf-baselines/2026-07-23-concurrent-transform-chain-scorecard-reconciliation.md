# Concurrent Transform Chain scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_transform_chain` in compiled and bytecode mode, refresh
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
| compiled | 8.1760 s | Go 0.0065 s | 1,257.846× | 1,374.355× |
| bytecode | 2.8300 s | Python 0.1628 s | 17.383× | 19.533× |
| bytecode | 2.8300 s | Ruby 0.2095 s | 13.508× | 15.533× |

The Python and Ruby reference cohorts were volatile, so the pooled result
keeps every sample. Both selected and pooled results are unambiguous target
misses.

## Current product frontier

The checked selection contains 99 rows: 53 compiled and 46 bytecode. The full
status scorecard contains 106 rows across 53 applications. The performance
frontier has eight snapshot meets, 91 misses, five established guards, zero
actionable groups, and `174.95463157894739` seconds of summed target excess.

Concurrent Transform Chain joins the closed compiled-concurrency and
bytecode-concurrency groups. Compiled profiles reproduce the rejected
`bridge.currentGID`/`runtime.Stack` wall. Bytecode profiles record 454,754
inline-call hits and zero misses, then split into already-reviewed Array,
integer, environment/cache, frame/return, dispatch, and arithmetic work.

The weighted interaction frontier still has minimum depth three, no
depth-zero or depth-one triples, and 163 improved triples. The targeted
concurrency × arrays/files × functions/closures interaction rises from three
to four applications. Concurrency × functions/closures × interface dispatch
is now the sole minimum-depth triple, represented by three applications and
adjacent to `43.580000` seconds of target excess.

The deterministic architecture dependency chain was regenerated in order.
Its decisions remain unchanged: no current local cross-engine mechanism, no
native-tier prototype, no portable foreign backend, no shared-runtime
semantic-ABI migration, and no closed-region production cutover is admitted.

## Closure and tooling reconciliation

Only `compiled-concurrency` and `bytecode-concurrency` changed membership and
evidence. Both were reviewed and advanced. The ledger contains 21 current
closures and zero invalidations, and now records the current selection and
input fingerprints when a reviewed partial advance is written. This repairs
provenance only; it does not broaden which closures an advance may change or
mask production-scope drift.

## Verification

- exact output parity across tree-walker, bytecode, compiled Able, Go, Python,
  and Ruby;
- two complete five-process cohorts per runtime lane, 50/50 verified;
- three clean compiled and three warmed bytecode profiles;
- catalog, selection, coverage, operation-depth, matrix, triple, scorecard,
  frontier, closure-ledger, and architecture dependency checks;
- repository-wide Go and v12 test handoff;
- JSON, source-identity, source-line, whitespace, and diff checks.

## Next recommendation

Complete `portable-concurrent-callable-interface-application-frontier`.

Why: concurrency × functions/closures × interface dispatch is the only
remaining minimum-depth interaction, and its adjacent applications account
for `43.580000` seconds of target excess. Combining the two dynamic boundaries
inside ordinary worker code is the best evidence-led way to search for a
general call/dispatch owner without retrying closed micro-optimizations.

What it entails: add one deterministic file-driven application whose workers
process nominal records through user-defined interface implementations and
then invoke first-class strategy functions or captured callbacks; provide
source-equivalent Able, Go, Python, and Ruby implementations and an exact
schedule-independent verifier; take two five-process cohorts per lane; and
profile only after parity is established. Admit runtime or compiler work only
for an exact owner repeated across at least three unlike applications. Update
canonical `able-stdlib` only for a reusable API or correctness defect. Do not
begin WASM work.
