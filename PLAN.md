# Able Project Roadmap (v10 focus)

## Scope
- Maintain a canonical Able v10 language definition across interpreters and tooling, with the written specification remaining the source of truth for behaviour.
- Prioritise the Go interpreter until it matches the TypeScript implementation feature-for-feature (the only intentional divergence is that Go uses goroutines/channels while TypeScript simulates concurrency with a cooperative scheduler).
- Keep the TypeScript and Go AST representations structurally identical so tree-sitter output can feed either runtime (and future targets like Crystal); codify that AST contract inside the v10 specification once validated.
- Document process and responsibilities so contributors can iterate confidently.
- Modularize larger features into smaller, self-contained modules. Keep each file under one thousdand (i.e. 1000) lines of code.

## Existing Assets
- `spec/full_spec_v10.md`: authoritative semantics.
- `interpreter10/`: Bun-based TypeScript interpreter + AST definition (`src/ast.ts`) and extensive tests.
- `interpreter10-go/`: Go interpreter and canonical Able v10 runtime. Go-specific design docs live under `design/` (see `go-concurrency.md`, `typechecker.md`).
- Legacy work: `interpreter6/`, assorted design notes in `design/`, early stdlib sketches. Do not do any work in these directories.

## Ongoing Workstreams
- **Spec maintenance**: keep `spec/full_spec_v10.md` authoritative; log discrepancies in `spec/todo.md`.
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

- **Plan the v11 directory layout**
  - Introduce a top-level `v11/` tree containing `spec/`, `design/`, `fixtures/`, `stdlib/`, `parser/`, and `interpreters/{ts,go}/`, plus versioned helper scripts. Keep the existing `interpreter10*/` directories frozen under `v10/` so future v12 work is a straightforward copy/rename.
  - Update CLI/build scripts to accept a `--version` (default `v11`) and resolve stdlib/modules from the matching subtree. Document the migration steps so contributors know where to land new code once the scaffolding exists.

- **V11 readiness sweep**
  - `bun interpreter10/scripts/run-module.ts run examples/...` now covers root programs, `examples/leetcode/`, and `examples/rosettacode/` (see `tmp/examples_run.log`, `tmp/examples_run_rosettacode.log`). Most non-trivial samples fail today because the v10 surface lacks `loop {}`, tuple returns, array/string helpers (`.size()`, `.push()`, `.substring()`, etc.), and module resolution for `able.*` packages.
  - Added these gaps to `spec/TODO_v11.md` so we have explicit spec work items for: `loop` syntax, tuple/multi-return semantics, a proper Array/String runtime API, mutable `=` assignment behaviour, and stdlib packaging/search paths. This list now mirrors the blockers uncovered while running the examples.

- **Next steps (v11 bring-up)**
  1. Fix mutable `=` assignment in the TS interpreter (and mirror in Go) so `while`/`for` loops can make progress; re-run a subset of Rosetta/LeetCode samples to confirm the hang is gone.
  2. Extend the parser + AST to recognise `loop {}`, triple-quoted strings, and textual `and`/`or`, then add fixtures to cover the new grammar.
  3. Define and implement the minimal Array/String builtin surface (size/push/get/set, substring/split/etc.) along with typed `=` declarations and contextual integer widening so the existing examples can execute without rewriting them.
  4. Bootstrap the `v11/` directory structure + stdlib packaging plan, then start porting the current interpreters/spec/tests into that versioned layout before any v11-only features land.
