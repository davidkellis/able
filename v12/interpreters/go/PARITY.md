# Go Interpreter Parity Checklist

The TypeScript interpreter (`v11/interpreters/ts/`) is currently our reference implementation for Able v11 semantics. The Go interpreter must eventually support the same behaviour (with goroutines/channels replacing the cooperative scheduler used in TypeScript).

This document tracks the remaining gaps between the two interpreters. For each feature the table lists the primary TypeScript test coverage and the current Go status.

## Outstanding Gaps (Go vs TypeScript)

The table above captures line-item status; the following themes summarise what still blocks full parity:

- **Control flow coverage** — Core while/range/if/elsif/else semantics and for-loop destructuring now mirror the TS suite (`interpreter_control_flow_test.go`, `interpreter_patterns_test.go`), leaving no outstanding control-flow gaps.
- **Modules & imports** — Static/dynamic imports (including nested re-exports) now align with TS.
- **Interfaces & generics** — Dispatch precedence, default methods, constraints, and named impl disambiguation mirror TS coverage.
- **Data access & operators** — Bitshift/range checks and mixed numeric comparisons now match TS semantics.
- **Error reporting** — Raise/rescue/or-else/rethrow diagnostics now mirror TS behaviour.
- **Concurrency** — Executor-backed `proc`/`spawn` landed, and both runtimes now share the executor contract while exercising the cancellation/yield fixtures through their respective harnesses.
- **Tooling parity** — Dyn-import metadata and breakpoint signalling are now covered; future work can focus on enhanced debugging hooks.

## Immediate next actions

1. Document the shared executor contract across Go and TypeScript runtimes (design notes + spec touchpoints).
2. Evaluate whether fairness-specific concurrency fixtures can be enabled once we add matching coordination helpers in Go.
3. Keep this backlog in sync with each milestone—when a theme moves forward, update the relevant checklist items, add fixtures, and note progress in `PLAN.md`.

_NOTE:_ The only intentional difference between the interpreters is the concurrency implementation detail (Go goroutines/channels vs TypeScript cooperative scheduler). All observable Able semantics must remain identical.
