# Bytecode Program-Switch Cache Policy Rejection

Date: 2026-07-16

## Outcome

Complete the active-program restoration attribution gate and retain no VM,
runtime, compiler, stdlib, fixture, or application behavior change. A generic
two-entry validation-cache policy improved the current five applications, but
regressed a valid three-program nested call pattern. It was therefore removed
rather than tuning the interpreter to the present benchmark mix.

Two focused cache-policy benchmarks remain. Temporary branch diagnostics were
fully removed. Two independent final builds produced the same restored binary:

```text
e6a3d5758cfd8cee193a9d727f18c068110774c32e597477b86621a4e8cb967b
```

No change was needed in canonical `../able-stdlib`. WASM remains deferred.

## Branch census

Temporary program-switch-only counters ran without the broader per-opcode
statistics overhead. Counts cover one measured application execution except
Array map, which used five:

| workload | switch calls | same program | lookup restored / reset | validation fast-nil / resolve | immediate known | i32 activate |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| String Split/Join | 1,233,174 | 0 | 616,587 / 616,587 | 413,072 / 820,102 | 1,233,174 | 0 |
| iterator collect | 328,092 | 0 | 164,046 / 164,046 | 280,056 / 48,036 | 328,092 | 0 |
| Array map, 5x | 360,200 | 0 | 180,100 / 180,100 | 0 / 360,200 | 360,200 | 0 |
| Distance Field | 8,000,000 | 0 | 4,000,000 / 4,000,000 | 6,000,000 / 2,000,000 | 8,000,000 | 0 |
| RMS Norm | 4,000,002 | 0 | 2,000,001 / 2,000,001 | 2,000,001 / 2,000,001 | 4,000,002 | 0 |

The switch wall is not discovery work in these programs. Slot-immediate tables
are known on every switch, i32-frame activation never occurs, and no switch is
same-program. The stable shared pattern is call/return alternation: half the
transitions restore the caller's lookup state and half reset for the callee.
Validation slices are consulted in every workload, from 14.6% of iterator
switches to 100% of Array-map switches.

## Candidate

`validatedIntegerConstSlots(...)` keeps hot and alternate program entries. An
alternate hit previously promoted the entry by swapping both program pointers
and both validation slices. Strict caller/callee alternation consequently
performed the swap on every lookup. The candidate returned the alternate slice
in place, preserving both entries without LRU promotion.

This was a generic cache-policy change. It did not inspect source names,
functions, nominal types, benchmarks, or call sites, and allocations remained
zero in the focused benchmarks.

## Application gate

Processes alternated baseline/candidate order. All samples remain in the means,
including visibly slow workstation samples:

| workload | baseline mean | candidate mean | wall change | allocation result |
| --- | ---: | ---: | ---: | --- |
| String Split/Join, 4 completed pairs | 1.779 s | 1.558 s | -12.39% | process-init bytes varied slightly |
| iterator collect, 5 pairs | 598.35 ms | 509.15 ms | -14.91% | exactly unchanged |
| Array map, 10 pairs at 20x | 88.43 ms | 88.59 ms | +0.18% | unchanged at 107/op |
| Distance Field, 3 pairs | 5.940 s | 5.834 s | -1.79% | unchanged at 26,000,125/op |
| RMS Norm, 3 pairs | 5.125 s | 4.437 s | -13.43% | unchanged at 20,000,125/op |

Split/Join contained two approximately 2.2-second baseline samples and RMS had
one 6.04-second baseline sample. Removing the paired outliers still leaves
iterator about 4.6% faster and RMS about 4.4% faster; Distance remains modestly
better and the longer Array gate is neutral. Thus the application evidence was
favorable, but it was not the final generality gate.

## Rejection on nested-program behavior

Same-source fixed-iteration microbenchmarks compared two valid access shapes:

| validation-cache shape | promoted baseline | stable alternate candidate | change |
| --- | ---: | ---: | ---: |
| strict `A, B, A, B` alternation | 9.728 ns | 4.172 ns | -57.1% |
| nested `A, B, C, B, A` | 27.036 ns | 29.744 ns | +10.0% |

The nested sequence models ordinary deeper calls. LRU promotion preserves the
recently returned middle program when the outer program is restored; the
candidate evicts it and falls through to the direct cache on the next descent.
Keeping the candidate would therefore make the current mostly two-program
benchmarks faster while making deeper real programs slower. That violates the
project's performance policy, so the candidate and its changed expectation
were reverted. Both microbenchmarks remain to guard future cache designs.

## Verification after revert

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_.*(ValidatedIntegerConstSlots|Const|ProgramSwitch|Return|InlineReturnRestoresCallerActiveLookupCaches)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.530s

go test ./pkg/interpreter -run '^(TestBytecodeVM|TestPrimitiveKernel)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  17.650s

go test ./pkg/runtime -count=1 -timeout 60s
ok  able/interpreter-go/pkg/runtime  0.056s
```

## Next recommendation

Return to the compiled-performance selection gate and reconcile generated
backing-slice growth across Reverse Complement, Lexical Rollup, and Array Slice
Window, with Base64 as an unlike competitive guard.

Why: three consecutive bytecode tranches have now rejected locally attractive
representation, map, return, and cache-policy changes because they traded one
valid program shape for another. The existing compiled selection evidence
already identifies capacity/backing growth across three unlike applications,
and compiler performance remains a co-equal 95%-of-Go project goal.

What it entails: collect matched generated-Go allocation/CPU attribution for
the three primaries; identify one shared generated capacity calculation,
append/copy sequence, or backing-store boundary; and change only the general
generated slice-growth machinery. Do not add an `Array`, named-container,
application, or benchmark lowering. Guard with Base64 plus unrelated compiled
applications, repeat volatile runs and average them, and retain only if the
generated allocation reduction translates into broad wall-time improvement.
