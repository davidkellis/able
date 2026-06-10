# Bytecode String-Map Attribution and Rejection

Date: 2026-07-16

## Outcome

Complete the five-profile string-key map attribution gate and retain no VM,
runtime, compiler, stdlib, fixture, or application behavior change. The one
exact shared caller was primitive integer metadata lookup, but both a full
metadata switch and a narrower membership classifier failed unlike application
guards. All production candidates were removed.

A focused `BenchmarkLookupIntegerInfo` remains as diagnostic coverage. The
restored binary exactly matches the captured same-test-source baseline:

```text
0e43745e644d314b785402261aaab0f44725de7306d740ff7e425a1b26779ab9
```

No change was needed in canonical `../able-stdlib`. WASM remains deferred.

## Exact caller attribution

The saved one-process profiles showed that `runtime.mapaccess2_faststr` is an
aggregate of several semantic lookup families. Its caller distribution was:

| workload | string-map cumulative | `lookupIntegerInfo` share of map wall | other material callers |
| --- | ---: | ---: | --- |
| String Split/Join | 11.78% | 35.90% | type matching, environment bindings, known-type caches |
| iterator collect | 6.58% | 40.00% | environment bindings, canonical type names, generic expansion |
| numeric Array map | 4.24% | 57.14% | type matching and environment bindings |
| Distance Field | 3.60% | 100% | none sampled |
| RMS Norm | 0.68% | 100% | none sampled |

Within Split/Join's integer-metadata samples, 64.29% came from
`isPrimitiveName(...)` and 35.71% from
`fastNamedStructTypeNameIsNonNominal(...)`. Thus the same fixed twelve-member
primitive suffix table was genuinely shared, while the remaining string maps
were disjoint and did not justify a common cache change.

## Rejected full metadata switch

The first representation replaced `map[IntegerType]integerInfo` access with a
fixed switch returning package-level immutable metadata values. Cold sequence
helpers retained the map. This was a primitive language-type optimization, not
a source-name, benchmark, or nominal-container rule.

Five fixed-iteration microbenchmark means improved with zero allocations:

| key | map baseline | switch candidate |
| --- | ---: | ---: |
| `i32` | 24.03 ns | 8.87 ns |
| `i128` | 19.61 ns | 8.03 ns |
| `isize` | 20.61 ns | 7.83 ns |
| miss | 15.48 ns | 6.69 ns |

The application gate contradicted the local result:

| workload | baseline mean | candidate mean | wall change | allocations |
| --- | ---: | ---: | ---: | --- |
| iterator collect, 5 alternating pairs | 549.00 ms | 588.29 ms | +7.16% | exactly unchanged |
| Array map, 5 alternating 20-iteration pairs | 85.21 ms | 88.35 ms | +3.68% | unchanged at 107/op |

Split/Join sampling became too workstation-volatile to interpret, ranging from
about 1.1 to 2.3 seconds. The two steadier unlike failures were already enough
to reject the candidate. Returning the large metadata struct from a larger
switch changed call/code-generation costs that the isolated benchmark did not
represent.

## Rejected membership-only refinement

Profile callers often discarded the metadata and only asked whether a suffix
was an integer. The refinement restored full metadata retrieval to the map and
used a boolean fixed switch only at membership-only call sites across type
matching, coercion, lowering, recurrence, and integer VM helpers.

The boolean classifier itself measured roughly 0.7-2.0 ns/op with no
allocations. Whole applications remained mixed:

| workload | baseline mean | candidate mean | wall change | allocations |
| --- | ---: | ---: | ---: | --- |
| iterator collect, 5 alternating pairs | 555.03 ms | 542.37 ms | -2.28% | exactly unchanged |
| Array map, 5 alternating 20-iteration pairs | 97.57 ms | 103.63 ms | +6.21% | unchanged at 107/op |

Four of five Array-map candidate samples were slower. The refinement therefore
still traded one workload for another and was removed. This closes fixed
string-switch replacement of integer metadata on the current compiler/runtime
layout; direct nanosecond measurements are not an admission substitute for
unlike program averages.

## Verification after reverts

```text
go test ./pkg/interpreter -run 'Test.*(Integer|Numeric|Cast|TypeMatch|TypeCoerc|NamedStruct)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  14.228s

go test ./pkg/interpreter -run '^(TestBytecodeVM|TestPrimitiveKernel)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  18.901s

go test ./pkg/runtime -count=1 -timeout 60s
ok  able/interpreter-go/pkg/runtime  0.058s
```

## Next recommendation

Attribute active bytecode-program restoration inside
`switchRunProgramWithActiveLookupState(...)` before changing another value,
map, or return ABI.

Why: it is the next exact shared descendant in all five refreshed profiles,
at 0.88%-3.25% cumulative. The surrounding return parent is much larger, but
prior return-guard and frame-result experiments were mixed. Program switching
contains separable work: instruction-slice replacement, integer-constant and
slot-immediate table selection, active lookup-cache restoration, and i32
register-frame activation.

What it entails: add opt-in counters or a focused benchmark that distinguishes
same-program returns, restored cached lookup state, cold table discovery, and
i32-frame activation; identify one repeated branch across at least three
unlike programs; then attempt at most one generic state-restoration reduction.
Run alternating averaged Split/Join, iterator, Array map, Distance, and RMS
guards with unchanged allocations, and close the wall without code if its
subpaths are disjoint.
