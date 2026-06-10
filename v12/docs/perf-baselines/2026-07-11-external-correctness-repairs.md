# BinaryTrees and Sudoku Correctness Repairs

## Decision

Keep the shared interpolation semantics repair and the generic compiler
label-reference correction. Keep the Sudoku benchmark-source correction, but
do not reintroduce Sudoku to timing comparisons yet: both Able modes still
exceed the 60-second guarded external-run limit.

No named-container lowering, benchmark-specific runtime branch, verifier
change, external-corpus edit, or canonical-stdlib change was made.

## Repairs

### Shared interpolated-string escapes

Backtick interpolation text now decodes the standard string escapes for
newlines, carriage returns, tabs, backspace, form feed, backslash, quotes,
slash, and Unicode. It retains its additional escaped backtick and dollar
forms. The decoder is shared with ordinary quoted literals; malformed
interpolation escapes now produce a parser error rather than silently
preserving a different meaning. The v12 spec sections 6.1.5 and 6.6 now state
that contract explicitly.

The executable interpolation fixture exercises the control escapes, fixed and
braced Unicode escapes, backslash, escaped dollar, and escaped backtick across
the tree-walker and bytecode VM. Focused parser and fixture tests pass.

The repaired BinaryTrees compiled run passes its unchanged public verifier:

| Mode | Validation | Stdout SHA-256 | Guarded real time |
| --- | --- | --- | ---: |
| compiled | verified | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 30.4100 s |

Its 11 output lines now contain tab byte 09, rather than the rejected
backslash-plus-t sequence.

### Exact compiler loop-label matching

While shaping the Sudoku source, a valid nested-loop form exposed an unrelated
compiler issue: generated break to temporary label 12 was treated as a
reference to temporary label 1 by prefix matching, producing an unused Go
label. The compiler now recognizes only an exact generated-label token.
Focused compiler tests cover both the false-prefix case and valid delimited
references.

### Sudoku input and solver

parse_board now maps a period to the solver empty-cell value zero before digit
conversion. The initial naive first-empty-cell backtracker consequently
performed the intended work but timed out. To keep the Able program
algorithmically comparable to the external exact-cover solver without a
runtime special case, the benchmark now uses ordinary Able bit masks and
minimum-remaining-values selection for rows, columns, and 3-by-3 squares.
It still solves the same boards with the same numeric output; the corpus has
one 81-character solution per puzzle.

Under the existing 60-second guard, the corrected bit-mask source emits a
byte-for-byte equal prefix of the external expected solution before timing
out:

| Mode | Captured lines | Prefix result | SHA-256 |
| --- | ---: | --- | --- |
| compiled | 497 | first 497 expected lines equal | 8232944dd62a00d64cef300834dfd744f8f94daa4f13645534f41ed6315882b9 |
| bytecode | 4 | first 4 expected lines equal | 6fbd0ebbe0a2e92cdbb349e07f5d19130281d5f8895b60416b6ce5b85849a5df |

The full verifier is deliberately reported as not run after timeout, not as a
pass. The prior naive source emitted only 209 compiled lines in the same
window; the bit-mask version improves real valid-work progress, but neither
mode completes the required 1,010 lines inside the scorecard bound.

## Verification

Passed:

- Focused compiler label-reference tests.
- Focused parser interpolation-escape tests.
- Tree-walker and bytecode executable interpolation fixture tests.
- BinaryTrees compiled public-verifier run under the 60-second external guard.

The retained build output, captured stdout, JSON summaries, and prefix
comparisons are under
v12/tmp/benchmark-correctness-repair-2026-07-11/.

## Next recommendation

Profile the corrected Sudoku program at a bounded workload and compare the
result with other typed i32 Array and recursive-call workloads before making
any runtime change. The 60-second prefix rates show a large general
array-access/mutation, recursive-call, and bitwise-control cost in both
compiled and bytecode modes, but do not yet identify one shared leaf.

The work entails one-process CPU and allocation profiles for the compiled and
bytecode paths using a deterministic reduced driver, plus controls such as
numeric Array-map and QuickSort. Only a repeated generic helper or lowering
boundary may become an optimization candidate; Sudoku must stay
timeout/unverified in the broad scorecard until it completes the unchanged
external verifier.
