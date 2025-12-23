# Able v11 Typechecker Plan (Go)

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
- TypeScript parity locked in for method-set diagnostics and typed pattern
  tolerance, with Bun unit tests guarding the new `Self` obligation errors and
  ensuring typed assignments remain advisory rather than hard failures.

## Next improvements

With spans flowing from both parsers and the diagnostics baseline stable, the
next passes should prioritise:

1. **Export surfaces & CLI integration** – finish shaping the Go
   `ProgramChecker` export metadata (privacy-respecting structs/interfaces/fns,
   impl and method-set registries) and surface the summaries through the CLI so
   package-qualified diagnostics and tooling hooks share a single schema.
   Mirror the export capture in the TypeScript checker so Bun consumers and
   future editors receive identical data.
2. **TypeScript parity & fixture enforcement** – close the remaining gaps
   (privacy/import diagnostics, outstanding constraint edges) so the Bun
   interpreter can run `ABLE_TYPECHECK_FIXTURES=warn|strict` without skips.
   Regenerate the shared baseline once parity lands and update CI to execute the
   checker alongside Go.
3. **Incremental checking** – after export surfaces stabilise, explore caching
   package summaries/inference tables to support future compiler/LSP reuse. This
   includes defining invalidation keys and documenting the session lifecycle.
4. **Fixture & doc coverage** – keep extending the shared AST fixtures and
   update design/spec notes whenever new diagnostics or export metadata land, so
   downstream tooling stays aligned.

**Status (2025-11-02):** The TypeScript checker now threads package summaries through the Bun fixture harness, reports unknown package/selector references, and enforces the shared `fixtures/ast/typecheck-baseline.json` expectations whenever `ABLE_TYPECHECK_FIXTURES=warn|strict` is set, keeping TS + Go fixture workflows aligned.

### Package summary schema & CLI behaviour

- **Canonical schema**: `PackageSummary` captures the public API for each
  module (name, exported symbol map, struct/interface/function metadata, and the
  implementation/method-set registries) along with any recorded obligations or
  where-clause requirements. The TypeScript definition lives in
  `v11/interpreters/ts/src/typechecker/diagnostics.ts` and mirrors the Go struct in
  `interpreter-go/pkg/typechecker/program_checker.go`.
- **Go CLI**: `interpreter-go/cmd/able/main.go` prints the summary block under
  `---- package export summary ----` whenever `able run` exits due to
  typechecker errors (see the assertions in `main_test.go`). This keeps users
  informed about which packages failed to export complete surfaces.
- **Bun CLI/harness**: the Bun fixture runner (`v11/interpreters/ts/scripts/run-fixtures.ts`)
  now emits the same summary block whenever diagnostics appear while running
  with `ABLE_TYPECHECK_FIXTURES=warn|strict`. This mirrors the Go output so Bun
  developers and tooling receive the same signals while we wire the broader CLI
  together.

Documenting the shared surface up front makes it easier to extend future CLI
commands (`able check`, `able test`) and editor tooling without re-deriving the
export metadata.

### Multi-package typechecking roadmap (v11 alignment)

- **Session model** – introduce a `ProgramChecker` (or similar) that owns a
  single `Checker`, a registry of package exports, and the dependency-ordered
  module list returned by `driver.Loader`. The session walks each module once,
  guaranteeing that imported packages have already populated their export
  surface before downstream modules are checked. **(Done: `typechecker.ProgramChecker`
  now backs CLI typechecking and caches exports across modules, including struct/interface/function metadata and impl/method-set registries.)**
- **Export surfaces** – extend declaration collection to emit public definitions
  (`struct`, `union`, `interface`, `fn`, constants) and impl/method metadata
  while respecting `private` visibility. The Go side needs the privacy filter
  completed; the TypeScript checker must mirror the capture so both runtimes
  share the same export schema.
- **Package namespace types** – model `import foo;` bindings by introducing a
  dedicated package namespace type that exposes the exported identifiers via
  member access (`foo.Bar`, `foo.main`). Selective and aliased imports bind the
  same export entries directly into scope, and `import foo.*;` splats every
  public symbol (detecting duplicates per v11 rules).
- **Impl propagation** – ensure implementations and method sets defined in one
  package remain available when another package imports it. Merge per-package
  tables into the session export data during capture so cross-package interface
  resolution follows the spec’s visibility and coherence rules.
- **CLI integration** – replace the per-module `Interpreter.EnableTypechecker`
  call with a `CheckProgram` pass inside `able run`/`able check`, failing fast on
  diagnostics and printing package-qualified paths. Update the CLI docs to
  describe the new behaviour.
- **Fixtures/tests** – add cross-package fixtures that exercise imports,
  privacy, aliasing, wildcard splats, and impl visibility. Gate them in both the
  Go CLI tests and the checker’s unit suite so parity with the v11 spec stays
  enforced, and mirror the cases in Bun once parity lands.

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
