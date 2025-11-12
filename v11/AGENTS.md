# Able Project — Agent Onboarding

This document gives contributors the context required to work across the Able v11 effort (TypeScript + Go interpreters, spec, and tooling) while keeping the original v10 artefacts untouched at the repo root.

Unconditionally read PLAN.md and spec/full_spec_v11.md (using spec/full_spec_v10.md as the frozen reference) before starting any work.

## Mission & Principles
- Keep the Able v11 language spec (`spec/full_spec_v11.md`) authoritative while referencing the frozen v10 doc when questions arise. The spec is the ultimate arbiter for every runtime feature.
- Treat the AST as part of the language contract with the Go definitions as canonical. Every interpreter must share the same structure and field semantics, and the v11 spec will codify this canonical AST form and its evaluation rules.
- Use the Go interpreter as the reference implementation while ensuring the TypeScript interpreter—and any future runtimes—match the same semantics defined in the spec.
- Prefer incremental, well-tested changes. Mirror behavior across interpreters whenever possible.
- Document reasoning (design notes in `design/`, issue trackers) so future agents can follow decisions.
- Keep shared fixtures (`fixtures/ast`) green in both interpreters; every fixture change must be exercised by `bun run scripts/run-fixtures.ts` and the Go parity tests (`go test ./pkg/interpreter`).
- Align code changes with the current design notes (e.g., `design/pattern-break-alignment.md`) and update `spec/todo.md`/`PLAN.md` when work lands.
- Modularize larger features into smaller, self-contained modules. Keep each file under one thousdand (i.e. 1000) lines of code.
- Defer AST mapping work until the parser produces the expected parse trees (as captured in the grammar corpus) for every feature under development; once grammar coverage is complete and stable, implement the mapping logic.

## Repository Map
- `interpreters/ts/`: Bun/TypeScript interpreter, AST definition, and comprehensive tests. This is the v11 copy of the v10 runtime.
- `interpreters/go/`: Go interpreter and canonical Able runtime. Go-specific design docs live under `design/` (see `go-concurrency.md`, `typechecker.md`).
- `parser/`: Tree-sitter grammar work copied from v10 so we can continue parser experiments in the v11 branch.
- `spec/`: Language specification (v1–v11); focus on `full_spec_v11.md` plus topic-specific supplements.
- `examples/`, `fixtures/`, `stdlib*/`: Sample programs, shared AST fixtures, and stdlib sketches used for conformance testing.
- `design/`: High-level architecture notes, historical context, and future proposals.

## Getting Started
1. Read `spec/full_spec_v11.md` (and skim `spec/full_spec_v10.md` for historical context) plus `interpreters/ts/README.md` to internalize semantics and existing architecture.
2. Review `PLAN.md` (project-level) and the interpreter-specific notes in `interpreters/ts/README.md` / `interpreters/go/README.md` for current priorities.
3. Set up tooling:
   - **Go**: Go ≥ 1.22, `go test ./...` inside `interpreters/go/`.
   - **TypeScript**: Bun ≥ 1.2 (`bun install`, `bun test`) inside `interpreters/ts/`.
4. Before changing the AST, confirm alignment implications for every interpreter and the future parser.
5. Use `./run_all_tests.sh` (from the `v11/` root) to run TypeScript unit tests, fixtures, and the Go suite together before handing work off.

## Collaboration Guidelines
- Update relevant PLAN files when you start/finish roadmap items. The current typechecker roadmap lives in `design/typechecker-plan.md`.
- Keep `spec/todo.md` current when implementation work exposes gaps that need spec wording updates.
- Treat the shared AST contract as canonical: when introducing new node structures or runtime semantics, implement them in both interpreters and update fixtures so every runtime interprets them identically.
- When adding Go features, port or mirror the corresponding TypeScript tests (or vice versa) to keep coverage consistent.
- When adding or modifying fixtures in `fixtures/ast`, update `interpreters/ts/scripts/export-fixtures.ts`, run the exporter + TS harness, and confirm the Go parity test (`(cd interpreters/go && go test ./pkg/interpreter)`) still passes.
- Fixture manifests can include an optional `setup` array when multi-module scenarios are required (e.g., dyn-import packages); both harnesses evaluate those modules before the entry `module.json`.
- Use concise, high-signal comments in code. Avoid speculative abstractions; match the TS design unless we have a strong reason to diverge.
- Update the spec (`spec/full_spec_v11.md`) once behaviour becomes canonical; check off items in `spec/todo.md`.
- At the end of every session: document progress, current state, and next steps; update PLAN/todo/docs accordingly; capture lessons/process adjustments in design notes so the next contributor can resume seamlessly.
- Mark off and remove completed items from the PLAN file once they are complete.
- Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines.
- Tests should run quickly; no test should take more than one minute to complete.

## Concurrency Expectations
- TypeScript interpreter uses a cooperative scheduler to emulate Able `proc`/`spawn` semantics; the helper functions `proc_yield()`, `proc_cancelled()`, `proc_flush()`, and the diagnostic `proc_pending_tasks()` are available inside Able code so fixtures/tests can drive (and introspect) the scheduler deterministically.
- Go interpreter must implement these semantics with goroutines/channels while preserving observable behavior (status, value, cancellation, cooperative helpers). The Go runtime exposes the same helper surface (including `proc_flush`) via native functions.
- Document any deviations or extensions; tests should exercise cancellation, yielding, and memoization scenarios.

## When in Doubt
- Ask for clarification if spec vs implementation conflicts arise.
- Capture decisions in `design/` and reference them from PLAN documents.
- Keep interoperability in mind: the eventual tree-sitter parser must emit ASTs that both interpreters can evaluate without translation loss.
