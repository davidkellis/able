# Able Project Log

## 2025-11-06 — Tree-Sitter Mapper Modularization Complete
- Go declarations/patterns/imports all run through the shared `parseContext`, removing the last raw-source helpers so both runtimes share one parser contract (`interpreter10-go/pkg/parser/{declarations,patterns,statements,expressions}_parser*.go`).
- TypeScript parser README now calls out the shared context, and `PLAN.md` logged the Step 6 regression sweep so future contributors know the refactor is locked in (`interpreter10/src/parser/README.md`, `PLAN.md`).
- Wrapper exports like `parseExpression(node, source)` / `parseBlock` / `parsePattern` were removed; all Go parser consumers now flow through the context pipeline.
- Tests: `./run_all_tests.sh --typecheck-fixtures=warn` and `cd interpreter10-go && GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser`.
- Follow-up: confirm any remaining helpers that still accept raw source (e.g., host-target parsing) genuinely require it before migrating them to `parseContext`.

## 2025-11-03 — AST → Parser → Typechecker Completion Sweep
Status: ✅ Completed. Parser coverage table reads fully `Done`, TypeScript + Go checkers share the bool/comparison semantics, and `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts` executes cleanly after wiring the remaining builtin signatures, iterator annotations, and error fixture diagnostics into the manifests.

Open items (2025-11-02 audit):
- [x] Iterator literals dropped both the optional binding identifier and optional element type annotation. Update both ASTs + parsers/interpreters so the metadata survives round-trips and execution.
- [x] Teach both typecheckers to honor the iterator element annotation so every `yield` matches the declared element type (TS `src/typechecker/checker.ts`, Go `pkg/typechecker/literals.go` + iterator tests).
- [x] Carry iterator element metadata through `for` loops so typed loop patterns validate across array/range/iterator inputs (TS `src/typechecker/checker.ts`, Go `pkg/typechecker/{control_flow,literals,type_utils}.go` + new cross-loop tests).
- [x] Give the TS checker parity with Go for block/proc/spawn typing and array/range literal inference so all three stages (AST → parser → typechecker) agree on the element/result metadata (§6.8, §6.10, §12.2).
- [x] Enforce async-only builtins (`proc_yield`) and add concurrency smoke tests so TS emits the same diagnostics as Go when authors call scheduler helpers outside `proc`/`spawn`.
- [x] Implement `if`/`while` diagnostics + inference in the TS checker so control-flow expressions match the Go implementation (§8.1/§8.2).
- [x] Mirror Go's match/rescue guard enforcement in the TS checker (§8.1.2 / §11.3).
- [x] Enforce package privacy + import diagnostics in the TS checker so private packages/definitions behave identically to Go (updated `imports.test.ts` + package summaries carry `visibility` metadata).

## Historical Status Notes

### 2025-10-30
- Comments are now ignored during parser → AST mapping for both interpreters.
  - ✅ Go: `ModuleParser` / helper utilities skip `comment`, `line_comment`, `block_comment` nodes and `TestParseModuleIgnoresComments` asserts the behaviour.
  - ✅ TypeScript: `tree-sitter-mapper` filters the same node types; `fixtures_mapper.test.ts` covers the mapping path and `fixtures_parser.test.ts` ensures the raw grammar parses comment-heavy sources.
- TODO: audit remaining parser/mapping gaps per `design/parser-ast-coverage.md` (pipes/topic combos, functional updates, etc.) and backfill fixtures/tests.
- DONE: comment skipping now wired through struct literals, struct patterns, and related mapper helpers across both runtimes.
- TODO: Build end-to-end coverage across **all three facets** (parsing, tree → AST mapping, AST evaluation) for both interpreters. Use the coverage table to drive fixture additions, parser assertions, and runtime tests until every spec feature is green.
- TODO: Extend the **typechecker** suites (Go + TS) so they verify type rules and inference across modules. Assemble an exhaustive inference corpus exercising expression typing, generics, interfaces/impls, and cross-module reconciliation; ensure these scenarios are evaluated alongside runtime fixtures.

