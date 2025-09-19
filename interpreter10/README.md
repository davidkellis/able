## interpreter10

Able v10 AST and reference interpreter (TypeScript), built with Bun.

### Quick start

```bash
bun install
bun test
```

Typecheck only:

```bash
bun run typecheck
```

### Language spec

For the complete v10 language definition and semantics, see: [full_spec_v10.md](../spec/full_spec_v10.md).

### What’s in this package

- `src/ast.ts`: v10 AST data model and DSL helpers to build nodes.
- `src/interpreter.ts`: v10 reference interpreter (single pass evaluator).
- `test/*.test.ts`: Jest-compatible Bun tests covering interpreter features.
- `index.ts`: exports `AST` and `V10` (interpreter) for external use.

### Interpreter architecture

The interpreter evaluates AST nodes directly (tree-walk). Key pieces:

- Runtime value union (`V10Value`): string, bool, char, nil, i32, f64, array, range, function, struct_def, struct_instance, error, bound_method, interface_def, union_def, package, impl_namespace, dyn_package, dyn_ref.
- `Environment`: nested lexical scopes with `define`, `assign`, and `get`.
- Control-flow signals: `ReturnSignal`, `RaiseSignal`, loop-only break signal, and labeled `BreakLabelSignal` for non-local jumps to `breakpoint`.
- Method lookup: inherent methods (`MethodsDefinition`) and interface `ImplementationDefinition` registered by type name. Named `impl` blocks are exposed as `impl_namespace` values (explicit calls only).

### Recent updates

- Concurrency handles now execute asynchronously, expose `ProcStatus` structs, and surface `ProcError` payloads through `value()` so downstream code can use `!`/pattern matching without special cases.
- Cooperative helpers `proc_yield()` and `proc_cancelled()` allow long-running tasks to yield control and observe cancellation from inside the task body.
- Cancellation requests queue the handle’s runner instead of flipping state immediately, so tasks can poll `proc_cancelled()` and clean up before `status()` reports `Cancelled`.
- Interfaces support default method bodies; impls inherit them automatically, with `Self` substituted for the concrete target type.
- Values typed as an interface are wrapped as dynamic `interface_value` instances so member access dispatches to the underlying implementation.
- Typed patterns and function parameters annotated with interfaces coerce values into these dynamic wrappers, enabling runtime type checks against interface names.

High-level evaluation flow:

1) Literals and identifiers return their corresponding runtime value or env binding.
2) Expressions: unary/binary ops, calls, blocks, ranges, indexing, string interpolation (prefers `to_string` on structs), member access (fields/methods) with UFCS fallback to free functions.
3) Control flow: if/or, while (with break), for (arrays/ranges).
4) Data: `StructDefinition`, `StructLiteral`, named/positional fields, member access, static methods.
5) Functions/lambdas: closures capture the defining environment; parameters support destructuring patterns.
6) Pattern matching: identifier, wildcard, literal, struct, array, typed patterns (minimal runtime checks).
7) Error handling: raise, rescue (with guards), or-else, propagation `expr!`, ensure, rethrow (with raise stack).
8) Modules/imports: executes body in a module/global env; selector imports and aliasing; privacy enforced for functions/types/interfaces/unions. Wildcard imports bring only public symbols. `import pkg as Alias` binds a `package` value exposing public members. `dynimport` binds late-resolving `dyn_ref`s or `dyn_package` aliases.
9) Concurrency: `proc` returns a lightweight handle (`status`, `value`, `cancel`) backed by `ProcStatus` (`Pending`, `Resolved`, `Cancelled`, `Failed`) and `ProcError`; `spawn` returns a memoizing future handle with the same `status`/`value` API. Tasks start asynchronously; `value()` blocks and returns `!T` (either the underlying value or an `error` whose payload is a `ProcError`, so `!`/pattern matching work naturally).

### Feature coverage (implemented)

- Primitives: string, bool, char, nil, i32, f64
- Arrays: literals, index read/write, member-access by integer `.n`
- Operators: arithmetic, comparison, logical, bitwise, ranges
- Control flow: if/or, while (break), for over arrays/ranges
- Functions/lambdas: closures, destructuring params, call arity checks
- Blocks/assignments: declarations `:=`, reassign `=`, destructuring assignment, compound assignments (`+=`, `-=`, `*=`, `/=`, `%=` and `&=`, `|=`, `^=`, `<<=`, `>>=`)
- Structs: definitions, literals (named/positional), member access, static methods
- Pattern matching: identifier, wildcard, literal, struct, array, typed patterns
- Error handling: raise, rescue (guards), or-else, propagation `!`, ensure, rethrow
- Modules/imports: selector import, aliasing, privacy for private functions/types/interfaces/unions; wildcard imports; package alias; dynimport selectors/alias/wildcard
- Methods/impls: inherent methods and interface impl methods with bound `self`, interface default methods, dynamic dispatch via interface-typed values (`interface`-typed bindings coerce to dynamic wrappers)
  - Resolution now ranks union-target impls by subset size, honours constraint supersets (including inherited interfaces) and surfaces richer ambiguity diagnostics.
  - Dynamic interface containers (e.g., arrays of `Show`) dispatch using the most specific impl in scope.

### Using the interpreter

Programmatic usage:

```ts
import { AST, V10 } from "./index";

const mod = AST.module([
  AST.functionDefinition(
    "add",
    [AST.functionParameter("a"), AST.functionParameter("b")],
    AST.blockExpression([
      AST.returnStatement(AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")))
    ])
  ),
  AST.assignmentExpression(
    ":=",
    AST.identifier("x"),
    AST.functionCall(AST.identifier("add"), [AST.integerLiteral(2), AST.integerLiteral(3)])
  ),
]);

const interp = new V10.InterpreterV10();
const result = interp.evaluate(mod as any); // { kind: 'i32', value: 5 }
```

