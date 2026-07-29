# Mode-aware benchmark contract and complete scorecard closure

Date: 2026-07-28

## Decision

Retain the general mode-aware benchmark argument/input contract and promote
all seven previously unranked bytecode rows. Compiled and Go rows continue to
use the broad canonical workloads; bytecode, Python, and Ruby use the
previously calibrated portable scales. No compiler, generated runtime,
runtime, interpreter, VM, canonical-stdlib, language, dependency, nominal
special case, or WASM production rule was changed.

The reviewed selection is now complete at 126 rows: 63 compiled and 63
bytecode.

## General contract

`bench_external_program_args` now accepts a benchmark mode. The comparison
driver resolves arguments inside the mode loop and fingerprints the resulting
program arguments, verifier, and declared inputs by `(benchmark, mode)`.
Reference refreshers explicitly request the compiled or bytecode contract.
Programs and equivalent references accept ordinary command-line workload
arguments rather than embedding a scorecard-only branch.

| Application | Compiled / Go | Bytecode / Python / Ruby |
| --- | --- | --- |
| Binary Trees | depth 21 | depth 15 |
| N-Body | 500,000 steps | 50,000 steps |
| Quick Sort | `numbers.txt` | first 500,000 values in `numbers-bytecode.txt` |
| Sudoku Masks | ten puzzles × ten passes | ten puzzles × one pass |
| TapeLang Alphabet | canonical ten-level delay | portable five-level delay |
| Fibonacci | `fib(45)` | `fib(40)` |
| Matrix Multiply | 1000×1000 | 400×400 |

The external public verifiers accept only the exact canonical or portable
result (or the existing N-Body numeric tolerance), and the per-application
READMEs document both modes. Quick Sort retains committed input/solution
identities, and TapeLang retains a committed portable program.

The full refresh driver also gained `--work-root`, which places its tagged
build tree below a caller-selected disk-backed directory. The retained
measurement used `/var/tmp`; no large build tree was placed in RAM-backed
`/tmp`.

## Validation before the full refresh

- Shell syntax and the mode-aware exact-argument contract tests passed.
- Selection, scorecard partition, and evidence tests passed.
- Modified Go references built after formatting; modified Python and Ruby
  references passed syntax checks.
- Fresh one-process canonical Go references passed 7/7.
- Fresh one-process portable Python/Ruby references passed 14/14.
- Able compiled smoke passed and verified 7/7.
- Able bytecode smoke passed and verified 7/7.

The smoke caught one real implementation defect before measurement: the Go
Quick Sort reference used `os.Args` without importing `os`.

## Complete repeated evidence

The authoritative refresh used CPUs 7–10 after a quiet-host preflight,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and a 90-second per-process cap. It collected:

- 126 Able rows × five successful processes = 630 verified Able processes;
- 63 Go rows × five successful processes = 315 verified Go processes;
- 126 Python/Ruby rows × five successful processes = 630 verified
  interpreter-reference processes.

All 1,575 timed processes completed. There were no timeouts, execution
failures, or verifier failures. The evidence checker confirmed 126 selected
rows, 126 full-status rows, 30 retained Able source reports, 30 retained
reference reports, five successful Able samples per row, and five successful
reference samples per comparison.

The aggregate is
`2026-07-28-mode-aware-benchmark-contract-refresh.{json,md}` and was promoted
to `external-scoreboard-current.{json,md}`. It records 126 measured Able
source fingerprints, 126 measured verifier/input contracts, and 189 measured
reference source fingerprints. The selection identity is
`0c72eaf2a1b12d3a5a2f88d00b3382a706f5c5c16977c24b55fb64214f8d429e`.

## Current performance frontier

- Compiled: 6/63 rows meet the 95%-of-Go target; the 63-ratio geometric mean
  is 5.575597× Able/Go and positive target excess totals 5.675368 seconds.
