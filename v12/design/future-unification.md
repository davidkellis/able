# Unified Future Model (Proc/Spawn Unification)

Date: 2026-01-23  
Status: Approved (spec + design alignment pending implementation)  
Owners: Able v12 maintainers

## Summary

Able v12 now defines a single async facility: `spawn` returns a `Future T` handle that also supports implicit evaluation to `T` when a value is required. This unifies the previously separate `Proc T` and `Future T` concepts into one coherent model with two views:

- **Handle view:** explicit `status()`, `value()`, `cancel()` methods.
- **Value view:** implicit blocking evaluation to `T` in `T`-typed contexts.

The unification removes the `proc` keyword and renames scheduler helpers to `future_*` for consistency.

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
- `future_flush(limit?: i32)`
- `future_pending_tasks() -> i32`

These are still executor diagnostics and must preserve cooperative semantics.

### Await Integration

Futures remain `Awaitable`. `await` continues to multiplex async arms; cancellation of the enclosing spawned task must propagate the same `FutureError` semantics used elsewhere.

### Non-Goals

- No new syntax beyond removing `proc`.
- No change to OS process handles (`ProcHandle` in stdlib/interop) or to host process APIs.

## Migration & Compatibility

Source changes required by users/tests:

- `proc` → `spawn`
- `Proc`/`ProcError`/`ProcStatus` → `Future`/`FutureError`/`FutureStatus`
- `proc_*` helpers → `future_*` helpers

Fixture and test names should migrate from `proc_*` to `future_*` to match the new model.

## Implementation Impact (High-Level)

- **Spec:** rewrite Section 12 to define unified Future semantics; remove `proc` from keywords and examples.
- **Parser/AST:** remove `ProcExpression` node; update placeholder scoping and parser tests.
- **Typechecker:** remove `ProcType`, add implicit-evaluation rules for `Future` in `T` contexts, update helper names and diagnostics.
- **Interpreters:** merge proc/future handle structures; implement cancellation on Future; update scheduler helpers and await integration.
- **Fixtures/Tests:** rename fixtures, adjust expected diagnostics, update parity harness and example programs.
- **Stdlib:** collapse `able.concurrent.proc` into `able.concurrent.future` (or move definitions into a single module); update helper names.

This work is fully enumerated in the updated `PLAN.md` work breakdown.
