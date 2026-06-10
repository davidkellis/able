# Bytecode Precomputed Package Identity

Date: 2026-07-16

## Outcome

Retain a generic static-call cache improvement. Runtime package values now
carry the canonical package identity already implied by their immutable import
metadata. The bytecode VM reuses that identity instead of rebuilding it with
`strings.Join(...)` at every imported static member call.

This is import/cache metadata, not a math, benchmark, source, function, or
stdlib nominal-type fast path. The fallback still derives the identity for
manually constructed and legacy package values. No `able-stdlib` change is
needed.

The exact pre-candidate and retained test binaries are:

```text
baseline  681ec93372866b0afef6b168a90143acf28a78899b23383cbb199be1ad9b3486
candidate 1631e7fa651fdef9d7b871a8f8a463afdc0f71734960123ae76f25aaf7f22db5
```

## Profile gate and rejected external-region input

Fresh post-region profiles covered Distance Field, RMS Norm, reduced NBody,
and `matrixmultiply_f64_small`. Before this candidate, package identity
construction through `strings.Join` accounted for about 16.6% of sampled
allocation objects in Distance and 19.8% in RMS, plus 4.1% in NBody. The same
identity was rebuilt for every imported static call even though a package's
name path cannot change.

The profiles first motivated a more ambitious typed-region experiment:
evaluate exact external float leaves in language order, put their values on a
region input stack, and consume them from the existing postfix region plan.
The implementation guarded exact runtime kinds, preserved slot reads before
calls, and rejected deferred division before later calls. It passed focused
tests but removed no allocations from any of the four numeric programs. The
call/binary boundary had already allocated the interface carrier before the
region received it. Wall results were mixed, so the experiment was fully
reverted. A restored binary exactly matched the baseline hash above.

## Retained implementation

`PackageValue` and `DynPackageValue` have an `IdentityKey` populated once when
normal or dynamic imports construct them. Its encoding is exactly the prior
cache encoding: the NUL-separated `NamePath`, or `Name` when no path exists.
Static receiver identity uses the precomputed string and falls back to the old
derivation when the field is empty.

A regression test proves that precomputed and fallback identities compare
equal and that extracting identity from a populated runtime package performs
zero allocations.

## Repeated broad gate

All processes used `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, the canonical
external stdlib, skipped benchmark typechecking, and a 55-second outer limit.
Wall values are arithmetic means of separate workstation processes. Positive
change is slower.

| workload | baseline mean | candidate mean | wall change | deterministic allocation result |
| --- | ---: | ---: | ---: | --- |
| Distance Field (3 each) | 6.815 s | 5.913 s | -13.23% | 464,051,560 B / 32,000,136 to 400,051,131 B / 28,000,129; 4.0M fewer allocations |
| RMS Norm (3 each) | 5.411 s | 4.492 s | -16.98% | 384,049,573 B / 26,000,140 to 320,049,080 B / 22,000,130; 4.0M fewer allocations |
| reduced NBody (3 each) | 1.560 s | 1.596 s | +2.29% | 97,572,467 B / 6,161,055 to 94,371,811 B / about 5,961,013; 200k fewer allocations |
| matrix (10 each) | 0.2983 s | 0.2907 s | -2.55% | identical: 47,636,112 B / 1,187,689 allocations |
| split/join control (15 each) | 1.0635 s | 1.0740 s | +0.98% | no deterministic change; fresh-load counts varied by less than 60 |
| iterator collect control (20 each) | 0.4167 s | 0.4272 s | +2.52% | identical: 8,671,264 B / 192,871 allocations |
| array/map control (20 each) | 0.0680 s | 0.0724 s | +6.46% | identical: 869,448 B / 239 allocations |

The workstation was particularly volatile. Distance baseline/candidate CVs
were 12.03%/8.07%; the candidate won every completed pair. RMS CVs were
9.57%/3.21%. NBody's 2.29% movement is below its candidate CV of 3.20%, while
its allocation reduction is exact. Matrix improves with identical allocation
shape.

The controls received 15-20 fresh processes because early load spikes were
order-biased. Their mean movements are below candidate CVs of 17.17%, 6.31%,
and 19.93%, respectively, with unchanged allocations. They are neutral guards,
not evidence of a representation or semantic regression. In contrast, the
primary allocation savings are deterministic and recur at exact imported
static-call counts in three unlike numeric applications.

A retained RMS allocation profile contains no `strings.Join`, builder-grow,
or `MakeNoZero` package-identity frame. Its measured allocation count is
22,000,189 with profile overhead, versus the unprofiled 22,000,130 shape.

## Verification

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_StaticMember|TestBytecodeVM_TypedFloatRegion|TestBytecodeVM_LoweringEmitsTypedFloatRegion|Test.*Import' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.555s

go test ./pkg/runtime -count=1 -timeout 60s
ok  able/interpreter-go/pkg/runtime  0.061s

go test ./pkg/interpreter -run '^TestBytecodeVM' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  20.368s
```

Every touched file remains below 1,000 lines.

## Next recommendation

Add raw-ABI coverage to the primitive numeric kernel natives, then measure the
existing exact-native raw result path before extending typed regions again.

Why: after package identity disappears, the retained RMS allocation profile
attributes 24.88% of allocation objects to `execAndFinishExactNativeCall`,
split almost evenly between the primitive sqrt implementation's boxed result
and copied call arguments. The rejected external-region input could not help
because this allocation happened before the region. The VM already has a
generic `NativeFunctionValue.RawImpl` and `runtime.RawValue` boundary that can
avoid both boxes for primitive-only kernel calls.

What it entails: audit primitive numeric host natives, add raw implementations
where their public argument/result types are exact primitives, preserve the
ordinary implementation for tree-walker and dynamic callers, and add parity,
wrong-kind, error-order, NaN/infinity, and mixed-call tests. Re-profile and run
repeated Distance, RMS, NBody, matrix, split/join, iterator, and array/map
gates. Retain only a shared primitive-kernel boundary win; do not add a math
package, source, benchmark, or nominal-container special case. Continue to
defer WASM.
