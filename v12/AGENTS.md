# Able Project — Agent Onboarding

This document gives contributors the context required to work across the Able v12 effort (Go tree-walker + bytecode interpreters, spec, and tooling) while keeping legacy artefacts untouched at the repo root.

Unconditionally read PLAN.md and spec/full_spec_v12.md (using archived specs as historical reference) before starting any work.

## Mission & Principles
- Keep the Able v12 language spec (`spec/full_spec_v12.md`) authoritative while referencing archived specs when questions arise. The spec is the ultimate arbiter for every runtime feature.
- Treat the AST as part of the language contract with the Go definitions as canonical. Every interpreter must share the same structure and field semantics, and the v12 spec will codify this canonical AST form and its evaluation rules.
- Use the Go interpreters as the reference implementations while ensuring any future runtimes match the same semantics defined in the spec.
- Prefer incremental, well-tested changes. Mirror behavior across interpreters whenever possible.
- Document reasoning (design notes in `design/`, issue trackers) so future agents can follow decisions.
- Keep shared fixtures (`fixtures/ast`) green across the Go interpreters; every fixture change must be exercised by the Go fixture harnesses (`go test ./pkg/interpreter`).
- Align code changes with the current design notes (e.g., `design/pattern-break-alignment.md`) and update `spec/todo.md`/`PLAN.md` when work lands.
- Modularize larger features into smaller, self-contained modules. Keep each file under one thousdand (i.e. 1000) lines of code.
- Defer AST mapping work until the parser produces the expected parse trees (as captured in the grammar corpus) for every feature under development; once grammar coverage is complete and stable, implement the mapping logic.

## Repository Map
- `interpreters/go/`: Go interpreters and canonical Able runtime. Go-specific design docs live under `design/` (see `go-concurrency.md`, `typechecker.md`).
- `parser/`: Tree-sitter grammar work copied from the archived workspace so we can continue parser experiments in the v12 branch.
- `spec/`: Language specification (v1–v12); focus on `full_spec_v12.md` plus topic-specific supplements.
- `examples/`, `fixtures/`, `stdlib*/`: Sample programs, shared AST fixtures, and stdlib sketches used for conformance testing.
- `design/`: High-level architecture notes, historical context, and future proposals.

## Getting Started
1. Read `spec/full_spec_v12.md` (and skim archived specs for historical context) plus `interpreters/go/README.md` to internalize semantics and existing architecture.
2. Review `PLAN.md` (project-level) and the interpreter-specific notes in `interpreters/go/README.md` for current priorities.
3. Set up tooling:
   - **Go**: Go ≥ 1.22, `go test ./...` inside `interpreters/go/`.
4. Before changing the AST, confirm alignment implications for every interpreter and the future parser.
5. Use `./run_all_tests.sh` (from the `v12/` root) to run the Go suite before handing work off.

## Collaboration Guidelines
- Update relevant PLAN files when you start/finish roadmap items. The current typechecker roadmap lives in `design/typechecker-plan.md`.
- Keep `spec/TODO_v12.md` current when implementation work exposes gaps that need spec wording updates.
- Treat the shared AST contract as canonical: when introducing new node structures or runtime semantics, implement them in both interpreters and update fixtures so every runtime interprets them identically.
- When adding or modifying fixtures in `fixtures/ast`, run the Go fixture harness (`(cd interpreters/go && go test ./pkg/interpreter)`).
- Fixture manifests can include an optional `setup` array when multi-module scenarios are required (e.g., dyn-import packages); the Go harness evaluates those modules before the entry `module.json`.
- Use concise, high-signal comments in code. Avoid speculative abstractions.
- Update the spec (`spec/full_spec_v12.md`) once behaviour becomes canonical; check off items in `spec/TODO_v12.md`.
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
