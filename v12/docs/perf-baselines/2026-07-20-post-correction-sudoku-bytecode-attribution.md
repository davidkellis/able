# Post-correction Sudoku bytecode attribution

## Decision

Keep no bytecode VM, compiler, runtime, canonical-stdlib, fixture, or
benchmark-source change from this tranche. The corrected Sudoku workload no
longer has the discarded three-element Array allocation wall, but its remaining
VM time does not expose one new material exact leaf shared with both an
array/member-heavy TapeLang control and an unrelated recursive integer/Array
QuickSort control.

The common work is dispatcher and operand/call transport already covered by
recent rejected broad gates. The concrete hot handlers split into tracked
`Array.get` plus bitwise search in Sudoku, named-struct/member plus Array-slot
dispatch in TapeLang, and fused Array-read comparisons in QuickSort. A new
raw-integer, call-frame, return, generic Array, benchmark, or named-nominal
shortcut is therefore not admitted.

## Workloads and bounds

The Sudoku source is the corrected canonical program, unchanged for this
measurement. Only its input corpus was bounded: the first canonical puzzle was
copied to a temporary `sudoku/sudoku.txt`; the program's existing ten-iteration
loop remained intact. A Ruby verifier checked all ten solutions, their givens,
and every row, column, and square. TapeLang and QuickSort used the checked-in
small benchmark fixtures unchanged.

| Workload | Source SHA-256 | Role |
| --- | --- | --- |
| corrected Sudoku Masks | `eef79a0bb49abd7533d23c4a749d6fa1540688d48b364a1f95ad8d4da4c1914b` | newly invalidated profile owner |
| TapeLang small | `583279a9e4900dc11f130a0fd610e1358e66a90088fa47f9daba96056721833b` | array/member and interpreter-shaped control |
| QuickSort file small | `2e40f4a3dbeb15c5dfe5b665aba2bba9a4448b177cdf5f8298bc774099081d15` | unrelated recursive integer/search control |

Every process used CPU 0, `GOMAXPROCS=1`, `GOGC=50`, a 1 GiB Go memory limit,
the canonical external `able-stdlib`, source-root-only loading, and a 55-second
cap. The CPU captures measured warmed `main()` calls only: 100 Sudoku calls,
8,000 TapeLang calls, and 1,000 QuickSort calls. They produced 4.73, 5.69, and
9.16 seconds of CPU samples, respectively. Diagnostic opcode/call traces used
one separate warmed call and are not timing evidence.

## Repeated wall and allocation evidence

Following the workstation sampling rule, Sudoku and QuickSort use five
independent benchmark processes. TapeLang's first five-process cohort contained
one visible scheduling outlier, so the final arithmetic mean uses ten fresh
independent processes without removing that outlier.

| Workload | Processes / calls per process | Mean ns/op | Mean B/op | Mean allocs/op |
| --- | ---: | ---: | ---: | ---: |
| corrected Sudoku | 5 / 10 | 45,322,492 | 849,119 | 10,500 |
| TapeLang | 10 / 500 | 675,796 | 15,362 | 187 |
| QuickSort | 5 / 50 | 8,282,325 | 460,588 | 16,951 |

Sudoku's five ns/op samples span 44,751,188-45,770,302 and QuickSort spans
8,269,109-8,319,424. TapeLang spans 652,568-782,150; the high sample is retained
in the reported mean. Allocation counts are stable across processes (Sudoku
10,500-10,502, TapeLang exactly 187, and QuickSort exactly 16,951).

## CPU leaf reconciliation

The following are the exact VM leaves visible in all three profile tables.
Percentages are flat CPU, not cumulative parent attribution.

