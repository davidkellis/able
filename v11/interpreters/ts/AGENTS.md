## Able v11 Interpreter — Agent Onboarding

This document gives agents the context needed to extend and maintain the Able v11 reference interpreter in `v11/interpreters/ts/`.

Note: The full v11 language specification lives at [full_spec_v11.md](../spec/full_spec_v11.md).

### Repository map (local to this package)

- `src/ast.ts`: v11 AST data model + DSL builders (e.g., `identifier`, `functionCall`, `structLiteral`, plus aliases like `id`, `str`, `int`, `block`, `assign`, `call`).
- `src/interpreter.ts`: single-file reference interpreter (tree-walk evaluator).
- `test/*.test.ts`: Bun test suites (Jest-compatible) verifying each feature.
- `index.ts`: exports `AST` (all builders/types) and `V11` (interpreter APIs).
- `PLAN.md`: implementation plan, current status, and prioritized next steps.
- `README.md`: human-focused overview and usage.

### Setup & commands

```bash
bun install            # install deps
bun test               # run all tests
bun test --watch       # watch mode
bun run typecheck      # TypeScript typecheck (no emit)
```

### Interpreter architecture (src/interpreter.ts)

- Runtime values are a tagged union `RuntimeValue` with kinds:
  - `string`, `bool`, `char`, `nil`, `i32`, `f64`
  - `array`, `range`
  - `function` (closure with `node` + `closureEnv`), `bound_method` (function + `self`), `native_function`, `native_bound_method`
  - `struct_def`, `struct_instance` (named/positional fields)
  - `interface_def`, `interface_value` (dynamic wrapper around a concrete value implementing the interface)
  - `proc_handle`, `future`
  - `error` (message, optional value / `ProcError` payload)
- `Environment` provides lexical scoping (`define`, `assign`, `get`) with nesting.
- Control-flow uses signals (Errors): `ReturnSignal`, `RaiseSignal`, internal `BreakSignal` shim, and `BreakLabelSignal` for labeled non-local jumps.
- Method lookup merges:
  - Inherent methods from `MethodsDefinition` per type.
  - Interface methods from `ImplementationDefinition` per type.
- Member access on structs first tries fields, then methods; UFCS fallback searches free functions; bound methods inject `self` at call. Interface-typed values forward member access to the underlying concrete type’s implementation.
- Pattern system helpers:
  - `tryMatchPattern` implements match-time patterning (id, wildcard, literal, struct, array, typed).
  - `assignByPattern` implements destructuring assignment and parameter binding (typed patterns coerce interface values automatically).
  - `matchesType` provides minimal runtime checks for `TypedPattern` (simple names, Array T, nullable, etc.), including interface names.

### Feature coverage (implemented)

- Literals, identifiers, blocks; declarations `:=` and assignment `=`
- Operators: arithmetic, comparison, logical, bitwise; ranges
- Arrays: literals, index read/write, and `.index` member-access on arrays
- Control flow: `if/or`, `while` (with `break`), `for` (arrays/ranges)
- Functions/lambdas: closures; destructuring parameters; arity checks; minimal runtime param type checks
- Structs: `StructDefinition`, named/positional `StructLiteral`, member access, static methods, mutation
- Pattern matching: identifier, wildcard, literal, struct, array, typed patterns
- Error handling: `raise`, `rescue` + guards, `else` (or-else), propagation `expr!`, `ensure`, `rethrow`
- Modules/imports: executes in global module env; selector imports + aliasing; private functions cannot be imported
- Methods/impls: inherent and interface-based, with precedence resolved by explicit lookup order. Interface signatures may supply default bodies; impls that omit them inherit the default. Interface-typed bindings coerce to dynamic wrappers so method calls dispatch to the underlying type’s implementation.
  - Union-target impls participate in resolution; smaller variant sets win over larger ones. Constraint supersets (inheriting interface requirements) rank higher, and ambiguity errors now surface candidate details. Dynamic interface containers (arrays/ranges) automatically pick the most specific impl during iteration.
- Concurrency handles: `proc`/`spawn` schedule work asynchronously and return handles exposing `status()`, `value()` (`!T` – success or `ProcError`), and `cancel()`.
  - We keep scheduling inside the interpreter (simple cooperative queue) rather than using JavaScript `async`/`await` so evaluation stays synchronous/ deterministic and cancellation can flip pending tasks immediately without relying on host promise semantics.
  - Cooperative helpers are available to Able code: `proc_yield()` triggers a `ProcYieldSignal`, re-queueing the current runner, and `proc_cancelled()` surfaces the `cancelRequested` flag on the active handle. The interpreter keeps an `asyncContextStack` so these helpers know which async value is executing.
  - `runProcHandle`/`runFuture` push the active handle on the stack before evaluating and pop it in `finally`. They also prevent cancellation from short-circuiting once evaluation has begun via a `hasStarted` flag.
  - `procHandleCancel` now schedules the existing runner instead of marking the handle immediately; cancellation is finalized inside the next `runProcHandle` tick so task code can poll `proc_cancelled()` and exit cooperatively. Tests may call `runProcHandle` directly (via `Interpreter` cast to `any`) when they need fine-grained control over staging.
  - When you add additional async helpers, mirror this pattern: throw a dedicated signal, schedule the existing runner if the task should resume, and avoid leaking the helper into normal synchronous execution.

