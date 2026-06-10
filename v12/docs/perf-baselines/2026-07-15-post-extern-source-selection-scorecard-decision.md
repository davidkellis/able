# Post extern/source-selection scorecard decision

## Scope

The shared compiler host-extern repair and `able setup` source-resolution
repair justified one fresh selection screen. `just bench-scorecard-refresh`
ran the complete current portable application scope on CPU 14 after its
immediate three-sample quiet-host preflight. Every timed process used one CPU,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, one run, and a 45-second cap.
The generated source reports and aggregate are retained as
`2026-07-15-post-extern-source-selection-*` in this directory.

The refresh rebuilt fresh Go references for compiled mode and fresh Python/Ruby
references for bytecode mode. Its aggregate is reproducible by passing its 16
source scorecards back to `bench_external_scoreboard`.

## Result

There are 5 of 21 rankable compiled rows at or better than the 95%-of-Go
target, and 3 of 15 rankable bytecode rows at or better than both the Python
and Ruby targets. Timeouts and missing matched references remain unranked.

Against the 2026-07-14 one-run screen, almost every existing target
classification is unchanged. Monte Carlo becomes a compiled pass, while the
newly rankable bytecode MatrixMultiply row is a pass because its current Python
and Ruby references are much slower; neither is evidence of a source-level
change. The current runs are generally faster than the prior one-run screen,
including unrelated controls, so their absolute deltas are host/run variation,
not a claimed regression or improvement.

The material misses still divide below their common entry frames:

- compiled async applications retain the previously profiled
  `bridge.currentGID` / `runtime.Stack` identity route, whose fixed
  execution-context ABI has already failed broad N-body and K-Nucleotide
  guards;
- K-Nucleotide map/counting and Reverse Complement tracked-byte work remain
  distinct descendants; and
- numeric, recursive-search, text, channel, Future, and Mutex applications do
  not repeat one new concrete VM/compiler/runtime leaf across three unlike
  verified programs.

## Decision

Keep no compiler, VM, generated-runtime, canonical-stdlib, or benchmark
source change. In particular, do not reopen the rejected execution-context
ABI, print-boundary, or named-container experiments from scorecard ratios or
parent-frame recurrence alone.

The host-extern and stdlib resolver changes are retained for correctness; this
screen establishes that they did not expose a new broad performance candidate.
Any next implementation tranche requires a material cross-cutting change or a
new spec-defined portable application, followed by bounded profiles that show
the same concrete descendant in at least three unlike verifier-backed
applications and broad compiled/bytecode controls.
