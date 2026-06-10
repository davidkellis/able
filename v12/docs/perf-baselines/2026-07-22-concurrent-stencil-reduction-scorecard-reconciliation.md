# Concurrent Stencil Reduction scorecard reconciliation — 2026-07-22

## Decision

Promote `concurrent_stencil_reduction` in compiled and bytecode mode, regenerate
the complete scorecard, performance frontier, per-engine and cross-engine
budgets, architecture decisions, and closure ledger, and retain no engine,
stdlib, language, dependency, or WASM implementation change.

The integrity rule requires every portable coverage application to have a
reviewed compiled row. Promotion therefore belongs to the application tranche
rather than being left as a failing follow-up.

## Promoted evidence

The chronologically first ordinary five-process cohort was promoted; the
independent second cohort remains visible in the application gate and pooled
means. Every promoted process and reference process verified.

| Mode | Able | Reference | Snapshot ratio | Pooled ten-process ratio |
| --- | ---: | ---: | ---: | ---: |
| compiled | 0.2140 s | Go 0.0051 s | 41.961× | 47.718× |
| bytecode | 1.8620 s | Python 0.0963 s | 19.335× | 15.469× |
| bytecode | 1.8620 s | Ruby 0.1240 s | 15.016× | 15.999× |

The choice of the first cohort is chronological, not favorable-sample
selection. Both the snapshot and pooled evidence are unambiguous misses.

## Current product frontier

The checked selection now contains 95 rows: 51 compiled and 44 bytecode. The
complete status scorecard contains 102 rows across 51 applications. The current
performance frontier has eight snapshot meets, 87 misses, five established
guards, zero actionable groups, and `162.31273684210527` seconds of summed
target excess.

Concurrent Stencil Reduction joins the closed compiled-concurrency and
bytecode-concurrency groups. Its compiled row adds `bridge.currentGID` evidence
to the rejected fixed-context family. Its bytecode row adds Array/raw-integer
evidence without a new exact removable child.

The architecture reconciliation remains
`no-go-current-cross-engine-local-mechanism`. Bytecode owns 85.320152% of the
current deficit. The native compiled-proxy model now has 44 common applications:
11 meet and 33 still miss. All dependent architecture artifacts were
regenerated in dependency order.

The deterministic gate exposed two general scorecard-tooling omissions, both
now fixed. `concurrent_stencil_reduction` belongs to the bounded
coverage-extra refresh partition, so future complete refreshes schedule it.
Strict retained-run validation now mirrors the producer's per-mode reference
contract when a source report contains multiple modes: compiled rows require
Go, while bytecode and other interpreter rows require the requested
Python/Ruby references.

## Closure reconciliation

Only `compiled-concurrency` and `bytecode-concurrency` changed membership and
semantics. The closure ledger was rebuilt after those reviews and the complete
architecture chain was current. It contains 21 closures and zero invalidations.
No earlier rejected candidate was reopened.

## Verification

- 51-application complete catalog and all feature/operation interaction tests;
- 95-row selection and 102-row current scorecard checks;
- exact scorecard evidence and performance-frontier checks;
- mixed-mode five-sample evidence and refresh-partition regression tests;
- complete architecture-budget and semantic-ABI dependency gate;
- 21 current closures and zero invalidations;
- JSON, source-line, whitespace, and diff checks.

The repository-wide handoff runner is separately red in accumulated compiler
work: boundary-marker ownership, Sudoku static Array writes, Future
cancellation parity, and dynamic package-object lowering fail, and one fixture
audit exceeds one minute before the compiler package reaches its 30-minute
aggregate timeout. These do not invalidate the application measurements or
deterministic scorecard/architecture gates, but they take correctness priority
over the next performance application.

## Next recommendation

First complete `compiler-full-suite-contract-reconciliation`, then
`portable-concurrent-interface-data-application-frontier`.

Why: the current compiler-suite contract failures make further performance
work unsafe to admit. Once they are green, the top minimum-depth interaction is
concurrency × arrays/files × interface dispatch. Existing local engine families
remain closed, so another micro-candidate would repeat rejected evidence; an
unlike interface-driven concurrent application can either expose a genuinely
shared semantic owner or strengthen the guard against narrow changes.

What it entails: first repair and individually bound the five compiler-suite
failures recorded above, then implement one deterministic file-driven numeric
or structured data application in Able, Go, Python, and Ruby; use user-defined
interfaces in the hot worker path; verify exact schedule-independent output;
take repeated compiled/bytecode/reference cohorts; and admit a candidate only
if one exact owner reaches three unlike applications with a material target
model. No WASM, benchmark branches, named-container rules, or non-primitive
nominal compiler special cases.
