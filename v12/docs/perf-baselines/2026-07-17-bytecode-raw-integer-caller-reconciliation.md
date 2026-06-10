# Bytecode raw-integer caller reconciliation — 2026-07-17

## Decision

Keep direct extraction of the dominant built-in `i32` and `i64` carriers in
`bytecodeDirectSameTypeSmallIntPair(...)`. The helper now handles raw i32 slot
values/stack cells and raw i64 result values/slot cells, including exact-kind
boxed small integers, before using the canonical all-integer fallback.

This is a runtime-carrier optimization at an existing same-type integer
boundary. It does not recognize an application, benchmark, nominal type, or
stdlib container. Exact suffix checks preserve mixed-width and signedness
semantics. Rare signed widths, unsigned carriers, wide integers, return
scratch, and generic raw cells remain on `bytecodeRawIntegerValueInfo(...)`.
No compiler, stdlib, language, or fixture source changed.

## Caller and carrier census

Fresh bounded profiles attributed the shared raw-extraction wall to different
callers, so temporary counters recorded calls, hits, carriers, and misses
before changing production code.

| Workload | Pair calls/hits | Direct-value calls/hits | General-integer calls/hits | Store calls/hits |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 786,432 / 786,432 | 1,108,032 / 714,816 | 802,824 / 786,440 | 401,884 / 393,317 |
| Run-length encode | 48 / 48 | 1,919,904 / 0 | 196 / 196 | 2,250,231 / 174,995 |
| Unicode Scalar Pipeline | 5,308,416 / 5,308,416 | 12,386,304 / 10,616,832 | 4 / 4 | 10,616,888 / 7,077,891 |
| Temporary custom nominal `Eq` | 0 / 0 | 524,288 / 0 | 0 / 0 | 2 / 2 |
| Iterator Collect | 8,004 / 8,004 | 124,000 / 124,000 | 60,012 / 60,012 | 148,018 / 60,014 |
| Numeric Array Map | 78,006 / 78,006 | 84,000 / 84,000 | 84,028 / 84,028 | 108,040 / 108,033 |

The pair helper is the only audited boundary with a 100% hit rate across four
unlike material consumers. Boolean and Unicode are dominated by raw i64 result
and slot cells; Numeric Array Map adds raw i32 values; Iterator also uses boxed
small integers. By contrast, the direct-value and store helpers receive
millions of intentional non-integer values in text workloads. They were not
globally prefixed with another switch.

`compareBytecodeCondition(...)` was also audited because its fallback can
repeat integer comparison work. It was reached only by Run-length in this
cohort: 959,952 calls, all ordinary fallbacks, with zero direct comparison,
direct-value, integer-fast, or float-fast hits. It therefore failed the
three-unlike-consumer admission rule. All counters and diagnostic output were
removed.

## Rejected alternative

A shared common-carrier helper called twice by the pair path was fully
reverted. Across five independent processes per side it regressed Boolean from
550.461 ms to 582.100 ms (5.75%) and Unicode from 4.3682 s to 4.7097 s (7.82%),
despite improving Numeric Array Map from 85.208 ms to 68.444 ms (19.67%). The
extra helper call and second type-switch layer cost more than they saved in the
i64-heavy programs.

The retained form performs the dominant carrier switch inline exactly once at
the existing pair boundary, then falls through to the unchanged extractor for
all other cases.

## Repeated workstation gate

Every timing is an independent process with one warmup and one measured call,
the canonical external `able-stdlib`, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, CPU 0, and skipped benchmark typechecking. Order alternates
between binaries. Every valid workstation outlier remains in the arithmetic
mean; volatile Boolean, Numeric Array Map, and Iterator cohorts were expanded.

| Workload | Samples/side | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 10 | 566.722 ms | 568.070 ms | 0.24% slower; neutral |
| Unicode Scalar Pipeline | 5 | 4.9653 s | 4.5764 s | 7.83% faster |
| Numeric Array Map | 15 | 87.209 ms | 89.975 ms | 3.17% slower |
| Run-length encode | 5 | 1.5111 s | 1.5331 s | 1.46% slower |
| Temporary custom nominal `Eq` | 5 | 460.739 ms | 420.256 ms | 8.79% faster |
| Iterator Collect | 25 | 493.253 ms | 514.174 ms | 4.24% slower |
| String Split/Join | 5 | 1.2100 s | 1.1909 s | 1.58% faster |
| Bounded Reverse Complement | 5 | 1.475 ms | 1.286 ms | 12.83% faster |

The nominal and Reverse Complement improvements contain large baseline
outliers and are treated only as passing controls, not claimed as attributable
wins. Numeric Array Map's initial five-sample result was 5.21% slower; keeping
all samples and expanding to fifteen moved the mean to 3.17% slower. Iterator
moved from 3.81% slower at five samples to 4.66% at fifteen and 4.24% at
twenty-five. No row crosses the 5% broad-regression guard.

An initial zsh timing loop accidentally treated the two-word order as one
candidate-only label. It was stopped, produced no side comparison, and is
excluded from this table. The corrected loop used explicit bash arrays.

## Mechanism and correctness

Post-change CPU sampling confirms that work left the intended boundary:

- Unicode's pair helper falls from 2.13% to 1.04% cumulative, while the total
  raw extractor falls from 5.54% to 4.79% flat.
- Numeric Array Map's pair helper falls from 5.88% to 1.41% cumulative in a
  longer candidate sample; total raw extraction falls from 14.71% to 9.86%
  flat.

Focused same-type pair, direct comparison, raw carrier, mixed-width,
signedness, bitwise, overflow, coercion, and runtime tests pass. Regression
coverage includes raw i32 and i64 carrier pairs and an i64/u64 mismatch. The
rare raw-u64 fallback is also locked. The shared worktree's previously recorded
unrelated fixture failures were not attributed to this change. Changed Go
files remain below 1,000 lines and `git diff --check` passes.

## Next recommendation

Profile and reconcile integer slot-store ownership, beginning at
`tryStoreRawIntegerSlotValue(...)` and its opcode callers, but segment calls by
lowered target type and value category before proposing a fast path.

Why: store work is material in Boolean, Unicode, and Numeric Array Map, but the
same helper also receives 2.08 million non-integer values in Run-length and
3.54 million in Unicode. A global type-switch reorder would repeat the failed
miss-heavy strategy. The next tranche should identify whether lowering already
knows an exact primitive integer target at three unlike hot store sites, then
route only those proven sites to the existing raw store operation. It entails
fresh opcode/caller profiles, temporary typed-target and carrier counters,
alias/reuse tests for i32/i64 plus mixed signed/unsigned fallbacks, and the same
expanded repeated controls. Continue to defer WASM.