See `README.md` for human-facing details and examples.

### Development workflow (strongly recommended)

1) Write focused tests under `test/` for the exact behavior you intend to add/change.
2) Implement the evaluator change in `src/interpreter.ts`:
   - Add/extend a `case` in the big `switch (node.type)` inside `evaluate`.
   - If a new runtime shape is needed, add a kind to `RuntimeValue`.
   - Prefer existing helpers (`assignByPattern`, `tryMatchPattern`, `matchesType`, `findMethod`, `valueToString`, `isTruthy`).
3) Run `bun test` and fix failures. Keep tests deterministic.
4) Lint/typecheck via `bun run typecheck` if you change types significantly.
5) Update `PLAN.md` if you complete a milestone or add new next steps. Update `README.md` if user-facing behavior changes.

### Iteration workflow used in this project

This is the concrete cadence we’ve been following to add features safely and quickly:

1) Plan the milestone
   - Create a short TODO list (one item per meaningful outcome).
   - Mark only the first item as in-progress.
   - Ensure there’s a clear acceptance criterion (tests passing, lints clean, docs updated if user-facing).

2) Write evaluation-first tests
   - Add focused tests under `test/` for the new feature.
   - Include at least one “evaluation” test that builds an AST using the new feature and asserts the interpreter’s result (e.g., addition evaluates to the expected sum).
   - Add both success and failure-path tests (e.g., type/shape mismatches, privacy violations).

3) Implement narrowly in `src/interpreter.ts`
   - Add or extend a `case` under `evaluate` for the new AST node(s).
   - Prefer reusing helpers (`assignByPattern`, `tryMatchPattern`, `matchesType`, `valueToString`, `isTruthy`, `findMethod`).
   - Keep the change localized and readable; avoid refactors unrelated to the feature.

4) Run and fix
   - `bun test` until green; keep tests deterministic.
   - `bun run typecheck` and fix type/lint issues.
   - If edits grew, add/adjust failure-path tests.

5) Update docs and plan
   - If behavior is user-facing, update `README.md` (examples, caveats).
   - Update `PLAN.md` (mark milestone complete; add the next prioritized steps).
   - Check off the TODO and set the next one to in-progress.

6) Repeat for the next milestone
   - Work from foundational features upward (primitives → ops/ranges → control flow → data → errors → modules → methods/impls → concurrency).

Per-feature checklist:
- [ ] Tests written (evaluation + failure paths)
- [ ] `evaluate` case added / updated
- [ ] Helpers reused; new helper added only if necessary
- [ ] `bun test` green; `bun run typecheck` clean
- [ ] Docs/plan updated; TODO advanced

### Conventions & guardrails

- Do not reformat unrelated code. Keep existing indentation and style.
- Prefer explicit, readable code paths over cleverness; match existing patterns in the file.
- When adding new features, ensure both success tests and failure-path tests (errors) exist.
- Enforce privacy only where specified (currently: private functions cannot be imported).
- Generic type arguments are accepted at runtime but not typechecked (tests document this contract).

### Common pitfalls

- Remember that blocks create a new `Environment`; `while`/`for` bodies re-evaluate per iteration with a child env.
- Bound methods inject `self`; ensure arity checks include injected args.
- `TypedPattern` is a minimal runtime check; it should not attempt full typechecking.
- Module evaluation currently uses the global env to persist top-level definitions.

### Extending with a new AST feature (checklist)

- [ ] Add or confirm the AST node exists in `src/ast.ts` and ensure builders exist for tests
- [ ] Add a `case` in `evaluate` and supporting helpers as needed
- [ ] If a new runtime value is required, extend `RuntimeValue` with a new `kind`
- [ ] Add tests (success + failure) in `test/`
- [ ] Run `bun test` to confirm green; keep lints clean
- [ ] Update `PLAN.md` status/next steps if applicable; update `README.md` if user-facing

### Next steps (short)

See `PLAN.md` for the prioritized backlog. Near-term items currently include:

- Concurrency ergonomics (yield hooks, cancellation observability, proc/future stress tests)
- Interface & impl completeness (higher-kinded constraint chains, mixed visibility cases, disambiguation guidance)
- Dynamic interface collections (ranges/maps of interface values, nested containers) and remaining privacy/import spec gaps

### Handy AST DSL examples

```ts
import { AST } from "./index";

// 2 + 3
const sum = AST.binaryExpression("+", AST.integerLiteral(2), AST.integerLiteral(3));

// fn add(a, b) { return a + b }
const addFn = AST.functionDefinition(
  "add",
  [AST.functionParameter("a"), AST.functionParameter("b")],
  AST.blockExpression([AST.returnStatement(sum)])
);

// Point { x: 1, y: 2 }
const pointDef = AST.structDefinition(
  "Point",
  [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
  "named"
);
const pointLit = AST.structLiteral([
  AST.structFieldInitializer(AST.integerLiteral(1), "x"),
  AST.structFieldInitializer(AST.integerLiteral(2), "y"),
], false, "Point");
```

If in doubt, search the tests in `test/` for ready-made examples of each feature.
