# Able Project Roadmap (v12 focus)

## Standard onboarding prompt
Read AGENTS, PLAN, and the v12 spec, then start on the highest priority PLAN work. Proceed with next steps. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. You have permission to run tests. Tests should run quickly; no test should take more than one minute to complete.

## Standard next steps prompt
Proceed with next steps as suggested; don't talk about doing it - do it. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. Tests should run quickly; no test should take more than one minute to complete.

## Scope
- Keep the frozen Able v10/v11 toolchains available for historical reference while driving all new language, spec, and runtime work in v12.
- Use the Go tree-walking interpreter as the behavioral reference and keep the Go bytecode interpreter in strict semantic parity.
- Preserve a single AST contract for every runtime so tree-sitter output can target both the historical branches and the actively developed v12 runtime; document any deltas immediately in the v12 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.

## Existing Assets
- `spec/full_spec_v10.md` + `spec/full_spec_v11.md`: authoritative semantics for archived toolchains. Keep them untouched unless a maintainer requests errata.
- `spec/full_spec_v12.md`: active specification for all current work; every behavioral change must be described here.
- `v10/` + `v11/`: frozen workspaces (read-only unless hotfix required).
- `v12/`: active development surface for Able v12 (`interpreters/go/`, `parser/`, `fixtures/`, `stdlib/`, `design/`, `docs/`).

## Ongoing Workstreams
- **Spec maintenance**: stage and land all wording in `spec/full_spec_v12.md`; log discrepancies in `spec/TODO_v12.md`.
- **Go runtimes**: maintain tree-walker + bytecode interpreter parity; keep diagnostics and fixtures aligned.
- **Tooling**: build a Go-based fixture exporter; update harnesses to remove TS dependencies.
- **Performance**: expand bytecode VM coverage; add perf harnesses for tree-walker vs bytecode.
- **WASM**: run the Go runtime in WASM with JS tree-sitter parsing and a defined host ABI.
- **Stdlib**: keep v12 stdlib aligned with runtime capabilities and spec requirements.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `v12/design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.

## Guardrails (must stay true)
- `./run_all_tests.sh` (v12 default) must stay green for the Go suites and fixtures.
- `./run_stdlib_tests.sh` must stay green (tree-walker + bytecode).
- Diagnostics parity is mandatory: tree-walker and bytecode interpreters must emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`.
- After regenerating tree-sitter assets under `v12/parser/tree-sitter-able`, force Go to relink the parser by deleting `v12/interpreters/go/.gocache` or running `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`.
- It is expected that some new fixtures will fail due to interpreter bugs/deficiencies. Implement fixtures strictly in accordance with the v12 spec semantics. Do not weaken or sidestep the behavior under test to "make tests pass".

## TODO (working queue: tackle in order, move completed items to LOG.md)
- WASM target
  - Prototype JS tree-sitter parsing that feeds AST into the Go/WASM runtime.
  - Build a minimal `ablewasm` runner (Node + browser harness) once the Go runtime builds to WASM.
  - Document the WASM deployment contract in `v12/docs/`.
- Regex parser + quantifiers (syntax-level)
  - Add regex AST nodes and grammar in tree-sitter (quantifiers, groups, classes, alternation).
  - Wire AST mapping for regex nodes in Go parser.
  - Add fixtures/tests for regex AST output and exec behavior; keep stdlib engine parity.
  - Add parser corpus cases that cover nested groups, alternation, and escaped quantifiers.
  - Update `spec/TODO_v12.md` with remaining regex syntax/semantics gaps.
  - Align stdlib regex implementation with parser outputs as grammar coverage expands.
- Go compiler (initial)
  - Continue expanding codegen beyond literal returns (comparisons, control flow, collections).
  - Tighten interop conversion coverage for structs and primitive coercions.
