# BinaryTrees and Sudoku Correctness Diagnostics

## Decision

Keep no compiler, bytecode-VM, tree-walker, or canonical-stdlib change in
this diagnostic tranche. The two failed external validations have independent
causes in the Able benchmark programs (one exposes a shared parser-semantic
gap); neither is evidence for a performance optimization or a nominal-type
lowering rule.

The retained reproductions are under
`v12/tmp/benchmark-correctness-diagnostics-2026-07-11/`.

## Method

Used the same pinned process guardrails as the validated generality pass, with
the external suite directories as the process working directories and their
unmodified Ruby verifiers:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  ./v12/bench_perf --runs 1 --timeout 60 --modes compiled \
    --executor goroutine --cpu-affinity 2 \
    --run-from ../benchmarks/binarytrees \
    --verify-ruby-script ../benchmarks/binarytrees/verify.rb --keep \
    --workdir v12/tmp/benchmark-correctness-diagnostics-2026-07-11/binarytrees \
    v12/examples/benchmarks/binarytrees.able

GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  ./v12/bench_perf --runs 1 --timeout 60 --modes compiled,bytecode \
    --executor goroutine --cpu-affinity 2 \
    --run-from ../benchmarks/sudoku \
    --verify-ruby-script ../benchmarks/sudoku/verify.rb --keep \
    --workdir v12/tmp/benchmark-correctness-diagnostics-2026-07-11/sudoku \
    v12/examples/benchmarks/sudoku/sudoku.able
```

The BinaryTrees bytecode process was also run only until it emitted its first
program line. The bounded tree-walker process did not reach that line within
20 seconds, so it is not reported as a completed execution result.

## BinaryTrees: interpolation escape decoding

The compiled program computes every required depth, iteration count, and tree
check correctly. Its complete 11-line stdout has hash
`50973ed148a92229db8201c391857024375cde9e3d214c5e4f60961114bb67b5`.
The sole difference is the separators:

| Location | Bytes |
| --- | --- |
| Ruby verifier expectation | tab: `09` |
| Able compiled stdout | backslash then `t`: `5c 74` |

For example, the actual first line is `stretch tree of depth 22\t check:
8388607`; all of its numeric text is otherwise equal to the verifier's
expectation. The bytecode process emits the same `5c 74` bytes before its
bounded long-running portion.

`binarytrees.able` uses a backslash-plus-`t` escape in backtick interpolation
text. The shared Go parser's `unescapeInterpolationText` currently decodes
only escaped backticks, dollar signs, and backslashes; it deliberately
preserves backslash-plus-`t`. The generated compiled program therefore
contains those same two separator characters. This is a parser/AST
semantic boundary shared by all execution modes, not a compiled-code result or
a future/task correctness problem. Section 6.1.5 of the v12 spec says ordinary
double-quoted strings support the standard character escapes, while its
backtick interpolation wording only calls out escaped backticks and dollars.
The next repair must settle and document the general interpolation-escape rule,
rather than adding a BinaryTrees output exception.

## Sudoku: empty-cell conversion in the benchmark program

Compiled and bytecode executions are byte-for-byte identical:

| Mode | Lines | Bytes | SHA-256 |
| --- | ---: | ---: | --- |
| compiled | 1,010 | 146,410 | `db3a4e6c4d0fc33f343909d555abd2f86cd97273b94f1f01ce6e2f0192d9bd7d` |
| bytecode | 1,010 | 146,410 | `db3a4e6c4d0fc33f343909d555abd2f86cd97273b94f1f01ce6e2f0192d9bd7d` |
| external expected solution | 1,010 | 82,820 | `70f621ee8eb896bb208139b31d46eca3c081fe9bdb4b55275e3c20b6db1bed29` |

The first actual line starts `6-279-213...`, whereas the external input starts
`6.79.13.4...`. This follows directly from `parse_board`: it converts every
input byte with `(byte as i32) - 48`. A digit byte is handled correctly, but a
period is byte 46 and becomes `-2`. `find_empty` searches only for `0`, so it
finds no vacant cell; `solve` immediately returns `true` and the benchmark
prints the original puzzle with each period rendered as `-2`.

The retained generated compiled source confirms the same generic `u8` match,
cast, and subtraction. Since compiled and bytecode have the same result, this
is a benchmark-program input-decoding defect, not a divergence in byte
iteration, matching, casts, or solver recursion. The external input, setup,
expected solution, and verifier remain unchanged.

## Next repair boundary

First make the interpolation escape rule explicit in the v12 spec, then teach
the shared parser to decode the ordinary escaped control characters inside
interpolation text and add parser plus cross-mode regression coverage. That is
a language-wide semantic repair and will correct BinaryTrees without a
benchmark-shaped compiler branch.

Separately correct Sudoku's source-level parser so `'.'` maps to the solver's
empty-cell value `0` before applying the digit conversion. Add a focused
benchmark regression and rerun its unmodified external verifier in compiled
and bytecode modes. This is a correctness repair to the benchmark program,
not a runtime optimization. Only after both rows validate should they be
reintroduced to performance comparisons and the full generality scorecard.
