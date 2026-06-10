# Bytecode post-provenance profile and Error-name gate

Date: 2026-07-22

## Decision

Keep the preceding typechecker-proven generic-union call optimization, but
retain no additional runtime, VM, compiler, stdlib, benchmark, fixture,
language, or WASM change from this tranche.

Four fresh, one-process steady-state CPU and sampled-allocation profiles moved
the frontier beyond generic-union method resolution. They admitted one exact,
general allocation candidate: `errorInterfaceNames` built the same one- or
two-element Error-protocol name slice on every match. Reusing immutable arrays
preserved bootstrap-before-canonical lookup order and removed objects in every
owner. It was nevertheless rejected because the expanded raw wall-time mean
for Binary Event Log regressed 2.66%.

## Profile method

The fixed profile binary SHA-256 was
`883a6ff6eaa828076a32dd5738f8844e2bec1803778bf2be86fb7e19ec650a00`.
Each workload loaded and typechecked normally, ran one untimed warmup, and
profiled one measured `main()` call with source-root-only loading under
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 59-second process cap.
All four processes passed.

| Workload | Profiled ns/op | Bytes/op | Objects/op |
| --- | ---: | ---: | ---: |
| Binary Event Log | 7,033,713,408 | 465,602,848 | 8,361,497 |
| Option/Result Config | 738,988,580 | 67,946,296 | 1,167,871 |
| Manifest Normalization | 1,273,848,104 | 63,650,512 | 916,244 |
| Policy Record Dispatch | 7,575,187,570 | 132,395,304 | 1,402,554 |

## Reconciliation

The following exact interpreter leaves occurred flat in at least three unlike
owners. Percentages are flat CPU samples in Binary, Option, Manifest, and
Policy order:

| Exact leaf | Binary | Option | Manifest | Policy | Disposition |
| --- | ---: | ---: | ---: | ---: | --- |
| `finishInlineReturn` | 0.29% | 5.33% | 2.36% | 1.86% | Shared; needs branch/frame census before another candidate |
| `popCallFrameFields` | 0.43% | 2.67% | 0.79% | 1.86% | Shared return-frame descendant |
| `pushCallFrame` | 0.43% | 2.67% | 1.57% | 0.27% | Shared but individually small |
| `releaseSlotFrameRawCells` | 0% | 2.67% | 0.79% | 0.53% | Three owners; representation-specific |
| `bytecodeRawIntegerValueInfo` | 2.57% | 0% | 0.79% | 1.59% | Three owners; mixed callers |

Generic Go map/hash and GC leaves also recur, but their interpreter callers
are heterogeneous. `runResumable` and `execCallOpcode` are cumulative dispatch
parents, not admissible implementation leaves.

The sampled allocation profiles exposed several exact shared constructors:

| Exact allocation leaf | Binary | Option | Manifest | Policy | Disposition |
| --- | ---: | ---: | ---: | ---: | --- |
| `errorInterfaceNames` | 344,074 | 49,153 | 16,384 | 49,153 | Admitted and gated |
| `NewStructInstancePositionalSized` | 212,194 | 12,289 | 155,667 | 151,569 | Application result objects; semantic allocations |
| `NewIdentifier` | 786,479 | 98,310 | 24,577 | 16,384 | Type-expression family previously rejected broadly |
| `NewUnionTypeExpression` | 740,611 | 58,986 | 19,662 | 26,216 | Type-expression family previously rejected broadly |
| `errorInterfaceNames` share | 4.55% | 4.08% | 1.14% | 2.66% | Same exact leaf in all four |

Allocation-profile counts are sampled estimates, not exact event counters.
The benchmark's `allocs/op` values below provide the exact candidate delta.

## Candidate

The candidate replaced a freshly allocated `[]string{"Error"}` plus optional
append with immutable package arrays representing the two legal lookup shapes:
bootstrap Error alone, or bootstrap Error followed by
`able.core.interfaces.Error`. No caller could mutate the arrays because the
helper is private and both uses only range over the result.

The candidate SHA-256 was
`d4dde8fb9c219cbfbbf81faa8ff9437568d2e57f531b7915c009e6e7c03ca1f3`.
The fixed control was the profile binary. Processes alternated order; Binary
and Policy were expanded to eight repetitions per variant after five-run means
were volatile. Option and Manifest used three per variant.

| Workload | Processes/variant | Control ns/op | Candidate ns/op | Time | Bytes | Objects |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 8 | 7,007,115,148 | 7,193,503,883 | +2.66% | -2.17% | -3.77% |
| Option/Result Config | 3 | 745,367,995 | 698,142,746 | -6.34% | -6.10% | -11.09% |
| Manifest Normalization | 3 | 1,296,911,302 | 1,262,638,992 | -2.64% | -0.80% | -2.01% |
| Policy Record Dispatch | 8 | 6,525,365,550 | 6,511,611,867 | -0.21% | -0.26% | -0.66% |

Binary's candidate median was faster (6,853,117,433 versus 6,957,126,058
ns/op), and six of eight paired samples favored the candidate. One
9,555,638,589 ns candidate process dominates its raw average. It remains part
of the result: the workstation policy is to repeat and average volatile runs,
not discard an inconvenient sample. Since the raw expanded mean regresses,
the candidate fails before unrelated-control admission. Split/join, iterator
collect, and numeric Array map were therefore not spent on a rejected owner
candidate.

Every one of the 22 candidate runs reduced allocations. The exact reductions
were 315,392 objects for Binary, 129,504 for Option, about 18,426 for Manifest,
and about 9,225 for Policy. Memory benefit alone does not override the broad
wall-time bar.

## Verification and restoration

The candidate's semantic order test and the existing generic-union, cache,
truthiness, and Error tests passed. After removal, the restored focused suite
passes:

```text
go test ./pkg/interpreter -run 'TestOptionResultConfigGenericUnionMethodsRemainResolvedWhenWarm|TestBytecodeGenericUnionCallCacheInvalidatesMemberChanges|TestExecFixtures/06_11_truthiness_boolean_context|TestRaiseConvertsValueToErrorStruct' -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 3.605s
```

All raw timing samples are retained in the companion JSON. Cleanup-eligible
profile and text artifacts remain under
`/tmp/able-post-provenance-profiles-20260722` and
`/tmp/able-error-interface-names-gate-20260722`.

## Next recommendation

Run a stats-only return-frame composition census across the same four owners
and the three unrelated controls before changing `finishInlineReturn` again.

Why: return handling is now the strongest exact interpreter-owned CPU family
shared by all four owners, but earlier blind guard reordering was neutral or
mixed. The profiles show that the cost is spread across return coercion,
frame-pop, slot-sidecar release, program switching, and operand restoration;
they do not say which semantic frame shape dominates.

What it entails: count full, self-fast, self-fast-minimal, slotless, raw-return,
coercion-required, program-switch, slot-frame-release, and already-balanced
return cases per instruction/site. Admit a candidate only if the same immutable
frame/return shape dominates at least three owners and does not occur
materially in the controls. Then change that general shape, preserve all
fallbacks and invalidation, and gate it with repeated alternating processes.
Continue to defer WASM.
