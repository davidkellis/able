# Bytecode non-integer scalar comparison rejection gate — 2026-07-17

## Decision

Keep an early non-integer scalar rejection in the bytecode integer-comparison
fallback. After the existing common-integer pair path misses, bool, char,
String, boxed float, and raw float operands now return to ordinary operator
dispatch without first traversing generic raw-integer extraction.

This is a primitive-category boundary, not a benchmark, application, or named
container special case. Integer operands still use the existing direct paths;
float operands still proceed to float comparison; custom nominal values retain
ordinary interface dispatch. No compiler, stdlib, language, or fixture source
change was required.

## Raw-integer audit

Fresh post-equality profiles put `bytecodeRawIntegerValueInfo(...)` at 1.69%
flat in Boolean, 1.58% in Run-length, 5.73% in Unicode, 1.75% in Iterator, and
10.77% in Numeric Array Map. Direct callers differed, so temporary counters
recorded every carrier and miss before selecting a candidate.

| Workload | Calls | Misses | Dominant miss |
| --- | ---: | ---: | --- |
| Boolean Reconciliation | 6,191,038 | 1,597,815 | 1,597,813 bool |
| Run-length encode | 8,030,576 | 7,835,045 | 7,834,656 char |
| Unicode Scalar Pipeline | 47,513,681 | 12,386,358 | 12,386,304 char |
| Temporary custom nominal `Eq` | 524,290 | 524,288 | 524,288 struct |
| Iterator Collect | 784,096 | 212,043 | struct/iterator sentinels |
| Numeric Array Map | 660,107 | 7 | Array setup tail |

The retained scalar rejection removes exactly one failed raw extraction per
primitive equality attempt: 393,216 Boolean calls, 959,952 Run-length calls,
and 1,769,472 Unicode calls. Candidate diagnostic totals were 5,797,822,
7,070,624, and 45,744,209 respectively. The nominal struct control is
unchanged at 524,290 calls, proving that the branch does not bypass custom
dispatch.

All counters, atomic operations, output plumbing, and diagnostic binaries were
removed after recording the result.

## Rejected alternatives

Two more general-looking candidates failed before retention:

- Splitting raw extraction into a common-carrier switch, `KindInteger` guard,
  and rare-carrier switch made raw `u64` extraction about 62% slower and String
  misses about 23% slower in five repeated microbenchmark processes. It was
  fully reverted.
- A `KindInteger` guard at the integer-comparison call site regressed Boolean
  12.30% across ten independent processes and Run-length 4.15%. Interface-kind
  calls cost more than the failed extraction they replaced. It was fully
  reverted.

The retained type switch lists only language primitive scalar categories and
runs after common integer pairs have already returned. Therefore integer-dense
programs do not pay the rejection branch.

## Repeated workstation gate

Every timing is an independent process with one warmup and one measured call,
the canonical external `able-stdlib`, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, CPU 0, and skipped benchmark typechecking. Order alternates
between sides. Every valid workstation outlier remains in the arithmetic mean.

| Workload | Samples/side | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 5 | 549.129 ms | 498.868 ms | 9.15% faster |
| Run-length encode | 5 | 1.4219 s | 1.3898 s | 2.26% faster |
| Unicode Scalar Pipeline | 5 | 4.6475 s | 4.6498 s | 0.05% slower; neutral |
| Temporary custom nominal `Eq` | 10 | 438.741 ms | 432.867 ms | 1.34% faster |
| Iterator Collect guard | 10 | 482.333 ms | 471.159 ms | 2.32% faster |
| Numeric Array Map guard | 10 | 78.608 ms | 80.513 ms | 2.42% slower |
| String Split/Join guard | 10 | 1.2527 s | 1.1575 s | 7.60% faster |
| Bounded Reverse Complement guard | 10 | 1.417 ms | 1.445 ms | 2.01% slower |

The baseline Split/Join mean retains a 1.593-second process. Reverse Complement
retains 2.04-2.26 millisecond processes on both sides. No sample was trimmed.
Boolean, custom nominal, Iterator, and Numeric Array Map allocation counts are
exactly identical. Allocation behavior is unchanged; the other rows differ
only in their existing small setup tails.

Post-change sampling reduces `bytecodeDirectIntegerValue(...)` from 1.32% to
0.90% cumulative in Run-length and from 3.63% to 3.01% in Unicode. The total
raw extractor remains material because valid integer casts, slot stores,
arithmetic, and array operations still use it; those are separate owners and
were not conflated with this candidate.

## Correctness and worktree state

- Focused direct integer comparison, primitive scalar rejection, raw integer
  carrier, mixed numeric, primitive/native equality, custom nominal equality,
  operator interface, hash/equality, and bytecode/tree-walker parity tests pass.
- Regression coverage includes bool, char, String, boxed f64, raw f32, and raw
  f64 rejection plus ordinary signed integer comparisons.
- `go test ./pkg/runtime -count=1 -timeout 55s` passes.
- The shared dirty worktree's previously recorded truthiness/cast fixture and
  diagnostic-message failures remain outside this tranche; they reproduce in
  the pre-candidate binary and are not reported as passing here.
- Changed Go files remain below 1,000 lines and `git diff --check` passes.

## Next recommendation

Re-profile and census the now hit-dominant raw-integer callers, especially
`bytecodeDirectSameTypeSmallIntPair(...)`, `bytecodeDirectIntegerValue(...)`,
`bytecodeIntegerValue(...)`, and `tryStoreRawIntegerSlotValue(...)`, across
Unicode, Numeric Array Map, Boolean, Iterator, Run-length, and the nominal
control. Track which opcode calls two or more of them for the same operand and
carrier.

Why: this tranche removed the broadly shared false-extraction wall. The
remaining extractor work represents real integers, so another global type
switch reorder is unlikely to help safely. The next work entails temporary
per-caller hit/carrier counters, call-tree reconciliation under casts, array
reads, comparisons, and slot stores, then a candidate only if redundant
extraction of the same operand repeats in at least three unlike programs. Gate
it with mixed-width/signedness tests and the same repeated controls. WASM
remains deferred.
