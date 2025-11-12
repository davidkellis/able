# Able Project Roadmap (v10 → v11 transition)

## Scope
- Maintain a canonical Able v10 language definition across interpreters/tooling **while** seeding the v11 fork so spec + runtime changes have a stable landing zone.
- Keep the Go interpreter as the behavioural reference and ensure the TypeScript runtime + future ports match feature-for-feature (the concurrency implementation strategy may differ).
- Preserve a single AST contract for every runtime so tree-sitter output can target both v10 and v11 branches; document any deltas immediately in the v11 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.

## Existing Assets
- `spec/full_spec_v10.md`: authoritative semantics, mirrored into `spec/full_spec_v11.md` as the base for upcoming edits.
- `interpreter10/`: Bun-based TypeScript interpreter + AST definition (`src/ast.ts`) and extensive tests.
- `interpreter10-go/`: Go interpreter and canonical Able v10 runtime. Go-specific design docs live under `design/` (see `go-concurrency.md`, `typechecker.md`).
- Legacy work: `interpreter6/`, assorted design notes in `design/`, early stdlib sketches. Do not do any work in these directories.

## Ongoing Workstreams
- **Spec maintenance**: keep `spec/full_spec_v10.md` authoritative while staging v11 edits in `spec/full_spec_v11.md`; log discrepancies in `spec/todo.md` and `spec/TODO_v11.md`.
- **Standard library**: coordinate with `stdlib/` efforts; ensure interpreters expose required builtin functions/types; track string/regex bring-up via `design/regex-plan.md` and the new spec TODOs covering byte-based strings with char/grapheme iterators.
- **Developer experience**: cohesive documentation, examples, CI improvements (Bun + Go test jobs).
- **Future interpreters**: keep AST schema + conformance harness generic to support planned Crystal implementation.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones now live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.

## Guardrails (must stay true)
- `interpreter10/scripts/run-parity.ts` is the authoritative entry point for fixtures/examples parity, and `run_all_tests.sh` must stay green (TS + Go unit tests, fixture suites, parity CLI).
- `interpreter10/testdata/examples/` remains curated; add programs only when parser/runtime support is complete (see `design/parity-examples-plan.md` for the active roster).
- Diagnostics parity is mandatory: both interpreters emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`. (✅ CLI now enforces diagnostics comparison; ✅ Go typechecker infers implicit `self`/iterator helpers so warn/strict runs match.)
- The JSON parity report (`tmp/parity-report.json`) must be archived in CI via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR` to keep regression triage fast.
- Track upcoming language work (channel select, advanced dynimport scenarios, interface dispatch additions) so fixtures/examples land immediately once syntax/runtime support exists.

## TODO (as items are completed, move them to LOG.md)

1. **Bootstrap the versioned workspace (blocking)**
   - Mirror every v10 asset into a new `v11/` tree (`spec/`, `design/`, `fixtures/`, `stdlib/`, `parser/`, `interpreters/{ts,go}/`, docs, helper scripts) without mutating the existing copies.
   - Update helper scripts/CLI entry points to accept a `--version` (default `v10` until v11 becomes primary) and explain the workflow in `README.md` + `AGENTS.md`.
   - Run the full v10 + v11 suites (`run_all_tests.sh`, Bun fixtures, Go parity tests) to prove the duplicate tree is green before landing any v11-specific edits.

2. **Freeze and document v10**
   - Move/alias the current `interpreter10*/`, `parser10/`, fixtures, and supporting assets under a `v10/` namespace so the repo clearly separates maintained versions.
   - Update onboarding/docs to spell out when work should target `v10/` vs `v11/`, and keep CI running the frozen v10 tests to catch regressions.
   - Track any residual v10 bugs or missing fixtures here so we can fix them before diverging too far.

3. **Expand the v11 specification**
   - Use `spec/TODO_v11.md` as the checklist for spec work (mutable `=`, map literals, struct updates, type aliases, safe navigation, typed declarations, literal widening, optional generic params, channel `select`, stdlib error reporting, `loop` keyword, Array/String APIs, stdlib packaging, regex/text docs).
   - Draft wording + examples in `spec/full_spec_v11.md`, referencing new fixtures/examples as they land.
   - Update `LOG.md`/`spec/todo.md` whenever a TODO graduates into the formal text.

4. **Implement v11 runtime features**
   - Apply the spec updates to the TypeScript interpreter first, mirror in Go, and keep `fixtures/ast` green via `bun run scripts/run-fixtures.ts` + `go test ./pkg/interpreter`.
   - Extend the parser/AST to cover the new grammar (`loop {}`, triple-quoted strings, textual `and`/`or`, map literals, safe member access) and regenerate fixtures.
   - Expand stdlib + runtime helpers (Array/String APIs, module resolution for `able.*`, regex/text helpers) and validate via the examples, RosettaCode, and LeetCode programs.
