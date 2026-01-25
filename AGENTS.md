# Able Project — Agent Onboarding

This document gives contributors the context required to work on the Able v12 effort (Go tree-walking + bytecode interpreters, spec, and tooling). The v10 and v11 workspaces are frozen for archival purposes; do **not** modify them unless a maintainer explicitly calls out a critical fix.

Unconditionally read PLAN.md plus `spec/full_spec_v12.md` before starting any work. `spec/full_spec_v10.md` and `spec/full_spec_v11.md` are reference material only.

## Mission & Principles
- Keep the v10/v11 specs authoritative for historical reference while staging all new language text in `spec/full_spec_v12.md`. Treat the v12 spec as the arbiter for new runtime work.
- Treat the AST as part of the language contract with the Go definitions as canonical. Every interpreter must share the same structure and field semantics; the spec codifies this contract.
- Use the Go interpreters (tree-walker + bytecode VM) as the reference implementations while ensuring any future runtimes match the same semantics defined in the spec.
- Prefer incremental, well-tested changes. Mirror behavior across interpreters whenever possible.
- Document reasoning (design notes in `design/`, issue trackers) so future agents can follow decisions.
- Keep shared fixtures (`fixtures/ast`) green across both Go interpreters; every fixture change must be exercised by the Go fixture harnesses (`go test ./pkg/interpreter`).
- Align code changes with the current design notes (e.g., `design/pattern-break-alignment.md`) and update `spec/todo.md`/`PLAN.md` when work lands.
- Modularize larger features into smaller, self-contained modules. Keep each file under one thousdand (i.e. 1000) lines of code.
- Defer AST mapping work until the parser produces the expected parse trees (as captured in the grammar corpus) for every feature under development; once grammar coverage is complete and stable, implement the mapping logic.

## Repository Map
- `v10/`: Frozen Able v10 workspace kept for historical context. Do not edit unless a maintainer assigns a blocking hotfix.
- `v11/`: Frozen Able v11 workspace kept for historical context.
- `v12/`: Active Able v12 workspace with the same structure (`interpreters/go/`, `parser/`, `fixtures/`, `stdlib/`, `design/`, `docs/`). All day-to-day work happens here.
- `spec/`: Language specs (v1–v12) and topic supplements.
- `interpreter6/`, `old/*`: Historical artifacts; do not modify.

## Getting Started
1. Read `spec/full_spec_v12.md` (skim `spec/full_spec_v10.md` and `spec/full_spec_v11.md` only for historical context) plus the versioned README under `v12/`.
2. Review `PLAN.md` for roadmap updates. The v10/v11 sub-plans are archival; only update the root PLAN when progressing v12 work.
3. Set up tooling:
   - **Go**: Go ≥ 1.22. Run `go test ./...` inside `v12/interpreters/go`.
   - **CLI wrappers**: use `./v12/abletw` for tree-walker runs and `./v12/ablebc` for bytecode runs.
4. If you regenerate tree-sitter assets in `v12/parser/tree-sitter-able`, force Go to relink the parser by deleting `v12/interpreters/go/.gocache` or running `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`.
5. Before changing the AST, confirm alignment implications for every interpreter and the future parser.
6. Use `./run_all_tests.sh` (defaults to v12) to run the Go suites before handing work off. Only run v10/v11 variants if a maintainer explicitly requests verification of a historical regression.

## Collaboration Guidelines
- Update relevant PLAN files when you start/finish roadmap items. The current typechecker roadmap lives in `design/typechecker-plan.md`.
- Keep `spec/TODO_v12.md` current when implementation work exposes gaps that need spec wording updates.
- Treat the shared AST contract as canonical: when introducing new node structures or runtime semantics, implement them in both v12 interpreters and update fixtures so every runtime interprets them identically.
- When adding or modifying fixtures, update `v12/fixtures` and run the Go fixture harness (`go test ./pkg/interpreter`).
- Fixture manifests can include an optional `setup` array when multi-module scenarios are required (e.g., dyn-import packages); the Go harness evaluates those modules before the entry `module.json`.
- Use concise, high-signal comments in code. Avoid speculative abstractions.
- Update the spec (`spec/full_spec_v12.md`) as behaviour becomes canonical; log gaps in `spec/TODO_v12.md`. v10/v11 spec edits should only happen for emergency errata.
- At the end of every session: document progress, current state, and next steps; update PLAN/todo/docs accordingly; capture lessons/process adjustments in design notes so the next contributor can resume seamlessly.
- Mark off and remove completed items from the PLAN file once they are complete.
- Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines.
- Tests should run quickly; no test should take more than one minute to complete.

## Concurrency Expectations
- Go interpreters must implement Able `spawn`/`Future` semantics with goroutines/channels while preserving observable behavior (status, value, cancellation, cooperative helpers).
- The runtime exposes `future_yield()`, `future_cancelled()`, `future_flush()`, and `future_pending_tasks()` so fixtures/tests can drive (and introspect) the scheduler deterministically.
- Document any deviations or extensions; tests should exercise cancellation, yielding, and memoization scenarios.

## When in Doubt
- Ask for clarification if spec vs implementation conflicts arise.
- Capture decisions in `design/` and reference them from PLAN documents.
- Keep interoperability in mind: the tree-sitter parser must emit ASTs that both Go interpreters can evaluate without translation loss.
