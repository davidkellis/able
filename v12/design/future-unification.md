# Unified Future Model (Proc/Spawn Unification)

Date: 2026-01-23
Status: Implemented in the Go tree-walker, bytecode VM, and AOT compiler; reconciled 2026-07-14
Owners: Able v12 maintainers

> **Active contract:** Section 12 of `spec/full_spec_v12.md` is normative.
> This document records the completed `proc`-to-`Future` unification decision;
> it is not a pending scheduler, performance, or language-feature backlog.

## Summary

Able v12 now defines a single async facility: `spawn` returns a `Future T` handle that also supports implicit evaluation to `T` when a value is required. This unifies the previously separate `Proc T` and `Future T` concepts into one coherent model with two views:

- **Handle view:** explicit `status()`, `value()`, `cancel()` methods.
- **Value view:** implicit blocking evaluation to `T` in `T`-typed contexts.

The unification removes the `proc` keyword and renames scheduler helpers to `future_*` for consistency.

The Go reference execution modes implement this contract and share coverage for
status, memoized `value()`, cancellation, cooperative helpers, `await`, and
nested waits. `ProcHandle` remains a distinct host-process interop type; it is
not the removed Able language `proc` feature.

## Motivation

The prior split (`proc` for handles, `spawn` for implicit results) led to:

- Ambiguity over which handle could be cancelled and how.
- Conflicting semantics in the spec (e.g., Future cancellation).
- Two overlapping abstractions that complicate the parser, typechecker, fixtures, and mental model.

The unified `Future T` model collapses those concepts into a single, predictable surface.

## Design Overview

### Syntax

- **Removed:** `proc` keyword.
- **Kept:** `spawn FunctionCall` / `spawn BlockExpression`.
- **Result type:** `Future T`.

### Types & Interfaces

`Future T` is the single async handle type. It provides:

- `status() -> FutureStatus` (non-blocking)
- `value() -> !T` (blocking, returns `Error | T`)
- `cancel() -> void` (best-effort cancellation)

Supporting types:

```
struct Pending;
struct Resolved;
struct Cancelled;
struct Failed { error: FutureError }
union FutureStatus = Pending | Resolved | Cancelled | Failed

struct FutureError { details: String }
impl Error for FutureError { ... }
```

### Handle View vs Value View

`Future T` is a single value with two interpretations:

- **Handle view** (explicit): Used whenever a `Future T` is expected. Method calls do *not* trigger evaluation.
- **Value view** (implicit): Used whenever a `T` is expected. The task blocks and yields `T` (or raises on failure/cancellation).

This distinction is critical for correctness and should be encoded in the typechecker and interpreter.

### Memoization

The result of a spawned task is memoized once it reaches a terminal state.
Both `value()` calls and implicit evaluation reuse the cached result or error.

### Cancellation

Calling `cancel()` requests cancellation. Cancellation is best-effort and race-permitted; the first terminal state wins. When cancellation wins:

- `value()` returns `FutureError`.
- Implicit evaluation raises `FutureError` (handle via `rescue`).

### Scheduler Helpers

Renamed to align with the unified Future model:

- `future_yield()`
- `future_cancelled()`
- `future_flush()`
- `future_pending_tasks() -> i32`

These are still executor diagnostics and must preserve cooperative semantics.

### Await Integration

Futures remain `Awaitable`. `await` continues to multiplex async arms; cancellation of the enclosing spawned task must propagate the same `FutureError` semantics used elsewhere.

### Non-Goals

- No new syntax beyond removing `proc`.
- No change to OS process handles (`ProcHandle` in stdlib/interop) or to host process APIs.

## Historical Migration & Compatibility

The completed source migration was:

- `proc` → `spawn`
- `Proc`/`ProcError`/`ProcStatus` → `Future`/`FutureError`/`FutureStatus`
- `proc_*` helpers → `future_*` helpers

Fixtures and test names use `future_*` to match the unified model.

## Completed Implementation Scope

- **Spec:** Section 12 defines unified Future semantics and no longer lists
  `proc` as a keyword.
- **Parser/AST:** the grammar and mapper expose `spawn`; there is no
  `ProcExpression` AST node.
- **Typechecker:** Future value-view rules and `future_*` diagnostics are
  covered by the shared fixture corpus.
- **Interpreters and compiler:** the Go tree-walker, bytecode VM, and AOT
  compiler implement Future status/value/cancel, helper, and `await` paths.
- **Fixtures/tests:** executable fixtures cover status/error/memoization,
  cancellation/fairness, handle/value views, `await`, and nested spawn.
- **Stdlib/interop:** host-process `ProcHandle` remains separate from the
  language Future model.

Future performance changes remain subject to the roadmap's cross-application
selection rule; this completed semantic unification does not authorize a
scheduler-specific optimization.