### Conventions and notes

- Integers default to i32; floats default to f64.
- Type arguments on calls are accepted but not typechecked at runtime.
- Privacy is enforced on import for functions (more privacy rules TBD).
- Destructuring and TypedPattern checks are best-effort runtime validations, not full typechecking. Function parameters with type annotations are validated minimally at call time via `matchesType`.
- Truthiness: `false`, `nil`, and any value of kind `error` are falsy; all others are truthy.

### How the new logic works (high level)

- Package registry and imports:
  - Modules with `package` declarations register top-level definitions in an internal registry keyed by package path. Qualified names (e.g., `pkg.name`) are placed in globals for selector imports. Wildcard imports copy only public symbols into the importing env. `import pkg as Alias` binds a `package` value whose `Alias.member` yields the symbol. `dynimport` binds `dyn_ref` and `dyn_package` placeholders that resolve at use-time.

- Named `impl` exposure:
  - Named `impl` blocks are exposed as `impl_namespace` values (not packages). Methods are accessed as `ImplName.method(...)`. Unnamed impls populate the implicit method tables for instance method calls.

- UFCS fallback:
  - If `receiver.name(...)` doesn’t match a field or method, the interpreter searches for a free function `name` in scope and returns a bound-method-like callable injecting `receiver` as the first argument.

- Labeled `breakpoint`/`break`:
  - `breakpoint 'label { ... }` evaluates the body and returns its last value unless a `break 'label expr` is encountered, in which case the breakpoint returns `expr`. Loop frames propagate labeled breaks to allow unwinding.

- Compound assignments and shifts:
  - Compound ops update identifiers, struct fields, and array indices with single-target evaluation. Left/right shifts validate the count is within 0..31 for `i32`.

- String interpolation and Display:

#### Concurrency handles

- `proc expr` captures the current environment, schedules the work asynchronously (microtask or `setTimeout(0)` fallback), and returns a handle with
  - `status()` → one of `Pending`, `Resolved`, `Cancelled`, `Failed` (struct instances stored in `ProcStatus`).
  - `value()` → blocks until the task finishes and returns `!T` (either the success value or an `error` containing a `ProcError { details }`). Cancels/failures therefore compose with `!` and pattern matching without special cases.
  - `cancel()` → best-effort cancellation that flips the status to `Cancelled` and memoizes a `ProcError` for subsequent `status`/`value` calls.
- `spawn expr` behaves similarly but returns a memoizing future handle. The first `value()` drives the computation; later calls reuse the cached success or failure.
- The interpreter manages a cooperative run queue instead of using JavaScript `async`/`await`. Keeping evaluation synchronous preserves deterministic ordering for tests, avoids turning every `evaluate` call into a promise, and lets cancellation flip pending tasks to `Cancelled` without depending on host-specific microtask behaviour.
- Long-running tasks can call `proc_yield()` to reschedule themselves and `proc_cancelled()` to check whether cancellation has been requested, making observability explicit in task bodies.
- `cancel()` now defers its terminal transition until the next scheduled evaluation so cooperative tasks can observe the flag, perform their own unwinding, and then allow the interpreter to flip the handle to `Cancelled`.

### Cooperative concurrency helpers

`proc_yield()` and `proc_cancelled()` are exposed as global native functions so Able code can cooperate with the interpreter’s scheduler:

- `proc_yield()` throws an internal `ProcYieldSignal`. The interpreter catches it, re-queues the current task’s runner, and unwinds to the scheduler. Use it inside tight loops to give other work an opportunity to progress.
- `proc_cancelled()` returns a boolean indicating whether cancellation has been requested on the current `proc` handle. Tasks can poll it and exit early with their own clean-up logic.

Under the hood the interpreter maintains an `asyncContextStack` so helper invocations can discover the active async value. Both helpers must run inside a `proc`/`spawn` body or they will raise an error. Tests in `test/proc_spawn.test.ts` exercise interleaving (`trace` becomes "ABC") and cooperative cancellation (`trace` becomes "wx" when a loop notices cancellation before the interpreter finalises the handle).

#### Interface defaults & dynamic dispatch

- Interface signatures may include a `defaultImpl` block. When an `impl Interface for Type` omits that method, the interpreter materializes a function using the default body, substituting every `Self` in parameters/return types/patterns with the concrete target type.
- Values typed as an interface (via typed bindings/patterns) are wrapped as `interface_value` runtime values. Member access on these wrappers consults the underlying type’s impl table (filtered to the interface) and dispatches dynamically, enabling the spec’s “interface as type” behaviour.
  - Interpolated expressions are formatted with `valueToStringWithEnv`. For struct instances, the interpreter attempts to call a `to_string(self)` method; if it returns a string, its value is used; otherwise a structural `{ field: value }` is emitted.

### Extending the interpreter

- New node: add to `V10Value` if needed, add a case in `evaluate`, and tests under `test/`.
- Performance: consider caching method lookups and optimizing environment chains.
- Concurrency: evolve `proc`/`spawn` to real async handles with `join`.

### Roadmap highlights

- Concurrency semantics: add cooperative yielding hooks, richer cancellation observability, and broaden proc/future stress tests.
- Interface completeness: cover higher-kinded/inherited constraint chains, mixed visibility scenarios, and strengthen tooling guidance for disambiguation.
- Dynamic collections: exercise ranges/maps of interface values so iteration keeps using the most specific available implementation.

### Development

```bash
bun test            # run all tests
bun test --watch    # watch mode
bun run typecheck   # TypeScript typecheck
```

This project was created using `bun init` in bun v1.2.19. [Bun](https://bun.sh/) is a fast all-in-one JavaScript runtime.
