# Sudoku Masks and TapeLang source-equivalence reconciliation

Date: 2026-07-20

## Decision

Retain one benchmark-source correction and no compiler, VM, generated-runtime,
or canonical-stdlib optimization. `sudoku_masks` now constructs its fixed-size
best-position result only when a candidate actually improves the current best.
The previous Able source constructed and discarded a three-element `Array i32`
for every empty cell examined, unlike all three current reference programs.

The correction is present in the canonical v12 source and the external
benchmark mirror. `Array.with_capacity(3)` also records the known result size
instead of growing the backing store three times. The external suite README now
makes the best-result replacement rule part of the algorithm contract.

TapeLang remains unchanged. Its implementations use different host-level
dispatch representations, but the differences favor Able and therefore cannot
explain Able's remaining deficit.

## Sudoku equivalence finding

All four implementations use the same bit-mask, most-constrained-cell search:

- scan the same first ten puzzles for ten passes;
- maintain row, column, and square masks;
- count available bits with `value &= value - 1`;
- choose the empty cell with the lowest candidate count; and
- recurse through candidate digits in ascending order.

Python constructs its result tuple and Ruby its result Array only inside
`count < best_count`. Go updates four scalar locals in that branch. Able
previously allocated its result before the branch, so non-improving cells
performed allocation work absent from every reference. The retained source
change moves creation into the branch while preserving result type, ordering,
solver behavior, and output.

## Allocation and CPU effect

Before the source correction, a verified sampled compiled profile reported
2.89 GB and 160,532,091 whole-process objects. Generated
`find_best_empty` owned 98.77% of bytes and 99.68% of objects. Exact allocation
instrumentation could not finish inside 55 seconds.

After the correction, exact instrumentation completes and reports
156,370,688 main-phase bytes, 7,802,594 allocations, and 11 GCs. Relative to
the pre-fix sampled totals, observable traffic falls about 94.6% in bytes and
95.1% in objects. The measurement methods differ, so these percentages are
directional; both identify the same removed source owner.

The current verified main-only CPU profile contains 1.72 seconds of samples:

| Owner | Flat | Cumulative |
| --- | ---: | ---: |
| generated `find_best_empty` | 25.58% | 86.05% |
| generated `square_index` | 24.42% | 37.79% |
| generated `bit_count` | 13.37% | 13.37% |
| checked signed multiply | 5.81% | 6.40% |
| checked signed shift | 3.49% | 4.65% |
| `runtime.mallocgc` | 0.58% | 9.30% |

The old allocation/GC wall is gone. Remaining costs are source search work and
distinct checked primitive helpers, not a general Array-allocation candidate.

## Repeated compiled evidence

Two independent five-process Able cohorts reused one generated binary per
application. Two independent five-process Go cohorts rebuilt and reused the
current reference binaries. Every one of the 40 outputs passed the public
verifier, and all samples are retained in the combined statistics.

| Application | Able mean | Able median | Able CV | Go mean | Go median | Go CV | Able / Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Sudoku Masks | 1.9080 s | 1.9000 s | 4.69% | 0.6536 s | 0.6238 s | 13.24% | 2.9193x |
| TapeLang Alphabet | 3.9780 s | 3.8750 s | 10.10% | 1.9322 s | 1.9532 s | 4.80% | 2.0588x |

Sudoku's preceding current-source five-run mean was 9.566 seconds. The aligned
ten-run mean is 80.1% lower. This is an algorithm-equivalence correction, not
a compiler speedup claim.

The formal scorecard replacement also reran the unchanged QuickSort control
and a third five-process Sudoku compiled cohort. It records QuickSort at
1.896 seconds versus 2.6422 seconds for Go and Sudoku at 1.788 seconds versus
0.6100 seconds for Go (2.93x). Both rows verified five times.

## Interpreter status

One ordinary aligned Sudoku bytecode process again reached the 55-second cap.
The source correction is therefore not enough to admit Sudoku to the selected
bytecode cohort. It remains a visible unranked `able_timeout` row rather than a
missing result or an inferred ratio.

