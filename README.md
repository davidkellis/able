# Able Language Workspace

Able is an experimental programming language. The repository now carries two versioned workspaces:

- `v10/` — the frozen Able v10 toolchain (spec, interpreters, fixtures, docs). Use this tree for maintenance releases only.
- `v11/` — the active Able v11 fork where new language/spec work lands.

Shared coordination docs (this `README.md`, `PLAN.md`, `AGENTS.md`) live at the repo root so contributors can quickly see the roadmap and version boundaries.

## Project Highlights
- **Spec-first**: `spec/full_spec_v10.md` remains the authoritative v10 document; `spec/full_spec_v11.md` is the evolving v11 draft. Keep the spec in sync with behaviour before/alongside code changes.
- **Versioned interpreters**: each workspace mirrors the same structure (`interpreters/{ts,go}`, `parser/`, `fixtures/`, `stdlib/`, `design/`, `docs/`). The Go runtime stays canonical; the TypeScript runtime mirrors its behaviour (different concurrency strategy, same semantics).
- **Canonical AST & semantics**: every interpreter consumes the same AST contract captured in the spec. Divergences are bugs.
- **Cross-version clarity**: freezing v10 inside `v10/` keeps the existing toolchain intact while `v11/` evolves toward the next spec. Future versions can follow the same pattern.

## Repository Layout
- `spec/` — Language specs (v1–v11) and supplemental notes.
- `v10/` — Frozen Able v10 workspace (`design/`, `docs/`, `fixtures/`, `interpreters/{ts,go}/`, `parser/`, `stdlib/`, `run_all_tests.sh`, etc.).
- `v11/` — Active Able v11 workspace with the same structure as `v10/`.
- Root docs (`README.md`, `PLAN.md`, `AGENTS.md`) describe the multi-version roadmap and onboarding steps.

## How We Work
1. Start with the relevant spec (`spec/full_spec_v10.md` or `spec/full_spec_v11.md`). Update wording (and the AST contract) before or alongside code.
2. Keep the AST structure identical across interpreters. The Go runtime is canonical; the TypeScript runtime mirrors it. Document and resolve mismatches immediately.
3. Mirror tests/fixtures across interpreters so behaviour stays consistent across versions.
4. Use the root `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance. Version-specific notes live under `v10/` and `v11/`.

## Getting Started
- **Go interpreter (v10)**: `cd v10/interpreter10-go && go test ./...`
- **Go interpreter (v11)**: `cd v11/interpreters/go && go test ./...`
- **TypeScript interpreter (v10)**: `cd v10/interpreter10 && bun install && bun test`
- **TypeScript interpreter (v11)**: `cd v11/interpreters/ts && bun install && bun test`
- **Specs**: edit `spec/full_spec_v11.md` for new work; keep `spec/full_spec_v10.md` authoritative for the frozen branch.

Combined test suites:

```bash
# Run TypeScript + Go tests and shared fixtures (default = v10)
./run_all_tests.sh

# Target a specific version
./run_all_tests.sh --version=v10 --typecheck-fixtures=strict
./run_all_tests.sh --version=v11 --typecheck-fixtures=warn
```

See `v11/docs/parity-reporting.md` (or the archived `v10/docs/` copy) for details on directing the parity JSON report into CI artifacts (`ABLE_PARITY_REPORT_DEST`, `CI_ARTIFACTS_DIR`) and consuming the machine-readable diffs.


## Contributing
- Follow the roadmap in `PLAN.md`; update it when work progresses.
- Record architectural decisions in `design/` with clear timestamps and motivation.
- Prefer incremental pull requests with mirrored tests across language implementations.
- Coordinate changes across interpreters before modifying the AST or spec.

## Questions?
Start with `AGENTS.md`, then dive into `v11/README.md`, `v11/AGENTS.md`, or the archived `v10/` docs as needed. Capture design decisions in the relevant `design/` tree so future contributors have context.

## Notes

Standard onboarding prompt:
```
Read AGENTS, PLAN, and the v10 spec, and then start on the higest priority PLAN work. proceed with next steps. we need to correct any bugs if bugs or broken tests are outstanding. we want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. I have given you permissions to run tests.
```

Standard next steps prompt:
```
Proceed with next steps as suggested; don't talk about doing it - do it. we need to correct any bugs if bugs or broken tests are outstanding.  we want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. tests should run quickly; no test should take more than one minute to complete.
```
