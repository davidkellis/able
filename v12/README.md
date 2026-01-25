# Able v12 Workspace

Able is an experimental programming language. This workspace hosts the actively developed Able v12 toolchain; frozen legacy artifacts live at the repository root for historical reference only.

## Project Highlights
- **Spec-first**: `spec/full_spec_v12.md` is the authoritative document for all new behaviour; consult archived specs only for historical context.
- **Go interpreters**: `interpreters/go/` houses both the tree-walking interpreter and the bytecode VM runtime. They must stay in strict semantic parity.
- **Canonical AST & semantics**: all runtimes consume the same AST contract. Any divergence between interpreters is treated as a bug or a spec gap that must be resolved immediately.
- **Parser & tooling**: the v12 workspace carries its own tree-sitter grammar (`parser/`) and docs so parser work can proceed without touching archived sources.
- **Manual & docs**: the manuals under `docs/` should describe v12 semantics; update them alongside spec changes so they reflect v12 behaviour.

## Repository Layout
- `spec/` — Language specs (v1–v12) plus TODO trackers. `full_spec_v12.md` is the active v12 document.
- `interpreters/go/` — Go interpreters (tree-walker + bytecode), CLI, and canonical AST/typechecker definitions.
- `parser/` — Tree-sitter oriented parser experiments copied from the archived workspace.
- `fixtures/`, `examples/`, `stdlib/` — Shared AST fixtures, curated example programs, runnable exec fixtures (`fixtures/exec`), and stdlib sketches. See `docs/exec-fixtures.md` for authoring/running exec programs.
- `design/`, `docs/`, `README.md`, `PLAN.md`, `AGENTS.md` — Copied documentation and roadmap files that we will update specifically for v12 work.

## How We Work
1. Start with `spec/full_spec_v12.md`. Update it (and the AST contract) before or alongside code so behaviour never drifts from the written spec.
2. Keep the AST identical across the Go interpreters. Treat the Go definitions as canonical, document any mismatches, and schedule fixes immediately.
3. Mirror features/tests across the tree-walker and bytecode runtimes so behaviour stays consistent.
4. Use `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance specific to the v12 effort.

## Getting Started
- **Go interpreters**: install Go ≥ 1.22 and run `go test ./...` inside `interpreters/go/`. Before handing off work, prefer `./run_all_tests.sh --version=v12` from the repo root (fixtures/typechecker default to strict).
- **CLI wrappers**: use `./v12/abletw` for tree-walker runs and `./v12/ablebc` for bytecode runs.
- **Specs**: edit `spec/full_spec_v12.md` for new behaviour; consult archived specs only to understand the baseline.

Combined test suites:

```bash
# Run the v12 Go test suite + fixtures
./run_all_tests.sh --version=v12

# Override fixture typechecking (warn logs diagnostics, off disables)
./run_all_tests.sh --version=v12 --typecheck-fixtures=warn
./run_all_tests.sh --version=v12 --typecheck-fixtures=off

# Fixture-only sweep (Go fixture runner)
./run_all_tests.sh --version=v12 --fixture
```


## Contributing
- Follow the roadmap in `PLAN.md`; update it when work progresses.
- Record architectural decisions in `design/` with clear timestamps and motivation.
- Prefer incremental pull requests with mirrored tests across the Go interpreters.
- Coordinate changes across interpreters before modifying the AST or spec.

## Questions?
Start with `AGENTS.md`, then dive into `interpreters/go/README.md` and `spec/full_spec_v12.md`. If something is unclear, capture an issue or design note so future contributors have answers.

## Notes

Standard onboarding prompt:
```
Read AGENTS, PLAN, and the v12 spec, and then start on the highest priority PLAN work. proceed with next steps. we need to correct any bugs if bugs or broken tests are outstanding. we want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. I have given you permissions to run tests.
```

Standard next steps prompt:
```
Proceed with next steps as suggested; don't talk about doing it - do it. we need to correct any bugs if bugs or broken tests are outstanding. we want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. tests should run quickly; no test should take more than one minute to complete.
```
