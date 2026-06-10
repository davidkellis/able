# Compiler allocation flat-cache tranche

Date: 2026-07-24

## Decision

Retain a general allocation reduction in compiler import and type-expression
lookup:

- collected static-import and source-re-export slices are returned directly
  during their read-only phase instead of being copied on every lookup;
- import resolution uses a comparable composite key instead of allocating a
  concatenated string key;
- imported-selector facts and normalized source-expression results share one
  flat expression/package cache instead of allocating separate nested maps.

No rule names a benchmark, stdlib container, or non-primitive nominal type.
The change affects compiler bookkeeping only; generated application semantics
and runtime carriers are unchanged.

## Selection profiles

Allocation profiles were captured in separate bounded processes for:

1. canonical `BigInt -> Result<i32> -> Expectation<Result<i32>>`;
2. canonical `Iterable`/`Iterator` specialization through `able.spec`;
3. an executable strict no-fallback `Vector<String>` application.

Before the candidate, the same owners repeated in all three:

| Flat allocation owner | Result/Matcher | Array/Iterator | Strict Vector |
|---|---:|---:|---:|
| copying `staticImportsForPackage` | 1,687 MB | 976 MB | 100 MB |
| nested source-normalization cache insertion | 1,390 MB | 815 MB | 230 MB |
| imported-selector expression cache | 1,265 MB | 729 MB | 186 MB |
| allocated import-resolution string keys | 269 MB | 166 MB | 56 MB |

These were compiler-wide lifecycle and representation costs, not a workload-
specific lowering opportunity.

## Allocation result

Final sampled allocation-space totals:

| Workload | Before | Retained | Change |
|---|---:|---:|---:|
| Result/Matcher | 9,304.62 MB | 5,639.90 MB | -39.39% |
| Array/Iterator | 5,788.69 MB | 3,857.73 MB | -33.35% |
| Strict Vector | 1,551.91 MB | 1,220.09 MB | -21.38% |

The copied-import and allocated string-key owners disappear from the final top
allocation set. The remaining shared compiler owners are imported-selector
fact discovery, structural normalization-key construction, generic binding-map
cloning, and interface self-binding construction.

## Repeated elapsed/RSS gate

The retained candidate was measured in three independent workstation
processes per workload.

### Result/Matcher

| Run | Test time | Wall time | Peak RSS |
|---:|---:|---:|---:|
| 1 | 35.607 s | 36.31 s | 1,419,596 KB |
| 2 | 37.073 s | 37.74 s | 1,315,432 KB |
| 3 | 38.402 s | 39.02 s | 1,389,732 KB |
| Mean | 37.027 s | 37.690 s | 1,374,920 KB |

The preserved pre-candidate three-run mean was 36.454 seconds test time,
38.773 seconds wall time, and 3,916,107 KB peak RSS. The retained result is
therefore +1.57% test time, -2.79% wall time, and -64.89% peak RSS. Runtime is
neutral within workstation variability while memory improves materially.

### Array/Iterator

| Run | Test time | Wall time | Peak RSS |
|---:|---:|---:|---:|
| 1 | 23.169 s | 23.85 s | 1,042,052 KB |
| 2 | 23.162 s | 23.85 s | 991,852 KB |
| 3 | 25.842 s | 26.50 s | 1,030,380 KB |
| Mean | 24.058 s | 24.733 s | 1,021,428 KB |

The immediate pre-candidate measurement was 32.993 seconds test time,
35.28 seconds wall time, and 2,176,456 KB peak RSS. The retained mean is
27.08% lower in test time, 29.89% lower in wall time, and 53.07% lower in peak
RSS. The pre-candidate row is a single bounded selection measurement; the
retained result is the required repeated cohort.

### Strict executable Vector

| Run | Test time | Wall time | Peak RSS |
|---:|---:|---:|---:|
| 1 | 5.407 s | 6.17 s | 330,644 KB |
| 2 | 5.605 s | 6.32 s | 332,420 KB |
| 3 | 5.425 s | 6.19 s | 339,604 KB |
| Mean | 5.479 s | 6.227 s | 334,223 KB |

The immediate pre-candidate measurement was 8.444 seconds test time,
9.26 seconds wall time, and 740,288 KB peak RSS. The retained mean is 35.11%
lower in test time, 32.76% lower in wall time, and 54.85% lower in peak RSS.
Again, the pre-candidate row is the bounded selection measurement and the
retained result is repeated.

## Correctness and lifecycle guards

Added a zero-allocation guard for reads of collected import/re-export slices.
The earlier negative-cache invalidation and source-expression identity tests
remain in force.

Passed:

```text
go test ./pkg/compiler -run \
  '^(TestCompilerImportResolutionCachesInvalidateWhenBindingsGrow|TestCompilerSourceReexportResolutionCacheInvalidatesWhenBindingsGrow|TestCompilerTypeNormalizationCachesBySourceExpressionIdentity|TestCompilerCollectedImportReadsDoNotCopyBindingSlices|TestCompilerConcreteEnumerableGenericMethodsStayNative|TestCompilerConcreteIteratorGenericMethodsStayNative|TestCompilerConcreteIteratorFilterMapStayNative|TestCompilerGenericInterfaceBoundaryHelperSynthesizesImplGenericConcreteAdapter|TestCompilerStdlibOptionResultMapSpecializationsStayNative)$' \
  -count=1 -timeout 60s

go test ./pkg/compiler -run \
  '^TestCompiler(Imported.*Alias|SpecializedGenericImportedShadowed.*Alias|SharedMapperRecoversImportedShadowed.*Alias|NativeInterfaceMethodShapes.*ImportedShadowedAlias)' \
  -count=1 -timeout 60s

go test ./cmd/ablec -count=1 -timeout 60s
```

The strict Vector workload also compiled, linked, executed, and verified its
expected output in every measured process.

No `able-stdlib`, runtime, interpreter, language, dependency, or WASM change
was needed.

## Next

The compiler build-scalability gate is no longer the immediate blocker:
all three governing builds complete below one minute and peak memory fell by
more than half. Imported-selector fact discovery and normalization-key
construction remain future compiler allocation targets, but optimizing them
now would delay the primary application-runtime objective.

Next refresh interpreter-free compiled CPU/allocation profiles for at least
three unlike application guards selected from the current scorecard. Choose
the largest generated-runtime or generated-code owner that repeats in all
three, excluding already-closed GC-ballast, execution-context, and named
container/nominal designs. Advance only a general lowering/runtime rule with
verifier-backed repeated A/B evidence, while keeping equivalent Go performance
as the governing target.
