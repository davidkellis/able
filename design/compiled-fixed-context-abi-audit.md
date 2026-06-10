# Compiled Fixed-Context ABI Audit

## Status

This is a design and call-site audit only. It authorizes no runtime change.

The retained opt-in execution-context prototype demonstrates a material
generic concurrent-call win, while the rejected all-helper variadic ABI made
Lexical-Rollup 9.4% slower. The next candidate must use fixed pointer
parameters on direct generated paths and retain no-context compatibility
entries. It must not add a branch for a benchmark, task count, `Channel`, or
any other nominal container.

## Audit result

The compiler emits 42 runtime `*_impl` helpers.

| Family | Helpers | Source-level direct helper map | Compatibility-only callers also audited |
| --- | ---: | ---: | --- |
| Array | 9 | 9 | Runtime array values; standalone IR array literals; extern adapters |
| HashMap | 10 | 10 | Map literals, runtime indexing, standalone IR map literals, extern adapters |
| String/character | 4 | 4 | Extern adapters |
| Numeric | 4 | 4 | Extern adapters |
| Channel/mutex | 13 | 13 | Awaitable callbacks and extern adapters |
| Await | 2 | 0 | Named runtime-call adapters only |
| **Total** | **42** | **40** | Every remaining call site is classified below |

The 40 direct helpers are selected in `runtimeHelperImpl`. The existing
`contextAwareConcurrencyHelper` name list is therefore not a complete ABI
description and must not be extended as the next optimization.

The audited names are:

- Array: `array_new`, `array_with_capacity`, `array_size`, `array_capacity`,
  `array_set_len`, `array_read`, `array_write`, `array_reserve`, `array_clone`.
- HashMap: `hash_map_new`, `hash_map_with_capacity`, `hash_map_get`,
  `hash_map_set`, `hash_map_remove`, `hash_map_contains`, `hash_map_size`,
  `hash_map_clear`, `hash_map_for_each`, `hash_map_clone`.
- String/numeric: `string_from_builtin`, `string_to_builtin`,
  `char_from_codepoint`, `char_to_codepoint`, `ratio_from_float`, `f32_bits`,
  `f64_bits`, `u64_mul`.
- Concurrency: `channel_new`, `channel_send`, `channel_receive`,
  `channel_try_send`, `channel_try_receive`, `channel_await_try_recv`,
  `channel_await_try_send`, `channel_close`, `channel_is_closed`, `mutex_new`,
  `mutex_lock`, `mutex_unlock`, `mutex_await_lock`.
- Compatibility-only await helpers: `await_default`, `await_sleep_ms`.

### Direct compiled-call sites

There are 16 static compiled-call target emitters outside the naming helpers.
Only static named and resolved-method calls currently route arguments through
`compiledCallArgs`; the remaining emitters would silently omit a hidden
parameter. A fixed ABI implementation must route every one through one common
compiled-call expression builder.

| Category | Emitters | Required fixed-context source |
| --- | --- | --- |
| Ordinary static function/method calls | `generator_static_named_calls.go`, `generator_resolved_method_calls.go`, `generator_index_static.go` | Active `compileContext.executionContextExpr` |
| Bound methods and expression helpers | `generator_bound_method_values.go`, `generator_exprs_helpers.go` | Capture the active expression in the generated closure |
| Native-interface lowering | `generator_native_interface_calls.go`, `generator_native_interface_generic_calls.go`, `generator_native_interface_generic_defaults.go`, `generator_native_interface_collect_arrays.go` | Active compile context or a captured native-call context |
| Rendered interface dispatch/adapters | `generator_render_interface_generic_dispatch.go`, `generator_render_interfaces.go`, `generator_native_interface_collect_arrays.go` | A context derived once from the receiving `NativeCallContext`, otherwise the legacy wrapper |
| Public/native wrappers | `generator_render_functions.go` function and method wrappers | `__able_context_from_native(ctx)` |
| Bodies, package entries, Go extern entries | `generator_render_bodies.go`, `generator_host_externs.go` | Their fixed local execution-context parameter |

### Runtime-helper compatibility sites

Every direct call to a helper name outside `runtimeHelperImpl` is accounted for:

