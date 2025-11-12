# TypeScript Typechecker Parity Plan

Date: 2025-10-20  
Owner: Able Agents

## Context

The Go typechecker recently relaxed several behaviours to mirror the runtime more closely:

- Declaration bindings are installed before analysing the right-hand side so recursive async/proc scenarios do not trip undefined errors (`interpreter10-go/pkg/typechecker/literals.go:203`).
- Wildcard `dynimport` and static wildcard imports seed placeholder bindings and enable dynamic name lookup (`interpreter10-go/pkg/typechecker/checker.go:90`, `interpreter10-go/pkg/typechecker/literals.go:255`).
- Builtins now include `print`, and async helper diagnostics match the runtime (`interpreter10-go/pkg/typechecker/decls.go:323`, `interpreter10-go/pkg/typechecker/builtins.go:21`).
- Struct functional updates inherit field/positional metadata from the source value (`interpreter10-go/pkg/typechecker/struct_literal.go:80`).
- Literal patterns contribute potential non-match types without forcing false negatives (`interpreter10-go/pkg/typechecker/patterns.go:28`).
- Member access recognises positional selectors (`pair.0`) and honours alias metadata from imports/dynimports (`interpreter10-go/pkg/typechecker/member_access.go:9`, `interpreter10-go/pkg/typechecker/literals.go:250`).

Before the TypeScript checker is enabled, we need parity so both runtimes agree on permissive behaviour and diagnostics.

## Plan of record

1. **Scaffold the TypeScript checker**
   - Establish a `interpreter10/src/typechecker/` module with `checker.ts`, `environment.ts`, `types.ts`, and `inference.ts` mirroring the Go structure.
   - Port declaration collection and builtin registration first so scopes/builtins align.

2. **Port permissive rules**
   - Recreate the pre-binding logic for `:=` assignments and destructuring so recursive proc/async definitions resolve.
   - Track `allowDynamicLookups` toggled by wildcard static/dynamic imports and return `Unknown` for identifiers when enabled.
   - Seed placeholders for import aliases and selectors in the local environment.
   - Register `print`, `proc_yield`, `proc_cancelled`, and `proc_flush` with signatures that match the Go checker, including async-context diagnostics.
   - When checking struct literals with a functional-update source, merge field and positional types from `StructInstance` / `StructType` sources before applying overrides.
   - Teach literal patterns to infer their literal type and avoid forcing non-match errors when the RHS type is still `Unknown`.
   - Extend member-access to accept integer literal selectors, resolve positional fields, and fallback to method lookup with import alias metadata.

3. **Diagnostics parity**
   - Ensure diagnostic messages use the same text as the Go checker so shared fixtures stay in sync. Reuse the `typechecker:` prefix for easier manifest comparison.
   - Surface diagnostics through a harness-friendly API (matching `InterpreterV10` once the checker integrates with the TS runner).

4. **Tests and fixtures**
   - Mirror existing Go checker unit tests in TypeScript once the checker lands (start with the fixtures touched above).
   - Update `scripts/run-fixtures.ts` to assert typechecker diagnostics once the TypeScript checker is wired up.

## Open questions

- Baseline coverage: do we gate TypeScript fixtures behind the same `ABLE_TYPECHECK_FIXTURES` flow or keep the checker opt-in until feature parity is complete?
- Shared diagnostic struct: should we define a JSON-serialisable diagnostic model so both harnesses can reuse the same manifest expectations?

## Next actions

1. Implement TypeScript checker scaffolding (env/types/inference map).
2. Port dynamic import/alias handling and assignment pre-binding.
3. Port struct literal seeding, literal pattern inference, and positional member access.
4. Add unit coverage mirroring the Go tests and keep fixtures green with `./run_all_tests.sh --typecheck-fixtures=strict`.
