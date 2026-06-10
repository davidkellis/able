# Bytecode Native Scalar-Result ABI Rejection

Date: 2026-07-16

## Outcome

Complete the compact native scalar-result experiment with no retained runtime,
VM, compiler, stdlib, fixture, or benchmark production change. The ordinary
borrowed native-call path remains the retained implementation.

The experiment confirmed that the boxed `f64` square-root result is shared by
Distance Field, RMS Norm, and reduced NBody. It also demonstrated why a
result-only ABI cannot remove that cost while the following operand or native
boundary still consumes `runtime.Value`: the compact result first becomes a
VM raw-float interface carrier and is then materialized again at its next use.
The candidate therefore added one allocation per square-root call and regressed
wall time.

No change was needed in canonical `../able-stdlib`. WASM remains deferred.

## Refreshed retained profiles

Fresh bounded allocation profiles used the retained borrowed-argument binary,
one warmed execution, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and the
canonical external stdlib. Their recurring leaves were:

| workload | normalized raw-float carrier | raw-float materialization | boxed sqrt result | other major leaf |
| --- | ---: | ---: | ---: | ---: |
| Distance Field | 44.71% | 33.09% | 11.12% | unary result 11.03% |
| RMS Norm | 43.00% | 14.39% | 14.28% | stack snapshot 14.40%; unary result 13.86% |
| reduced NBody | 34.34% | 28.19% | 2.26% | stack snapshot 32.85% |

The approximate boxed-result sample counts were 2.0 million, 2.0 million, and
100,000 respectively, matching the applications' square-root call counts.
That was sufficient cross-program evidence to prototype one generic contract.

## Rejected candidate

The candidate added a 16-byte tagged scalar carrier supporting every Able
integer type through 64 bits plus `f32` and `f64`. An optional native callback
consumed the existing ordinary borrowed argument view and returned that compact
carrier. Ordinary implementations remained authoritative for tree-walker,
dynamic, public-materialization, and bound-call fallback paths.

Four primitive-to-primitive kernel natives opted in: `f32` bits, `f64` bits,
`f64` square root, and wrapping `u64` multiplication. The nominal Ratio
constructor did not. Focused tests covered carrier size, integer kinds,
unsigned high bits, float bit patterns, exact raw continuation, and bound-call
fallback. They passed before measurement.

The exact binaries were:

```text
baseline  bca8de4923380f0ce0ed8a28b9c7aac36a42ec92610b56514ff36c2ba6773a33
candidate e67007eb628dd5572e805abc8e1852919129db4835bd14490f85d25cb4a86cbb
```

## Early-stop gate

Separate workstation processes were repeated and averaged. The allocation
changes were exact in every completed process, so the candidate met the early
rejection condition before the unlike controls were needed.

| workload | retained mean | candidate mean | wall change | retained allocation shape | candidate allocation shape |
| --- | ---: | ---: | ---: | ---: | ---: |
| Distance Field (2 each) | 5.111 s | 5.855 s | +14.56% | 368,050,816 B / 26,000,125 | 384,051,056 B / 28,000,127 |
| RMS Norm (2 each) | 4.173 s | 4.792 s | +14.84% | 288,048,792 B / 20,000,125 | 304,049,040 B / 22,000,128 |

The RMS candidate samples were volatile at 4.356 and 5.228 seconds, but both
were slower and both had the identical +2,000,003 allocation shape. Distance
candidate samples were 5.773 and 5.937 seconds versus retained samples of
5.127 and 5.094 seconds. A third profiled Distance execution reproduced
384,051,064 B / 28,000,127 allocations.

Reduced-NBody processes were attempted under the same guards but were killed
without benchmark output during the workstation's memory pressure. Repeating
them could not make the candidate retainable after two primary programs showed
the same deterministic regression.

## Failure attribution

The candidate Distance allocation profile removed the old square-root
implementation leaf, but raw-float materialization rose from about 5.94
million to 7.93 million sampled objects. The normalized raw-float carrier also
remained. End to end, each call now did:

1. materialize the raw argument for the ordinary native callback;
2. return the compact scalar and box it as the VM's raw-float interface
   carrier; and
3. materialize that carrier when the next call or dynamic consumer required
   `runtime.Value`.

The old path performed steps one and one result box. The candidate performed
steps one, two, and three. A result-only boundary therefore moves the result
box and adds a later materialization; it cannot pay off without a consumer
that keeps the value raw across multiple operations.

The candidate and all candidate-only tests were removed. The rebuilt reverted
test binary is byte-for-byte identical to the captured baseline:

```text
bca8de4923380f0ce0ed8a28b9c7aac36a42ec92610b56514ff36c2ba6773a33
```

## Verification after revert

```text
go test ./pkg/interpreter -run '^(TestBytecodeVM|TestHashHelperBuiltins|TestPrimitiveKernel)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  18.099s

go test ./pkg/runtime -count=1 -timeout 60s
ok  able/interpreter-go/pkg/runtime  0.074s
```

## Next recommendation

Prototype a typed-float region input for compact primitive native operations,
instead of another standalone call-result carrier.

Why: retained typed-float regions already win broadly by amortizing type proof
and representation work over several arithmetic operations. This tranche
proves a compact native result is useful only when its consumer remains in the
same raw region; putting it back on the boxed operand stack immediately loses
the benefit. Distance, RMS, and NBody all cross the same primitive square-root
boundary, so the opportunity is shared rather than application-specific.

What it entails: define one compact typed primitive callback shape, such as a
generic raw `f64 -> f64` operation, and let statically proven float-region
plans invoke eligible callbacks directly from raw region operands. Resolve
eligibility from primitive types and native metadata, never source, function,
benchmark, or nominal-type names. Preserve ordinary callbacks for all dynamic
paths; guard language evaluation/error order, wrong kinds, f32 normalization,
NaN/infinity/signed zero, nested calls, and cold fallback. Gate repeated
Distance, RMS, NBody, and matrix runs plus split/join, iterator collect, and
array/map controls, and revert unless both wall time and allocations improve
broadly.