Two independent reference cohorts completed and verified:

| Reference | Ten-run mean | Median | CV |
| --- | ---: | ---: | ---: |
| Python 3.14 | 17.9866 s | 17.8458 s | 3.38% |
| Ruby 4.0 | 22.0246 s | 21.8866 s | 2.80% |

The scorecard status replacement uses one bounded QuickSort probe and one
bounded Sudoku probe because neither bytecode row is selected or completes.
The five-run evidence rule still holds for all 65 selected rows; timeout rows
are status, not timing claims.

## TapeLang equivalence audit

All four current programs parse the same checked-in 275-byte source and emit
the same 27-byte reverse alphabet plus newline. The public verifier hash is
`a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149`.

Able builds flat parallel kind/value/jump arrays and buffers output. Go,
Python, and Ruby build recursive operation trees and emit output one character
at a time. Able and Go start with an unused cell before the program's leading
move; Python and Ruby start one cell earlier. The program never observes that
absolute offset. All use the same relative moves and cell updates, and none
executes input.

A neutral compiled counter over the flat semantic program reports 193 parsed
operators and 953,344,873 executed dispatches:

| Operation | Executions |
| --- | ---: |
| `+` | 288,891,823 |
| `-` | 288,891,813 |
| `<` | 28,890,202 |
| `>` | 28,890,202 |
| `[` | 28,888,993 |
| `]` | 288,891,813 |
| output | 27 |
| input | 0 |

Recursive trees and per-character output add host work to the references;
Able's flat jumps and buffered write remove that work. Aligning references to
Able would make the comparison harder for Able, while changing Able to the
recursive representation would deliberately discard an idiomatic improvement.
The current 2.06x compiled deficit is therefore not an artifact that favors
Able's competitor. No Tape source change is justified.

## Scorecard reconciliation and verification

The prior QuickSort/Sudoku compiled and bytecode source partitions were
replaced as units so no unrelated row disappeared. The promoted scorecard now
records:

- Sudoku compiled: verified, 1.788 seconds, 2.93x Go, target miss;
- Sudoku bytecode: one bounded timeout, unranked;
- Tape compiled: unchanged current row, 3.770 seconds, 1.96x Go, target miss;
  and
- Tape bytecode: unchanged timeout status, unranked.

Verification:

- `just bench-catalog-check` passes with 36 portable applications, 79 bounded
  fixtures, and 115 combined programs;
- `just bench-selection-check` passes with 65 selected rows;
- `just bench-scoreboard-check` passes after source-fingerprint reconciliation;
- strict evidence passes with five successful Able/reference samples for all
  65 selected rows and 72 full-status rows;
- all repeated compiled and interpreter-reference samples passed their public
  verifiers;
- exact aligned allocation and current CPU-profile outputs passed the Sudoku
  verifier; and
- no WASM work was performed.

## Next recommendation

Run a bounded post-correction bytecode attribution gate for Sudoku Masks using
a reduced but semantically identical workload, then reconcile its exact leaves
against current unlike-program profiles before admitting any VM candidate.

Why: the full bytecode program still exceeds 55 seconds, while the source
correction removed about 95% of compiled allocation traffic and completely
changed the compiled hot distribution. The retained reduced Sudoku bytecode
profile predates this correction and is no longer a reliable picture of the
current program. The compiler now ends in source-level `square_index`,
`bit_count`, checked multiply, and checked shift; the bytecode VM may expose a
general primitive, call, or Array leaf beneath those operations, but one
application alone cannot authorize a change.

What it entails: derive the reduced input only through iteration/puzzle-count
parameters or a diagnostic copy of the same algorithm; collect warmed main-only
CPU, exact allocation, opcode, and call-site evidence under 55-second process
caps; and compare concrete descendants with current integer/control workloads
such as TapeLang's bounded fixture and one unrelated selected numeric or search
application. Advance a candidate only if the same non-nominal leaf is material
in at least three unlike programs, then gate it with repeated full-program
averages and selected bytecode controls. Do not add Sudoku, Tape, Array-shape,
or named nominal special cases, and do not begin WASM.
