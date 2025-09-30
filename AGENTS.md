# Able Project — Agent Onboarding

Welcome! This document gives contributors the context required to work across the Able v10 effort (TypeScript + Go interpreters, spec, and tooling).

## Mission & Principles
- Keep the Able v10 language spec (`spec/full_spec_v10.md`) authoritative; report divergences immediately. The spec is the ultimate arbiter for every runtime feature.
- Treat the AST as part of the language contract with the Go definitions as canonical. Every interpreter must share the same structure and field semantics, and the v10 spec will eventually codify this canonical AST form and its evaluation rules.
- Use the Go interpreter as the reference implementation while ensuring the TypeScript interpreter—and any future runtimes—match the same v10 semantics defined in the spec.
- Prefer incremental, well-tested changes. Mirror behavior across interpreters whenever possible.
- Document reasoning (design notes in `design/`, issue trackers) so future agents can follow decisions.
- Keep shared fixtures (`fixtures/ast`) green in both interpreters; every fixture change must be exercised by `bun run scripts/run-fixtures.ts` and the Go parity tests (`go test ./pkg/interpreter`).
- Align code changes with the current design notes (e.g., `design/pattern-break-alignment.md`) and update `spec/todo.md`/`LOG.md` when work lands.

## Repository Map
- `interpreter10/`: Bun/TypeScript interpreter, AST definition, and comprehensive tests. Source of inspiration and a compatibility target.
- `interpreter10-go/`: Go interpreter under active development; this implementation is the canonical reference for Able v10 semantics.
- `spec/`: Language specification (v1–v10); focus on `full_spec_v10.md` plus topic-specific supplements.
- `examples/`, `stdlib*/`: Sample programs and stdlib sketches used for conformance testing.
- `design/`: High-level architecture notes, historical context, and future proposals.

## Getting Started
1. Read `spec/full_spec_v10.md` and `interpreter10/README.md` to internalize semantics and existing architecture.
2. Review `PLAN.md` (project-level) and `interpreter10/PLAN.md` (TS-specific) for current priorities.
3. Set up tooling:
   - **Go**: Go ≥ 1.22, `go test ./...` inside `interpreter10-go/`.
   - **TypeScript**: Bun ≥ 1.2 (`bun install`, `bun test`).
4. Before changing the AST, confirm alignment implications for every interpreter and the future parser.

## Collaboration Guidelines
- Update relevant PLAN files when you start/finish roadmap items.
- Keep `spec/todo.md` current when implementation work exposes gaps that need spec wording updates.
- Treat the shared AST contract as canonical: when introducing new node structures or runtime semantics, implement them in both interpreters and update fixtures so every runtime interprets them identically.
- When adding Go features, port or mirror the corresponding TypeScript tests (or vice versa) to keep coverage consistent.
- When adding or modifying fixtures in `fixtures/ast`, update `interpreter10/scripts/export-fixtures.ts`, run the exporter + TS harness, and confirm the Go parity test (`go test ./pkg/interpreter`) still passes.
- Fixture manifests can include an optional `setup` array when multi-module scenarios are required (e.g., dyn-import packages); both harnesses evaluate those modules before the entry `module.json`.
- Use concise, high-signal comments in code. Avoid speculative abstractions; match the TS design unless we have a strong reason to diverge.
- Update the spec (`spec/full_spec_v10.md`) once behaviour becomes canonical; log the change in `LOG.md` and check off items in `spec/todo.md`.
- At the end of every session: document progress, current state, and next steps; update PLAN/todo/docs accordingly; capture lessons/process adjustments in `LOG.md` or design notes so the next contributor can resume seamlessly.

## Concurrency Expectations
- TypeScript interpreter uses a cooperative scheduler to emulate Able `proc`/`spawn` semantics.
- Go interpreter must implement these semantics with goroutines/channels while preserving observable behavior (status, value, cancellation, cooperative helpers).
- Document any deviations or extensions; tests should exercise cancellation, yielding, and memoization scenarios.

## When in Doubt
- Ask for clarification if spec vs implementation conflicts arise.
- Capture decisions in `design/` and reference them from PLAN documents.
- Keep interoperability in mind: the eventual tree-sitter parser must emit ASTs that both interpreters can evaluate without translation loss.
