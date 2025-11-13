# Able Project Roadmap (v10 → v11 transition)

## Scope
- Maintain a canonical Able v10 language definition across interpreters/tooling **while** seeding the v11 fork so spec + runtime changes have a stable landing zone.
- Keep the Go interpreter as the behavioural reference and ensure the TypeScript runtime + future ports match feature-for-feature (the concurrency implementation strategy may differ).
- Preserve a single AST contract for every runtime so tree-sitter output can target both v10 and v11 branches; document any deltas immediately in the v11 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.

## Existing Assets
- `spec/full_spec_v10.md`: authoritative semantics, mirrored into `spec/full_spec_v11.md` as the base for upcoming edits.
- `v10/interpreters/ts/`: Bun-based TypeScript interpreter + AST definition (`src/ast.ts`) and extensive tests for the frozen v10 workspace.
- `v10/interpreters/go/`: Go interpreter and canonical Able v10 runtime. Go-specific design docs live under `v10/design/` (see `go-concurrency.md`, `typechecker.md`).
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
- `v10/interpreter10/scripts/run-parity.ts` (and its v11 counterpart) remain the authoritative entry points for fixtures/examples parity; `./run_all_tests.sh --version=<v10|v11>` must stay green (TS + Go unit tests, fixture suites, parity CLI).
- `v10/interpreter10/testdata/examples/` remains curated; add programs only when parser/runtime support is complete (see the corresponding `design/parity-examples-plan.md` for the active roster).
- Diagnostics parity is mandatory: both interpreters emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`. (✅ CLI now enforces diagnostics comparison; ✅ Go typechecker infers implicit `self`/iterator helpers so warn/strict runs match.)
- The JSON parity report (`tmp/parity-report.json`) must be archived in CI via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR` to keep regression triage fast.
- Track upcoming language work (awaitable orchestration, advanced dynimport scenarios, interface dispatch additions) so fixtures/examples land immediately once syntax/runtime support exists.

## TODO (as items are completed, move them to LOG.md)

1. **Expand the v11 specification**
   - Use `spec/TODO_v11.md` as the checklist for spec work (mutable `=`, map literals, struct updates, type aliases, safe navigation, typed declarations, literal widening, optional generic params, await/async coordination, stdlib error reporting, `loop` keyword, Array/String APIs, stdlib packaging, regex/text docs).
   - Draft wording + examples in `spec/full_spec_v11.md`, referencing new fixtures/examples as they land.
   - Update `LOG.md`/`spec/todo.md` whenever a TODO graduates into the formal text.
   - ✅ Mutable `=` semantics captured in §5.3.1, map literal syntax/semantics documented in §6.1.9, struct functional update rules expanded in §4.5.2, type alias coverage added in §4.7, safe navigation documented in §6.3.4, typed `=` declarations captured in §5.1–§5.1.1, literal widening/contextual typing documented in §6.1.1/§6.3.2, optional generic parameter inference documented in §7.1.5/§4.5/§10.1, the `await` + error surface defined in §12.6–§12.7, `loop`/Array/String runtime APIs documented in §8.2.3/§6.8/§6.1.5, stdlib packaging & search-path rules captured in §13.6–§13.7, and the regex/text modules plus string/grapheme iterators documented in §14.2/§6.12.1. Keep `spec/TODO_v11.md` current as new language work is scheduled.

2. **Implement v11 runtime features**
   - Apply the spec updates to the TypeScript interpreter first, mirror in Go, and keep `fixtures/ast` green via `bun run scripts/run-fixtures.ts` + `go test ./pkg/interpreter`.
   - Extend the parser/AST to cover the new grammar (`loop {}`, triple-quoted strings, textual `and`/`or`, map literals, safe member access) and regenerate fixtures.
   - Expand stdlib + runtime helpers (Array/String APIs, module resolution for `able.*`, regex/text helpers) and validate via the examples, RosettaCode, and LeetCode programs.
