# Able Language Workspace

Able is an experimental programming language. The repository now carries multiple versioned workspaces, but only the v12 tree accepts day-to-day changes:

- `v10/` — frozen Able v10 toolchain (spec, interpreters, fixtures, docs). Read-only unless a maintainer asks for a critical hotfix.
- `v11/` — frozen Able v11 workspace kept for historical reference.
- `v12/` — active Able v12 workspace where new language, spec, and runtime work lands.

Shared coordination docs (`README.md`, `PLAN.md`, `AGENTS.md`) live at the repo root so contributors can quickly see the roadmap and the “v12-only” expectation.

## Project Highlights
- **Spec-first**: `spec/full_spec_v12.md` is the active specification and must reflect behaviour before/alongside code. Older specs are reference-only.
- **Go-first runtimes**: the v12 workspace ships Go interpreters (tree-walker + bytecode VM). These must stay in semantic lockstep.
- **Canonical AST & semantics**: every runtime consumes the same AST contract captured in the spec. Divergences are bugs.
- **Cross-version clarity**: freezing v10/v11 preserves historical toolchains while v12 evolves the language.

## Repository Layout
- `spec/` — language specs (v1–v12) and supplemental notes.
- `v10/` — frozen Able v10 workspace.
- `v11/` — frozen Able v11 workspace.
- `v12/` — active Able v12 workspace (`interpreters/go/`, `parser/`, `fixtures/`, `stdlib/`, `design/`, `docs/`).
- Root docs (`README.md`, `PLAN.md`, `AGENTS.md`) describe the multi-version roadmap and onboarding steps.

## How We Work
1. Start with `spec/full_spec_v12.md`. Update wording (and the AST contract) before or alongside code.
2. Keep AST structure identical across the v12 interpreters. Divergences are bugs.
3. Mirror tests/fixtures across the v12 interpreters so behaviour stays consistent.
4. Use the root `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance. Version-specific notes live under `v12/`.

## Getting Started
- **Go interpreter tests (v12)**: `cd v12/interpreters/go && go test ./...`
- **Parser regen gotcha (Go)**: after running `tree-sitter generate/build` in `v12/parser/tree-sitter-able`, Go's cgo cache can reuse stale `parser.c` when `GOCACHE` points at `v12/interpreters/go/.gocache`. Fix by deleting `v12/interpreters/go/.gocache` or running `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`.
- **Frozen toolchains (v10/v11)**: run tests only if a maintainer explicitly requests verification of an archival regression.

Combined test suites:

```bash
# Run the v12 Go test suite + fixtures (default = v12)
./run_all_tests.sh

# Target legacy toolchains only when directed
./run_all_tests.sh --version=v10
./run_all_tests.sh --version=v11
```

## Contributing
- Follow the roadmap in `PLAN.md`; update it when work progresses.
- Record architectural decisions in `v12/design/` with clear timestamps and motivation.
- Prefer incremental pull requests with mirrored tests across v12 interpreters.
- Coordinate changes across interpreters before modifying the AST or spec.

## Questions?
Start with `AGENTS.md`, then dive into `v12/README.md`, `v12/AGENTS.md`, or specific `v12/design/` notes. Reference the archived `v10/` and `v11/` docs only when you need historical context.
