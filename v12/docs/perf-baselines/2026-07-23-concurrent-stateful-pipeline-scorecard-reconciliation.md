# Concurrent Stateful Pipeline scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_stateful_pipeline` in compiled and bytecode mode, refresh
the current scorecard, performance and interaction frontiers, architecture
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
| compiled | 0.8140 s | Go 0.0044 s | 185.000× | 187.849× |
| bytecode | 0.3500 s | Python 0.0616 s | 5.682× | 5.712× |
| bytecode | 0.3500 s | Ruby 0.0492 s | 7.114× | 7.110× |

Both selected and pooled results are unambiguous target misses. The two
cohorts moved by at most 3.33% in any lane, so no workstation outlier
selection was necessary.

## Current product frontier

The checked selection contains 103 rows: 55 compiled and 48 bytecode. The full
status scorecard contains 110 rows across 55 applications. The performance
frontier has eight snapshot meets, 95 misses, five established guards, zero
actionable groups, and `176.92063157894736` seconds of summed target excess.

Concurrent Stateful Pipeline joins the closed compiled-concurrency and
bytecode-concurrency groups. Compiled profiles reproduce the rejected
`bridge.currentGID`/`runtime.Stack` wall beneath the three interface-selected
stages and their captured reducer callbacks. Bytecode profiles record 57,363
inline-call hits and zero misses plus 45,061 resolved-member inline hits and
three fallbacks, then split into already-reviewed lock/cache, integer,
arithmetic, nominal-field, frame/return, and dispatch work.

The weighted interaction frontier now has minimum depth five, no depth-zero or
depth-one triples, and all 165 triples improved over the staged baseline. Two
interactions tie at minimum depth:

- concurrency × functions/closures × interface dispatch, adjacent to
  `45.546000` seconds of target excess;
- functions/closures × inherent methods × interface dispatch, adjacent to
  `51.164526` seconds of target excess.

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
- repository-wide Go and v12 test handoff;
- JSON, source-identity, source-line, whitespace, and diff checks.

## Next recommendation

Add one materially different sixth portable application that covers both
minimum-depth interactions with independent spawned state machines.

Why: both shallowest triples are now at depth five, and one well-shaped
application can raise both without adding another worker pool or ordered
channel-pipeline variant. Their adjacent applications represent `45.546000`
and `51.164526` seconds of current target excess, so this is the highest-value
remaining semantic breadth gap.

What it entails: build source-equivalent deterministic Able, Go, Python, and
Ruby implementations where independently spawned state machines use nominal
state methods, interface-selected handlers, and first-class callbacks. Avoid
files, a shared worker queue, and this tranche's serial stage topology. Add an
exact schedule-independent verifier, establish six-lane parity, take two
five-process cohorts per lane, and profile only after correctness is fixed.
Admit runtime or compiler work only for an exact generic owner repeated across
at least three unlike applications. Update canonical `able-stdlib` only for a
reusable API or correctness defect, and do not begin WASM work.
