# Bytecode single-pass raw-integer store gate — 2026-07-17

## Decision

Keep a single-pass implementation of
`tryStoreRawIntegerSlotValue(...)`. Untyped slot stores now classify every
supported raw or boxed small-integer carrier directly instead of calling
`bytecodeRawIntegerValueInfo(...)` and traversing a second type switch.
Dominant raw i32 and i64 carriers proceed directly to their existing owned-slot
stores; rare signed/unsigned carriers retain the canonical range check and
generic raw store.

This is a primitive runtime-carrier optimization at the existing untyped slot
boundary. It does not recognize an application, benchmark, nominal type, or
stdlib container. Typed exact stores, i32 opcodes, big integers, non-integer
values, and all ordinary fallback storage retain their existing semantics. No
compiler, stdlib, language, or benchmark fixture source changed.

## Store-site census

Fresh profiles put `execStoreSlot(...)` at 3.63% cumulative in Run-length,
6.65% in Unicode, 7.38% in Split/Join, and a coarsely sampled 11.11% in Numeric
Array Map. Temporary counters then correlated declaration/reassignment,
lowering-time simple type, runtime carrier, and successful raw stores.

| Workload | Calls | Raw-store hits | Dominant successful carriers | Dominant misses |
| --- | ---: | ---: | --- | --- |
| Boolean Reconciliation | 401,884 | 393,317 | 393,216 raw i64 | 8,565 bool/other |
| Run-length encode | 2,250,231 | 174,995 | 154,848 boxed integers; 20,000 raw i32 | 2,074,896 char |
| Unicode Scalar Pipeline | 10,616,888 | 7,077,891 | 5,308,416 raw i64; 1,769,475 other integers | 3,538,944 char |
| Iterator Collect | 148,018 | 60,014 | 30,004 raw i64; 30,010 boxed integers | 88,004 nominal/sentinel values |
| Numeric Array Map | 108,040 | 108,033 | 72,000 raw i32; 36,006 raw i64 | 7 setup values |
| String Split/Join | 2,121,877 | 1,505,265 | 18,006 raw i32/i64; 1,487,259 other integers | bool, String, and nominal/text values |

The proposed lowering-only route failed admission. Nearly all material integer
hits were statically `unknown`; exact `i32` metadata was material only for
6,000 Split/Join stores. `AnyInteger` metadata cannot safely choose a suffix.
The general repeated shape was instead a stable runtime primitive carrier at
the same untyped store boundary.

All diagnostic fields, counters, atomics, output plumbing, and the diagnostic
binary were removed after the census.

## Rejected first candidate

An inline common-carrier switch in `execStoreSlot(...)` improved the
integer-heavy programs but made uncertain nominal values traverse that switch
and then the original extractor switch. Iterator Collect regressed in every
sample and failed the broad gate.

| Workload | Samples/side | Baseline mean | First candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 15 | 528.081 ms | 519.210 ms | 1.68% faster |
| Unicode Scalar Pipeline | 5 | 4.2485 s | 4.0107 s | 5.60% faster |
| Numeric Array Map | 5 | 71.094 ms | 70.059 ms | 1.46% faster |
| Run-length encode | 5 | 1.2893 s | 1.3284 s | 3.03% slower |
| String Split/Join | 5 | 1.0214 s | 1.0402 s | 1.84% slower |
| Iterator Collect | 5 | 445.121 ms | 495.141 ms | 11.24% slower; rejected |

The inline switch was fully reverted. The retained candidate moves the
complete carrier classification into the existing helper, so hits and misses
both traverse exactly one switch.

## Retained repeated workstation gate

Every timing is an independent process with one warmup and one measured call,
the canonical external `able-stdlib`, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, CPU 0, and skipped benchmark typechecking. Order alternates
between binaries. Every workstation outlier remains in the arithmetic mean;
volatile Boolean and Unicode cohorts and the previously failed Iterator
control were expanded.

| Workload | Samples/side | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 10 | 654.879 ms | 669.927 ms | 2.30% slower |
| Unicode Scalar Pipeline | 10 | 4.7897 s | 4.8324 s | 0.89% slower; neutral |
| Run-length encode | 5 | 1.5380 s | 1.4823 s | 3.62% faster |
| Numeric Array Map | 5 | 79.788 ms | 78.170 ms | 2.03% faster |
| String Split/Join | 5 | 1.1588 s | 1.0916 s | 5.80% faster |
| Iterator Collect | 10 | 491.762 ms | 486.855 ms | 1.00% faster |
| Bounded Reverse Complement | 10 | 1.282 ms | 1.277 ms | 0.37% faster; neutral |

The Boolean mean retains 807-896 ms baseline processes and a 927 ms candidate
process in its first five pairs, plus a 920 ms candidate in its second five.
Unicode retains candidate processes from 4.36 to 5.37 seconds. No row crosses
the 5% broad-regression guard. Allocation spot checks are not used as a claim:
Numeric Array Map matched exactly, while Iterator differed by 40 allocations
in one noisy process despite this allocation-free classifier change.

## Mechanism and correctness

Post-change CPU sampling confirms that work left the intended boundary:

- Run-length `execStoreSlot(...)` falls from 3.63% to 2.16% cumulative and the
  raw-store helper from 1.55% to 0.72%.
- Unicode `execStoreSlot(...)` falls from 6.65% to 4.87%, the helper from 3.55%
  to 2.54%, and total raw extraction from 5.54% to 1.69% flat.
- Split/Join's helper falls from 6.56% to 3.45%; total raw extraction falls
  from 8.20% to 4.31% flat.

Tests cover every supported small raw/result/slot/scratch/boxed carrier,
dominant i32/i64 direct stores, i64 source-slot alias independence, rare u64
fallback, primitive non-integer ordinary storage, typed exact stores, float
slot ownership, binary integer paths, and runtime values. Focused interpreter
and runtime suites pass. Changed files remain below 1,000 lines and
`git diff --check` passes.

## Next recommendation

Refresh bounded post-store CPU profiles for Boolean, Unicode, Run-length,
Split/Join, Iterator Collect, and Numeric Array Map, then reconcile the
remaining `bytecodeRawIntegerValueInfo(...)` callers with call-tree ownership.

Why: this tranche removed the nested extractor specifically beneath untyped
slot stores. The residual extractor remains 1.69% flat in Unicode and 4.31% in
Split/Join, but those samples now belong to casts, comparisons, array/index
work, and other integer consumers rather than one shared store wall. The next
tranche should identify a concrete descendant repeated across at least three
unlike programs before changing code. It entails new one-process profiles,
caller/opcode attribution without global hot-path counters, duplicate-
extraction checks for the same operand, mixed-width tests, and the same
expanded repeated controls. Continue to defer WASM.
