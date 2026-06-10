# Benchmark semantic-work equivalence normalized

Date: 2026-07-26

Decision: retain benchmark/reference normalization for Tapelang Alphabet and
Sudoku Masks. This tranche changes no compiler, generated runtime, interpreter,
VM, stdlib, language, dependency, or WASM code. The resulting ratios are
benchmark corrections, not compiler speedups.

## Why normalization was required

The preceding audit showed that the two largest selected native misses did not
compare equivalent source work:

- Tapelang used a flat jump-table interpreter in Able and a recursive
  instruction tree in Go.
- Sudoku allocated a dynamic three-element Able Array for every improved
  candidate position while Go returned three scalar values.

The implementation audit also found a more serious Tapelang problem in the old
Go reference: its parser declared a zero-valued `INC` opcode and appended an
operation even when the input byte matched no instruction. The 275-byte input
contains 82 ignored comment/whitespace bytes. Able correctly ignores those
bytes; the old Go reference turned them into increments. The earlier inferred
Able dispatch count therefore inherited operations from a semantically
different reference and is superseded by this record.

## Retained representation contracts

Tapelang now uses the same representation and execution model in both
languages:

- three flat `i32`/`int32` vectors for opcode, operand, and jump target;
- one instruction-pointer dispatch loop;
- paired bracket jump targets;
- identical ignore-unknown-token parsing;
- identical growable integer tape and buffered byte output.

An instrumented copy of the retained Go reference measured the operations that
the structurally identical Able loop performs:

| Operation | Count |
| --- | ---: |
| increments | 577,783,636 |
| moves | 57,780,404 |
| inputs | 0 |
| prints | 27 |
| loop starts/tests | 288,891,814 |
| loop ends/backedges | 288,891,813 |
| total dispatches | 1,213,347,694 |

Sudoku now uses slice-backed `i32` boards and masks in both languages. Each
recursive solve call creates one reusable three-field position record and
passes it to `find_best_empty` as an output parameter. Candidate improvements
mutate that record rather than allocating a new dynamic Array.

The search work remains unchanged and agrees across the paired sources:

| Operation | Count |
| --- | ---: |
| solve/find calls | 1,918,450 |
| scanned cells | 155,394,450 |
| empty cells | 64,090,010 |
| bit-count loop iterations | 166,923,250 |
| best-position updates | 3,893,780 |
| tried choices | 1,918,350 |
| backtracks | 1,912,510 |

This removes the prior 7,787,560 source-chosen heap allocations: one Array
carrier and one backing slice for every best-position update. Go escape
analysis reports `best does not escape`; generated-Able escape analysis reports
the hot `&Position{...}` does not escape. The generated conversion helpers for
dynamic boundaries still exist, but the hot direct calls do not invoke them.

The external benchmark's `sudoku-masks/run.able` is synchronized with the
canonical v12 source so its Docker lane uses the same retained representation.

## Strict boundary and correctness evidence

- Both canonical sources compile with `--no-fallbacks`.
- Each final generated graph has 96 packages and omits `pkg/interpreter`.
- The extracted Tapelang execute and Sudoku find/solve hot functions contain
  no `runtime.Value`, bridge, wrapper, thunk, or interpreter calls.
- Feature coverage, operation-depth coverage, and the complete benchmark
  catalog checks pass without metadata changes.
- Both public verifiers accept every measured process. Tapelang stdout is
  `a8ac3a10...766d149`; Sudoku stdout is `35a81e44...659ec`.

Fib remains unchanged as the equal-algorithm compiler-effect sentinel.

## Repeated performance

The authoritative project harness rebuilt both references with Go 1.26.4 and
both Able binaries with `--no-fallbacks`, pinned every serial process to CPU 0,
and verified five processes per lane:

| Benchmark | Able mean | Go mean | Able/Go | Target |
| --- | ---: | ---: | ---: | --- |
| Tapelang Alphabet | 3.7060s | 2.9649s | 1.2500x | miss |
| Sudoku Masks | 1.5740s | 0.7027s | 2.2399x | miss |

All 20 processes verified. Both Able rows reported zero GC cycles. A separate
alternating five-pair cohort also verified all 20 processes and measured
1.2218x for Tapelang and 2.1379x for Sudoku. The independent cohort confirms
the classification despite normal workstation noise.

The changed Go references and canonical Sudoku Able source invalidated old
scoreboard fingerprints, as intended. The two compiled rows were promoted into
a reconciled full compiled source report and the current external scoreboard
was regenerated. Sudoku's unselected bytecode status row was also repeated
under its existing one-run, 55-second contract; it remains a timeout. The
current scoreboard now validates against the normalized sources.

For context, the preceding unequal-work cohort measured 1.9439x for Tapelang
and 2.9014x for Sudoku. The ratio reductions are not compiler improvements:
Tapelang's equivalent flat Go reference does more appropriate interpreter
work, and Sudoku's Able source no longer chooses millions of heap allocations
while the Go source now uses comparable dynamic carriers.

Neither normalized row meets the 1.052632x Go target. Their remaining excess is
now eligible for compiler investigation because the known semantic-work and
compiled/interpreted-boundary confounders are closed.

The complete `./run_all_tests.sh` handoff passes every scoreboard/coverage
contract, all non-compiler packages, all 32 compiler batches, and the final
bytecode fixture corpus (80.691s).

Machine-readable evidence:
`2026-07-26-benchmark-semantic-work-equivalence-normalized.json`.
The promoted measurements are preserved in
`2026-07-26-semantic-equivalence-normalized-{go-reference,compiled}.json`;
the reconciled full source reports are
`2026-07-26-semantic-equivalence-normalized-full-{go-reference,compiled}.json`
and
`2026-07-26-semantic-equivalence-normalized-full-bytecode-status.json`.

## Next

Profile the normalized Tapelang and Sudoku binaries alongside the unchanged
equal-algorithm Fib sentinel. Compare generated source, assembly, CPU profiles,
and exact allocation evidence against Go, and advance only one exact general
generated-code rule that is material across all three unlike applications.

Why: these applications now perform equivalent work and their strict graphs
never cross into the interpreter. Any repeated residual owner is therefore a
credible native-lowering cost rather than a benchmark algorithm artifact.

What it entails: freeze current exact binaries, collect repeated profiles and
call/check/allocation counts, intersect concrete generated symbols or emitted
patterns, then run verifier-backed five-or-more-pair A/B/Go measurements if a
general candidate clears the three-program gate. Retain no code if no shared
owner clears it.

Why it is important: this is the shortest evidence-backed route from normalized
applications to the mission of matching native Go performance without adding
benchmark-specific, named-container, non-primitive nominal, or boundary ABI
special cases.
