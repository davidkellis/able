# Bytecode Shared VM Wall Rejection

Date: 2026-07-16

## Outcome

Complete the five-workload profile refresh and reject both generic candidates.
No VM, runtime, compiler, stdlib, fixture, or application behavior change is
retained. One focused benchmark for the existing raw-integer extractor remains
so future representation work has a direct local guard.

The restored test binary is byte-for-byte identical to the captured baseline:

```text
38b571de1d50f13e11903adffe33c1d31393cf602e351f25375c924d997799f4
```

No change was needed in canonical `../able-stdlib`. WASM remains deferred.

## Refreshed bounded profiles

Each profile used one process with `GOMAXPROCS=1`, `GOGC=50`, a 1 GiB Go
memory limit, stdlib typechecking skipped, and only the requested measured
iterations. CPU and sampled cumulative-allocation profiles were collected for
five unlike programs:

| workload | iterations | profiled runtime | bytes/op | allocs/op |
| --- | ---: | ---: | ---: | ---: |
| String Split/Join | 3 | 1.122 s | 51,001,378 | 579,861 |
| linked-list iterator collect | 5 | 457.1 ms | 8,434,694 | 192,598 |
| numeric Array map | 20 | 83.87 ms | 810,243 | 110 |
| Distance Field | 1 | 5.881 s | 368,082,664 | 26,000,186 |
| RMS Norm | 1 | 4.508 s | 288,080,504 | 20,000,184 |

`bytecodeRawIntegerValueInfo(...)` was a repeated flat leaf in the three
integer-heavy programs: 5.74% in Split/Join, 3.51% in iterator collect, and
10.91% in Array map. `finishInlineReturn(...)` was a wider cumulative parent:
20.24%, 6.14%, 10.30%, 10.27%, and 9.26% respectively. The generic
`popCallFrameFields(...)` leaf itself was about 1.3%-4.0% where sampled.

String-key map access also repeated broadly. `runtime.mapaccess2_faststr`
accounted for 11.78% cumulative in Split/Join, 6.58% in iterator collect,
4.24% in Array map, and 3.60% in Distance. Its callers differ enough that this
tranche did not assume those map reads are one optimization opportunity.

## Rejected raw-integer extractor split

The first candidate handled boxed integers and the common raw-i32 carrier
before a cold helper containing the remaining raw carriers. This was a generic
representation-based change with no source or nominal-type rule. It was
removed after the direct microbenchmark disproved its premise.

Five fixed-iteration runs produced these means, all at zero allocations:

| input | unified baseline | split candidate | result |
| --- | ---: | ---: | ---: |
| boxed i32 | 7.458 ns | 10.734 ns | slower |
| raw i32 | 1.901 ns | 6.397 ns | slower |
| raw u64 | 1.909 ns | 7.316 ns | slower |
| non-integer miss | 2.140 ns | 6.979 ns | slower |

The existing unified type switch is substantially better optimized by Go than
the apparently smaller hot/cold split. The retained
`BenchmarkBytecodeRawIntegerValueInfo` covers boxed i32, raw i32, raw u64, and
a miss so this shape does not need to be rediscovered through application
profiles alone.

## Rejected call-frame result carrier

The second candidate replaced the ten-result `popCallFrameFields(...)` ABI
with a caller-owned result struct and a boolean return. It was generic to every
inline call-frame kind and preserved all restoration semantics. Focused
call/return/frame tests passed before benchmarking.

Repeated application processes alternated baseline and candidate order.
Short controls used five pairs; Array map used ten iterations per process.
The workstation became visibly noisy during the latter cohort, so all samples
remain reported and Array map is treated as inconclusive rather than as proof
of a win.

| workload | baseline mean | candidate mean | wall change | allocation result |
| --- | ---: | ---: | ---: | --- |
| String Split/Join, 5 pairs | 978.93 ms | 993.51 ms | +1.49% | process initialization varied slightly |
| iterator collect, 5 pairs | 525.41 ms | 536.44 ms | +2.10% | exactly unchanged |
| Array map, 5 pairs | 127.19 ms | 108.67 ms | -14.56% | exactly unchanged; highly volatile |
| Distance Field, 2 completed pairs | 5.413 s | 5.766 s | +6.53% | exactly unchanged |
| RMS Norm, 1 completed pair | 4.981 s | 4.685 s | -5.94% | exactly unchanged |

The attempted longer expensive cohort was stopped by external process exits,
but no additional sampling was needed for admission: three unlike workloads
already regressed, while the two favorable readings disagreed with them and
the shortest favorable control was extremely volatile. Go already returns
large result sets through caller-owned return space; making that ownership
explicit moved code generation rather than removing semantic work.

## Verification after both reverts

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_.*(CallFrame|Inline|Return|SlotFrame|ValueSlot|ArrayOwnership)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.533s

go test ./pkg/interpreter -run '^(TestBytecodeVM|TestPrimitiveKernel)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  17.056s

go test ./pkg/runtime -count=1 -timeout 60s
ok  able/interpreter-go/pkg/runtime  0.051s
```

## Next recommendation

Attribute the repeated string-key map-access wall by exact caller and lookup
purpose across the five saved profiles before changing another return or raw
value representation.

Why: both representation candidates in this tranche made locally plausible
code slower, while `runtime.mapaccess2_faststr` is already material in four
unlike programs. The aggregate symbol is not sufficient evidence because it
may combine environment bindings, member caches, type metadata, and unrelated
program maps.

What it entails: build caller-weight tables from the saved profiles; if
needed, add an opt-in sampled lookup-purpose counter with zero normal-run work;
identify one redundant hash, string-key construction, version check, or
duplicate map read shared by at least three unlike programs; and implement at
most one runtime/type-driven candidate. Guard it with alternating averaged
Split/Join, iterator, Array map, Distance, and RMS processes and unchanged
allocation shapes. If caller attribution remains disjoint, close the map wall
without code and select a different shared leaf.
