# Go Interpreter Concurrency Notes

Date: 2025-10-19  
Author: Able Agents

## Overview

The Go interpreter is the reference implementation for Able v12 concurrency. It
expands `spawn` expressions into goroutine-backed tasks that share a
thread-safe runtime environment. The goal is to make parallel execution safe
for the interpreter itself while leaving synchronisation of user-managed data
structures in the programmer’s hands.

Key properties:

- **Goroutine executor** – Production mode uses `GoroutineExecutor`, which runs
  each spawned task on its own goroutine with a cancellable `context`.
- **Serial executor** – Tests default to `SerialExecutor`, a single worker loop
  that preserves deterministic ordering. `New()` wires this executor so unit
  tests retain stable traces; production callers can swap in the goroutine
  executor explicitly.
- **Thread-safe environments** – `runtime.Environment` now owns an
  `RWMutex`. Reads and writes are guarded so multiple goroutines can touch
  shared scopes without corrupting the interpreter’s bookkeeping.
- **Per-task interpreter state** – Breakpoint and raise stacks are stored in an
  `evalState` carried inside each async payload. The interpreter no longer
  serialises async evaluation with a global mutex; goroutines can run in
  parallel while still reporting cancellation and raises accurately.
- **User-controlled shared state** – The runtime guarantees the integrity of its
  own environments. Protecting shared structs/arrays remains the user’s
  responsibility (for example, by wrapping mutations with a mutex exposed via a
  native helper). Tests demonstrate that multiple tasks can update a shared
  string safely when user code locks around the critical section.
- **Helpers** – `future_yield`, `future_cancelled`, and `future_flush` are defined as
  plain native functions. `future_cancelled` inspects the current async payload,
  and now raises an error when called from outside an async task – a behaviour
  covered by dedicated tests.

## Implications for parser & fixtures

- Shared AST fixtures should assume that `spawn` bodies can run on
  different goroutines. Any fixture that relies on deterministic interleaving
  must use explicit coordination (e.g., native locks or channel helpers).
- `future_flush` in Go simply blocks until the executor’s queue drains. There is
  no additional scheduling beyond the goroutine scheduler, so fairness fixtures
  from the TypeScript interpreter must be reviewed before being ported.
- Parser work that emits Able code should not depend on TypeScript’s cooperative
  scheduler behaviour; the Go semantics here are authoritative.

## Open follow-ups

- Expand Go-side fixtures to cover `future_flush` semantics now that the executor
  no longer serialises tasks automatically.
- Audit shared fixtures under `fixtures/ast/concurrency` to ensure they are
  compatible with the goroutine model; keep any TS-specific scheduling tests out
  of the Go parity harness.
- Parser alignment roadmap:
  - Emit `spawn` nodes exactly as the Go AST helpers expect so runtime
    evaluation does not need parser-specific shims.
  - Surface cooperative helper imports (`future_yield`, `future_cancelled`,
    `future_flush`) in generated modules when users opt into concurrency helpers.
  - Add parser-driven integration tests that run Able source through the new
    parser + Go interpreter pipeline to verify goroutine semantics end-to-end.
