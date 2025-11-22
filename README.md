# Able Language Workspace

Able is an experimental programming language. The repository now carries two versioned workspaces, but only the v11 branch accepts day-to-day changes:

- `v10/` — the frozen Able v10 toolchain (spec, interpreters, fixtures, docs). Treat this tree as read-only unless a maintainer asks for a critical hotfix.
- `v11/` — the active Able v11 workspace where new language, spec, and runtime work lands.

Shared coordination docs (this `README.md`, `PLAN.md`, `AGENTS.md`) live at the repo root so contributors can quickly see the roadmap and the “v11-only” expectation.

## Project Highlights
- **Spec-first**: `spec/full_spec_v11.md` is the active specification and must reflect behaviour before/alongside code. `spec/full_spec_v10.md` stays untouched unless a maintainer requests errata.
- **Versioned interpreters**: each workspace mirrors the same structure (`interpreters/{ts,go}`, `parser/`, `fixtures/`, `stdlib/`, `design/`, `docs/`). The Go runtime stays canonical; the TypeScript runtime mirrors its behaviour (different concurrency strategy, same semantics). Build new features only under `v11/`.
- **Canonical AST & semantics**: every interpreter consumes the same AST contract captured in the spec. Divergences are bugs, even when only working in v11.
- **Cross-version clarity**: freezing v10 inside `v10/` keeps the historical toolchain intact while `v11/` evolves toward the next spec. Future versions can follow the same pattern once v11 stabilises.

## Repository Layout
- `spec/` — Language specs (v1–v11) and supplemental notes.
- `v10/` — Frozen Able v10 workspace (`design/`, `docs/`, `fixtures/`, `interpreters/{ts,go}/`, `parser/`, `stdlib/`, `run_all_tests.sh`, etc.) retained for archival purposes.
- `v11/` — Active Able v11 workspace with the same structure as `v10/`.
- Root docs (`README.md`, `PLAN.md`, `AGENTS.md`) describe the multi-version roadmap and onboarding steps.

## How We Work
1. Start with `spec/full_spec_v11.md`. Update wording (and the AST contract) before or alongside code. Reference `spec/full_spec_v10.md` only to clarify historical behaviour.
2. Keep the AST structure identical across interpreters. The Go runtime is canonical; the TypeScript runtime mirrors it. Document and resolve mismatches immediately.
3. Mirror tests/fixtures across the v11 interpreters so behaviour stays consistent.
4. Use the root `PLAN.md` for roadmap updates and `AGENTS.md` for onboarding guidance. Version-specific notes live under `v11/`; the `v10/` copies are archival.

## Getting Started
- **Go interpreter (v11)**: `cd v11/interpreters/go && go test ./...`
- **TypeScript interpreter (v11)**: `cd v11/interpreters/ts && bun install && bun test`
- **Frozen toolchain (v10)**: run `v10/` tests only if a maintainer explicitly requests verification of an archival regression.
- **Specs**: edit `spec/full_spec_v11.md` for new work; keep `spec/full_spec_v10.md` untouched unless errata are required.

Combined test suites:

```bash
# Run TypeScript + Go tests and shared fixtures (default = v11)
./run_all_tests.sh

# Target the legacy toolchain only when directed
./run_all_tests.sh --version=v10 --typecheck-fixtures=strict

# Run the full v11 suite with custom flags
./run_all_tests.sh --version=v11 --typecheck-fixtures=warn
```

See `v11/docs/parity-reporting.md` (or the archived `v10/docs/` copy) for details on directing the parity JSON report into CI artifacts (`ABLE_PARITY_REPORT_DEST`, `CI_ARTIFACTS_DIR`) and consuming the machine-readable diffs.


## Contributing
- Follow the roadmap in `PLAN.md`; update it when work progresses.
- Record architectural decisions in `design/` with clear timestamps and motivation.
- Prefer incremental pull requests with mirrored tests across language implementations.
- Coordinate changes across interpreters before modifying the AST or spec.

## Questions?
Start with `AGENTS.md`, then dive into `v11/README.md`, `v11/AGENTS.md`, or specific `v11/design/` notes. Reference the archived `v10/` docs only when you need historical context. Capture new design decisions in the relevant `v11/design/` tree so future contributors have context.

## Notes

Standard onboarding prompt (v11-only):
```
Read AGENTS, PLAN, and the v11 spec, then start on the highest priority PLAN work. Proceed with next steps. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. You have permission to run tests. Tests should run quickly; no test should take more than one minute to complete.
```

Standard next steps prompt:
```
Proceed with next steps as suggested; don't talk about doing it - do it. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. Tests should run quickly; no test should take more than one minute to complete.
```
