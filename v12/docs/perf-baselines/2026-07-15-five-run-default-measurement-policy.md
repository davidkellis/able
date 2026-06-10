# Five-run default measurement policy — 2026-07-15

## Decision

All active performance entry points now default to five independent process
samples: `bench_perf`, `bench_suite`, `bench_compare_external`,
`bench_refresh_go_refs`, `bench_refresh_interpreter_refs`, and
`bench_refresh_external_scorecard`.

This implements the workstation policy directly. A quiet CPU preflight remains
an optional precision tool, not a requirement for collecting evidence. Every
successful comparison and reference process still runs its canonical Ruby
verifier, and the existing memory/GC/one-P guards remain unchanged.

## Why five samples

One sample cannot distinguish a source change from normal shared-workstation
variation, and three samples have previously exposed material spread near
targets. Five independent launches provide an average suitable for the current
scorecard schema while keeping every individual process below its existing
timeout. The per-run records remain available for median and spread analysis
through `just bench-variance-report` whenever a result is close enough to
influence a decision.

The comparison and foreign-reference reports retain each run's status, timing,
and verified-output hash. A single modern five-run comparison can therefore
report Able and reference timing spread directly. It intentionally does not
invent per-run Able/reference ratios: those processes are independently
sampled, so ratio variation still needs multiple matched comparison reports.

The commands still accept `--runs N`. An explicit lower count is allowed for a
quick smoke check, but it must not be used to choose, retain, or claim a
compiler/VM performance improvement. Candidate experiments still require the
same concrete non-nominal descendant in three unlike verifier-backed
applications and broad-control validation.

## Scope

This is measurement infrastructure only. It changes no Able source semantics,
compiler lowering, bytecode VM behavior, canonical stdlib source, benchmark
workload, target, or timeout. It does not rerun an unchanged scorecard and it
does not relax the no-benchmark-specific-optimization rule.
