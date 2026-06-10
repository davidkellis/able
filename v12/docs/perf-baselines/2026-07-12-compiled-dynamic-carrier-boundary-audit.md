# Compiled Dynamic-Carrier Boundary Audit (2026-07-12)

## Decision

Keep the compiler and runtime unchanged. The staged-AOT carrier audit found no
shared, material `runtime.Value` escape in the sampled static application
workers. The visible dynamic helpers are mostly emitted support for explicit
dynamic interop, host calls, scheduler/error payloads, and registration; their
presence in a generated source file does not mean the application worker calls
them.

This closes the investigation recommended by the canonical runtime-value
architecture decision. It does not relax the AOT policy: an unclassified
static carrier crossing would still be a compiler defect. It means that the
current compiler-performance misses cannot be explained by replacing
`runtime.Value` globally or by adding a named nominal/container fast path.

## Method

The audit used two views:

1. The production compiler generator tree contains 134 non-test Go files and
   1,541 textual `runtime.Value` occurrences. This is an inventory only; it
   includes emitters for runtime helpers and dynamic ABI wrappers.
2. Fresh static `ablec -main` output was generated under
   `ABLE_SOURCE_ROOT_ONLY=1` for Fib, MatrixMultiply, Word Frequency, and
   Fixed Width 128. All temporary generated trees were removed after
   inspection. The full `compiled.go` counts were 1,942, 2,482, 5,829, and
   3,617 respectively. They are intentionally not treated as a performance
   metric because every output includes canonical kernel/stdlib definitions,
   dynamic-call wrappers, and callable registration whether or not its static
   worker reaches those helpers.

The audit then inspected the generated functions selected by each application
and categorized each dynamic carrier occurrence against the four staged-AOT
categories in the v12 specification. Existing no-bootstrap boundary markers
were run at the same revision.

## Generated static-worker findings

| Application | Native worker evidence | Dynamic-carrier occurrences in selected application code | Classification |
| --- | --- | --- | --- |
| Fib | `__able_compiled_fn_fib(n int32) -> int32` is direct checked scalar recursion | `main` returns `runtime.Value` and sends its final value to named `print` | host/output boundary; worker has none |
| MatrixMultiply | `build_matrix` and `matmul` use `*__able_array_array_f64`, `*__able_array_f64`, `float64`, and `int32` | `main` returns `runtime.Value` and boxes only the final printed f64 | host/output boundary; kernels have none |
| Word Frequency | `lookup` receives a generated native `Map String i32` interface and returns `int32`; the line/string/map hot path stays native | `main` has its return/output boundary plus one `runtime.Value` loop-result temporary | output is host ABI; loop result is a cold static-control residual, discussed below |
| Fixed Width 128 | modular addition and ordered selection use `*UInt128`, `uint64`, `int32`, and direct compiled calls | printing boxes the final text; generic numeric/ratio helpers are emitted but not selected by these workers | host/output boundary; hot `UInt128` worker has no generic carrier crossing |

The result matters more than the raw generated-file count: Fib recursion,
Matrix typed arrays, Word's native map interface lookup, and Fixed Width's
native `UInt128` pointer path are already distinct native lowering families.
The slow programs remain slow for their attributed arithmetic, allocation,
iterator/map, or scheduler work, not because a common dynamic carrier sits in
the observed worker loops.

## Boundary classification

| Generated family | AOT category | Audit result |
| --- | --- | --- |
| `__able_call_named`, `__able_call_value`, method lookup, environment lookup, callable thunks, and compiled-function registrars | explicit dynamic crossing or residual polymorphic adaptation | emitted as shared capability/registration support; no fallback calls in the focused static fixtures |
| Native function/extern adapters and final `print` arguments | host/ABI conversion | required at the host/output edge; not a hot worker path in the samples |
| Future, await, scheduler, error, and diagnostic helpers | runtime service payload | required only when the application uses those semantics; not evidence for an ordinary static local carrier |
| Typed Array/Map/struct/union/interface generated carriers | host-native static representation | selected static kernels use these carriers directly; no generic object-model replacement is warranted |
| Word Frequency loop-result temporary | disallowed static-control residual, but cold | its `runtime.Value` is initialized before the outer loop and written/read on loop exit to preserve untyped break/error result semantics; it is not allocated or updated per word/iteration |

The last row is a legitimate future correctness/cleanup opportunity: a
statement-context loop whose break forms are provably valueless could use a
native void result. It is not a performance candidate in this tranche. It
needs a complete proof for break values, typed loop expressions, pattern
mismatch errors, rescue/ensure control, and diagnostics, and it did not recur
as a hot boundary in the independent native worker samples. Rewriting it now
would optimize cold control bookkeeping rather than application execution.

## Verification

The following focused compiler gate passed in 14.939 seconds with bounded Go
memory and a 90-second test limit:

```text
go test ./pkg/compiler -run 'TestCompilerBoundaryExplicitHelperSetSourceAudit|TestCompilerNoBootstrapStaticFixturesStayBoundaryClean|TestCompilerStaticKernelHelpersUseDirectImplCalls|TestCompilerNativeInterfaceExecutes|TestCompilerExperimentalMonoArraysMatrixMultiplyBuilds' -count=1 -timeout 90s
```

It validates the explicit-boundary emitter allowlist, no-bootstrap static
fixture behavior, direct static kernel calls, native interface execution, and
typed MatrixMultiply array lowering. The audit also generated the four sampled
applications with `ABLE_SOURCE_ROOT_ONLY=1`; the temporary output was deleted
after inspection.

## Outcome and historical next recommendation

No compiler, VM, runtime, fixture, benchmark, or `able-stdlib` source changes
are authorized. A global carrier replacement is rejected by the architecture
decision, and the sampled static workers do not expose a repeated dynamic
carrier cost to remove.

The debug-only telemetry mode and verifier-backed 22-application sweep are
complete. The implementation is invoked by `just
bench-compiled-boundary-audit`, emits no telemetry code in normal binaries,
and records its durable result in
`2026-07-12-compiled-dynamic-boundary-reachability-decision.md`. The sweep
rejects global dynamic-carrier work but identifies repeated generic static
`print` calls in I-Before-E and PiDigits as the sole material shared host
boundary. The next profile gate is static primitive/String print lowering with
unrelated compiled application guards.
