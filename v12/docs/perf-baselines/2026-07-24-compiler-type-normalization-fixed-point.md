# Compiler type-normalization fixed-point tranche

Date: 2026-07-24

## Decision

Retain general import-resolution and immutable type-expression normalization
caches. They make the canonical stdlib `BigInt -> Result<i32> ->
Expectation<Result<i32>>` compile finish within the required one-minute test
budget without changing generated-program semantics or adding any nominal,
stdlib-container, or benchmark-specific lowering rule.

Do not retain the separate native-interface direct-adapter compatibility
candidate. It did not materially improve the governing compile and was removed.

## Root cause

The original handoff attributed the timeout primarily to repeated native
interface adapter refresh. A bounded CPU profile confirmed that refresh was
material, but exposed a more general ancestor:

- compileability, join inference, generic interface dispatch, and adapter
  discovery repeatedly normalized the same immutable AST type expressions;
- every normalization recursively checked selector imports and source
  re-exports;
- source re-export walks repeatedly rebuilt recursion sets and rescanned
  bindings;
- normalized expressions were repeatedly rendered back to structural strings.

The compiler already had a structural normalization cache, but reaching that
cache required repeating much of the expensive work used to build its string
key.

## Retained implementation

- Cache selector type-alias resolution, including misses.
- Cache source re-export resolution, including misses and ambiguity.
- Cache whether a source type expression contains an imported selector alias.
- Invalidate all import-resolution caches whenever static import or source
  re-export bindings grow.
- Cache normalization directly by immutable source expression identity and
  package before building the structural cache key.
- Cache structural type-expression strings by normalized expression identity.
- Invalidate the identity and string caches with the existing normalization
  cache invalidation point.

Three focused tests cover negative-cache invalidation and source-expression
identity reuse. Existing imported-shadowed-alias tests cover the broader alias
and package-resolution semantics.

## Governing measurement

Before the retained caches, the canonical test repeatedly exceeded its limit:

- one bounded run timed out at 55.045 seconds while still compiling;
- the preceding handoff also recorded a timeout beyond 60 seconds.

After the retained caches, three independent workstation runs completed:

| Run | Go test time | Process wall time |
|---:|---:|---:|
| 1 | 37.022 s | 42.88 s |
| 2 | 36.908 s | 37.47 s |
| 3 | 35.432 s | 35.97 s |
| Mean | 36.454 s | 38.773 s |

Because the old run did not finish, its true completion time is unknown.
Relative to the 55-second cutoff alone, the retained mean is at least 33.7%
lower.

## Verification

Passed:

```text
go test ./cmd/ablec -count=1 -timeout 60s

go test ./pkg/compiler -run \
  '^(TestCompilerImportResolutionCachesInvalidateWhenBindingsGrow|TestCompilerSourceReexportResolutionCacheInvalidatesWhenBindingsGrow|TestCompilerTypeNormalizationCachesBySourceExpressionIdentity|TestCompilerConcreteEnumerableGenericMethodsStayNative|TestCompilerConcreteIteratorGenericMethodsStayNative|TestCompilerConcreteIteratorFilterMapStayNative|TestCompilerGenericInterfaceBoundaryHelperSynthesizesImplGenericConcreteAdapter|TestCompilerStdlibOptionResultMapSpecializationsStayNative)$' \
  -count=1 -timeout 60s

go test ./pkg/compiler -run \
  '^TestCompiler(Imported.*Alias|SpecializedGenericImportedShadowed.*Alias|SharedMapperRecoversImportedShadowed.*Alias|NativeInterfaceMethodShapes.*ImportedShadowedAlias)' \
  -count=1 -timeout 60s
```

The canonical expectation test passed in all three measured runs. No
`able-stdlib`, runtime, interpreter, language, dependency, or WASM change was
needed.

## Remaining wall and next gate

The canonical generated Go file is approximately 8.6 MB, yet completed
processes peak around 3.7-4.0 GB RSS. A traced finalization pass had 59 native
interface carriers, 222 already-rendered adapters, and 178 specialized
functions before boundary stabilization. The high peak is therefore
intermediate discovery/render allocation churn, not simply retention of the
final source bytes.

Next profile allocations across canonical Result/Matcher, Array/Iterator, and
one unrelated strict application compile. Attribute retained bytes and
allocation objects separately across compileability, interface-adapter
refresh, warm-to-discard boundary rendering, and final rendering. Advance only
a general memoization or streaming/dry-run finalization rule whose owner
repeats across all three and that preserves late-specialization correctness.
Gate it with repeated elapsed/RSS means plus the current semantic matrix before
returning to runtime-profile selection.
