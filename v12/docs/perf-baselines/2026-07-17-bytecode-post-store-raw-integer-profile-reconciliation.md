# Bytecode post-store raw-integer profile reconciliation — 2026-07-17

## Decision

Do not retain another raw-integer optimization from this tranche. Fresh
profiles confirm that the single-pass slot-store change removed the previous
shared nested-extraction wall. The residual
`bytecodeRawIntegerValueInfo(...)` samples are now divided among arithmetic,
casts, typed-pattern matching, comparisons, and return handling. No material
concrete owner repeats across three unlike benchmark programs.

This is an evidence-backed admission decision, not an assertion that raw
integer extraction is finished. The two-program families remain useful future
leads, but changing a shared primitive helper for one of them now would violate
the broad-applicability gate. No interpreter, compiler, stdlib, fixture, or
language source changed in this tranche.

## Fresh bounded profiles

The retained interpreter was built once, then each workload ran in a separate
profiled process with the canonical external `able-stdlib`, `GOMAXPROCS=1`,
`GOGC=50`, `GOMEMLIMIT=1GiB`, CPU 0, and skipped benchmark typechecking. Small
programs used repeated iterations in the one process to obtain at least about
1.5 seconds of CPU samples; longer programs ran once. These are attribution
runs rather than candidate timing comparisons.

| Workload | Iterations | Measured ns/op | CPU samples |
| --- | ---: | ---: | ---: |
| Boolean Reconciliation | 3 | 544,846,646 | 1.63 s |
| Unicode Scalar Pipeline | 1 | 4,911,128,098 | 4.88 s |
| Run-length encode | 1 | 1,438,373,069 | 1.43 s |
| String Split/Join | 1 | 1,332,312,177 | 1.32 s |
| Iterator Collect | 3 | 501,387,571 | 1.50 s |
| Numeric Array Map | 20 | 127,947,424 | 2.55 s |

## Direct caller ownership

`pprof` call trees were focused on `bytecodeRawIntegerValueInfo(...)`, and its
direct parents were reconciled with their source flows. Percentages below are
of the corresponding whole profile.

| Workload | Raw extractor | Material direct owners |
| --- | ---: | --- |
| Boolean Reconciliation | 60 ms / 3.68% | `bytecodeDirectIntegerValue` 30 ms; `bytecodeIntegerValue` 20 ms; `isNumericValue` 10 ms |
| Unicode Scalar Pipeline | 120 ms / 2.46% | `bytecodeDirectIntegerValue` 100 ms; `isIntegerValue` 10 ms; `isNumericValue` 10 ms |
| Run-length encode | 20 ms / 1.40% | typed-pattern matching 10 ms; `isNumericValue` 10 ms |
| String Split/Join | 30 ms / 2.27% | typed-pattern matching 20 ms; same-type integer pair 10 ms |
| Iterator Collect | 60 ms / 4.00% | immediate comparison 30 ms; same-type integer pair 20 ms; return append 10 ms |
| Numeric Array Map | 220 ms / 8.63% | casts 120 ms; `bytecodeIntegerValue` 100 ms |

The ownership families therefore stop at two unlike programs:

- direct integer materialization: Boolean and Unicode;
- generic integer conversion: Boolean and Numeric Array Map;
- typed-pattern matching: Run-length and Split/Join;
- same-type pair extraction: Split/Join and Iterator Collect;
- immediate comparison, return append, and cast conversion: one program each.

`isNumericValue(...)` is the only direct parent present in three profiles, but
it contributes exactly one 10 ms sample in each. Source inspection also shows
that the arithmetic dispatcher can validate numeric operands before its exact
integer path extracts them again. That repeated work is real, but the fresh
profiles do not show it as a material three-program wall. A speculative
combined validation/extraction API was therefore not admitted.

## Broader wall selected for the next tranche

The same profiles point away from raw extraction and toward string-keyed map
lookup. `runtime.mapaccess2_faststr` is cumulative in all six workloads:

| Workload | Cumulative string-map lookup |
| --- | ---: |
| Boolean Reconciliation | 160 ms / 9.82% |
| Unicode Scalar Pipeline | 500 ms / 10.25% |
| Run-length encode | 150 ms / 10.49% |
| String Split/Join | 150 ms / 11.36% |
| Iterator Collect | 110 ms / 7.33% |
| Numeric Array Map | 120 ms / 4.71% |

These totals include many semantic maps, so they do not by themselves justify
a cache or representation change. They do establish a larger, repeated wall
whose concrete Able call-tree owners should be separated next. The relevant
question is whether global/method version lookup, type/match metadata,
call/member caches, or another semantic lookup family repeats across at least
three programs.

## Verification and cleanup

The focused raw-integer and runtime suites pass after the no-change decision.
The temporary test binary and six CPU profiles were removed after recording
the measurements. No canonical stdlib change was needed because this tranche
only reconciled VM primitive-carrier costs.

## Next recommendation

Attribute the shared `runtime.mapaccess2_faststr` wall by concrete interpreter
owner across the same six workloads, then admit a candidate only for one
lookup family repeated materially in at least three unlike programs.

Why: string-keyed map access now consumes roughly 5–11% cumulatively in every
profile, whereas raw-integer ownership has fragmented into smaller one- and
two-program families. This is the best current chance to remove general VM
overhead without encoding benchmark- or nominal-container-specific behavior.
The tranche should use focused call trees and temporary per-owner counters,
distinguish successful lookups from misses and version checks, test cache
invalidation and dynamic-definition semantics, and then apply the usual
repeated workstation benchmark gate with all outliers retained. Continue to
defer WASM.
