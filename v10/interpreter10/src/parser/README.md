# Parser module layout

The TypeScript mapper now mirrors the modular plan in `PLAN.md`. Each helper file registers itself onto the shared `ParseContext` so recursive helpers only talk through the interface instead of importing one another directly.

- `shared.ts` – foundational helpers (`MapperError`, span/origin utilities, `ParseContext` factory, plus `setActiveParseContext`/`parseLabel`).
- `imports.ts` – package/import clauses + qualified identifier parsing.
- `literals.ts` – literal expressions that do not depend on other helpers.
- `types.ts` – type expressions, generics, and where-clause helpers.
- `patterns.ts` – pattern parsing (identifier/struct/array/typed/literal) wired through `ParseContext`.
- `statements.ts` – statement + block parsing, delegating definitions via the context.
- `expressions.ts` – all expression forms (calls, pipes, match/if, rescue/ensure, iterators, etc.).
- `definitions.ts` – structs, unions, interfaces, impls, functions, extern/prelude helpers.
- `tree-sitter-mapper.ts` – the thin orchestrator: instantiate a context, register the helper modules, and map a source file into an AST module.
- The Go parser in `interpreter10-go/pkg/parser` is adopting the same `ParseContext` contract; imports, declarations, patterns, and statements now call context methods to keep both runtimes aligned.

When adding new helpers, follow the same pattern:

1. Extend `ContextFns`/`ParseContext` in `shared.ts` if new entry points are needed.
2. Register the helper in its module (`registerXYZParsers(ctx)`).
3. Updates should only communicate across modules via `ctx.parse*` to keep boundaries explicit.

## Spec Cross-References

- Expressions, statements, and control-flow helpers align with §6–§8 of [`spec/full_spec_v10.md`](../../spec/full_spec_v10.md).
- Type/generic helpers map to §4 and §10 references in the same spec file; keep TODOs in sync with `spec/todo.md`.
- Concurrency/rescue helpers mirror §11–§12 semantics; document any delta in `design/channels-mutexes.md`.
