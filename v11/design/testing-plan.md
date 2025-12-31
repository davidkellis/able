# Unified Testing and Spec Tooling Plan (Draft)

## Scope

This document reconciles the current testing assets (fixtures, parity harnesses,
stdlib testing modules, CLI skeleton) into a single plan. It defines:
- the testing model with a strict separation between language implementation
  tests and user-facing test frameworks,
- the stdlib testing framework contract and namespace split,
- the `able test` CLI responsibilities,
- and the spec/doc updates required to make the model explicit.

This plan intentionally keeps the Go interpreter canonical and ensures the
TypeScript runtime mirrors behavior.

## Current Assets (Source of Truth)

Design docs:
- `v11/design/testing-cli-protocol.md` (framework protocol)
- `v11/design/testing-cli-design.md` (CLI behavior)
- `v11/docs/testing-matchers.md` (matcher catalog)

Implementation (quarantine; planned rename):
- `v11/stdlib/quarantine/src/testing/*` (protocol, registry, harness, spec-style DSL,
  reporters, matchers) → split into `able.test.*` + `able.spec`
- `v11/stdlib/quarantine/tests/*.test.able` (early tests using the spec-style DSL)

Tooling:
- `v11/interpreters/ts/scripts/run-module.ts` (able test CLI skeleton)
- `v11/interpreters/ts/test/cli/run_module_cli.test.ts` (CLI skeleton tests)
- `v11/fixtures/` + `v11/docs/conformance-plan.md` (fixture and spec coverage)

Spec:
- `spec/full_spec_v11.md` has a TODO placeholder for "Tooling: testing framework".

## Separation (Non-Negotiable)

We draw a hard line between two unrelated concerns:

1. **Language implementation tests and conventions**
   - Purpose: validate parser, AST mapping, interpreter parity, and spec
     coverage.
   - Mechanism: AST fixtures, exec fixtures, and parity harnesses wired into
     `run_all_tests.sh`.
   - Audience: interpreter and spec maintainers, not end users.

2. **User-facing testing framework and conventions**
   - Purpose: let Able users test their own programs.
   - Mechanism: `able test` + `able.spec` DSL (backed by `able.test.*` protocol).
     matchers).
   - Audience: Able library/application authors.

These two areas are intentionally unrelated. Fixtures do not inform the public
test framework, and the public framework does not replace fixtures or parity
tests.

## Unified Model

We keep two distinct layers:

1. **Conformance fixtures (spec coverage)**
   - AST fixtures (parser/mapping) and exec fixtures (end-to-end semantics).
   - Used for cross-interpreter parity and spec coverage.
   - Already wired into `run_all_tests.sh`.

2. **Able test framework (user and stdlib tests)**
   - Tests authored in Able using a stdlib framework (spec-style DSL).
   - The CLI orchestrates discovery, filtering, execution, and reporting.
   - Frameworks are pluggable, but the default standard library framework is
     `able.spec`.

## Test Module Conventions

- A file ending in `.test.able` or `.spec.able` is a test module.
- Test modules live in the same package directories as production code.
- Standard build/run/check omit test modules unless `--with-tests` is enabled.
- `able test` implicitly enables `--with-tests` and typechecks all modules in
  one session to allow private API access in tests.

## Stdlib Testing Framework Contract

Namespace split:
- `able.test.*` holds the protocol and harness used by the CLI.
- `able.spec` is the default user-facing DSL that implements the protocol and
  registers itself on import.

The stdlib exposes a protocol and harness (currently in quarantine):

- `able.test.protocol` defines:
  - `Framework` interface with `discover` and `run`.
  - `Reporter` interface + `TestEvent` stream.
  - `TestDescriptor`, `TestPlan`, `RunOptions`, `Failure`.
- `able.test.registry` registers frameworks on import.
- `able.test.harness` provides:
  - `discover_all`, `run_plan`, `run_all`.
- `able.spec` provides the default DSL and matcher suite.
- `able.test.reporters` provides Doc/Progress reporters.

Minimum framework behavior:
- A framework must register itself during module evaluation.
- `discover` returns descriptors for all tests in that framework.
- `run` streams structured events to a `Reporter`.

## CLI Responsibilities (`able test`)

The CLI is responsible for orchestration, not semantics:

1. Locate test modules in scope.
2. Load modules (triggering framework registration).
3. Build `DiscoveryRequest` from CLI filters.
4. Call `able.test.harness.discover_all`.
5. Filter/plan; list or execute based on flags.
6. Create a `Reporter` from `--format` (doc, progress, tap, json).
7. Call `run_plan` with `RunOptions`.
8. Map events to exit codes: 0 pass, 1 failures, 2 discovery/runtime errors.

The TypeScript CLI skeleton already parses flags and prints a plan summary. The
next step is wiring it to the stdlib harness, plus adding the same CLI flow in
the Go interpreter.

## Determinism and Concurrency

Tests may use `proc_flush()` and `proc_pending_tasks()` for deterministic
coordination. The CLI should not attempt to manage scheduling directly; it only
controls ordering and repeat/shuffle. Runtime determinism remains the
interpreter's job (cooperative scheduler in TS, goroutine executor in Go with a
serial mode for tests).

## Spec and Documentation Updates

We should add a minimal "Tooling: Testing framework" section to the v11 spec
that codifies:
- test module suffixes,
- `--with-tests` behavior,
- `able test` enabling test modules and sharing package scope,
- and the requirement that runtimes expose the `able.test` protocol types and
  helpers, with `able.spec` as the default user-facing framework.

The detailed CLI and DSL behavior remains in docs/design notes, not the spec.

## Proposed Implementation Order

Phase 1: Promote stdlib testing modules out of quarantine.
- Move `stdlib/quarantine/src/testing` to `stdlib/src/test` and split:
  - `able.test.*` (protocol, registry, harness, reporters)
  - `able.spec` (DSL + matchers)
- Update stdlib manifests as needed.
- Port or mirror existing quarantine tests into `stdlib/tests`.

Phase 2: Wire `able test` in TypeScript.
- Load test modules, call `discover_all`, run via harness.
- Implement reporter integration and exit codes.
- Add CLI tests that run a tiny test suite.

Phase 3: Wire `able test` in Go.
- Mirror CLI behavior and harness invocation.
- Add Go CLI tests or small integration coverage.

Phase 4: Spec + docs.
- Add spec text under "Tooling: Testing framework".
- Update manual/tooling docs with `able test` usage.

## Open Decisions

1. Should `able test` run only test modules by default, or also allow explicit
   inclusion of non-test modules as targets?
2. Should matchers live only under `able.spec`, or also be exposed via
   `able.test.matchers` for reuse by other frameworks?

## Next Steps

- Decide on the remaining open items above.
- Once decisions land, update the CLI skeleton and promote stdlib modules.
