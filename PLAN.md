# Able Project Roadmap (v11 focus)

## Standard onboarding prompt
Read AGENTS, PLAN, and the v11 spec, then start on the highest priority PLAN work. Proceed with next steps. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. You have permission to run tests. Tests should run quickly; no test should take more than one minute to complete.

## Standard next steps prompt
Proceed with next steps as suggested; don't talk about doing it - do it. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. Tests should run quickly; no test should take more than one minute to complete.

## Scope
- Keep the frozen Able v10 toolchain available for historical reference while driving all new language, spec, and runtime work in v11.
- Keep the Go interpreter as the behavioural reference and ensure the TypeScript runtime + future ports match feature-for-feature (the concurrency implementation strategy may differ).
- Preserve a single AST contract for every runtime so tree-sitter output can target both the historical v10 branch and the actively developed v11 runtime; document any deltas immediately in the v11 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.

## Existing Assets
- `spec/full_spec_v10.md`: authoritative semantics for the archived toolchain. Keep it untouched unless a maintainer requests an errata fix.
- `spec/full_spec_v11.md`: active specification for all current work; every behavioural change must be described here.
- `v10/interpreters/{ts,go}/`: Frozen interpreters that match the v10 spec. Treat them as read-only unless a blocking support request lands.
- `v11/interpreters/{ts,go}/`, `v11/parser`, `v11/fixtures`, `v11/stdlib`: active development surface for Able v11.
- Legacy work: `interpreter6/`, assorted design notes in `design/`, early stdlib sketches. Do not do any work in these directories.

## Ongoing Workstreams
- **Spec maintenance**: stage and land all wording in `spec/full_spec_v11.md`; log discrepancies in `spec/TODO_v11.md`. Reference the v10 spec only when clarifying historical behaviour.
- **Standard library**: coordinate with `stdlib/` efforts; ensure interpreters expose required builtin functions/types; track string/regex bring-up via `design/regex-plan.md` and the new spec TODOs covering byte-based strings with char/grapheme iterators.
- **Developer experience**: cohesive documentation, examples, CI improvements (Bun + Go test jobs).
- **Future interpreters**: keep AST schema + conformance harness generic to support planned Crystal implementation.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones now live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.
- Status: interface alignment is in place (Apply/Index/Awaitable + stdlib helpers); parity harness is green with the latest sweep saved to `tmp/parity-report.json`; kernel/stdlib loader auto-detection is bundled; Ratio/numeric conversions are landed; exec fixtures are spec-prefixed with coverage enforced via `scripts/check-exec-coverage.mjs` in `run_all_tests.sh`; regex work is stdlib-only (host hooks removed) with literal + quantifier parsing, match/find_all, and streaming scan now available; full v11 test + stdlib suites green as of 2026-01-22.

## Guardrails (must stay true)
- `./run_all_tests.sh --version=v11` must stay green (TS + Go unit tests, fixture suites, parity CLI); it may take up to 10 minutes to run
- `./run_stdlib_tests.sh` must stay green; it may take up to 10 minutes to run
- `v11/interpreters/ts/scripts/run-parity.ts` remains the authoritative entry point for fixtures/examples parity. Run the v10 suite only when explicitly asked to investigate archival regressions.
- `v11/interpreters/ts/testdata/examples/` remains curated; add programs only when parser/runtime support is complete (see the corresponding `design/parity-examples-plan.md` for the active roster).
- Diagnostics parity is mandatory: both interpreters emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`. (✅ CLI now enforces diagnostics comparison; ✅ Go typechecker infers implicit `self`/iterator helpers so warn/strict runs match.)
- The JSON parity report (`tmp/parity-report.json`) must be archived in CI via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR` to keep regression triage fast.
- After regenerating tree-sitter assets under `v11/parser/tree-sitter-able`, force Go to relink the parser by deleting `v11/interpreters/go/.gocache` or running `cd v11/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`.
- Track upcoming language work (awaitable orchestration, advanced dynimport scenarios, interface dispatch additions) so fixtures/examples land immediately once syntax/runtime support exists.
- It is expected that some new fixtures will fail due to interpreter bugs/deficiencies.We should implement fixtures strictly in accordance with the v11 spec semantics. Do not weaken or sidestep the behavior under test to "make tests pass". If a fixture fails under a given interpreter, follow up by fixing the interpreter so the implementation honors the spec. The interpreters should perfectly implement all the semantics described in the v11 spec. You need to figure out whether each of the fixture failures is due to a bug in the typechecker, or an issue with the expectation encoded in the test fixture, because you have encoded a bunch of expectations in the fixtures to assume that the typechecker will emit an error that in the past you had just ignored as a means to work around a limitation in the typechecker. We are in a state now where the fixtures didn't trust the typechecker to produce a correct results, and so you  worked around those previous typechecker limitations by just "baselining" them with an understanding that errors would be emitted but you would ignore them and just  run things without the typechecker. We made a lot of progress like that, but now it's time to correct that approach. We should never ignore errors or warnings and just work around them like we did in the past. Now it is time to correct those mistakes. When we observe failures, you need to figure out if the failure is due to a bug in the typechecker or a bug in the interpreter or if the expectation encoded in the test fixture (or elsewhere) is an incorrect expectation that should no longer be expected or assumed, because the most important thing is that we have correct language semantics and we judge correctness by conformance to the v11 spec. All behavior must conform to the v11 spec. Any test expectation that doesn't conform to the v11 spec's well-defined semantics should be changed to conform to the v11 spec.

## TODO (working queue: tackle in order, move completed items to LOG.md)
- Compiler/interpreter vision: typed core IR + runtime ABI implementation track
  - Define the concrete IR node set and type system (values, blocks, control flow, call/impl dispatch, async/await surfaces).
  - Add an IR serialization format (JSON + stable schema) and a small loader for tests/tooling.
  - Implement a lowering pass from AST → IR with type annotations (start with literals, bindings, blocks, calls).
  - Add an IR interpreter or verification pass to validate execution semantics against the tree-walker.
  - Add conformance fixtures that execute both AST and IR paths (shared harness).
  - Update `v11/design/compiler-interpreter-vision.md` as the IR contract stabilizes.
- Interpreter performance track: bytecode VM expansion
  - Expand bytecode instruction set (control flow, functions/closures, structs, arrays, member access, interface dispatch).
  - Lower more AST nodes to bytecode; keep tree-walker fallback for unsupported nodes.
  - Add VM/AST parity tests for each newly supported feature (unit + fixtures).
  - Gate VM behind a CLI/runtime flag and add perf harness to compare VM vs tree-walker.
  - Document the bytecode format + calling convention in `v11/design/compiler-interpreter-vision.md`.
- Regex parser + quantifiers (syntax-level)
  - Add regex AST nodes and grammar in tree-sitter (quantifiers, groups, classes, alternation).
  - Wire AST mapping for regex nodes in TS + Go parsers.
  - Add fixtures/tests for regex AST output and exec behavior; keep stdlib engine parity.
  - Update `spec/TODO_v11.md` with remaining regex syntax/semantics gaps.
  - Align stdlib regex implementation with parser outputs as grammar coverage expands.
