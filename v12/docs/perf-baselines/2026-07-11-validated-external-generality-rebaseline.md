# Validated External Generality Rebaseline

## Decision

Keep no compiler, runtime, canonical-stdlib, or benchmark-source performance
change. The first full generality pass through output validation proves that
some formerly timed rows are invalid correctness results, while the remaining
verified misses span distinct algorithms and runtime boundaries. A one-run
matrix is a selection map, not a performance-candidate verdict.

## Method

Ran the complete 15-program `generality` suite once in compiled and bytecode
modes with the external-scorecard validation gate:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  ./v12/bench_compare_external --suite generality \
    --modes compiled,bytecode --runs 1 --timeout 45 --cpu-affinity 2 --keep \
    --workdir v12/tmp/validated-generality-2026-07-11/work \
    --output-json v12/tmp/validated-generality-2026-07-11/generality.json \
    --output-md v12/tmp/validated-generality-2026-07-11/generality.md
```

Each benchmark ran from its exact external suite directory, including setup
hooks and program arguments. Successful output was verified when the suite
provided `verify.rb`; every successful capture retained a SHA-256 hash. The
full machine-readable report, Markdown table, and command log are under
`v12/tmp/validated-generality-2026-07-11/`.

## Status matrix

| Result class | Rows | Meaning |
| --- | ---: | --- |
| Verified and completed | 18 | Valid one-run comparison evidence. |
| Completed, verifier unavailable | 5 | Timing retained, but not eligible to select a performance candidate. |
| Verification failed | 3 | Invalid correctness rows; timing is intentionally absent. |
| Timed out | 4 | Bounded status only; no ratio or output claim. |

The three invalid rows are:

| Benchmark | Mode | Validation result |
| --- | --- | --- |
| BinaryTrees | compiled | `failed`; stdout hash retained. |
| Sudoku | compiled | `failed`; stdout hash retained. |
| Sudoku | bytecode | `failed`; same deterministic stdout hash as compiled. |

The bounded bytecode timeouts remain BinaryTrees, QuickSort, NBody, and
Tapelang. MatrixMultiply, JSON, and compiled Tapelang completed but are
explicitly verifier-unavailable.

## Interpreting the valid comparisons

Verified completed rows confirm several misses, but they are not one common
implementation wall: I-Before-E is text search/bridge work, Base64 is codec
work, Monte Carlo Pi is scalar random/numeric work, Pidigits is big-integer
work, Mandelbrot is float-loop work, and Reverse Complement is byte scanning.
The broad one-run data therefore does not authorize a compiler fast path,
runtime shortcut, nominal-container branch, or stdlib special case.

The correctness failures take priority over those performance readings. Until
BinaryTrees and Sudoku produce externally verified output, their old timing
rows must not be compared to Go, Ruby, or Python or used to select an
optimization.

## Next recommendation

Diagnose the BinaryTrees and Sudoku verifier mismatches before another
performance experiment. Why: they are deterministic invalid rows in the
fresh generality gate, and a fast incorrect binary would undermine the
benchmark-driven optimization program. The work entails preserving the exact
external inputs, capturing expected versus Able output, reproducing the
behavior in treewalker, bytecode, and compiled modes where applicable, then
adding a shared fixture/regression before implementing any semantic repair.
The fix must restore language semantics across runtimes; do not weaken a
verifier, change an external corpus, or add a BinaryTrees/Sudoku-specific
lowering rule.
