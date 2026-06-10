# Compiled residual call-path reachability decision — 2026-07-15

## Scope

This audit adds an opt-in generated-binary call-path census. It is not a
timing or allocation measurement: a telemetry build includes atomic counters,
so its execution time must not be compared with normal compiled binaries.
Normal compiler output contains none of the counters, branch checks, or
atomic operations.

The `coverage` catalog was run through the existing one-process,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, 45-second-cap path. Each
successful row passed its canonical Ruby verifier. The audit completed 32
verified rows; Sudoku remains status-only after its existing 45-second cap.
The machine-readable row evidence is in
`2026-07-15-compiled-call-path-reachability.json` and the table is in the
matching Markdown report.

## Results

`generic_union_method_call` reached only Option/Result Configuration
(147,456 calls), and `generic_union_fallback` reached no application. This
rules out an Option/Result- or generic-union-fallback-specific performance
change.

`fast_method_call` reached seven applications (165,484 total calls). Four
were materially exercised:

| Application | Calls | Shape |
| --- | ---: | --- |
| Option/Result Configuration | 147,456 | Generic named-union method table |
| await-channel-mux | 10,752 | Await/channel composition |
| mutex-await-journal | 6,196 | Mutex plus awaiting work |
| future-await-race | 1,059 | Future race/awaiting work |

The other three entries have only 4–9 calls and are selection guards rather
than profile targets. Unlike the generic-union counter, the fast method-value
helper crosses distinct language features and program shapes, so it clears the
reachability threshold for a *profile*—not an optimization.

## Decision

Keep no compiler, generated-runtime, VM, canonical-stdlib, or benchmark
change. The next eligible performance tranche is quiet-host CPU-only profiling
of normal compiled binaries for Option/Result Configuration,
await-channel-mux, mutex-await-journal, and future-await-race. Capture enough
samples to resolve descendants of `__able_call_value_fast` and admit a change
only if the same concrete leaf, rather than the helper parent, recurs across at
least three of those verified applications. Do not tune generic-union fallback,
Option/Result names, Future, channel, or Mutex types as separate cases.
