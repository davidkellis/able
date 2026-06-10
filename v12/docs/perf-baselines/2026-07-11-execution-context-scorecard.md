# Experimental Execution-Context Scorecard

## Scope

This tranche made the generated-call execution-context ABI selectable through
every normal benchmark entry point:

- `v12/bench_perf --experimental-execution-context`
- `v12/bench_suite --experimental-execution-context`
- `v12/bench_compare_external --experimental-execution-context`

`able build` now forwards that explicit opt-in to `compiler.Options`. JSON and
Markdown reports record whether it was requested. The mode affects compiled
programs only; bytecode measurements are a separate VM scorecard.

The external wrapper also now runs each program from its own benchmark
directory. The previous rule that switched every legacy `run.able` directory
to the shared benchmark root exposed unrelated packages and made `nbody` fail
before measurement. The repaired wrapper measures the same canonical target
with its own relative inputs; it is not a program-specific workaround.

## Compiled ABI Decision

The paired fixture gate used `fixture-generality`, compiled mode, two serial
runs per program, a 90-second run cap, and a 240-second build cap. It covers
arrays, channels/futures, queues/deques/heaps, persistent structures, numeric
loops, regex/string work, and graph algorithms.

| Result | Default | Candidate |
| --- | ---: | ---: |
| Programs | 15 | 15 |
| Successful runs | 30/30 | 30/30 |
| Candidate/default geometric mean | — | 1.0594x |
| Strictly slower programs | — | 7 |
| Strictly faster programs | — | 4 |
| Equal at timer resolution | — | 4 |

The candidate regressed shared numeric and collection cases, including
`heap_i32_small` (1.60x), `random_lcg_i64_small` (1.39x), and
`nbody_small` (1.25x). It also improved some workloads, but the objective is a
broad improvement rather than moving a few benchmarks.

The independent external generality corpus then ran 15 programs pinned to CPUs
`2-3`, one bounded run each. The candidate completed all 15 programs but was
1.0141x candidate/default geometrically, slower in 10/15 programs, faster in
2, and equal in 3. This single-run lane is directional; the repeated fixture
gate is the decisive measurement for small differences.

Decision: keep `ExperimentalExecutionContext` opt-in. Do not change the
production compiler default and do not pursue a benchmark-shaped patch to
rescue it.

## External Baseline

The corrected default report covers compiled and bytecode execution against
the checked-in Go, Ruby, and Python reference results:

- Compiled Able completed all 15 programs. Of the 13 with a Go reference, 3
  were within the 95%-of-Go time threshold for this snapshot (`quicksort`,
  `base64`, and `json`).
- Bytecode completed 11/15 programs and timed out at 120 seconds on
  `binarytrees`, `quicksort`, `nbody`, and `tapelang_alphabet`. Against the
  Ruby references with available successful measurements, 5/8 are within the
  95% threshold; against Python, 5/7 are within it.

These are scorecard facts, not excuses for per-program lowering. They identify
the next generic VM wall: allocation/dispatch pressure that repeats across the
timeout and slow numeric/recursive workloads.

## Artifacts

- `2026-07-11-compiled-execution-context-fixture-generality-default.{json,md}`
- `2026-07-11-compiled-execution-context-fixture-generality-candidate.{json,md}`
- `2026-07-11-external-generality-default.{json,md}`
- `2026-07-11-external-generality-execution-context-candidate.{json,md}`

No Able program source or external stdlib source changed in this tranche.