| Category | Files | Fixed-path treatment |
| --- | --- | --- |
| Map literal lowering | `generator_collections.go` | Emit fixed calls with the enclosing compile context; map-spread callback captures that pointer. |
| Array/map standalone IR emission | `ir_codegen.go` | Keep the standalone public IR ABI on legacy wrappers in the first candidate; it has no generated execution-context type or caller context. |
| Generated runtime array/hash-map helpers | `generator_render_runtime_arrays.go`, `generator_render_runtime.go` | Keep legacy-wrapper calls unless their enclosing generated function has a native context to convert. |
| Awaitable callbacks | `generator_render_runtime_concurrency.go` | Preserve legacy behavior first; a later phase may derive a fixed context from the callback's `NativeCallContext`. |
| External adapters | `generator_render_runtime_{arrays,hash_maps,strings,numeric,concurrency}.go` | Keep their callable signatures and invoke legacy wrappers. |
| Await named adapters | `generator_render_runtime_calls_tail.go`, `generator_render_runtime_await.go` | Keep legacy wrappers; they are not in the source direct-helper map. |

This split is intentional. A compatibility caller may pay the old lookup until
its API supplies an environment or native call context. A normal statically
compiled call must never take that path.

## Proposed experimental ABI

The option remains opt-in. Default generated output and existing external
callable shapes remain unchanged.

For each runtime helper, generate paired entries:

```go
func __able_hash_map_set_ctx(
    args []runtime.Value,
    execCtx *__able_execution_context,
) (runtime.Value, error) {
    // Current helper body. It reads environment/payload only from execCtx.
}

func __able_hash_map_set_impl(args []runtime.Value) (runtime.Value, error) {
    return __able_hash_map_set_ctx(args, nil) // compatibility root
}
```

`nil` means compatibility mode, not an absence of semantics. The context
normalizer may use the legacy runtime environment only in this wrapper path.
The direct compiler path calls `_ctx(args, __able_exec_ctx)` with one ordinary
pointer parameter. It never constructs a variadic slice and never consults a
goroutine identity.

Compiled functions and entries follow the same pairing:

```go
func __able_compiled_fn_work_ctx(x int32, execCtx *__able_execution_context) (...)
func __able_compiled_fn_work(x int32) (...) {
    return __able_compiled_fn_work_ctx(x, __able_context_root())
}
```

Same-package static code targets the body `_ctx`; cross-package code targets
the entry `_ctx`. Registered native wrappers use
`__able_context_from_native(ctx)` and call the `_ctx` entry. Root main and
extern boundaries construct a root context once from their supplied
environment. The current `__able_extern_call` function-type ABI stays
unchanged because it calls the legacy `*_impl` entry.

The direct target name and its arguments must be emitted together by one
helper, rather than by separate `compiledCallTargetName` and ad-hoc
`fmt.Sprintf` calls. That helper is the migration enforcement point for all
16 direct-call emitters.

## Migration order

1. Add internal naming helpers for fixed body, entry, and runtime-helper
   variants; retain all current names as compatibility wrappers.
2. Add context constructors for an explicit environment and a
   `NativeCallContext`. A non-nil native context must not consult
   `Runtime.Env()`.
3. Migrate compiled bodies, entries, function wrappers, method wrappers, main,
   spawn, and Go extern entries to fixed parameters.
4. Replace all 16 direct compiled-call emitters with the common expression
   builder. Add a source guard that fails if a direct target is formatted by
   hand outside that builder.
5. Move the 42 helper bodies to `_ctx` entries. Migrate the 40 source-level
   helper-map calls and map-literal lowering to fixed calls. Preserve the
   classified compatibility calls above through their legacy wrappers.
6. Run semantic/race coverage before timing. Do not combine dynamic bridge
   `...In` APIs with this candidate; dynamic fallback is a separate phase once
   the fixed static/helper path is neutral on serial controls.

## Required acceptance coverage

The implementation is incomplete until all of these pass:

- Generated-source guards show fixed pointer parameters, legacy wrappers, and
  no variadic execution-context parameter on normal compiled/helper paths.
- A source fixture exercises direct array, map, string, numeric, channel, and
  mutex helper calls from a spawned child; an independent map-literal/spread
  fixture covers captured context.
- Existing standalone `EmitIRFunction` array/map tests still compile without
  importing execution-context implementation details.
- Native function and method wrapper tests prove a supplied
  `NativeCallContext` reaches a fixed entry without a runtime environment
  lookup.
- Existing nested-spawn race, blocked-flush, cancellation, await, I/O,
  cross-package, and bridge-frame tests remain green.
- Fresh three-run Channel-Rollup, Lexical-Rollup, and BinaryTrees controls all
  preserve expected output. Retain only if Lexical and BinaryTrees are not
  materially slower and the concurrent gain survives.

## Explicit non-goals

- No default-ABI rollout in this tranche.
- No `able-stdlib` change: this is generated compiler/runtime plumbing.
- No dynamic bridge rewrite in the first fixed-pointer candidate.
- No specialization by benchmark, source file, task count, or nominal type.
