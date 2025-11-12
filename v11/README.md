# Able Language Workspace

Able is an experimental programming language. This workspace (`/v11`) is a versioned copy of the Able v10 toolchain so we can begin Able v11 development while keeping the original v10 artefacts untouched at the repo root.

## Project Highlights
- **Spec-first**: `spec/full_spec_v11.md` currently mirrors the v10 semantics; it is the document we will extend for the v11 cycle while `spec/full_spec_v10.md` stays as the frozen reference.
- **Go reference interpreter**: `interpreters/go/` is the v11 copy of the canonical Go runtime; it defines the shared AST (`pkg/ast`), runs the static typechecker by default, and must remain aligned with the spec.
- **TypeScript interpreter**: `interpreters/ts/` is the Bun-based runtime that mirrors the Go implementation; its AST (`src/ast.ts`) must stay structurally aligned with the Go AST so both follow the same spec-defined contract.
- **Canonical AST & semantics**: every Able interpreter (v10 and v11) consumes the same AST shapes and must produce identical observable behaviour per the spec; divergence is treated as a spec or implementation bug.
- **Future parser**: With both runtimes mirrored here, we can bring up the tree-sitter grammar under `parser/` for the v11 effort without disturbing the v10 sources.
- **Quick language tour**: read the [Able manual](docs/manual/index.html); it currently documents v10 semantics and will be updated alongside the v11 spec.

## Repository Layout
- `spec/` — Language specs (v1–v11) plus TODO trackers. `full_spec_v11.md` is the active v11 document.
- `interpreters/ts/` — Bun/TypeScript interpreter, tests, CLI tooling, and AST builders.
- `interpreters/go/` — Go interpreter, CLI, and canonical AST/typechecker definitions.
- `parser/` — Tree-sitter oriented parser experiments copied from the v10 workspace.
- `fixtures/`, `examples/`, `stdlib/` — Shared AST fixtures, curated example programs, and stdlib sketches.
- `design/`, `docs/`, `README.md`, `PLAN.md`, `AGENTS.md` — Copied documentation and roadmap files that we will update specifically for v11 work.

## How We Work
1. Start with the spec. Treat it as source of truth—if behaviour changes, update the spec (including the canonical AST contract) before or alongside code.
2. Keep the AST structure identical across interpreters, with the Go definitions serving as canonical while remaining faithful to the spec. Document any mismatches and schedule fixes immediately.
3. Mirror features and tests between interpreters so behavior stays consistent.
4. Use `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance.

## Getting Started
- **Go interpreter (canonical)**: install Go ≥ 1.22, run `go test ./...` inside `interpreters/go/`, and prefer `./run_all_tests.sh --typecheck-fixtures=strict` before sending code for review.
- **TypeScript interpreter**: inside `interpreters/ts/`, run `bun install`, `bun test`, and `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts` so the checker stays in lockstep with the Go runtime before sharing changes.
- **Specs**: edit and review `spec/full_spec_v11.md` (the v11 fork) while consulting `spec/full_spec_v10.md` for the frozen reference.

Combined test suites:

```bash
# Run TypeScript + Go tests and shared fixtures
./run_all_tests.sh

# Include Go fixture typechecking (warn logs diagnostics, strict enforces them)
./run_all_tests.sh --typecheck-fixtures=warn
./run_all_tests.sh --typecheck-fixtures=strict
```

See `docs/parity-reporting.md` for details on directing the parity JSON report into CI artifacts (`ABLE_PARITY_REPORT_DEST`, `CI_ARTIFACTS_DIR`) and consuming the machine-readable diffs. The instructions still reference the v10 paths—update them as you touch those docs.


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
Read AGENTS, PLAN, and the v10 spec, and then start on the higest priority PLAN work. proceed with next steps. we need to correct any bugs if bugs or broken tests are outstanding. we want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. I have given you permissions to run tests.
```

Standard next steps prompt:
```
Proceed with next steps as suggested; don't talk about doing it - do it. we need to correct any bugs if bugs or broken tests are outstanding.  we want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. tests should run quickly; no test should take more than one minute to complete.
```
