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
| For loops / array iteration | `for.test.ts`, `for_destructuring.test.ts` | ✅ limited |
| While loops | `while.test.ts` | ⚠️ implemented but untested |
| Ranges (`1..10`) | `ops_range.test.ts` | ⚠️ implemented but untested |
| Break/continue (with labels) | `break.test.ts`, `breakpoint_labeled.test.ts` | ✅ Unlabeled continue supported; labeled continue rejected per spec |
| If/or & else | `if_or.test.ts` | ⚠️ implemented but untested |
| String interpolation | `string_interpolation.test.ts` | ✅ Covered (`interpreter_test.go`, fixtures/ast/strings) |
| Match / pattern matching | `match.test.ts` | ✅ Covered (`interpreter_test.go`, fixtures/ast/match) |
| Error handling (`raise`, `rescue`, `ensure`, `or else`, `rethrow`) | `error_handling.test.ts`, `rethrow.test.ts` | ⚠️ partial (core expressions done; remaining diagnostics) |
| Modules & imports | `module.test.ts`, `import_alias.test.ts`, `wildcard_import.test.ts`, `package_alias.test.ts`, `dynimport.test.ts` | ⚠️ selectors/alias, wildcard (fixtures), and dyn import metadata covered; re-export chains still pending |
| Privacy (modules/interfaces/methods) | `privacy_import.test.ts`, `privacy_methods.test.ts`, `privacy_interface_import.test.ts` | ❌ missing |
| Structs & functional update | `structs.test.ts`, `struct_functional_update.test.ts` | ✅ |
| Static methods | `static_methods.test.ts` | ✅ |
| Interfaces & impls | `methods_impls.test.ts`, `interface_dynamic_dispatch.test.ts`, `interface_default_methods.test.ts`, `impl_resolution.test.ts`, `impl_generic_constraints.test.ts`, `impl_ambiguity.test.ts` | ⚠️ default methods, dynamic dispatch, union specificity, and constraint precedence covered; generic introspection tooling pending |
| Generics & constraints | `generic_constraints.test.ts`, `method_generic_constraints.test.ts`, `generic_type_introspection.test.ts`, `type_args.test.ts` | ❌ missing |
| UFCS | `ufcs.test.ts` | ✅ |
| Index assignment/read | `index_assign.test.ts`, `index_read.test.ts` | ✅ basic cases |
| Compound assignments | `compound_assign.test.ts` | ✅ basic cases |
| Bitshifts & range validation | `bitshift_range.test.ts` | ❌ missing |
| Breakpoints | `breakpoint.test.ts` | ❌ missing |
| Concurrency (`proc`, `spawn`, futures) | `proc_spawn.test.ts` | ❌ missing |
| Dyn import privacy & metadata | `privacy_interface_import.test.ts`, `dynimport.test.ts` | ❌ missing |

## Outstanding Gaps (Go vs TypeScript)

The table above captures line-item status; the following themes summarise what still blocks full parity:

- **Control flow coverage** — While unlabeled `continue` and labeled `break` now align, Go still lacks thorough testing for while/range edge cases, labeled `continue` rejection scenarios, and the full `if/or` short-circuit suite.
- **Modules & imports** — Re-export chains, nested manifest setups, and dyn-import metadata aren’t implemented yet. Privacy diagnostics for interfaces/methods must match TS strings exactly.
- **Interfaces, impls, and generics** — Go is missing UFCS, default methods, advanced impl resolution, generic constraints/introspection, and named/unnamed impl ambiguity checks exercised in TS.
- **Data access & operators** — Bitshift operators and range validation remain, but index reads/writes and compound assignments now have basic coverage.
- **Error reporting** — Deeper rescue/raise diagnostics and payload metadata still diverge from TS expectations.
- **Concurrency** — No goroutine-based scheduler exists yet; `proc`/`spawn` handles, futures, cancellation, and cooperative helpers remain unimplemented.
- **Tooling parity** — Breakpoint signalling and dyn-import privacy metadata are absent, so debugger fixtures and parity diffing can’t run end-to-end.

## Parity backlog (theme-by-theme)

### Control flow & iteration
- [ ] Port the full `for.test.ts`, `for_destructuring.test.ts`, and `while.test.ts` suites into Go, including guard rails for loop-local scopes and remaining edge cases. We currently exercise key fixtures; Go unit coverage has expanded but still misses some TS scenarios.
- [x] Confirmed that labeled `continue` is not part of Able v10; both interpreters raise `"Labeled continue not supported"` to stay aligned with the specification.
- [ ] Harden range evaluation so inclusive/exclusive and reverse ranges match TS diagnostics (`ops_range.test.ts`). Add fixtures that cover descending ranges and non-i32 endpoints once supported.
- [ ] Ensure `if/or` evaluation matches TS short-circuit semantics and port focused assertions from `if_or.test.ts` (truthiness checks, else fallthrough).

