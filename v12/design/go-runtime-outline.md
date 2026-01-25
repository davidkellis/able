# Able v11 Go Runtime Outline

This document sketches the runtime design for the Go reference interpreter (Phase 2 of the project plan). It targets Go 1.22+ and mirrors the semantics in `spec/full_spec_v11.md` and the existing TypeScript interpreter while embracing Go-native concurrency primitives.

## 1. Core Value Representation

Define a discriminated union for runtime values:

```go
type Kind int

const (
    KindString Kind = iota
    KindBool
    KindChar
    KindNil
    KindI32
    KindF64
    KindArray
    KindRange
    KindFunction
    KindNativeFunction
    KindStructDef
    KindStructInstance
    KindInterfaceDef
    KindInterfaceValue
    KindUnionDef
    KindPackage
    KindFuture
    KindError
    KindBoundMethod
    KindNativeBoundMethod
    KindImplNamespace
)

type Value interface{
    Kind() Kind
}
```

- Prefer concrete structs per kind (e.g., `StringValue`, `ArrayValue`) backed by immutable data except where the spec allows mutation (arrays, struct fields).
- `Range` wraps start/end/inclusive plus cached iterator state.
- `FunctionValue` stores the AST node plus an environment pointer for closures.
- `NativeFunction` covers runtime primitives (I/O, scheduler ops) and should expose arity metadata for error reporting.
- `StructInstance` uses a map keyed by field identifiers for named structs and index-based slice for positional structs; store a pointer to the `StructDefinition` for metadata.
- `InterfaceValue` holds the concrete value plus the resolved method table (similar to the TypeScript “dynamic wrapper”).

## 2. Environment & Scope

Implement lexical scoping with parent pointers:

```go
type Environment struct {
    values map[string]Value
    parent *Environment
}
```

- Methods: `Define(name, value)`, `Assign(name, value)`, `Get(name) (Value, error)`.
- Support lexical shadowing and module-level persistence for packages.
- Provide helpers for block scope creation and cloning captured environments for closures.

## 3. Signals & Control Flow

Model non-local exits using custom error types:

- `returnSignal` carries an optional value.
- `raiseSignal` wraps an `ErrorValue` (Able `error` union instances, `FutureError`, etc.).
- `breakSignal` carries optional label + value.
- `breakLabelSignal` for `break some_label` semantics.

Use Go errors (`error` interface) for propagation, but keep the types internal so user-facing errors remain spec-compliant.

## 4. Evaluator Skeleton

Expose two entry points:

- `Evaluate(expr ast.Expression, env *Environment) (Value, error)`
- `InterpretModule(module *ast.Module, opts Options) (*Environment, error)`

`InterpretModule` should:

1. Initialize the global environment with builtins.
2. Apply package statement (sets current module path, privacy defaults).
3. Execute imports before body, respecting visibility rules.
4. Evaluate each top-level statement sequentially.

Within `Evaluate`, dispatch on `expr.NodeType()` and implement behavior matching the spec (use helper methods for binary ops, pattern matching, etc.).

## 5. Pattern Matching & Assignment

- Implement `TryMatch(pattern Pattern, value Value, env *Environment) (bool, error)`.
- Separate `BindPattern` for destructuring assignment and parameter binding.
- Ensure typed patterns perform runtime checks (primitives, arrays, structs, interfaces) and coerce interface bindings into dynamic wrappers, mirroring TypeScript behaviour.

## 6. Functions & Closures

- When evaluating `LambdaExpression` or `FunctionDefinition`, capture the defining environment.
- Enforce arity and generic parameter counts at call time.
- Bind `Self`/`self` during method invocation via `boundMethod` values that inject the receiver.

## 7. Concurrency (`spawn` / `Future`)

Adopt Go primitives while keeping observable semantics identical to the spec:

- `spawn` launches a goroutine that writes into a `Future` handle containing:
  - `status` (enum: Pending | Resolved | Cancelled | Failed)
  - `result` channel for success value
  - `error` channel or field carrying `FutureError`
  - synchronization via `sync.Mutex` + `sync.Cond` for deterministic blocking semantics
- Cancellation: store a `cancelRequested` flag and expose `Cancel()`; cooperative checks happen via `future_cancelled()` native function.
- Memoize the result on first completion so repeated `value()` calls and implicit evaluations reuse the cached value/error.
- Provide scheduler utilities similar to TypeScript’s cooperative queue so unit tests remain deterministic (e.g., maintain an internal run loop when evaluating top-level expressions). Evaluate whether a simple goroutine launch with buffered channels is sufficient or if a custom scheduler is needed to preserve ordering.

Helper natives:

- `future_yield()` causes the task to reschedule (can be modelled via channels + re-queuing on an internal worker loop).
- `future_cancelled()` inspects the current handle’s `cancelRequested` flag.

## 8. Builtins & Modules

- Seed the global environment with runtime functions (`print`, numeric ops, `future_yield`, etc.) as specified.
- Implement module privacy rules: track `is_private` flags on functions/structs/interfaces/unions, enforce during import resolution, expose packages as `package` values (map of exported symbols).
- `DynImportStatement` should produce a lazy handle (e.g., `DynRef`) that resolves members on demand; structure mirrors TypeScript’s `dyn_import` handling.

## 9. Error Handling

- Represent Able `error` values as dedicated structs. `Raise` and `Rescue` propagate via the signal errors mentioned above.
- `PropagationExpression` triggers early return when encountering error variants; align with spec’s `!` semantics.
- `EnsureExpression` executes its cleanup block regardless of success/failure.

## 10. Testing Strategy

- Mirror TypeScript tests where practical by converting AST builders to Go DSL calls.
- Add concurrency-specific tests covering cancellation, yielding, and memoization.
- Include round-trip tests for pattern matching, struct updates, and interface dispatch to guard against regressions.

This outline provides the scaffold for implementing the Go runtime. Subsequent design documents can deep-dive into specific subsystems (e.g., scheduler, module system) as we approach their milestones.
