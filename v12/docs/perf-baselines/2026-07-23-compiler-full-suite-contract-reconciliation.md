# Compiler full-suite contract reconciliation

## Decision

**rebaseline-reviewed-compiler-and-shared-semantics-scopes**.

The compiler correctness tranche changed the generated compiler runtime and
shared interpreter concurrency semantics, invalidating all 21 performance
closures by scope hash. The changes repair general language/runtime contracts;
they do not introduce a benchmark-selected optimization. A final-code,
24-batch whole-fixture audit and two independent five-process performance
cohorts over affected unlike applications support rebasing both reviewed
scopes without changing any closure disposition.

This is an evidence rebase, not a performance win. It admits no implementation
candidate and makes no claim from the observed timing movement.

## Contract repairs

- Removed the unowned boundary marker and restored primitive Array write
  lowering without routing a discarded write result through `runtime.Value`.
- Aligned goroutine-executor cancellation completion with tree-walker
  semantics, including the generated runtime.
- Restored local dynamic-import statement lowering and its lexical environment
  ownership.
- Restored discarded tail-match lowering, control-transfer facade handling,
  and conservative result/type inference for unannotated and nominal paths.
- Lowered discarded conditionals and blocks directly as statements, and
  inferred an implicit `void` result only when an unannotated conditional tail
  cannot produce a value. Recursive static calls such as Quicksort therefore
  retain their native carrier without weakening value-returning inference.
- Restricted interface identity qualification to implementations whose
  interface ownership is actually resolved.
- Made the generated error bridge preserve both the exact raised Able value and
  its original diagnostic/error chain. This fixes singleton errors rescued
  after crossing a generic-union method boundary without special-casing the
  union or error type.
- Split/deduplicated the expensive reporters audit so each fixture remains
  independently diagnosable.
- Added `--build-timeout` to `bench_compare_external`, forwarding the existing
  preparation timeout separately from the per-run timeout. Compiling a program
  can no longer consume the runtime measurement allowance.

No named container, non-primitive nominal, benchmark, WASM, or external-stdlib
special case was added.

## Correctness evidence

The final consolidated audit ran
`TestCompilerInterfaceLookupBypassForStaticFixtures` with
`ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all` in all 24 zero-based batches.
That audit compiles and executes every selected fixture and simultaneously
requires:

- exact stdout, stderr, and exit status;
- the strict-dispatch marker;
- zero boundary fallback markers;
- zero interface-lookup fallback markers; and
- zero global-environment fallback markers.

All 24 batches passed on final code. Four bounded workers used
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`; aggregate batch elapsed times
ranged from 197 to 378 seconds because each batch contains many separately
bounded generated-program builds and executions.

Focused repeated checks also passed:

| Check | Result | Elapsed arithmetic mean |
| --- | --- | ---: |
| generated error carrier, singleton rescue, generic-union call | 3/3 | 13.27s |
| `06_12_30_stdlib_option_result` compiled execution | 3/3 | 5.16s |
| strict/interface diagnostic preservation | 3/3 | 16.92s |
| reporters interface audit | 3/3 | 57.78s |
| implicit-void/static recursive carrier group | 3/3 | 30.08s |

The first error-carrier repetition included a cold Go build; all individual
focused invocations completed within one minute.

The former monolithic compiler-package handoff could exceed its aggregate
30-minute timeout even when no individual test was stalled. The default
`run_all_tests.sh` handoff now schedules the same short-mode compiler tests in
bounded batches of 25, retaining per-command timeouts and independently
diagnosable failures. The final bounded default handoff passed all noncompiler
packages, all 31 compiler batches, and the complete bytecode fixture pass.

## Repeated-process performance evidence

Each ranked benchmark/mode has two independent five-process cohorts. All 100
ranked processes passed their exact output verifier; zero failed or timed out.
The legacy `sudoku` program timed out in all five first-cohort compiled runs and
is explicitly unranked. `sudoku_masks` is the maintained Sudoku guard.

`Cohort spread` is the absolute difference between cohort means divided by
their mean. Rows above 15% are marked volatile.

| Benchmark | Mode | Cohort means (s) | Pooled mean (s) | Cohort spread | Volatile |
| --- | --- | ---: | ---: | ---: | --- |
| `quicksort` | compiled | 1.9500 / 2.1140 | 2.0320 | 8.1% | no |
| `sudoku_masks` | compiled | 2.1500 / 2.0780 | 2.1140 | 3.4% | no |
| `nbody` | compiled | 0.1760 / 0.1560 | 0.1660 | 12.0% | no |
| `dependency_plan` | compiled | 0.1180 / 0.1000 | 0.1090 | 16.5% | yes |
| `future_pipeline` | compiled | 0.4060 / 0.3680 | 0.3870 | 9.8% | no |
| `await_channel_mux` | compiled | 0.4020 / 0.4160 | 0.4090 | 3.4% | no |
| `concurrent_stencil_reduction` | compiled | 0.2620 / 0.2740 | 0.2680 | 4.5% | no |
| `future_pipeline` | bytecode | 0.4860 / 0.5200 | 0.5030 | 6.8% | no |
| `await_channel_mux` | bytecode | 0.2580 / 0.2500 | 0.2540 | 3.1% | no |
| `concurrent_stencil_reduction` | bytecode | 1.9700 / 1.8700 | 1.9200 | 5.2% | no |

These rows cover primitive Array/numeric lowering, iterator/control behavior,
and the shared concurrency executor. Error construction and rescue are cold
failure paths and are governed by the repeated semantic checks rather than a
timing claim.

The final Quicksort pooled mean is `1.011x` its stored `2.0100s` Go reference,
inside the project's 95%-of-Go target band. This is a contract guard for the
generic static-carrier repair, not evidence that the overall compiler target
is complete.

## Interpretation

The compiler and shared-semantics scope changes are now exercised by broad
correctness contracts and verifier-backed repeated processes. The one volatile
row is a very short application whose 18 ms cohort difference is consistent
with workstation scheduling noise. Arithmetic pooling prevents one process
from controlling the result, while the volatility label prevents that pooled
number from being presented as a causal change.

The repaired paths are general compiler/runtime mechanisms. No exact hot owner
was selected from these descriptive cohorts, so all prior closure dispositions
remain unchanged.

## Next recommendation

Complete `portable-concurrent-interface-data-application-frontier`.

Why: concurrency × arrays/files × user-defined interface dispatch is the
highest-ranked remaining minimum-depth interaction. Only three portable
applications cover it, and its adjacent current rows retain about 36.917
seconds of target excess. Strengthening that interaction gives the next
profile selection a better chance of finding a general mechanism used by real
programs rather than a benchmark-shaped shortcut.

What it entails: add one deterministic file-driven application whose workers
process numeric or structured Array data through ordinary user-defined
interfaces; provide source-equivalent Able, Go, Python, and Ruby programs and
one exact verifier; run two five-process cohorts per relevant lane; collect
bounded compiled and bytecode profiles; and admit a candidate only if the same
exact generic owner is material in at least three unlike applications.

Why this scope: it improves weak feature-interaction coverage and measures both
performance targets while preserving the project gates. It excludes WASM,
benchmark branches, named-container lowering, and non-primitive nominal
special cases.
