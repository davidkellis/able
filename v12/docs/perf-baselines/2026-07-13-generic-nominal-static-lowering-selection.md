# Generic nominal static-lowering selection — 2026-07-13

## Decision

Keep no compiler, bytecode VM, runtime, canonical-stdlib, fixture, or
benchmark-source change. The historical next-category note for generic
container/nominal carrier reduction no longer names an active candidate.

The v12 AOT policy requires native carriers for statically representable
arrays, nominal values, generic interfaces, direct dispatch, and static
control flow. A new lowering cut would be justified only by a residual static
carrier crossing that is both concrete and repeated in unlike programs. The
current selection audit found none.

## Audit scope

The audit deliberately samples different shared-lowering shapes rather than
one named container implementation:

| Static shape | Native-carrier evidence |
| --- | --- |
| Generic `Array` wrapper/boundary | Static `Array` parameters and returns remain `*Array`; the `runtime.ArrayValue` plumbing is limited to the explicit wrapper/runtime conversion boundary. |
| Generic default methods | `LinkedList -> Enumerable -> Iterator` `map`, `filter`, `reduce`, and `collect<C>()` call specialized compiled helpers with native iterator and user-defined accumulator carriers. |
| Nominal map families | `TreeMap` and `PersistentMap` locals, parameters, returns, construction, and `set` calls remain direct native pointer carriers. |
| Generic interface specialization | Pure-generic interface assignment and late concrete-adapter rendering preserve generated native interface carriers. |

The source-level selection guard passed under the normal one-core/1-GiB
settings:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 go test ./pkg/compiler -run '^(TestCompilerLoweringFacadeSourceAudit|TestCompilerGenericArrayWrapperRawBoundaryStaysDirect|TestCompilerExpectRuntimeValueExprLinesGenericArrayRawBoundaryStaysDirect|TestCompilerConcreteEnumerableGenericMethodsStayNative|TestCompilerConcreteIteratorGenericMethodsStayNative|TestCompilerConcreteIteratorCollectGenericNominalAccumulatorStaysNative|TestCompilerLazySeqIteratorCarrierStaysNative|TestCompilerLinkedListIterableAdapterStaysNative|TestCompilerTreeMapStaticCarrierStaysNative|TestCompilerPersistentMapStaticCarrierStaysNative|TestCompilerPureGenericInterfaceAssignmentUsesNativeCarrier|TestCompilerGenericInterfaceBoundaryHelperRendersLateSpecializedConcreteAdapter)$' -count=1 -timeout 60s
```

The no-fallback executable generic iterator controls also passed:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 go test ./pkg/compiler -run '^(TestCompilerConcreteIteratorGenericMethodsExecute|TestCompilerConcreteIteratorCollectGenericNominalAccumulatorExecutes)$' -count=1 -timeout 60s
```

The first control executes the generic iterator pipeline and prints `15`; the
second uses the same shared `collect<C>()` mechanism with a user-defined
`Default + Extend` accumulator and prints `21`. Neither succeeds by falling
back to dynamic dispatch.

## Why no performance experiment follows

The result aligns with the broader current-source dynamic-carrier audit:
Fib, MatrixMultiply, Word Frequency, and Fixed Width 128 already have native
hot workers; any dynamic carrier occurrences are output, explicit dynamic,
host/ABI, or cold control bookkeeping. The external coverage corpus now adds
independent numeric, text, iterator, map, concurrency, and public regex
applications, without exposing a shared residual container carrier.

A special path for a named collection would contradict both the AOT policy and
the evidence. Re-running a local container microbenchmark would only measure
a family whose static carrier boundary is already closed, not identify a new
application-wide cost.

## Next recommendation

Select a new spec-defined behavior only when it has both shared fixture
coverage and a portable, verifier-backed application shape. First prove its
ordinary semantics across tree-walker, bytecode, and compiled execution; then
profile it against the 28-application coverage catalog only if two unlike
applications expose one new concrete descendant. This prevents a nominal
container or benchmark-shaped optimization from being mistaken for general
performance work.
