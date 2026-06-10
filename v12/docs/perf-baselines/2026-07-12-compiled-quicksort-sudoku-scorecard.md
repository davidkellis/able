# Compiled QuickSort / Sudoku scorecard validity check (2026-07-12)

## Method

The planned independent compiler comparison rebuilt Go 1.26 references for
QuickSort, Sudoku, and JSON. Each reference used ten verifier-backed processes
pinned to CPU 2 with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a
45-second cap.

| Benchmark | Fresh Go real mean (s) | Valid runs |
| --- | ---: | --- |
| QuickSort | 2.8749 | 10/10 |
| Sudoku | 0.1600 | 10/10 |
| JSON | 1.6533 | 10/10 |

The current compiled Able QuickSort row also completed and verified on all ten
runs at `2.0420s`: `0.71x` fresh Go. It is therefore a healthy control, not a
material compiler miss.

## Sudoku exclusion

The current Able Sudoku source is locally modified: it replaces the committed
row/column/subgrid validity scan with mutable masks, best-empty selection, and
bit-counting. That is a source/algorithm change in the benchmark under test,
not a compiler change. Its first current generated process exceeded the
45-second cap; after the second process had begun, the remaining redundant
runs were stopped. The reference Go source remained valid at
`0.1600s`, but those two sources are no longer a stable, comparable scorecard
pair.

Do not call this a compiler regression, profile it as a general compiler wall,
or change the user-owned Sudoku implementation from this evidence. The JSON
Able control was not rerun after the invalid pair was detected; the immediately
preceding fresh scorecard remains its valid comparison.

## Decision

No code change and no `able-stdlib` change. This tranche rejects the selected
pair as a shared optimization probe: one program is already faster than Go and
the other no longer represents a clean benchmark lane. No generated `main`
profile is warranted because the prerequisite of two material, comparable
misses is not met.

## Next recommendation

First restore benchmark-suite comparability as an explicit maintenance task:
either retain the Sudoku algorithm rewrite and update equivalent reference
implementations/baselines, or leave the benchmark's committed algorithm intact
and measure the rewrite separately. Why: language performance claims require
algorithmically comparable, versioned programs. The work entails a maintainer
decision about that user-owned source, followed by verifier-backed fresh rows
for every affected language. Only after the suite is comparable should compiler
profiling resume.
