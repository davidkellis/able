# Bytecode Typed-Native Float Region Rejection

Date: 2026-07-16

## Outcome

Complete the typed `f64 -> f64` native-region experiment with no retained VM,
runtime, compiler, stdlib, fixture, or benchmark production change. The
experiment produced large repeatable wins in square-root-heavy programs, but
every representation that preserved those wins also imposed a measurable cost
on at least one unrelated control. That fails the project's broad-benefit bar.

No change was needed in canonical `../able-stdlib`. WASM remains deferred.

## Confirmed opportunity

The retained baseline binary was:

```text
bca8de4923380f0ce0ed8a28b9c7aac36a42ec92610b56514ff36c2ba6773a33
```

A statically proven float-region step invoked an exact primitive native with a
scalar `float64` operand and kept its result in the region's scalar stack. It
preserved ordinary dynamic callbacks, binding replacement fallback, error
context, `f32` normalization rules, and the bit patterns of signed zero,
infinity, and NaN.

Five alternating baseline/candidate pairs for the allocation-safe direct
scalar path showed:

| workload | baseline mean | candidate mean | wall change | allocation change |
| --- | ---: | ---: | ---: | ---: |
| Distance Field | 5.505 s | 5.129 s | -6.82% | about 368.05 MB to 336.05 MB |
| RMS Norm | 4.455 s | 3.874 s | -13.05% | about 288.05 MB / 20.0M to 208.05 MB / 18.0M |
| reduced NBody | 1.906 s | 1.634 s | -14.27% | about 92.77 MB / 5.86M to 88.77 MB / 5.76M |

The NBody cohort contained one slow baseline and one slow candidate sample;
its medians were 1.766 and 1.600 seconds, still a 9.4% candidate improvement.
These results confirm that a scalar primitive operation is valuable when its
input and consumer stay inside the same typed region.

## Rejected representations

Four implementation families were tested and removed:

1. A separate scalar callback field enlarged every `NativeFunctionValue`.
   Iterator collect changed from exactly 8,671,264 bytes to about 9.66 MB even
   though it never used the new callback.
2. Reusing the ordinary raw callback with a one-element `RawValue` slice kept
   the struct compact, but the slice escaped at the opaque callback boundary.
   Distance rose from about 368 MB / 26.0M allocations to 496 MB / 28.0M.
3. Widening the existing raw callback ABI to carry both ordinary values and a
   scalar avoided the slice escape, but charged every unrelated raw-native
   call for the wider arguments and results. Ten array/map pairs regressed by
   roughly 17% with an unchanged allocation shape.
4. A compact metadata ID fit existing struct padding and preserved the
   ordinary raw ABI. Moving its callback to plan-level metadata also avoided
   enlarging every float-region step. However, discovering eligible calls in
   generic lowering remained visible in unrelated short controls. Successive
   narrowing reduced but did not remove the effect: the final 15-pair
   array/map gate averaged 79.73 ms baseline versus 83.20 ms candidate
   (`+4.36%`), with identical 869,448-byte / 239-allocation shapes. Separate
   ten-pair matrix gates remained about 2.6-3.3% slower, while iterator collect
   was allocation-identical and effectively neutral in the final cohort.

The final allocation shapes after reverting are again exactly 8,671,264 bytes
/ 192,871 allocations for iterator collect and 869,448 bytes / 239 allocations
for array/map. The rebuilt test binary differs from the saved baseline hash
because touched source and test files were reformatted during the experiments;
search and focused tests confirm that no candidate opcode, callback metadata,
registry, raw-float return bypass, or altered raw-native ABI remains.

## Verification after revert

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_(TypedFloatRegion|LoweringEmitsTypedFloatRegion|StructCallableFieldRawImpl|ExactNativeRaw|CallMemberInterfaceIteratorNextFastPathRequiresIteratorNative|CallMemberNext|.*Return|SlotlessInline|MinimalReturn|InlineReturnRestoresCallerActiveLookupCaches)|TestPrimitiveKernelNativesBorrowArguments|TestBytecodeVM_PrimitiveKernelBorrowedArgsAreNotRetained' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.608s

go test ./pkg/interpreter -run '^(TestBytecodeVM|TestPrimitiveKernel)' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  18.212s

go test ./pkg/runtime -count=1 -timeout 60s
ok  able/interpreter-go/pkg/runtime  0.073s
```

## Next recommendation

Refresh bounded CPU and allocation profiles for split/join, iterator collect,
array/map, Distance, and RMS, then select a larger shared VM wall that appears
in at least three unlike workloads—prefer raw integer extraction, environment
or map lookup, or remaining return/type matching over another native-call ABI.

Why: this tranche proved that primitive native regions can win locally while
even tiny structural or lowering costs erase the benefit elsewhere. A retained
next change should remove work from an already universal hot path instead of
adding metadata discovery to every call or widening a shared value contract.

What it entails: collect one-process profiles under the existing 1 GiB guard,
rank exact repeated leaves across the five workloads, implement only one
type-/runtime-driven candidate with no source-name or nominal-container rule,
and run alternating repeated cohorts (at least ten for short volatile
controls). Revert unless unlike controls are allocation-neutral and their
averages remain within noise while multiple primary programs improve.
