# Concurrent Signal Dispatch scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_signal_dispatch` in compiled and bytecode mode, regenerate
the selected scorecard, performance frontier, interaction frontier,
architecture decisions, and closure ledger, and retain no engine, stdlib,
language, dependency, or WASM implementation change.

The portable-coverage integrity rule requires a reviewed compiled row for every
application. Both rows therefore belong to this application tranche rather
than remaining as unpromoted measurements.

## Promoted evidence

The chronologically first ordinary five-process cohort was promoted. The
independent second cohort remains in the application gate and pooled arithmetic
means. Every promoted and pooled process verified.

| Mode | Able | Reference | Snapshot ratio | Pooled ten-process ratio |
| --- | ---: | ---: | ---: | ---: |
| compiled | 0.2700 s | Go 0.0052 s | 51.923× | 53.905× |
| bytecode | 1.6140 s | Python 0.0612 s | 26.373× | 26.403× |
| bytecode | 1.6140 s | Ruby 0.0922 s | 17.505× | 19.385× |

The first cohort is selected chronologically, not by result. Both the
snapshot and pooled evidence are unambiguous misses. Across all lanes, the
two cohort means differ by 1.90%-14.05%, below the 15% workstation volatility
guard.

## Current product frontier

The checked selection contains 97 rows: 52 compiled and 45 bytecode. The full
status scorecard contains 104 rows across 52 applications. The current
performance frontier has eight snapshot meets, 89 misses, five established
guards, zero actionable groups, and `164.12684210526317` seconds of summed
target excess.

Concurrent Signal Dispatch joins the closed compiled-concurrency and
bytecode-concurrency groups. Its compiled profiles reproduce
`bridge.currentGID`/`runtime.Stack` at 93.75% cumulative beneath ordinary
interface dispatch. Its bytecode profiles reproduce Array, raw-integer,
binary, call, member, and type-matching families without a new exact removable
child.

The weighted interaction frontier still has minimum depth three, no
depth-zero or depth-one triples, and 163 improved triples. Concurrency ×
arrays/files × interface dispatch rises from three to four portable
applications. The new highest-ranked minimum-depth triple is concurrency ×
arrays/files × functions/closures, represented by three applications and
adjacent to `34.721474` seconds of target excess.

The architecture reconciliation remains
`no-go-current-cross-engine-local-mechanism`. Bytecode owns 85.321237% of the
current deficit. The compiled-equivalent proxy now has 45 common applications:
11 meet and 34 miss. All dependent architecture artifacts were regenerated in
dependency order; this is evidence maintenance only and does not reopen
foreign backend or WASM work.

## Closure reconciliation

Only `compiled-concurrency` and `bytecode-concurrency` changed membership and
evidence. Both closures were advanced after review. The checked ledger
contains 21 current closures and zero invalidations; no rejected candidate was
reopened.

## Verification

- 52-application catalog and feature/operation interaction checks;
- 97-row selection and 104-row current scorecard checks;
- exact scorecard evidence and performance-frontier checks;
- two complete five-process cohorts per runtime lane, 50/50 verified;
- three clean compiled and three warmed bytecode profiles;
- complete architecture-budget and semantic-ABI dependency gate;
- 21 current closures and zero invalidations;
- repository-wide handoff: all noncompiler packages, all 31 bounded compiler
  batches, and the complete bytecode fixture pass;
- JSON, source-line, whitespace, and diff checks.

## Next recommendation

Complete `portable-concurrent-callable-data-application-frontier`.

Why: the interface/data gap is now strengthened and every current local or
backend mechanism remains closed. Concurrency × arrays/files ×
functions/closures is the highest-ranked remaining minimum-depth interaction.
A first-class-callable workload is more likely to expose a reusable
call/frame/closure owner than another micro-candidate chosen from cumulative
VM parents.

What it entails: add one deterministic file-driven application whose workers
transform numeric or structured Array data through first-class functions or
closures; provide source-equivalent Able, Go, Python, and Ruby implementations
plus an exact schedule-independent verifier; take two five-process cohorts per
lane; and profile only an exact owner already present in two unlike
applications. Admit a change only if it clears broad application guards.
Update canonical `able-stdlib` only for a reusable API or correctness defect.
Do not begin WASM work.
