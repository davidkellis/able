# Testing Architecture: Active Contract

## Status and authorities

Able has two intentionally separate testing systems. Their contracts are live,
not a pending consolidation plan:

1. **Implementation verification** validates the language/toolchain through Go
   package tests, AST fixtures, exec fixtures, interpreter parity, compiler
   audits, and verifier-backed performance gates.
2. **User and stdlib tests** are Able test modules run by `able test` through
   the canonical external `able.test` protocol and `able.spec` framework.

The v12 specification §17 is authoritative for test-module and `able test`
semantics. [testing-cli-protocol.md](./testing-cli-protocol.md) describes the
implemented command and framework boundary. The external `../able-stdlib/src`
tree owns user-test protocol/framework source and `../able-stdlib/tests` owns
canonical stdlib test modules. The in-tree deprecated stdlib is not a source
of truth.

## Implementation verification

Go tests under `v12/interpreters/go` cover the runtime, parser, typechecker,
tree-walker, bytecode VM, compiler, CLI, and bridge behavior. Shared AST and
exec fixtures exercise the common Able semantics:

- AST fixtures live under `v12/fixtures/ast` and are consumed by the Go
  fixture/parity harnesses.
- Exec fixtures live under `v12/fixtures/exec`; each manifest defines
  observable stdout, stderr, exit, and optional typecheck expectations.
- Tree-walker/bytecode parity runs compare those observable outcomes. Compiler
  fixture suites add generated-source and static/dynamic-boundary checks.
- `v12/fixtures/exec/coverage-index.json` is checked by
  `v12/scripts/check-exec-coverage.mjs`; a new exec fixture must update it.

`v12/run_all_tests.sh` is the normal project gate. It checks fixture coverage,
the checked-in performance scoreboard/threshold controls, runs Go packages,
and performs the bytecode fixture pass. Its default fixture typechecking mode
is strict; `ABLE_TYPECHECK_FIXTURES`/script options select the documented
strict, warn, or off behavior for an intentional diagnostic test.

Implementation fixtures are evidence for spec parity and runtime correctness.
They do not register user frameworks, and user test modules do not replace or
configure them.

## User and stdlib testing

`.test.able` and `.spec.able` modules share package scope with production code
but are excluded from ordinary build/run/check unless `--with-tests` is
enabled. `able test` selects that profile, discovers modules, evaluates
framework registration, and delegates discovery/execution to `able.test`.

The standard library supplies:

- `able.test.protocol` for descriptors, plans, options, failures, events,
  reporters, and framework interfaces;
- `able.test.registry` and `able.test.harness` for registration, discovery,
  focus filtering, and grouped execution;
- `able.test.reporters` for human-readable reporters; and
- `able.spec` for the default spec-style DSL and matchers.

The Go CLI owns command parsing, loading, interpreted/compiled dispatch, and
rendering. External canonical source owns framework semantics. Changes that
cross this boundary require both CLI coverage and canonical stdlib tests.

## Maintenance rules

- Add focused Go tests with runtime/compiler/CLI changes; use fixtures for
  language-observable behavior and maintain their manifests/index.
- Keep tree-walker and bytecode behavior in parity; add compiler coverage when
  static lowering is in scope.
- Keep user-test documentation aligned with spec §17 and the active CLI
  protocol. Do not describe `able test`, TAP/JSON, compiled test mode, or the
  external stdlib framework as planned work.
- Performance benchmarks are not a substitute for semantic fixtures. Apply
  the performance candidate gate before selecting source changes from a
  benchmark result.

## Deferred scope

No TypeScript/Bun test-runner work, test-framework redesign, remote execution,
manifest opt-outs, or new test CLI feature is currently selected. Such work
requires a concrete user/tooling requirement and must preserve the established
fixture/user-test separation.

The completed quarantine promotion, old TypeScript CLI skeleton, and former
implementation phases are retained in
[the historical testing-plan record](./testing-plan-historical.md).
