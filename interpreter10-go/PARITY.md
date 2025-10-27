# Go Interpreter Parity Checklist

The TypeScript interpreter (`interpreter10/`) is currently our reference implementation for Able v10 semantics. The Go interpreter must eventually support the same behaviour (with goroutines/channels replacing the cooperative scheduler used in TypeScript).

This document tracks the remaining gaps between the two interpreters. For each feature the table lists the primary TypeScript test coverage and the current Go status.

| Feature area | TypeScript coverage | Go status |
| --- | --- | --- |
| Literals & identifiers | `primitives.test.ts` | ✅ Covered (`interpreter_test.go`) |
| Blocks, scopes | `primitives.test.ts` | ✅ |
| Binary/unary ops | `ops_range.test.ts` | ✅ basic arithmetic/logic only |
| Assignments (`:=`, `=`) | `assignments.test.ts` | ✅ basic cases |
| Array destructuring | `destructuring_assign.test.ts` | ✅ basic cases |
| For loops / array iteration | `for.test.ts`, `for_destructuring.test.ts` | ✅ Covered (`interpreter_control_flow_test.go`, `interpreter_patterns_test.go`) |
| While loops | `while.test.ts` | ✅ Covered (`interpreter_control_flow_test.go`) |
| Ranges (`1..10`) | `ops_range.test.ts` | ✅ Loop coverage (`interpreter_control_flow_test.go`, `interpreter_patterns_test.go`) |
| Break/continue (with labels) | `break.test.ts`, `breakpoint_labeled.test.ts` | ✅ Unlabeled continue supported; labeled continue rejected per spec |
| If/or & else | `if_or.test.ts` | ✅ Covered (`interpreter_control_flow_test.go`) |
| String interpolation | `string_interpolation.test.ts` | ✅ Covered (`interpreter_test.go`, fixtures/ast/strings) |
| Match / pattern matching | `match.test.ts` | ✅ Covered (`interpreter_test.go`, fixtures/ast/match) |
| Error handling (`raise`, `rescue`, `ensure`, `or else`, `rethrow`) | `error_handling.test.ts`, `rethrow.test.ts` | ✅ Covered (`interpreter_errors_test.go`) |
| Modules & imports | `module.test.ts`, `import_alias.test.ts`, `wildcard_import.test.ts`, `package_alias.test.ts`, `dynimport.test.ts` | ✅ Re-exports and nested packages covered (`interpreter_imports_test.go`) |
| Privacy (modules/interfaces/methods) | `privacy_import.test.ts`, `privacy_methods.test.ts`, `privacy_interface_import.test.ts` | ✅ Covered (`interpreter_privacy_test.go`) |
| Structs & functional update | `structs.test.ts`, `struct_functional_update.test.ts` | ✅ |
| Static methods | `static_methods.test.ts` | ✅ |
| Interfaces & impls | `methods_impls.test.ts`, `interface_dynamic_dispatch.test.ts`, `interface_default_methods.test.ts`, `impl_resolution.test.ts`, `impl_generic_constraints.test.ts`, `impl_ambiguity.test.ts` | ✅ Covered (dispatch precedence & ambiguity) |
| Generics & constraints | `generic_constraints.test.ts`, `method_generic_constraints.test.ts`, `generic_type_introspection.test.ts`, `type_args.test.ts` | ✅ Covered (`interpreter_generics_test.go`) |
| UFCS | `ufcs.test.ts` | ✅ |
| Index assignment/read | `index_assign.test.ts`, `index_read.test.ts` | ✅ basic cases |
| Compound assignments | `compound_assign.test.ts` | ✅ basic cases |
| Bitshifts & range validation | `bitshift_range.test.ts` | ✅ Covered (`interpreter_numeric_test.go`) |
| Breakpoints | `breakpoint.test.ts` | ✅ Covered (`interpreter_control_flow_test.go`) |
| Concurrency (`proc`, `spawn`, futures) | `proc_spawn.test.ts` | ✅ Executor + handle/future semantics implemented; cancellation/yield fixtures run via goroutine parity harness and TypeScript cooperative executor |
| Dyn import privacy & metadata | `privacy_interface_import.test.ts`, `dynimport.test.ts` | ✅ Metadata captured via package/dyn-package handles |

## Outstanding Gaps (Go vs TypeScript)

The table above captures line-item status; the following themes summarise what still blocks full parity:

- **Control flow coverage** — Core while/range/if-or semantics and for-loop destructuring now mirror the TS suite (`interpreter_control_flow_test.go`, `interpreter_patterns_test.go`), leaving no outstanding control-flow gaps.
- **Modules & imports** — Static/dynamic imports (including nested re-exports) now align with TS.
- **Interfaces & generics** — Dispatch precedence, default methods, constraints, and named impl disambiguation mirror TS coverage.
- **Data access & operators** — Bitshift/range checks and mixed numeric comparisons now match TS semantics.
- **Error reporting** — Raise/rescue/or-else/rethrow diagnostics now mirror TS behaviour.
- **Concurrency** — Executor-backed `proc`/`spawn` landed, and both runtimes now share the executor contract while exercising the cancellation/yield fixtures through their respective harnesses.
- **Tooling parity** — Dyn-import metadata and breakpoint signalling are now covered; future work can focus on enhanced debugging hooks.

## Parity backlog (theme-by-theme)

