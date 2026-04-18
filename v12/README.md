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
- `fixtures/`, `examples/`, `stdlib-deprecated-do-not-use/` — Shared AST fixtures, curated example programs, runnable exec fixtures (`fixtures/exec`), and archived in-tree stdlib snapshot (non-canonical). Canonical stdlib source is the external `able-stdlib` repository. See `docs/exec-fixtures.md` for authoring/running exec programs.
- `design/`, `docs/`, `README.md`, `PLAN.md`, `AGENTS.md` — Copied documentation and roadmap files that we will update specifically for v12 work.

## How We Work
1. Start with `spec/full_spec_v12.md`. Update it (and the AST contract) before or alongside code so behaviour never drifts from the written spec.
2. Keep the AST identical across the Go interpreters. Treat the Go definitions as canonical, document any mismatches, and schedule fixes immediately.
3. Mirror features/tests across the tree-walker and bytecode runtimes so behaviour stays consistent.
4. Use `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance specific to the v12 effort.

## Getting Started
- **Go interpreters**: install Go ≥ 1.22 and run `go test ./...` inside `interpreters/go/`. Before handing off work, prefer `./v12/run_all_tests.sh` (fixtures/typechecker default to strict).
- **CLI wrappers**: use `./v12/abletw` for tree-walker runs and `./v12/ablebc` for bytecode runs.
- **Stdlib bootstrap**: run `./v12/able setup` once to install/cache canonical stdlib + kernel roots under `$ABLE_HOME/pkg/src`.
- **Stdlib gate**: `./run_stdlib_tests.sh` now self-bootstraps stdlib + kernel into an isolated `ABLE_HOME` when no sibling `able-stdlib` checkout or cached stdlib is present.
- **Canonical stdlib resolution**:
  - `able setup` pins the default stdlib version into `$ABLE_HOME/pkg/src/able/<version>/src` and records the resolved stdlib/kernel sources in `$ABLE_HOME/setup.lock`.
  - `able override add https://github.com/davidkellis/able-stdlib.git <local-path>` redirects the canonical stdlib git dependency to a local checkout for development.
  - Active tooling now prefers that explicit override/cached install path first; sibling `able-stdlib` discovery is only a fallback for repo-local development helpers.
- **Specs**: edit `spec/full_spec_v12.md` for new behaviour; consult archived specs only to understand the baseline.
- **Perf harness**: use `./v12/bench_suite` for machine-readable benchmark snapshots (`fib`, `binarytrees`, `matrixmultiply`, `quicksort`, `sudoku`, `i_before_e`) across compiled/treewalker/bytecode modes.

Combined test suites:

```bash
# Run the v12 Go test suite + fixtures
./v12/run_all_tests.sh

# Override fixture typechecking (warn logs diagnostics, off disables)
./v12/run_all_tests.sh --typecheck-fixtures=warn
./v12/run_all_tests.sh --typecheck-fixtures=off

# Run focused subsets
./v12/run_all_tests.sh --treewalker
./v12/run_all_tests.sh --bytecode
./v12/run_all_tests.sh --compiler

# Details, overrides, and CI workflow inputs:
# v12/docs/compiler-full-matrix.md

# Benchmark harness details + JSON schema:
# v12/docs/performance-benchmarks.md
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