| Exact leaf | Sudoku | TapeLang | QuickSort | Admission result |
| --- | ---: | ---: | ---: | --- |
| `runResumable` dispatch body | 12.90% | 8.79% | 9.93% | broad dispatcher/layout family; no isolated semantic operation |
| `appendSlotStackValueChecked` | 3.59% | 2.11% | 0.87% | not material in the third control; earlier generic variants failed broad guards |
| `bytecodeRawIntegerValueInfo` | 1.69% | 2.64% | 4.26% | small and carrier-divergent; raw extraction gates are already closed |
| `pushCallFrame` | 1.69% | 2.28% | 0.76% | not material in QuickSort |
| `popCallFrameFields` | 1.48% | 1.58% | 0.98% | transport-scale only |
| `finishInlineReturn` | 2.11% | 0.88% | 0.66% | not material in either control; recent return guard failed |
| `execBinary` | 1.69% | 0.88% | 0.66% | parent splits into bitwise, comparisons, and arithmetic |

The workload-specific descendants are much clearer:

- Sudoku spends 32.98% cumulative in `execCallOpcode`, 7.19% in its direct
  same-type integer bitwise path, and 11.21% in `execCallMemberArrayGet`.
- TapeLang spends 30.93% cumulative in `execCallMemberArraySlot`, 15.99% in
  general member calls, and 6.85% in `mapaccess2_fast64`.
- QuickSort spends 24.89% cumulative in
  `execJumpIfArrayReadSlotCompareSlotFalse`, 23.69% in checked Array-slot reads,
  and 8.41% in Array-slot index extraction.

These are distinct consumers beneath the shared dispatch/transport shell.

## Exact opcode and call-site evidence

Sudoku executes 334,826 `LoadSlot`, 157,736 `Pop`, 144,893 `Const`, and 119,160
`CallMemberArrayGet` operations per bounded call. Its largest traced site is
the tracked `Array.get` at `sudoku_masks.able:70`, with 77,760 hits; `bit_count`
and `square_index` contribute 11,280 inline calls each.

TapeLang executes 2,658 `LoadSlot`, 1,478 `LoadSlotStructField`, and 1,342
`CallMemberArraySlot` operations. Its largest traced site is a monomorphic i32
`read_slot`, with 278 hits, followed by inline `inc` and Array read/write sites.

QuickSort is led by 31,775 `Jump`, 25,617 `Pop`, 18,097
`StoreSlotBinaryIntSlotConst`, and 13,247 fused
`JumpIfArrayReadSlotCompareSlotFalse` operations. Its trace is dominated by
two monomorphic i32 `read_slot` sites with 7,032 and 6,215 hits, plus one u8
input read with 5,386 hits.

All three already have zero operand-stack capacity growths. Their maximum call
depths are 50, 4, and 24, respectively. The apparent high-level overlap in
loads, calls, and Array access therefore does not identify a common missed
cache, capacity growth, or exact handler.

## Verification

- The bounded corrected Sudoku ordinary-bytecode run completed in 0.38 seconds
  and passed the ten-solution verifier.
- Ordinary bytecode TapeLang printed `2048` and exited successfully.
- Ordinary bytecode QuickSort printed the expected sorted low/high boundary
  values and exited successfully.
- All 20 repeated warmed benchmark processes completed within the cap.
- No production file changed, so no candidate A/B or broad code regression run
  was warranted.

Temporary input, profile, trace, and JSON files were removed after the figures
above were recorded.

## Next recommendation

Refresh corrected Sudoku's compiled CPU profile and compare it directly with
the Go reference plus unrelated compiled QuickSort and Binary Trees controls.
The source correction reduced compiled Sudoku from about 9.57 seconds to about
1.91 seconds and invalidated the old generated-code ownership picture, while
the promoted row still misses Go by about 2.93x. This is now the largest
changed, well-bounded compiler opportunity not already explained by the stale
allocation wall.

The work should preserve one Able and one Go binary per source fingerprint,
collect repeated verifier-backed wall means plus main-only CPU and exact
allocation profiles, inspect generated source at the material leaves, and
admit a candidate only if the same primitive or general nominal-lowering cost
is material in Sudoku and both unlike controls. It must not add Sudoku,
solver-shape, Array-layout, or named-container lowering and must not begin WASM.
