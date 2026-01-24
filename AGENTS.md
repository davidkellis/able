# Able Project — Agent Onboarding

This document gives contributors the context required to work on the Able v11 effort (TypeScript + Go interpreters, spec, and tooling). The v10 workspace is frozen for archival purposes; do **not** modify it unless a maintainer explicitly calls out a critical fix.

Unconditionally read PLAN.md plus `spec/full_spec_v11.md` before starting any work. `spec/full_spec_v10.md` is reference material only.

## Mission & Principles
- Keep the v10 spec (`spec/full_spec_v10.md`) authoritative for historical reference while staging all new language text in `spec/full_spec_v11.md`. Treat the v11 spec as the arbiter for new runtime work.
- Treat the AST as part of the language contract with the Go definitions as canonical. Every interpreter must share the same structure and field semantics; the spec codifies this contract.
- Use the Go interpreter as the reference implementation while ensuring the TypeScript interpreter—and any future runtimes—match the same semantics defined in the spec.
- Prefer incremental, well-tested changes. Mirror behavior across interpreters whenever possible.
- Document reasoning (design notes in `design/`, issue trackers) so future agents can follow decisions.
- Keep shared fixtures (`fixtures/ast`) green in both interpreters; every fixture change must be exercised by `bun run scripts/run-fixtures.ts` and the Go parity tests (`go test ./pkg/interpreter`).
- Align code changes with the current design notes (e.g., `design/pattern-break-alignment.md`) and update `spec/todo.md`/`PLAN.md` when work lands.
- Modularize larger features into smaller, self-contained modules. Keep each file under one thousdand (i.e. 1000) lines of code.
- Defer AST mapping work until the parser produces the expected parse trees (as captured in the grammar corpus) for every feature under development; once grammar coverage is complete and stable, implement the mapping logic.

## Repository Map
- `v10/`: Frozen Able v10 workspace kept for historical context. Do not edit unless a maintainer assigns a blocking hotfix.
- `v11/`: Active Able v11 workspace with the same structure (`interpreters/{ts,go}/`, `parser/`, `fixtures/`, `stdlib/`, `design/`, `docs/`). All day-to-day work happens here.
- `spec/`: Language specs (v1–v11) and topic supplements.
- `interpreter6/`, `old/*`: Historical artifacts; do not modify.

## Getting Started
1. Read `spec/full_spec_v11.md` (skim `spec/full_spec_v10.md` only for historical context) plus the versioned README under `v11/`.
2. Review `PLAN.md` for roadmap updates. The v10 sub-plan is archival; only update the root PLAN when progressing v11 work.
3. Set up tooling:
   - **Go**: Go ≥ 1.22. Run `go test ./...` inside `v11/interpreters/go`.
   - **TypeScript**: Bun ≥ 1.2. Run `bun install && bun test` inside `v11/interpreters/ts`.
4. If you regenerate tree-sitter assets in `v11/parser/tree-sitter-able`, force Go to relink the parser by deleting `v11/interpreters/go/.gocache` or running `cd v11/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`.
5. Before changing the AST, confirm alignment implications for every interpreter and the future parser.
6. Use `./run_all_tests.sh --version=v11` to run the TypeScript + Go suites together before handing work off. Only run the v10 variant if a maintainer explicitly requests verification of a historical regression.

## Collaboration Guidelines
- Update relevant PLAN files when you start/finish roadmap items. The current typechecker roadmap lives in `design/typechecker-plan.md`.
- Keep `spec/todo.md` current when implementation work exposes gaps that need spec wording updates.
- Treat the shared AST contract as canonical: when introducing new node structures or runtime semantics, implement them in both v11 interpreters and update fixtures so every runtime interprets them identically.
- When adding Go features, port or mirror the corresponding TypeScript tests (or vice versa) within `v11/interpreters`.
- When adding or modifying fixtures, update `v11/fixtures`, run the exporter + TS harness (`bun run scripts/export-fixtures.ts`), and confirm the Go parity test (`go test ./pkg/interpreter`) still passes.
- Fixture manifests can include an optional `setup` array when multi-module scenarios are required (e.g., dyn-import packages); both harnesses evaluate those modules before the entry `module.json`.
- Use concise, high-signal comments in code. Avoid speculative abstractions; match the TS design unless we have a strong reason to diverge.
- Update the spec (`spec/full_spec_v11.md`) as behaviour becomes canonical; log gaps in `spec/TODO_v11.md`. v10 spec edits should only happen for emergency errata.
- At the end of every session: document progress, current state, and next steps; update PLAN/todo/docs accordingly; capture lessons/process adjustments in design notes so the next contributor can resume seamlessly.
- Mark off and remove completed items from the PLAN file once they are complete.
- Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines.
- Tests should run quickly; no test should take more than one minute to complete.

## Concurrency Expectations
- TypeScript interpreter uses a cooperative scheduler to emulate Able `spawn`/`Future` semantics; the helper functions `future_yield()`, `future_cancelled()`, `future_flush()`, and the diagnostic `future_pending_tasks()` are available inside Able code so fixtures/tests can drive (and introspect) the scheduler deterministically.
- Go interpreter must implement these semantics with goroutines/channels while preserving observable behavior (status, value, cancellation, cooperative helpers). The Go runtime exposes the same helper surface (including `future_flush`) via native functions.
- Document any deviations or extensions; tests should exercise cancellation, yielding, and memoization scenarios.

## When in Doubt
- Ask for clarification if spec vs implementation conflicts arise.
- Capture decisions in `design/` and reference them from PLAN documents.
- Keep interoperability in mind: the eventual tree-sitter parser must emit ASTs that both interpreters can evaluate without translation loss.
