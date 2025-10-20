# Able v10 Typechecker Plan (Go)

Date: 2025‑10‑19  
Owner: Able Agents

## Goals

- Build a static typechecker that walks the existing AST without requiring
  structural changes.
- Provide reusable diagnostics for both the Go interpreter and future compiler.
- Keep inferred type metadata external to the AST (side tables) to preserve the
  serialisable contract shared with other runtimes.

## Architecture sketch

### Package layout

- `pkg/typechecker`
  - `checker.go` – entry point (`CheckModule(*ast.Module) ([]Diagnostic, error)`).
  - `env.go` – symbol tables (scoped maps of names → `TypeInfo`).
  - `types.go` – definitions for value types, interface constraints, generics,
    and utility builders.
  - `diagnostics.go` – diagnostic struct with message, optional source span, and
    severity.
  - `constraints.go` – trait/where-clause solving utilities (likely shared with
    interpreter impl resolution).

### Phases

1. **Declaration collection** – Walk module statements to register structs,
   unions, interfaces, and function signatures. Produces a global environment.
2. **Implementation collection** – Validate inherent/trait impl headers and
   record available methods for later resolution.
3. **Body checking** – For each function/proc/spawn, create a scoped environment
   and recursively check expressions/statements, populating an inference map.
4. **Constraint solving** – After body checks, ensure where-clause constraints
   and trait obligations hold. Reuse logic from `impl_resolution.go` where
   possible.

### Data structures

- `TypeInfo` union covering primitives, structs with generics, unions, function
  types, interface references, and proc/future handles.
- `InferenceMap` keyed by `ast.Node` (pointer) storing resolved `TypeInfo`.
- `Diagnostic` capturing message, optional `Span` (when available), and context
  (e.g., offending identifier).

### Error handling

- Checker should accumulate diagnostics and continue where safe.
- Runtime execution can still proceed without typechecking, so the API should
  return diagnostics but no runtime errors unless the AST is malformed.

## Immediate progress

- Literal, control-flow, async, and aggregate expressions now feed precise type
  inference across the core language surface.
- Declaration collection captures generics, where clauses, interface method
  signatures, and flags duplicate definitions prior to body checking.
- Pattern typing supports identifiers, wildcards, struct/array patterns, and
  typed wrappers; tests assert inference for match/assignment scenarios.
- Diagnostics cover undefined identifiers, duplicate declarations, arity/type
  mismatches in calls, control-flow misuse, async helper constraints, and
  now `Self`-scoped method-set where clauses (e.g., `Formatter<string>`).

## Next improvements

With the v10 surface covered, future iterations should focus on:

1. **Source spans & tooling hooks** – plumb parser-provided span data into
   diagnostics once the tree-sitter grammar lands so IDEs/linters can reuse the
   checker.
2. **Incremental checking** – explore module-level caching and invalidation so
   future compiler/LSP components can reuse inference results quickly.
3. **TypeScript parity** – mirror the Go checker behaviour in the Bun
   interpreter so `ABLE_TYPECHECK_FIXTURES` can be enforced on both runtimes.

## Dependencies / assumptions

- AST contract is frozen as described in `design/ast-contract.md`.
- No source span data yet; diagnostics should tolerate missing spans. If we add
  spans later, extend `nodeImpl` without breaking existing fixtures.
- Interpreter remains the execution reference; typechecker integration should be
  optional until completeness is achieved.

## Open questions

- **Proc/future typing** – represent handles as nominal `Proc<T>`/`Future<T>` to
  mirror runtime behaviour.
- **Interfaces vs traits** – ensure checker reuses the same resolution order as
  `impl_resolution.go` to avoid divergent semantics.
- **Interop with compiler** – capture enough metadata (inference map, resolved
  impls) so the future compiler can reuse the same results.

## Deliverables checklist

- [x] Declaration collection pass wired to `CheckModule`.
- [x] Expression/type coverage beyond literals for the core language surface.
- [x] Diagnostics surfaced for redefinitions and undefined symbols (extend to
      trait/impl violations).
- [x] Integration harness (optional flag) that runs checker before interpreter.
- [x] Documentation updates (README/design) as features land.
