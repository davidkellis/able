# Testing Architecture: Historical Consolidation Plan

## Status

The active Go-first testing architecture is in
[testing-plan.md](./testing-plan.md). This record preserves the former
consolidation proposal; it does not leave a testing rollout queue.

## Completed chronology

The earlier plan correctly separated implementation fixtures from user-facing
tests and proposed a stdlib `able.test.*` protocol plus `able.spec` DSL. That
work is complete: canonical source was promoted to `../able-stdlib/src/test`
and `src/spec.able`, canonical suites live in `../able-stdlib/tests`, and the
Go `able test` command loads the harness in interpreted and compiled modes.

The prior phases—move quarantine modules, wire a TypeScript CLI skeleton, wire
the Go CLI, then add spec/docs—are historical. v12 spec §17 now supplies the
test-module and command contract, and the Go CLI/protocol record owns the
current behavior.

## Resolved decisions

Explicit test modules and ordinary source-file targets are both supported by
the implemented CLI. Default matchers are exposed by `able.spec`; alternative
frameworks use the shared `able.test` protocol. Fixture tests remain unrelated
to user framework registration.

## Deferred ideas

Remote execution, additional test-framework APIs, and a future non-Go runner
need an explicit new requirement. They are not implied by this completed plan,
and they must not bypass the current canonical stdlib and Go fixture contracts.
