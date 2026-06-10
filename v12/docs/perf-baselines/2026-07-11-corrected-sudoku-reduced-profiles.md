# Corrected Sudoku Reduced-Workload Profiles

## Decision

Keep no runtime, compiler, canonical-stdlib, or benchmark-source performance
change. The corrected Sudoku source has a real bytecode cost, but its
dominant work does not repeat as one material generic VM leaf or generated
lowering boundary in independently shaped typed-array and ordinary-recursion
controls.

In particular, do not add a Sudoku, nested-Array, bit-mask, MRV, `Array
i32`, QuickSort, or recurrence-specific optimization. The existing Fibonacci
recurrence kernel was measured only to reject it as an ordinary-recursion
control; it is not evidence for a new change.

## Method

The disposable driver at
`v12/tmp/sudoku-reduced-profile-2026-07-11/sudoku_once.able` copies the
corrected benchmark's parsing, mask, solving, and board formatting paths. It
uses the first three fixed corpus inputs. The fourth input was excluded after
it exhausted a non-bounded bytecode calibration process before completion;
the retained runs use `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.

One bytecode process ran the three boards once and completed in 12.22 seconds
(11.98 seconds of CPU samples); all three emitted solutions. A second,
identical bounded process produced the cumulative allocation profile.

The compiler profile used the same three-board workload, repeated twenty
times solely to accumulate samples. It completed in 3.26 seconds and
produced 3.08 seconds of registered-main CPU samples. The source driver is
restored to one pass after collection.

Steady-state bytecode controls used the normal runtime benchmark harness with
the same memory/CPU guard:

| Control | Sampling interval | Reading |
| --- | ---: | --- |
| `array_map_i32_small` | 11.41 CPU s | 61,862,112 ns/op, 846,519 B/op, 300 allocs/op |
| `quicksort_file_small` | 12.00 CPU s | 9,514,674 ns/op, 458,431 B/op, 16,938 allocs/op |
| `fib_i32_small` diagnostic only | 11.78 CPU s | 2,313 ns/op, 416 B/op, 3 allocs/op; the existing i32 recurrence kernel executes it |

The retained profiles and benchmark output are under
`v12/tmp/sudoku-reduced-profile-2026-07-11/`.

## Evidence

### Bytecode Sudoku

The reduced Sudoku profile is allocation/GC heavy around generic execution:

- `runResumable` is 10.65 s cumulative (88.9%).
- `execCallOpcode` is 4.43 s cumulative, while `execBinary` and
  `execStoreSlot` are only 1.22 s and 1.04 s.
- `ArrayStoreNew` is 1.30 s cumulative, and allocation attribution is 495.08
  MB cumulative: `ArrayStoreNew` 144.62 MB, tracked-array registration 62.60
  MB, Array value views 51.50 MB, capacity growth 44 MB, and raw-i32 cache
  length growth 15 MB.

Those allocations arise from the source's nested tracked Arrays, the
temporary position Array created while searching cells, and normal
Array/method semantics. They are not a single missing VM operation.

### Typed Array and ordinary recursion controls

The Array/map control repeats some broad infrastructure but has a different
CPU balance: raw integer value extraction is 0.49 s flat, Array slot calls
1.90 s cumulative, Array get 1.20 s, and inline-call setup/return about 1.0
s each. Its retained heap is dominated by ordinary Array capacity and
raw-i32 cache storage.

QuickSort instead concentrates on mono primitive array reads and index
checks: `arrayReadSlotValue` 2.74 s cumulative,
`readArraySlotValueFastChecked` 3.19 s, and the fused array-read comparison
jump 2.71 s. It shares `bytecodeRawIntegerValueInfo` (0.73 s flat), but the
Sudoku profile has only 0.16 s there. Call completion and return are also
small in both relative to their workload-specific Array paths.

Thus raw integer extraction and generic call/return are real shared
infrastructure, but neither is a common dominant wall. Changing them from
this evidence would optimize a small fraction of two controls and risk the
same narrow regression pattern rejected in previous tranches.

### Compiled Sudoku

Generated compiled code has no material compiler bridge/lowering helper in
the hot path. `solve_with_masks` is 2.74 s cumulative (89.0%), led by the
source function `find_best_empty` at 2.71 s cumulative. The latter also
drives allocation (`growslice` 1.08 s cumulative and `mallocgc` 1.53 s).
The next leaves are source arithmetic helpers, `square_index` (0.27 s
cumulative) and `bit_count` (0.22 s).

This is an algorithm/data-structure cost in the benchmark's own generic Able
source, not a repeated generated-runtime boundary. Altering the source merely
to improve its score would not establish a generally faster compiler.

## Next recommendation

Refresh the validated external scorecard's remaining comparable miss families
with bounded profiles, beginning with I-Before-E and Base64 as independent
text/byte controls and a scalar numeric guard. Why: corrected Sudoku now
rules out a shared array/call/compiler leaf, while the scorecard still has
valid bytecode misses in text and codec families. The work entails one
process profile per completed workload under the same guardrails, followed by
a candidate only when the same concrete helper and caller recur in at least
two different families with an unrelated regression guard. Sudoku remains
timeout/unverified until it completes the unchanged external verifier.
