# Fixed-context ABI default scorecard gate — 2026-07-13

## Scope

This gate tested the existing opt-in fixed-pointer execution-context ABI
(`ablec -experimental-execution-context`) against the default compiled ABI.
It is a compiler-wide generated-call ABI experiment, not a named-container,
task-count, source-shape, or benchmark rule.

The purpose was to decide whether the option can become the default after the
paired Channel Rollup/Future Pipeline profile evidence. It must not: the ABI
has real concurrent benefits but a verified material regression in an
independent serial application.

No production source changes are made in this gate. The default remains the
legacy ABI and the fixed-context ABI remains explicitly opt-in.

## Method

The primary selection screen compiled every external `generality` application
twice from the same sources, once per ABI. Each application ran as a separate
CPU-15-pinned process with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a
45-second cap. Every completed output passed its canonical Ruby verifier.

One run per application identifies large candidate losses; it is not used as
a release timing claim. Any apparent material regression was then repeated
three times per ABI under the same guards. The normal-scheduler concurrency
screen deliberately omitted `GOMAXPROCS=1`, because forcing one P alters the
goroutine behavior that the context ABI is intended to improve.

## Broad selection screen

| Application | Default | Fixed context | Selection delta | Status |
| --- | ---: | ---: | ---: | --- |
| Fib | 3.67s | 3.38s | -7.9% | verified |
| BinaryTrees | 32.83s | 34.04s | +3.7% | verified |
| MatrixMultiply | 1.34s | 1.26s | -6.0% | verified |
| QuickSort | 1.98s | 1.86s | -6.1% | verified |
| Sudoku | cap | cap | n/a | symmetric timeout |
| Sudoku Masks | 10.00s | 12.28s | +22.8% | repeated below |
| I-Before-E | 0.21s | 0.13s | -38.1% | verified |
| Base64 | 2.77s | 2.48s | -10.5% | verified |
| JSON | 0.82s | 0.83s | +1.2% | verified |
| Monte Carlo Pi | 0.21s | 0.21s | 0.0% | verified |
| PiDigits | 1.36s | 1.46s | +7.4% | verified |
| Mandelbrot | 0.13s | 0.21s | +61.5% | repeated below |
| Reverse Complement | 0.11s | 0.19s | +72.7% | repeated below |
| K-Nucleotide | 3.75s | 4.10s | +9.3% | repeated below |
| N-body | 0.50s | 0.74s | +48.0% | repeated below |
| TapeLang Alphabet | 3.78s | 3.96s | +4.8% | verified |

The selection results intentionally include both wins and losses. A default
ABI must improve or retain broad behavior; averaging fast text or concurrency
rows over a meaningful serial regression would be invalid.

## Matched serial regression repeats

The five suspicious rows were rebuilt and measured three times per ABI. Every
one of the 30 successful launches was accepted by its canonical verifier.

| Application | Default mean | Fixed-context mean | Delta | Default GC | Context GC | Result |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Sudoku Masks | 9.1833s | 9.2800s | +1.1% | 184.33 | 181.00 | neutral |
| K-Nucleotide | 4.4900s | 4.0633s | -9.5% | 59.33 | 65.33 | better |
| N-body | 0.4267s | 0.6600s | **+54.7%** | 3.00 | 8.00 | regression |
| Mandelbrot | 0.1567s | 0.1300s | -17.0% | 3.00 | 3.00 | better |
| Reverse Complement | 0.1233s | 0.1133s | -8.1% | 4.00 | 3.33 | better |

The large N-body loss is a stable, disqualifying serial regression. The
one-run Sudoku Masks, Mandelbrot, and Reverse Complement losses did not hold,
which confirms that a one-process scorecard is a selection gate rather than
rollout authority.

## N-body generated-main profile

Paired phase-local CPU captures used the same one-core guards and verified
output:

- Default main: 330 ms samples, 0.5000s wall, 3 GCs.
- Fixed-context main: 550 ms samples, 0.6500s wall, 8 GCs.

The default `advance` loop spends its time in the existing general `sqrt` and
`abs` generated bodies (240 ms and 110 ms cumulative respectively), with only
40 ms cumulative in bridge environment swapping. The fixed-context version
adds 130 ms (23.6% of the profile) in
`__able_context_with_environment` / `runtime.mallocgc`, alongside longer
`sqrt_ctx` and `abs_ctx` work.

This identifies a generic ABI flaw, not an N-body or math special case. A
context-aware package entry currently rebinds a `*__able_execution_context` to
the callee package environment by allocating a new context object whenever
that environment differs. N-body has a dense static cross-package call path,
so those transient context allocations dominate and raise GC count. The same
mechanism is emitted for every package entry; it must be fixed, if at all, at
the generated ABI boundary.

Retained artifacts:

- `v12/interpreters/go/.profiles/20260713_fixed_context_nbody_default_phases/main.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_fixed_context_nbody_context_phases/main.cpu.pprof`

## Normal-scheduler concurrency confirmation

One CPU-15-pinned process per ABI, without `GOMAXPROCS=1`, verified all three
concurrency applications:

| Application | Default | Fixed context | Selection delta |
| --- | ---: | ---: | ---: |
| BinaryTrees | 38.69s | 35.36s | -8.6% |
| Channel Rollup | 1.29s | 1.29s | 0.0% |
| Future Pipeline | 0.72s | 0.76s | +5.6% |

This is a one-run scheduler screen, not the primary concurrency proof. The
preceding matched three-run gate already measured Channel Rollup 11.5% lower
and Future Pipeline 12.2% lower with the fixed-context ABI. Together the
results confirm the candidate's purpose while showing that a default decision
cannot rely only on concurrency applications.

## Decision

- Do **not** enable the fixed-context ABI by default. N-body's verified 54.7%
  serial regression violates the broad-performance requirement.
- Keep it as an opt-in compiler experiment. It retains a real generic remedy
  for `bridge.currentGID` / `runtime.Stack` overhead in concurrent generated
  programs, without any benchmark-specific rule.
- Keep no bytecode VM, default compiler, bridge, runtime, or canonical
  `able-stdlib` source change in this tranche.

## Next gate

Prototype a generic allocation-free package-environment linkage for the
execution-context ABI. The immutable task payload/context must remain safe
across spawned goroutines, while a static package call must not allocate a new
context merely to expose that package's dynamic-boundary environment. A viable
design must keep dynamic and native compatibility entries correct, preserve
per-task payload isolation, and avoid a `sqrt`, N-body, or package-name
special case. Gate it first with generated-source and nested-spawn/dynamic
boundary tests, then repeat the N-body profile, serial five-row gate, and
Channel Rollup/Future Pipeline pair before considering a new broad scorecard.
