# Bytecode call-execution reconciliation — 2026-07-17

## Decision

Close `execCallOpcode(...)` as a dispatcher parent on the current five-program
cohort and keep no production change from this tranche. Exact opcode and route
counts show that its apparent shared cost separates into different call
families. The only material opcode repeated in three unlike programs is the
already-specialized canonical Array slot-call path; a general validated-push
shell candidate regressed the broad controls and was fully reverted.

No compiler, stdlib, fixture, language, or benchmark-specific path changed.
All temporary counters, candidate code, binaries, profiles, and copied source
links are removed after this record is written.

## Clean profile baseline

One preserved post-`IteratorEnd` test binary supplied fresh CPU-only profiles.
Each workload ran in a separate process with canonical external `able-stdlib`,
CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and skipped benchmark
typechecking.

| Workload | Iterations | ns/op | CPU samples | `execCallOpcode` cumulative |
| --- | ---: | ---: | ---: | ---: |
| Unicode Scalar Pipeline | 1 | 3,560,451,940 | 3.76 s | 29.79% |
| Run-length encode | 1 | 983,880,723 | 1.14 s | 40.35% |
| String Split/Join | 1 | 1,065,009,365 | 1.22 s | 26.23% |
| Iterator Collect | 3 | 429,711,942 | 1.98 s | 47.98% |
| Numeric Array Map | 20 | 79,247,022 | 1.86 s | 25.27% |

These percentages confirm materiality but do not identify a removable shared
operation. The exact route census below is the admission evidence.

## Exact call-shape reconciliation

The repository's opt-in main-only bytecode counters recorded one CLI process
per workload. Counters perturb timing and were used only for exact frequency
and route attribution.

| Workload | Material call opcodes and routes |
| --- | --- |
| Unicode | 3,538,965 `CallName`, of which 3,538,947 are exact native; 1,769,488 specialized `CallMemberNext` |
| Run-length | 960,048 specialized `CallMemberNext`; 329,889 `CallMember`; only 148 `CallName` |
| Split/Join | 826,083 `CallName`, split into 221,558 exact-native, 505,782 direct-slot inline, and 98,743 direct-stack inline; 493,728 Array-get; 104,796 Array-slot |
| Iterator Collect | 186,033 static-member; 64,010 `CallName`, almost all direct-stack inline; 30,000 direct `Call`; 30,000 Array-slot |
| Numeric Array Map | 78,012 Array-slot; 36,000 direct `Call`; 36,000 Array-get |

`CallName` therefore repeats by opcode in Unicode, Split/Join, and Iterator,
but not by execution route: exact native is material in the first, mixed
native/direct-slot in the second, and direct-stack inline in the third.
`CallMemberNext` belongs to two string-character iterator programs;
static-member belongs materially to Run-length and Iterator; direct `Call` and
Array-get each occur materially in only two programs.

The one exact three-program overlap is `CallMemberArraySlot`: 104,796
Split/Join, 30,000 Iterator, and 78,012 Array Map executions. The existing
canonical cache already hits 104,788, 29,996, and 78,008 times respectively,
or more than 99.98% in every program. CPU attribution shows that its remaining
work is the actual Array operation and carrier handling rather than missed
dispatch. Array push appears beneath the path in all three profiles, while
Array Map also performs material `len` work and has different numeric backing
and raw-value costs.

## Rejected generic candidate

The trial kept the same kernel Array boundary and semantics. Once the outer
canonical Array-slot path had proved the opcode shape, stack bounds, receiver,
and cache identity, it passed the already-proven Array receiver directly into
the push body rather than repeating arity, bounds, VM, and receiver-type
checks. Generic/member-cache fallback and every non-push call retained the old
path.

Focused Array-slot, Array-push, and member-call tests passed, and the candidate
push hot-path microbenchmark remained at zero allocations/op. Repeated
independent processes nevertheless rejected the layout. Pairs alternated order
and every valid workstation outlier remains in the means:

| Workload | Pairs | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Unicode Scalar Pipeline | 5 | 3.59596 s | 3.50735 s | 2.46% apparent improvement; zero candidate exposure |
| Run-length encode | 5 | 1.32216 s | 1.31253 s | 0.73% apparent improvement; zero candidate exposure |
| String Split/Join | 5 | 1.84612 s | 2.09522 s | 13.49% slower; reject |
| Iterator Collect | 3 | 887.050 ms | 917.345 ms | 3.42% slower |
| Numeric Array Map | 3 | 116.387 ms | 93.157 ms | misleading improvement from one 175.423 ms baseline outlier |

The two zero-exposure rows demonstrate the current workstation noise floor.
Split/Join crosses the 5% broad regression guard by a wide margin and both of
its final order-balanced pairs are slower. Array Map's other two baseline
samples are about 86.9 ms, below every candidate sample, so its arithmetic
mean is not positive evidence for the change even though the outlier remains
in the required mean. The candidate was fully reverted without expanding the
cohort further.

## Correctness and closure

Focused Array-slot, Array-push, call-member, and call-dispatch tests pass after
the revert. The source tree contains no candidate helper or temporary counter,
all touched files remain below 1,000 lines, and `git diff --check` passes.

This closes the current call-dispatch aggregate rather than all future call
optimization. A new call candidate still requires one concrete route or leaf
that is material in at least three unlike programs. Do not retry a generic
`execCallOpcode` rearrangement, the validated Array-push helper split, or a
single route inferred from the aggregate parent.

## Next recommendation

Reconcile named-struct field identity and index resolution across Unicode,
Run-length, and Iterator Collect, centered on
`structDefinitionNamedFieldIndex(...)`, `bytecodeI32StructField(...)`, and
`structNamedFieldValue(...)`.

Why: after excluding the closed call parent, this is the next exact semantic
family sampled in three unlike programs. The combined field-resolution subtree
is 4.26% cumulative in Unicode, 12.28% in Run-length, and 2.02% in Iterator
Collect. Unlike the call aggregate, all three enter the same definition/name
index machinery, although their consumers differ.

What it entails: collect temporary counts by definition identity, named versus
positional storage, field name/index, read/write/call consumer, cache outcome,
and primitive carrier. Determine whether a general immutable definition-level
field-index table or lowering-time index transport is safe and material in all
three. Admit a candidate only if it applies to multiple nominal definitions
and preserves dynamic definitions, field callables, aliases, user-defined
structs, and both Go interpreters. Use the same repeated independent-process
workstation gate, add no named-structure special case, and continue to defer
WASM.
