# Able Project — Agent Onboarding

Welcome! This document gives contributors the context required to work across the Able v10 effort (TypeScript + Go interpreters, spec, and tooling).

## Mission & Principles
- Keep the Able v10 language spec (`spec/full_spec_v10.md`) authoritative; report divergences immediately.
- Treat the AST as part of the language contract with the Go definitions as canonical. Every interpreter must share the same structure and field semantics.
- Prefer incremental, well-tested changes. Mirror behavior across interpreters whenever possible.
- Document reasoning (design notes in `design/`, issue trackers) so future agents can follow decisions.

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
- When adding Go features, port or mirror the corresponding TypeScript tests (or vice versa) to keep coverage consistent.
- Use concise, high-signal comments in code. Avoid speculative abstractions; match the TS design unless we have a strong reason to diverge.

## Concurrency Expectations
- TypeScript interpreter uses a cooperative scheduler to emulate Able `proc`/`spawn` semantics.
- Go interpreter must implement these semantics with goroutines/channels while preserving observable behavior (status, value, cancellation, cooperative helpers).
- Document any deviations or extensions; tests should exercise cancellation, yielding, and memoization scenarios.

## When in Doubt
- Ask for clarification if spec vs implementation conflicts arise.
- Capture decisions in `design/` and reference them from PLAN documents.
- Keep interoperability in mind: the eventual tree-sitter parser must emit ASTs that both interpreters can evaluate without translation loss.
