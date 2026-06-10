# Bytecode Primitive-Kernel Borrowed Arguments

Date: 2026-07-16

## Outcome

Retain argument borrowing for the five Go kernel natives initialized together
by `initRatioBuiltins`: ratio construction, `f32` and `f64` bit extraction,
`f64` square root, and wrapping `u64` multiplication. Each implementation
consumes its argument slice synchronously and retains neither the slice nor its
elements. Marking that existing ownership fact lets exact bytecode calls pass
their already-stable argument view directly instead of allocating a copy.

This is native-call ownership metadata, not a benchmark, source, math-package,
or nominal-container lowering. Ordinary tree-walker and dynamic calls keep the
same implementations and values. No `able-stdlib` change is needed.

The exact binaries used for the retained comparison were:

```text
baseline  1631e7fa651fdef9d7b871a8f8a463afdc0f71734960123ae76f25aaf7f22db5
candidate bca8de4923380f0ce0ed8a28b9c7aac36a42ec92610b56514ff36c2ba6773a33
```

## Rejected raw ABI experiment

The first candidate followed the preceding recommendation literally and added
`NativeFunctionValue.RawImpl` coverage to the four primitive-to-primitive
operations. It preserved large unsigned values, signed zero, NaN/infinity,
kind errors, and raw VM results in focused tests. The raw square-root callback
itself allocated zero times.

The end-to-end path nevertheless failed the broad wall-time bar. Three
fresh-process means were:

| workload | baseline | raw candidate | wall change |
| --- | ---: | ---: | ---: |
| Distance Field | 5.330 s | 5.742 s | +7.71% |
| RMS Norm | 4.356 s | 4.475 s | +2.74% |
| reduced NBody | 1.630 s | 1.650 s | +1.25% |

Raw calls reduced Distance/RMS bytes by 16 MB and NBody bytes by 0.8 MB, but
did not remove the hot allocation count. Converting each argument to the wide
general `runtime.RawValue`, filling and clearing the raw scratch slice, and
then boxing a raw float stack carrier replaced rather than eliminated the
per-call object. Distance's regression exceeded both cohorts' variation. The
raw candidate and all representation-only tests were fully removed.

## Retained borrowing design

`BorrowArgs` was already the runtime contract used by Array, String, iterator,
hash-map, and other synchronous native implementations. Applying it here keeps
the ordinary primitive ABI and removes only `copyCallArgs(...)`. A focused
guard asserts that all five kernel natives advertise borrowing, have no raw
callback, and return an independent result after the caller reuses its
argument slice.

The retained RMS allocation profile no longer contains a `copyCallArgs`
frame. Unprofiled RMS falls from about 22.0M to 20.0M allocations. The next
exact-native leaf is the ordinary boxed square-root result itself.

## Repeated broad gate

All processes used the canonical external stdlib, skipped benchmark
typechecking, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 55-second
outer limit. Times are arithmetic means of separate workstation processes;
the short and volatile controls received ten samples per binary.

| workload | baseline mean | candidate mean | wall change | allocation result |
| --- | ---: | ---: | ---: | --- |
| Distance Field (3 each) | 5.520 s | 5.337 s | -3.31% | about 400,051,104 B / 28,000,129 to 368,050,843 B / 26,000,125; 2.0M fewer allocations |
| RMS Norm (3 each) | 4.587 s | 4.171 s | -9.05% | about 320,049,035 B / 22,000,129 to 288,048,781 B / 20,000,125; 2.0M fewer allocations |
| reduced NBody (3 each) | 1.736 s | 1.685 s | -2.97% | about 94,371,795 B / 5,961,013 to 92,771,469 B / 5,860,993; 100k fewer allocations |
| matrix (10 each) | 0.2989 s | 0.2932 s | -1.92% | identical: 47,636,112 B / 1,187,689 allocations |
| split/join (10 each) | 1.1059 s | 1.1111 s | +0.46% | no deterministic allocation change |
| iterator collect (10 each) | 0.4255 s | 0.4073 s | -4.27% | identical: 8,671,264 B / 192,871 allocations |
| array/map (10 each) | 0.1028 s | 0.1102 s | +7.17% | identical: 869,448 B / 239 allocations |

The primary candidates' CVs are 2.37%, 0.51%, and 2.98%; their wins accompany
exact allocation reductions. Matrix improves inside normal short-process
variation. Split/join is neutral. The array/map movement is far below its
candidate CV of 21.61%, has an exactly unchanged allocation shape, and does
not call the changed primitive kernel boundary in its hot path; it is a noisy
guard rather than a measured regression. Iterator collect improves with an
unchanged allocation shape.

## Verification

```text
go test ./pkg/interpreter -run '^(TestBytecodeVM|TestHashHelperBuiltins|TestPrimitiveKernel)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  24.311s

go test ./pkg/runtime -count=1 -timeout 60s
ok  able/interpreter-go/pkg/runtime  0.089s
```

The two touched Go files are 164 and 56 lines.

## Next recommendation

Profile and prototype a compact result-only scalar ABI for synchronous
primitive natives, without converting their arguments to `runtime.RawValue`.

Why: after argument borrowing, the RMS profile attributes about 14.4% of
allocation objects to the square-root implementation returning a boxed
`FloatValue`. The rejected full raw ABI proves that wide raw argument
conversion costs more wall time than it saves, while this tranche proves the
ordinary borrowed argument view is fast and stable. A result-only boundary can
target the remaining box without reintroducing that rejected work.

What it entails: first confirm the boxed primitive result leaf in fresh
Distance, RMS, and NBody profiles. If it repeats, add one compact tagged scalar
result contract shared by compatible primitive kernel natives while retaining
their ordinary implementations for tree-walker/dynamic calls. Guard signed
zero, NaN/infinity, wide unsigned bits, error ordering, nested calls, and public
materialization. Repeat the same four numeric and three non-float cohorts, and
revert unless wall time and allocations improve broadly. Continue to defer
WASM.
