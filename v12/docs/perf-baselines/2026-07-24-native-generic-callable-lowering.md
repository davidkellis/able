# Native generic-callable lowering

Date: 2026-07-24

## Decision

Retain the general generic named-union method specialization path.

Statically known calls on structurally represented generic named unions now
resolve their nominal method before dynamic dispatch, bind the receiver and
method type parameters, and reuse the existing specialized nominal-method and
native-callable machinery. Dynamic calls and incompletely bound calls retain
the existing runtime path.

This is not an `Option`, `Result`, container, or benchmark special case. The
selection rule applies to any generic named union whose complete structural
receiver type and call bindings are known.

## Root cause

The compiler already supported native generated callables such as
`__able_fn_int64_to_int64` and already monomorphized ordinary generic nominal
methods. Generic named unions are represented by their structural union
carrier, however, so `methodForReceiver` could not recover the nominal method
from the Go carrier name. Calls therefore reached
`__able_static_generic_union_method_call` before nominal specialization.

For `Result i64.map<i64>`, that selected the erased
`runtime.Value -> runtime.Value` callback. The callback's primitive arithmetic
then used `__able_binary_op`, which was both slow and the semantic blocker to
removing the interpreter package.

The retained lowering:

1. identifies a generic named-union method from the complete structural
   receiver;
2. binds target generics from the receiver and method generics from explicit
   type arguments, callback arguments, and the expected result;
3. creates or reuses the normal specialized nominal method;
4. marks that dispatch result as already concretely resolved so the call does
   not specialize the specialized method a second time.

## Generated-code evidence

Validated Job Pipeline now emits:

```go
func __able_compiled_method_Result__map_spec(
    self __able_union_int64_or_runtime_ErrorValue,
    f __able_fn_int64_to_int64,
) (__able_union_int64_or_runtime_ErrorValue, *__ableControl)
```

Its `transform` closure receives `value int64`; multiplication, addition, and
remainder lower to the generated checked integer helpers. The hot function has
no `__able_binary_op`, runtime-value callback, or static generic-union method
call.

Dependency Wave Validation independently emits the same native `i64 -> i64`
Result mapping shape. Option/Result Config emits
`Option<i32>.map<i32>` as `__able_fn_int32_to_int32`. Existing native Array,
Enumerable, Iterator map, and Iterator filter-map tests remain green, proving
that the new named-union route composes with the already-native generic
callable paths rather than replacing them.

No canonical `able-stdlib`, runtime, bytecode VM, tree-walker, language, or
WASM change was required.

## Repeated application gate

Every candidate row used five independent workstation processes and the
sibling verifier.

| Application | Prior verified mean | Candidate mean | Change | Candidate verification |
|---|---:|---:|---:|---|
| Validated Job Pipeline | 1.1367 s | 0.7660 s | -32.61% | 5/5 |
| Dependency Wave Validation | 1.4440 s | 0.9440 s | -34.63% | 5/5 |
| Option/Result Config | 0.1760 s | 0.0740 s | -57.95% | 5/5 |

Validated Job Pipeline's current Go reference averaged 0.0200 seconds across
five verified processes. The candidate is therefore still about 38.3x slower
on this application; this tranche removes one large compiled/runtime
transition but does not satisfy the 95%-of-Go project target.

Evidence:

- `2026-07-24-native-generic-callable-validated-candidate.json`
- `2026-07-24-native-generic-callable-dependency-wave-candidate.json`
- `2026-07-24-native-generic-callable-option-result-candidate.json`
- `2026-07-24-native-generic-callable-validated-go-reference.json`

## Correctness and boundary gate

Passing focused tests:

```text
go test ./pkg/compiler \
  -run 'TestCompiler(GenericNamedUnionMethodSpecializationStaysNative|StdlibOptionResultMapSpecializationsStayNative|GenericNominalMethodSpecializationStaysNative|HeapGenericMethodSpecializationStaysNative|StandaloneGenericNamedUnionMethods)$' \
  -count=1 -timeout 60s

go test ./pkg/compiler \
  -run 'TestCompiler(ConcreteEnumerableGenericMethodsStayNative|ConcreteIteratorGenericMethodsStayNative|ConcreteIteratorFilterMapStayNative)$' \
  -count=1 -timeout 60s

go test ./pkg/compiler \
  -run '^TestCompilerStaticGenericUnionKnownMethod' \
  -count=1 -timeout 60s

go test ./pkg/compiler \
  -run '^TestCompilerIteratorPipelineBenchmarkShapeExecutesFromNonMainSourcePackage$' \
  -count=1 -timeout 60s

go test ./cmd/ablec -count=1 -timeout 60s
```

The broad `go test ./pkg/compiler -short -timeout 60s` invocation reached the
suite-level timeout after sixty seconds while an individual executable test
had run for nine seconds. That exact test passed alone in 9.869 seconds; this
was aggregate suite duration, not a test failure or an over-one-minute test.

Strict no-fallback builds and verified execution passed for Option/Result
Config, Validated Job Pipeline, and Dependency Wave Validation. The final
Validated dynamic-boundary audit is retained as
`2026-07-24-native-generic-callable-validated-boundary-audit.json`. It records
three explicit dynamic calls, two residual polymorphic calls, one host ABI
call, and five runtime-service calls; none is the eliminated `Result.map`
callback route.

The final binaries still link the interpreter package because
`__able_binary_op` and `__able_unary_op` retain their general
`interpreter.Apply*Fast` escape hatches. This tranche removes the known
semantic user of that escape hatch but does not yet cut the linker roots.

## Next

Reapply the two-root interpreter-package cut and run a widened strict,
verifier-backed coverage sweep before timing it.

This is next because the only concrete semantic blocker found in the previous
cut is now natively lowered. The sweep must include the three applications
above, the six strict core applications, Array/Iterator guards, generic
unions, maps, strings, files, and concurrency. If another boxed primitive
operator remains reachable, retain the current package roots and use the first
failing generated function to drive the next general lowering fix. If all
applications verify, remove the interpreter import and operator roots, then
repeat the six performance guards and binary-size comparison. This ordering
separates semantic completeness from timing and prevents a fast but incorrect
package cut.