### 2025-10-31
- Regenerated the tree-sitter-able artifacts with the freshly rebuilt grammar (interface-composition fix now baked into `parser.c`/`.wasm`) using the local Emscripten toolchain; no diff surfaced, confirming the repo already carried the correct bits.
- Cleared local Go build caches (`.gocache`, `interpreter10-go/.gocache`) and re-ran `GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser` to mimic CI picking up the refreshed grammar without stale entries.
- ACTION: propagate the cache-trim guidance to CI docs if flakes recur; otherwise move on to the remaining parser fixture gaps (`design/parser-ast-coverage.md`).
- Mirrored the TypeScript placeholder auto-lift guardrails inside the Go interpreter so pipe placeholders evaluate eagerly, keeping the shared `pipes/topic_placeholder` fixture green.
- Parser sweep: both the TypeScript mapper and Go parser now skip inline comments when traversing struct literals, call/type argument lists, and struct definitions, with fresh TS/Go tests guarding the behaviour.
- TypeScript checker scaffold landed: basic environment/type utilities exist under `interpreter10/src/typechecker`, exported via the public index, and fixtures respect `ABLE_TYPECHECK_FIXTURES` ahead of the full checker port.
- TypeScript checker now emits initial diagnostics for logical operands, range bounds, struct pattern validation, and generic interface constraints so the existing error fixtures pass under `ABLE_TYPECHECK_FIXTURES=warn`.

### 2025-11-01
- TypeScript checker grew a declaration-collection sweep that registers interfaces, structs, methods, and impl blocks before expression analysis, mirroring the Go checker’s phase ordering.
- Implementation validation now checks that every `impl` supplies the interface’s required methods, that method generics/parameters/returns mirror the interface signature, and flags stray definitions; only successful impls count toward constraint satisfaction.
- Canonicalised type formatting to the v10 spec (`Array i32`, `Result string`) and keyed implementation lookups by the fully-instantiated type so generic targets participate in constraint checks.
- Extended the TypeScript checker’s type info to capture `nullable`, `result`, `union`, and function signatures, and taught constraint resolution to recognise specialised impls like `Show` for `Result string`, with fresh tests covering the new cases.
- Added focused Bun tests under `test/typechecker/` plus the `ABLE_TYPECHECK_FIXTURES` harness run to lock in the new behaviour and guard future regressions.
- Mirrored the Go checker and test suite to use the same spec-facing type formatter and wrapper-aware constraint logic so diagnostics now reference `Array i32`, `string?`, etc., and added parity tests (`interpreter10-go/pkg/typechecker/*`).

### 2025-11-02
- The Bun CLI suite (`interpreter10/test/cli/run_module_cli.test.ts`) now covers multi-file packages, custom loader search paths via the new `ABLE_MODULE_PATHS` env, and strict vs warn typecheck enforcement so the ModuleLoader refactor stays covered without pulling `stdlib/` into every run.
- Introduced the `able test` skeleton inside `scripts/run-module.ts`: it parses the planned flags/filters, materialises run options + reporter selection, and prints a deterministic plan summary before exiting with code `2` while the stdlib testing packages remain unparsable. (See `design/testing-cli-design.md` / `design/testing-cli-protocol.md`.)
- Extracted the shared package-scanning helpers (`discoverRoot`, `indexSourceFiles`, etc.) into `scripts/module-utils.ts` so other tooling (fixtures runner, future harnesses) can reuse the multi-module discovery logic without duplicating it.
- **Deferral noted:** full stdlib/testing integration is still on pause until the parser accepts `stdlib/v10/src/testing/*`; once that unblocks, wire the CLI skeleton into the able.testing harness per the design notes.

### 2025-11-05
- Step 6 regression sweep ran end-to-end: `./run_all_tests.sh --typecheck-fixtures=warn` stayed green post-refactor, and `GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser` uncovered/validated the Go-side AST gaps.
- Go AST parity improvements: `FunctionCall.arguments` now serialises as an empty array when no args exist, `NilLiteral` carries `value: null`, and break/continue statements omit label/value when not supplied (matching the TS mapper contract).
- Parser helper normalisation no longer nukes empty slices before fixture comparison, and parameter lists default to `[]`, which brought the channel/mutex fixtures back in sync.
- Fixture `structs/functional_update` gained explicit `\"isShorthand\": false` flags so both interpreters agree on struct literal metadata going forward.

### 2025-11-06
- Go CLI now exposes `able check` alongside `able run`, sharing the manifest/target resolution pipeline. Both commands surface ProgramChecker diagnostics + package summaries and fail fast when typechecking reports issues, keeping the TypeScript + Go CLIs aligned.
- Added dedicated CLI tests to cover `able check` success/failure cases so future refactors keep the new mode wired through manifest resolution, typechecker reporting, and exit codes.
