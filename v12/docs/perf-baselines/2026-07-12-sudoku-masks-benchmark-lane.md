# Sudoku Masks benchmark lane (2026-07-12)

## Purpose

The stable `sudoku` benchmark retains its committed scan-based solver, with a
small correctness normalization that treats `.` as a blank just like `0` in
the canonical corpus. The locally developed mask-based solver is now a
separate `sudoku_masks` lane instead of silently changing the historical
comparison. It uses row, column, and subgrid bit masks, most-constrained
empty-cell selection, recursion, and
file input. The external `full` / `generality` catalog includes the lane, while
the legacy `core` suite remains unchanged.

The retained scan-based core lane now parses the corpus correctly but reaches
the current 45-second compiled guard, rather than emitting malformed output
for `.` blanks. Keep that as the honest legacy-core status; the new bounded
lane does not replace or conceal it.

The new suite deliberately shares the canonical `sudoku/sudoku.txt` input
without duplicating it. It solves the first ten valid puzzles ten times (100
solves). The full 101-puzzle corpus completed in Go but exceeded the
45-second guard in preliminary Ruby and Python runs, so the fixed prefix keeps
all reference languages measurable under the same guard. The suite-local
verifier checks givens, rows, columns, and subgrids; it does not require one
particular valid solution for an ambiguous puzzle.

## Fresh pinned baseline

Each reference used one CPU-2-pinned, verifier-backed process with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second cap. Go 1.26.4,
Ruby 4.0.5, and Python 3.14.5 emitted the identical 100-line solution stream
(`35a81e…659ec`).

| Runtime | Real time (s) | Status |
| --- | ---: | --- |
| Go 1.26.4 | 0.6652 | verified 1/1 |
| Ruby 4.0.5 | 25.0302 | verified 1/1 |
| Python 3.14.5 | 22.0067 | verified 1/1 |
| Compiled Able | 9.5400 | verified 1/1 |
| Bytecode Able | n/a | timed out at 45s |

Compiled Able is `14.34x` the Go reference, `0.38x` Ruby, and `0.43x` Python.
This is an important honest coverage row: it is already faster than the
dynamic-language references, but does not meet the compiler target. Bytecode
has no valid ratio because it did not complete.

## Decision

Keep the benchmark and make no runtime, compiler, or `able-stdlib` change.
This is one newly separated application shape, not proof of a reusable
hot-path candidate. Do not add a Sudoku solver, mask-array, or named-container
optimization. Any future candidate must recur in an independent existing
array/integer/recursive workload and remain neutral on the text and numeric
controls.

## Next recommendation

Profile the current compiled and bytecode mask lanes only alongside an
independent recursive-array workload that is also materially below its
reference. Why: the new lane demonstrates a real feature combination, but one
application cannot establish a general cost. The work entails bounded CPU and
allocation captures for both programs, then a candidate only for the same
runtime or generated-code helper in each; retain broad Go, Python, and Ruby
guards before landing any change.
