# Able Language Workspace

Able is an experimental programming language. This workspace hosts the Able v10 language specification, reference interpreters, and supporting tooling.

## Project Highlights
- **Spec-first**: `spec/full_spec_v10.md` captures the current Able v10 semantics and codifies the canonical AST structure plus its evaluation rules.
- **Go reference interpreter**: `interpreter10-go/` is the canonical v10 runtime; it defines the shared AST in Go (`pkg/ast`), runs the static typechecker by default, and must remain in lockstep with the written spec.
- **TypeScript interpreter**: `interpreter10/` remains a mature implementation and source of design inspiration; its AST definition (`src/ast.ts`) must stay structurally aligned with the Go AST so both follow the same spec-defined contract.
- **Canonical AST & semantics**: every Able v10 interpreter is expected to consume the same AST shapes and produce identical observable behaviour per the spec; divergence is treated as a spec or implementation bug.
- **Future parser**: With the Go runtime solid, the next milestone is building the tree-sitter grammar that emits compatible nodes for all runtimes.

## Repository Layout
- `spec/` — Language specs (v1–v10) and topic supplements.
- `spec/todo.md` — Open items that must land in the canonical spec (keep in sync with design notes).
- `interpreter10/` — Bun/TypeScript interpreter, tests, and AST builders.
- `interpreter6/`, `old/compiler/`, `old/design/` — Historical artifacts and explorations.
- `examples/`, `stdlib*/` — Sample programs and standard library sketches used in conformance work.
- `README.md`, `PLAN.md`, `AGENTS.md` — Workspace docs you are reading now.

## How We Work
1. Start with the spec. Treat it as source of truth—if behaviour changes, update the spec (including the canonical AST contract) before or alongside code.
2. Keep the AST structure identical across interpreters, with the Go definitions serving as canonical while remaining faithful to the spec. Document any mismatches and schedule fixes immediately.
3. Mirror features and tests between interpreters so behavior stays consistent.
4. Use `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance.

## Getting Started
- **Go interpreter (canonical)**: install Go ≥ 1.22, run `go test ./...` inside `interpreter10-go/`, and prefer `./run_all_tests.sh --typecheck-fixtures=strict` before sending code for review.
- **TypeScript interpreter**: inside `interpreter10/`, run `bun install` then `bun test`.
- **Specs**: browse `spec/full_spec_v10.md`.

Combined test suites:

```bash
# Run TypeScript + Go tests and shared fixtures
./run_all_tests.sh

# Include Go fixture typechecking (warn logs diagnostics, strict enforces them)
./run_all_tests.sh --typecheck-fixtures=warn
./run_all_tests.sh --typecheck-fixtures=strict
```

## Contributing
- Follow the roadmap in `PLAN.md`; update it when work progresses.
- Record architectural decisions in `design/` with clear timestamps and motivation.
- Prefer incremental pull requests with mirrored tests across language implementations.
- Coordinate changes across interpreters before modifying the AST or spec.

## Questions?
Start with `AGENTS.md`, then dive into `interpreter10/README.md` and `spec/full_spec_v10.md`. If something is unclear, capture an issue or design note so future contributors have answers.
