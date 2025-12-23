# Able v11 Workspace

Able is an experimental programming language. This workspace hosts the actively developed Able v11 toolchain; frozen legacy artifacts live at the repository root for historical reference only.

## Project Highlights
- **Spec-first**: `spec/full_spec_v11.md` is the authoritative document for all new behaviour; consult archived specs only for historical context.
- **Go reference interpreter**: `interpreters/go/` remains the canonical runtime, defining the shared AST (`pkg/ast`) and typechecker semantics that every other implementation must match.
- **TypeScript interpreter**: `interpreters/ts/` mirrors the Go implementation (with Bun as the host runtime) and must stay structurally aligned with the Go AST/semantics.
- **Canonical AST & semantics**: all runtimes consume the same AST contract. Any divergence between interpreters is treated as a bug or a spec gap that must be resolved immediately.
- **Parser & tooling**: the v11 workspace carries its own tree-sitter grammar (`parser/`) and docs so parser work can proceed without touching archived sources.
- **Manual & docs**: the manuals under `docs/` should describe v11 semantics; update them alongside spec changes so they reflect v11 behaviour.

## Repository Layout
- `spec/` — Language specs (v1–v11) plus TODO trackers. `full_spec_v11.md` is the active v11 document.
- `interpreters/ts/` — Bun/TypeScript interpreter, tests, CLI tooling, and AST builders.
- `interpreters/go/` — Go interpreter, CLI, and canonical AST/typechecker definitions.
- `parser/` — Tree-sitter oriented parser experiments copied from the archived workspace.
- `fixtures/`, `examples/`, `stdlib/` — Shared AST fixtures, curated example programs, runnable exec fixtures (`fixtures/exec`), and stdlib sketches. See `docs/exec-fixtures.md` for authoring/running exec programs.
- `design/`, `docs/`, `README.md`, `PLAN.md`, `AGENTS.md` — Copied documentation and roadmap files that we will update specifically for v11 work.

## How We Work
1. Start with `spec/full_spec_v11.md`. Update it (and the AST contract) before or alongside code so behaviour never drifts from the written spec.
2. Keep the AST identical across interpreters. Treat the Go definitions as canonical, document any mismatches, and schedule fixes immediately.
3. Mirror features/tests across Go and TypeScript so behaviour stays consistent.
4. Use `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance specific to the v11 effort.

## Getting Started
- **Go interpreter**: install Go ≥ 1.22 and run `go test ./...` inside `interpreters/go/`. Before handing off work, prefer `./run_all_tests.sh --version=v11 --typecheck-fixtures=strict` from the repo root.
- **TypeScript interpreter**: inside `interpreters/ts/`, run `bun install`, `bun test`, and `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts` to keep the checker aligned with Go.
- **Specs**: edit `spec/full_spec_v11.md` for new behaviour; consult archived specs only to understand the baseline.

Combined test suites:

```bash
# Run TypeScript + Go tests and shared fixtures for v11
./run_all_tests.sh --version=v11

# Include Go fixture typechecking (warn logs diagnostics, strict enforces them)
./run_all_tests.sh --version=v11 --typecheck-fixtures=warn
./run_all_tests.sh --version=v11 --typecheck-fixtures=strict

# Fixture-only sweep (TS fixtures + parity + Go fixture runner)
./run_all_tests.sh --version=v11 --fixture
```

See `docs/parity-reporting.md` for details on directing the parity JSON report into CI artifacts (`ABLE_PARITY_REPORT_DEST`, `CI_ARTIFACTS_DIR`) and consuming the machine-readable diffs.


## Contributing
- Follow the roadmap in `PLAN.md`; update it when work progresses.
- Record architectural decisions in `design/` with clear timestamps and motivation.
- Prefer incremental pull requests with mirrored tests across language implementations.
- Coordinate changes across interpreters before modifying the AST or spec.

## Questions?
Start with `AGENTS.md`, then dive into `interpreters/ts/README.md`, `interpreters/go/README.md`, and `spec/full_spec_v11.md`. If something is unclear, capture an issue or design note so future contributors have answers.

## Notes

Standard onboarding prompt:
```
Read AGENTS, PLAN, and the v11 spec, and then start on the highest priority PLAN work. proceed with next steps. we need to correct any bugs if bugs or broken tests are outstanding. we want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. I have given you permissions to run tests.
```

Standard next steps prompt:
```
Proceed with next steps as suggested; don't talk about doing it - do it. we need to correct any bugs if bugs or broken tests are outstanding. we want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. tests should run quickly; no test should take more than one minute to complete.
```