### Control flow & iteration
- [x] Mirror `while.test.ts` and the key `if_or.test.ts` scenarios in Go (`interpreter_control_flow_test.go`).
- [x] Port the remaining `for.test.ts` / `for_destructuring.test.ts` coverage (array sums, struct destructuring) via `interpreter_control_flow_test.go` and `interpreter_patterns_test.go`.
- [x] Confirmed that labeled `continue` is not part of Able v10; both interpreters raise `"Labeled continue not supported"` to stay aligned with the specification.
- [x] Harden range evaluation so inclusive/exclusive and reverse ranges match TS diagnostics (`ops_range.test.ts`), including finite endpoint guards (`interpreter_control_flow_test.go`).

### Modules, imports & privacy
- [x] Flesh out dynamic import support (`dynimport.test.ts`) using the shared package registry; hydrate fixtures once semantics match TS.
 - [x] Enforce privacy filters for functions, structs, methods, and interfaces on selector and alias imports (`privacy_import.test.ts`, `privacy_methods.test.ts`, `privacy_interface_import.test.ts`). New coverage lives in `interpreter_privacy_test.go` and mirrors the TS error strings.
- [x] Extend coverage for module wiring beyond the happy-path cases: nested package paths and multi-hop re-exports are now exercised in `interpreter_imports_test.go`.
- [x] Capture dyn-import metadata (package path, privacy) within `runtime.PackageValue`/`DynPackageValue`; alias tests now assert names and privacy flags.

### Interfaces, generics & dispatch
- [x] Implement UFCS fallback so free functions act as methods when missing inherent/impl coverage (`ufcs.test.ts`).
- [x] Implement remaining interface method lookup nuances and dispatch precedence (`impl_resolution_test.go`, `interpreter_generics_test.go`) so inherent/generic/union ordering matches TS.
- [x] Add support for default interface methods and override priority (`interface_default_methods.test.ts`, `interface_dynamic_dispatch.test.ts`).
- [x] Port the impl resolution + ambiguity diagnostics (`impl_resolution_test.go`, `interpreter_generics_test.go`) so conflict errors line up (union specificity, constraint precedence, named impl disambiguation).
- [x] Introduce generic type parameters for functions and structs, along with constraint enforcement (`generic_constraints.test.ts`, `method_generic_constraints.test.ts`, `generic_type_introspection.test.ts`, `type_args.test.ts`, `struct_generic_constraints.test.ts`).
- [x] Map type metadata on runtime values so introspection helpers used in TS tests have Go equivalents (`T_type` bindings now covered in `interpreter_generics_test.go`).

### Data access & operators
- [x] Add compound assignment operators (`+=`, `-=`, `*=`, `/=`, etc.) for identifiers, members, and indexed accesses (`compound_assign.test.ts`).
- [x] Implement index read/write on arrays (`index_read.test.ts`, `index_assign.test.ts`); extend to strings/maps when those runtime types arrive.
- [x] Complete the bitshift operators and range validation diagnostics (`bitshift_range.test.ts`).
- [x] Expand arithmetic/comparison coverage to handle float/integer mixing consistent with TS semantics.

### Error handling & diagnostics
- [x] Align nested rescues, rethrows, and propagation diagnostics with TS; coverage lives in `interpreter_errors_test.go`.
- [x] Ensure raise-stack metadata surfaces identical payloads for nested errors so typed-pattern fixtures remain consistent across runtimes.

### Concurrency & scheduling
- [x] Design and implement the goroutine-based executor covering `proc` handles, `spawn`, futures, cancellation hooks, and memoisation (`proc_spawn.test.ts` parity baseline).
- [x] Add runtime structs for `ProcStatus`, `ProcError`, and plumbing for `value()`, `status()`, and `cancel()` so tests observe the same state transitions as TS.
- [x] Extend fixtures in the Go harness to cover the remaining concurrency helpers now that handles/futures are wired up (`concurrency/proc_cancel_value`, `concurrency/future_memoization`, `concurrency/proc_cancelled_outside_error`, `concurrency/proc_cancelled_helper`). Skip TS-only fairness traces (e.g. `ABC` scheduler ordering) unless we introduce equivalent coordination helpers on the Go side; TypeScript parity will follow once the executor lands there.
- [x] Port the executor semantics to the TypeScript interpreter and mirror the new fixture coverage.
- [x] Align TypeScript channel/mutex helpers with Go-style blocking semantics (waiting queues, close semantics, cancellation cleanup) and cover them via Bun tests (`test/concurrency/channel_mutex.test.ts`).

### Tooling & observability
- [x] Implement breakpoint signalling (`breakpoint.test.ts`) so debugger fixtures can validate hook behaviour.
- [x] Record dyn-import privacy metadata in logs/fixtures as TS does, ensuring parity harnesses can diff results.

## Immediate next actions

1. Document the shared executor contract across Go and TypeScript runtimes (design notes + spec touchpoints).
2. Evaluate whether fairness-specific concurrency fixtures can be enabled once we add matching coordination helpers in Go.
3. Keep this backlog in sync with each milestone—when a theme moves forward, update the relevant checklist items, add fixtures, and note progress in `PLAN.md`.

_NOTE:_ The only intentional difference between the interpreters is the concurrency implementation detail (Go goroutines/channels vs TypeScript cooperative scheduler). All observable Able semantics must remain identical.
