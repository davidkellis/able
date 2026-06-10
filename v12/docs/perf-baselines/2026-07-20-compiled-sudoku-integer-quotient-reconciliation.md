# Compiled Sudoku integer-quotient reconciliation

Date: 2026-07-20

## Decision

Keep the Sudoku source-equivalence correction from floating `/` plus an `i32`
cast to integer `//`. Keep no compiler, interpreter, runtime, fixture, or
canonical-stdlib optimization from this gate.

The correction is semantic, not a performance claim. In Able, `/` is floating
division and `//` is Euclidean integer quotient. The Go, Python, and Ruby
references all perform integer quotient for subgrid coordinates. The canonical
Able benchmark and its external mirror now have the same source SHA-256:
`88294708698dd72bd6ac6a6249633cc7fddf4274a33587930f8e932b00b199a5`.

## Measurement contract

- Every timed output passed the benchmark verifier.
- Serial Sudoku and QuickSort were pinned to CPU 0. Goroutine Binary Trees used
  CPUs 0-3 and its catalog budget of four logical CPUs.
- Preserved binaries were built before timing. Each application received two
  independent five-run cohorts; all samples, including workstation variation,
  are included in the means.
- CPU profiles cover only the registered Able `main` phase and merge three
  independently verified launches per application.
- Allocation snapshots subtract `main-start` from `main-end`. The Binary Trees
  allocation run reached the 55-second cap and is reported only as bounded
  status; its merged CPU profile is complete.

## Repeated compiled results

The preserved pre-correction binary establishes the controls used for
ownership reconciliation:

| Application | Cohort 1 | Cohort 2 | Pooled mean | Spread | Fresh Go | Able/Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Sudoku Masks | 1.616 s | 1.670 s | 1.643 s | 3.34% | 0.5334 s | 3.08x |
| QuickSort | 1.670 s | 1.702 s | 1.686 s | 1.92% | 2.3475 s | 0.72x |
| Binary Trees | 6.986 s | 7.158 s | 7.072 s | 2.46% | 10.1473 s | 0.70x |

The source correction was then measured in ten alternating preserved-binary
pairs, reversing order for each pair:

| Sudoku binary | Mean | Median | Range | Mean user |
| --- | ---: | ---: | ---: | ---: |
| Pre-correction | 1.703 s | 1.700 s | 1.69-1.75 s | 1.662 s |
| Integer quotient | 1.711 s | 1.710 s | 1.70-1.74 s | 1.672 s |

The +0.47% wall difference is neutral workstation noise. The correction is
retained because it makes the compared programs equivalent, not because it
improves the timing.

The refreshed formal scorecard uses five verified samples. Corrected Sudoku is
1.772 s versus Go at 0.5334 s, or 3.32x and a miss. QuickSort remains a meet at
1.688 s versus Go at 2.3475 s, or 0.72x. Full bytecode Sudoku and QuickSort each
reach the 55-second cap; the existing Python and Ruby reference means remain
visible rather than being converted into unsupported ratios.

## CPU ownership

Correcting the operation moves cost within `square_index`, but does not reduce
that function's total ownership:

| Main-only leaf or owner | Pre-correction Sudoku | Corrected Sudoku | QuickSort | Binary Trees |
| --- | ---: | ---: | ---: | ---: |
| `find_best_empty` cumulative | 86.15% | 87.01% | — | — |
| `square_index` cumulative | 35.64% | 35.63% | — | — |
| `__able_divmod_signed` cumulative | — | 15.35% | — | — |
| `__able_checked_mul_signed` cumulative | 6.31% | 12.40% | 11.92% | — |
| `quicksort` cumulative | — | — | 75.56% | — |
| `make_tree` cumulative | — | — | — | 78.32% |
| `runtime.mallocgc` cumulative | 9.98% | 11.81% | below 1% | 71.29% |

The old generated path paid for floating conversion, `math.IsNaN`,
`math.IsInf`, truncation, and range checks. The corrected path pays for the
general Euclidean division helper and signed bounds. `square_index` remains
35.6% either way. Checked multiplication is shared by Sudoku and QuickSort but
not Binary Trees, and its QuickSort samples belong to input parsing rather than
the sorting loop. Binary Trees is dominated by allocation and GC. Therefore no
material exact leaf repeats across all three unlike applications.

## Allocation ownership

Both Sudoku binaries allocate exactly 156,370,688 bytes in 7,802,594 main-phase
allocations with 11 collections. Allocation-space profiles assign 94.4% to
`find_best_empty` in both builds. QuickSort allocates 296,952,336 bytes in only
4,327 main-phase allocations, with 64.7% of allocation space in input parsing
and 32.4% in file reading. Binary Trees' bounded allocation snapshot timed out,
while its complete CPU profile independently attributes 78.32% cumulative CPU
to `make_tree` and 71.29% to `runtime.mallocgc`.

These are different ownership shapes. A Sudoku Array-allocation special case,
solver-shape lowering, or named-container branch would violate the broad
optimization bar and the nominal-lowering rules.

## Interpreter and correctness status

The unchanged full Sudoku workload still exceeds the 55-second bytecode cap.
A bounded one-puzzle source retaining the benchmark's existing ten-iteration
loop passes the Sudoku verifier in both interpreters: tree-walker in 1.75 s and
bytecode in 0.32 s. This separates correctness from the full-workload status.

Targeted Go checks pass:

```text
go test ./pkg/compiler -run 'TestCompilerDivModConcreteCarrierStaysNative|TestCompilerDiagnosticsParity' -count=1 -timeout 50s
ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_division_ops go test ./pkg/compiler -run '^TestCompilerExecFixtures$' -count=1 -timeout 50s
go test ./pkg/profilehook -count=1 -timeout 50s
just bench-catalog-check
```

## Next admission gate

Run a bounded quotient-only ownership census before changing lowering. Corrected
Sudoku makes `__able_divmod_signed` material at 15.35%, but neither control uses
that exact helper materially and the existing positive-input fast path has
already landed. Census at least three unlike applications that use `//`, and
separate quotient-only `//`, remainder-only `%`, and paired `/%` sites. Preserve
their binaries, collect repeated verified wall means and main-only profiles,
and record divisor/sign proofs available to general primitive lowering.

Only admit a candidate if the same quotient-only cost is material in all three.
Any candidate must preserve Euclidean negative operands, division by zero, and
minimum-integer divided by -1. This is the smallest next experiment that can
tell whether Sudoku exposed a general primitive cost or merely moved work
inside one benchmark.
