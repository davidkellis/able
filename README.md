# Able Language Workspace

Able is an experimental programming language. This workspace hosts the Able v10 language specification, reference interpreters, and supporting tooling.

## Project Highlights
- **Spec-first**: `spec/full_spec_v10.md` captures the current Able v10 semantics.
- **Go reference interpreter**: `interpreter10-go/` (in progress) will be the canonical implementation of the v10 semantics, using Go-native concurrency primitives while matching the spec exactly.
- **TypeScript interpreter**: `interpreter10/` remains a mature implementation and source of design inspiration; its AST definition (`src/ast.ts`) must stay structurally aligned with the Go AST.
- **Future parser**: Once both interpreters align on the AST, a tree-sitter grammar will emit compatible nodes for all runtimes.

## Repository Layout
- `spec/` — Language specs (v1–v10) and topic supplements.
- `interpreter10/` — Bun/TypeScript interpreter, tests, and AST builders.
- `interpreter6/`, `old/compiler/`, `old/design/` — Historical artifacts and explorations.
- `examples/`, `stdlib*/` — Sample programs and standard library sketches used in conformance work.
- `README.md`, `PLAN.md`, `AGENTS.md` — Workspace docs you are reading now.

## How We Work
1. Start with the spec. Treat it as source of truth and update it when semantics change.
2. Keep the AST structure identical across interpreters, with the Go definitions serving as canonical. Document any mismatches and schedule fixes immediately.
3. Mirror features and tests between interpreters so behavior stays consistent.
4. Use `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance.

## Getting Started
- **Go interpreter**: follow `PLAN.md` to help build the reference implementation. Once scaffolded, run `go test ./...` inside `interpreter10-go/`.
- **TypeScript interpreter**: inside `interpreter10/`, run `bun install` then `bun test`.
- **Specs**: browse `spec/full_spec_v10.md`

## Contributing
- Follow the roadmap in `PLAN.md`; update it when work progresses.
- Record architectural decisions in `design/` with clear timestamps and motivation.
- Prefer incremental pull requests with mirrored tests across language implementations.
- Coordinate changes across interpreters before modifying the AST or spec.

## Questions?
Start with `AGENTS.md`, then dive into `interpreter10/README.md` and `spec/full_spec_v10.md`. If something is unclear, capture an issue or design note so future contributors have answers.