- Bytecode: 4/63 rows meet both the Python and Ruby targets; the 126-ratio
  geometric mean is 12.780200× and positive target excess totals 221.503684
  seconds.
- Compiled Binary Trees averaged 10.0680 seconds versus Go at 10.4076;
  Quick Sort averaged 1.7640 versus 2.7073; both exceed Go.
- Bytecode Fibonacci, Matrix Multiply, JSON, and Pidigits meet both
  interpreter targets.
- The seven promoted bytecode rows are all ranked. Fibonacci and Matrix
  Multiply meet both targets; Binary Trees, N-Body, Quick Sort, Sudoku Masks,
  and TapeLang remain measured misses rather than timeouts.

The largest absolute bytecode target excess remains concentrated in
K-Nucleotide, Sudoku Masks, TapeLang, Binary Trees, Quick Sort, and N-Body.
The portable-scale calibration already profiled the last five and found no
new shared owner beyond dispatch and closed raw-integer/frame/Array routes, so
that work must not be repeated unchanged.

## Stability and derived-evidence closure

A second independent current-contract cohort repeated every snapshot pass
with five verifier-backed Able processes and five matched reference processes:
30 compiled Able and 30 Go executions on the same `2,8,13,0` CPU pool, plus
20 bytecode Able and 40 Python/Ruby executions on CPU 6. Both retained lanes
passed the quiet-host preflight. All 120 retained timed processes passed;
there were no timeouts, failures, or verifier failures.

All ten snapshot passes also meet in the second cohort and are therefore
established current-contract guards. The pooled limiting ratios are:

| Mode | Benchmark | Pooled Able/reference |
| --- | --- | ---: |
| compiled | Base64 | 0.894995× Go |
| compiled | Binary Trees | 0.991676× Go |
| compiled | JSON | 0.421027× Go |
| compiled | Monte Carlo Pi | 0.786045× Go |
| compiled | PiDigits | 0.914921× Go |
| compiled | Quick Sort | 0.677263× Go |
| bytecode | Fibonacci | 0.052996× Ruby |
| bytecode | JSON | 0.524205× Ruby |
| bytecode | Matrix Multiply | 0.320929× Ruby |
| bytecode | PiDigits | 0.598706× Python |

One earlier compiled measurement completed and verified but was excluded
because its Able and Go rows used different CPU pools; the comparison harness
rejected that execution-contract mismatch before emitting evidence.

The cross-mode frontier now covers all 126 selected rows. The five calibrated
bytecode misses form an explicit `bytecode-portable-workload-admission`
closure, while the ten repeated passes form current target-guard groups.
After reviewing the benchmark-only source/row-definition drift, the
performance-evidence ledger was regenerated at 22 closures with zero
invalidations. No compiler, runtime, interpreter, VM, stdlib, language, or
WASM scope changed.

## Retention rationale

This is benchmark infrastructure, source portability, and evidence promotion,
not a benchmark-specific compiler optimization. The catalog rule is mode
based, every implementation exposes the same ordinary arguments, compiled
coverage remains canonical, and verifiers preserve equivalent algorithms and
outputs. No named container or non-primitive nominal type receives special
lowering.

## Next tranche

Build a current source-identity coverage map for CPU and allocation profiles
across all 59 bytecode misses, reconcile it against the existing closure
records, and select the largest absolute-excess rows that remain genuinely
unprofiled in at least three unlike application families. Then collect fresh
complete profiles and advance only an exact shared general VM/runtime owner.

Why: 63/63 bytecode coverage now makes the broad frontier measurable, but the
largest five newly promoted misses already failed the shared-owner gate.
What it entails: map current rows to retained profile identities, exclude
closed dispatch/raw-integer/frame/stack/Array/call/launch routes, profile only
the uncovered unlike rows, and require verifier-backed repeated A/B evidence
before retaining code. Why it matters: this is the shortest evidence-backed
path toward the 95%-of-Python/Ruby goal without repeating closed experiments
or introducing benchmark/container special cases.
