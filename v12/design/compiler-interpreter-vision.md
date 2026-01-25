# Compiler + Interpreter Vision (v12)

This document captures the long-term execution strategy: fast compiled programs with dynamic fallbacks, plus a faster interpreter. It should stay consistent with `spec/full_spec_v12.md` and the runtime semantics.

## Goals
- Keep full Able expressiveness (dynamic interface values, metaprogramming, concurrency).
- Compiled output should be close to host performance (Go first, then other targets).
- Interpreted execution should move toward the performance of modern interpreted languages.
- A single semantic core must drive both interpreter and compiler backends.

## Recommended execution model
- Build a typed core IR from the AST.
- Two backends:
  1) Bytecode VM interpreter (fast, portable).
  2) Host-language codegen (Go first), emitting calls into a shared runtime library.
- Compiled artifacts bundle the runtime and (optionally) the VM for dynamic fallbacks.
- Dynamic features route to runtime or VM entry points when static compilation is not possible.

## Interface dispatch in compiled + interpreted code
- Interface values carry dictionaries (see `v12/design/interface-dispatch-dictionaries.md`).
- Static code uses direct calls where the concrete type is known.
- Dynamic interface calls use dictionary dispatch; default methods can be inlined or invoked through the dictionary entry.
- Dictionaries are constructed at interface coercion time and can be cached by the runtime.

## Interpreter modernization direction
- Move from tree-walking to bytecode or SSA-based VM:
  - Lower AST -> typed IR -> bytecode.
  - Use inline caches for member/method lookups and dictionary dispatch.
  - Keep value representations shared with the runtime to reduce conversions.
- Start with a stack-based bytecode to minimize compiler complexity, then consider register-based once stabilized.
- Maintain determinism for concurrency semantics (`spawn` / `Future` handles) while improving throughput.

## Compiler direction (Go first)
- Emit Go code from typed IR.
- Generate explicit runtime calls for:
  - interface coercion + dictionary dispatch
  - dynamic metaprogramming (expr-eval)
  - concurrency scheduling (`spawn`)
- Keep the runtime ABI stable so the interpreter and compiler stay in lockstep.

## Typed Core IR (proposed)
The core IR is a typed, SSA-like, control-flow graph used by both the VM and
codegen. It is intentionally small and explicit about runtime calls so we can
preserve v12 semantics while keeping backends simple.

### IR units
- **Program**: list of packages + entrypoints.
- **Package**: symbol table, type table, functions, constants.
- **Function**: generic params + typed params + return type + CFG.
- **Block**: ordered instruction list + single terminator.

### Type model
- **Primitive**: `bool`, `char`, `i32`, `i64`, `u64`, `f64`, `String`, `nil`.
- **Struct**: nominal struct with field list + field types.
- **Union**: tagged union (sum) with variant types.
- **Array / Map**: runtime-provided containers with element/key/value types.
- **Function**: `(T1, T2, ...) -> R`.
- **Interface**: interface name + type arguments (fully bound for existentials).
- **Future**: `Future T` handle type (per spec §12).
- **Result/Option**: `!T` and `?T` are unions in the IR.

### Values and variables
- **SSA temps**: immutable values produced by instructions.
- **Locals**: explicit mutable slots for `:=`/`=` binding semantics.
- **Type refs**: runtime type tokens when needed (e.g., `C.default()`).

### Core instruction set (representative)
- **Constants**: `const <T> value`.
- **Aggregate**:
  - `struct_new Type { field: value, ... }`
  - `struct_get value.field`
  - `struct_set value.field = new_value`
  - `array_new`, `array_get`, `array_set`, `array_push`
  - `map_new`, `map_get`, `map_set`, `map_delete`
- **Control flow**:
  - `branch cond -> block_true, block_false`
  - `jump block`
  - `match value -> [case/tag -> block]`
  - `return value`
  - `raise value`
- **Calls**:
  - `call fn(args...)`
  - `call_method receiver method(args...)`
  - `call_static Type.method(args...)`
  - `call_interface iface_value method(args...)` (dictionary dispatch)
- **Operators**:
  - arithmetic/compare/logical ops resolve to explicit calls or runtime helpers
  - `coerce_interface value -> interface` (build dictionary)
- **Concurrency**:
  - `spawn fn(args...) -> Future<T>`
  - `await value` (calls Awaitable protocol)
- **Dynamic**:
  - `dyn_eval expr_string` (or host-specific hook)
  - `dyn_import package`

### Notes
- Method sugar, UFCS, and implicit `self` are lowered before IR generation.
- Interface dictionary entries are bound during `coerce_interface` so runtime
  calls are direct and deterministic.
- All effectful operations (I/O, concurrency, metaprogramming) remain explicit
  runtime calls for parity across backends.

## Runtime ABI (stable surface for VM + codegen)
The runtime ABI is the shared contract implemented by the Go runtime and the VM
host layer. ABI calls are used by compiled code and the VM; the tree-walker can
reuse the same surface for consistency.

### Core services
- **Type metadata**: construct and compare runtime type tokens; used for
  generics (`T_type`) and diagnostics.
- **Errors**:
  - Construct `Error` values and standard error structs.
  - Propagate `!T` and `?T` via explicit helpers.
- **Strings**: create from bytes, slice, concat, interpolation helpers.
- **Collections**: allocate arrays/maps and perform primitive ops.

### Interface dispatch
- **Dictionary build**: `iface_build(value, iface_name, type_args) -> dict`
- **Dictionary lookup**: `iface_call(dict, method_name, args...)`
- **Default methods**: either stored in dict or resolved via interface metadata.

### Concurrency
- **Spawn**: `future_spawn(fn, args...) -> Future<T>`
- **Future handle**:
  - `future_status(handle) -> FutureStatus`
  - `future_value(handle) -> !T`
  - `future_cancel(handle) -> void`
  - `future_yield/flush/pending_tasks` helpers per spec
- **Await**: `await_poll(value)` drives `Awaitable` protocol in cooperative runtimes.

### Dynamic + externs
- **Dynimport**: load package/module objects dynamically.
- **Externs**: host function call boundary with value marshaling and error
  propagation.

### Diagnostics
- Shared formatting helpers for runtime/typechecker diagnostics so backends emit
  identical messages.

## Immediate next steps
- Define a typed core IR and document it.
- Prototype a minimal bytecode VM (values, calls, control flow).
- Add conformance tests that run on both tree-walker and VM backends.
- Sketch the Go codegen layer for a small subset of IR instructions.
