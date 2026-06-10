# Unranked scorecard reconciliation — 2026-07-15

## Scope

This reconciles the unranked rows in the promoted 32-application external
scorecard. An unranked row is never a target pass or miss; it means a bounded
Able or matched-reference run was incomplete.

## Current causes

| Mode | Rows | Cause |
| --- | --- | --- |
| compiled | Sudoku | Able exceeded the 45-second guard. The current Go reference completed. |
| bytecode | Binary Trees, QuickSort, Sudoku, Sudoku Masks, N-Body, TapeLang Alphabet, Regex Suffix Audit | Able exceeded the 45-second guard. Their matched Python/Ruby references are recorded. |
| bytecode | K-Nucleotide | Mixed result: one bounded Able process completed and others timed out, so the row is deliberately `incomplete` and unranked. |
| bytecode | Fib | Able and Ruby completed; the existing Python implementation exceeded the reference guard, so the required Python-and-Ruby comparison is unavailable. |

The remaining current scorecard rows have the matched references required by
their target. Adding another application, a duplicate reference source, or a
benchmark-only timeout exception would not make an existing unranked Able row
rankable.

## Decision

Keep the scorecard classification unchanged. The coverage gap is runtime
performance, not missing portable benchmark/reference infrastructure. Resolve
an unranked Able row only through a generic implementation improvement that
passes the normal cross-application candidate gate; do not loosen the cap,
replace a reference algorithm, or report a partial row as a target result.

The current current scorecard is
`external-scoreboard-current.{json,md}`. This reconciliation adds no timing,
VM, compiler, canonical-stdlib, fixture, benchmark-source, or reference-source
change.
