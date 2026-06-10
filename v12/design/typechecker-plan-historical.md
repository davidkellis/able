# Able v12 Typechecker Roadmap — Historical Record

> Archived on 2026-07-14. The active Go contract is
> [`typechecker-plan.md`](typechecker-plan.md). This record retains the reason
> the original roadmap existed without presenting retired work as current.

## 2025 bootstrap

The original 2025-10 plan proposed a static checker over the existing shared
AST, with diagnostics and inference kept outside that AST. Its staged design
was declaration collection, implementation collection, body checking, and
constraint solving. That work landed in the Go `pkg/typechecker` package,
including checker-specific environments, type representations, inference maps,
diagnostics, generic/where-clause obligations, and focused tests.

Subsequent milestones added program-wide checking in loader dependency order,
public/private package summaries, imported prelude construction, package
selector/privacy diagnostics, source hints, implementation/method-set export
metadata, and integration through the Go CLI, interpreter, fixtures, and
compiler analysis. Those are current behaviour, not pending stages.

## Retired work

The original roadmap also assigned the following work:

- TypeScript/Bun typechecker parity, fixture baselines, and a TypeScript CLI
  summary display;
- an export-surface rollout to support that second runtime;
- a staged proposal for incremental checking and eventual editor/LSP reuse;
- documentation written while source locations and program checking were still
  incomplete.

v12 is now Go-first and its TypeScript interpreter has been removed from the
active toolchain. The first two items are retired. Incremental/editor work is
not selected: it needs a real long-running tooling consumer, lifecycle and
invalidation contract, and source-backed benefit before it can enter the active
plan. The old `ABLE_TYPECHECK_FIXTURES=warn|strict` discussion remains relevant
only as a Go fixture diagnostic policy, not as a cross-runtime parity mechanism.

## Lessons retained

- Keep inferred data in side tables and preserve the shared AST contract.
- Check whole programs for import, privacy, re-export, and implementation
  visibility semantics; a standalone module check is not a substitute.
- Accumulate diagnostics where safe, carry package/file context, and preserve
  runtime checks for semantics that remain dynamic.
- Add a checker feature only after the specification and a focused source-level
  regression establish what should be diagnosed.