### Modules, imports & privacy
- [x] Flesh out dynamic import support (`dynimport.test.ts`) using the shared package registry; hydrate fixtures once semantics match TS.
- [ ] Enforce privacy filters for functions, structs, methods, and interfaces on selector and alias imports (`privacy_import.test.ts`, `privacy_methods.test.ts`, `privacy_interface_import.test.ts`). Go must surface the same error strings and binding behaviour.
- [ ] Extend coverage for module wiring beyond the happy-path cases: more complex re-export chains, nested package manifests, and explicit dyn-import metadata (basic re-export fixture added under `fixtures/ast/imports/static_reexport`; richer chains from `module.test.ts` still pending).
- [ ] Capture dyn-import metadata (package path, privacy) within `runtime.PackageValue` so TS ↔ Go fixtures can assert parity (`dynimport.test.ts`).

### Interfaces, generics & dispatch
- [x] Implement UFCS fallback so free functions act as methods when missing inherent/impl coverage (`ufcs.test.ts`).
- [ ] Implement remaining interface method lookup nuances and dispatch precedence (`methods_impls.test.ts`). The Go runtime needs the same trait resolution order as TS.
- [x] Add support for default interface methods and override priority (`interface_default_methods.test.ts`, `interface_dynamic_dispatch.test.ts`).
- [ ] Port the impl resolution + ambiguity diagnostics (`impl_resolution.test.ts`, `impl_ambiguity.test.ts`, `named_impls.test.ts`) so conflict errors line up (union specificity, where-clause precedence, and interface inheritance now covered; spec-introspection hooks still outstanding).
- [ ] Introduce generic type parameters for functions and structs, along with constraint enforcement (`generic_constraints.test.ts`, `method_generic_constraints.test.ts`, `generic_type_introspection.test.ts`, `type_args.test.ts`, `struct_generic_constraints.test.ts`).
- [ ] Map type metadata on runtime values so introspection helpers used in TS tests have Go equivalents.

### Data access & operators
- [x] Add compound assignment operators (`+=`, `-=`, `*=`, `/=`, etc.) for identifiers, members, and indexed accesses (`compound_assign.test.ts`).
- [x] Implement index read/write on arrays (`index_read.test.ts`, `index_assign.test.ts`); extend to strings/maps when those runtime types arrive.
- [ ] Complete the bitshift operators and range validation diagnostics (`bitshift_range.test.ts`).
- [ ] Expand arithmetic/comparison coverage to handle float/integer mixing consistent with TS semantics.

### Error handling & diagnostics
- [ ] Align remaining error messaging for nested rescues, rethrows, and propagation to match TS assertions (`error_handling.test.ts`, `rethrow.test.ts`). Some fixtures cover the happy paths; diagnostics still differ.
- [ ] Ensure raise-stack metadata surfaces identical payloads for nested errors so typed-pattern fixtures remain consistent across runtimes.

### Concurrency & scheduling
- [ ] Design and implement the goroutine-based scheduler covering `proc` handles, `spawn`, futures, cancellation, and memoisation (`proc_spawn.test.ts`). This includes mapping Able’s cooperative helpers to deterministic Go primitives.
- [ ] Add runtime structs for `ProcStatus`, `ProcError`, and plumbing for `value()`, `status()`, and `cancel()` so tests observe the same state transitions as TS.
- [ ] Extend fixtures under `fixtures/ast` to cover basic proc scenarios once the Go runtime can execute them; mirror expectations in manifests.

### Tooling & observability
- [ ] Implement breakpoint signalling (`breakpoint.test.ts`) so debugger fixtures can validate hook behaviour.
- [ ] Record dyn-import privacy metadata in logs/fixtures as TS does, ensuring parity harnesses can diff results.

## Immediate next actions

1. Prioritise the modules/import privacy work: port the remaining TS cases around re-exports and dyn-import metadata, keeping selector/alias privacy aligned.
2. Draft the concurrency design note (goroutine scheduler, channels, proc handles) so implementation can begin once module privacy lands.
3. Keep this backlog in sync with each milestone—when a theme moves forward, update the relevant checklist items, add fixtures, and note progress in `LOG.md` and `PLAN.md`.

_NOTE:_ The only intentional difference between the interpreters is the concurrency implementation detail (Go goroutines/channels vs TypeScript cooperative scheduler). All observable Able semantics must remain identical.
