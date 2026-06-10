# External Sudoku Size-Ladder Attribution

## Decision

Keep no compiler, bytecode-runtime, canonical-stdlib, or benchmark-source
change. The current full Sudoku timeouts are not an input, output, or verifier
contract mismatch: bounded prefixes produce the exact upstream solutions in
both execution modes. The evidence instead exposes a possible general
bytecode live-allocation/lifetime boundary, which needs an independently
shaped control before any runtime candidate is considered.

## Method

The disposable driver at
`v12/tmp/sudoku-size-ladder-2026-07-11/driver/sudoku_ladder.able` differs from
the benchmark only by accepting an argument whose byte length limits the
number of input lines. It preserves the benchmark's ten outer iterations,
parsing, solving, formatting, imports, and external-suite working directory.
The driver is diagnostic-only and does not modify the benchmark source.

For each completed prefix, the retained Ruby oracle derives the expected
output from the corresponding prefix of the upstream `solution.txt`, repeated
ten times, and requires byte-for-byte equality. Processes ran on CPU 2 with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second guard. The
unmodified upstream `verify.rb` remains the acceptance gate for the full
101-line program; the current generality scorecard records both full modes as
timeouts before that verifier can run.

## Verified prefix ladder

| Input puzzles | Compiled | Bytecode | Verification |
| ---: | --- | --- | --- |
| 1 | 0.09 s, 3 GCs | 0.55 s, 8 GCs | both exact upstream-derived output |
| 2 | 1.65 s, 31 GCs | timeout at 45.04 s | compiled exact; bytecode incomplete |
| 3 | 1.81 s, 34 GCs | not rerun after the two-puzzle timeout | compiled exact |
| 10 | 9.31 s, 173 GCs | not run; two puzzles already exceed the guard | compiled exact |
| 101 (unchanged benchmark) | timeout at 45 s | timeout at 45 s | upstream verifier not reached |

The four completed compiled prefix hashes and both completed one-puzzle
hashes are retained alongside runner rows under
`v12/tmp/sudoku-size-ladder-2026-07-11/`. The full rows are in
`v12/tmp/scorecard-generality-refresh/sudoku.json`.

## Attribution

The two-puzzle bytecode process emits only a partial seven-line (and, on an
earlier identical attempt, three-line) output before the guard. In both
attempts it reaches its 32nd GC near 42 seconds; the last reported live heap
is 454--480 MB and the concurrent-mark interval alone is 1.7--1.9 seconds.
The process remains CPU-active for 44.5 seconds, so this is neither a launch
failure nor an output verifier failure.

This agrees with the earlier reduced Sudoku profile's allocation-heavy nested
Array path, but does not establish the cause. It could be normal work in this
solver, bytecode object/ArrayStore lifetime across loop iterations, or their
interaction. The prior final-GC ArrayStore checks do not answer that
within-one-`main` lifetime question. In contrast, compiled prefixes validate
through ten puzzles, making compiler output correctness and the source parser
unlikely explanations for the current status gap.

Do not special-case Sudoku, masks, MRV, nested arrays, a loop shape, or a
solver helper. No `able-stdlib` change is warranted.

## Next recommendation

Measure live bytecode ownership over repeated allocation in two independently
shaped controls: a generic nested-Array construction/release fixture and an
existing non-Sudoku Array-heavy benchmark. Why: the two-puzzle trajectory is
large enough to merit investigation, but one solver cannot establish a shared
runtime defect. The work entails an opt-in bounded live-store/heap probe,
output checks for both controls, and comparison before and after each loop
iteration. Consider a runtime change only if the same ownership or retention
edge recurs in both; otherwise retain Sudoku as timeout/status evidence.
