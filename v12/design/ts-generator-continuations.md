# TypeScript Generator Continuations

Status: Draft – implementation pending  
Owners: Able v11 TypeScript interpreter track

## Goals
- Add lazy iterator semantics to the TypeScript interpreter that match §6.7 / §14 of the v11 spec and the new Go runtime implementation.
- Support generator literals (`Iterator { gen => ... }`) and the `Iterator`/`Iterable` protocol without forcing eager precomputation of elements.
- Keep the implementation incremental: surface the iterator runtime primitives first, then port `for` loops and stdlib helpers to consume them lazily.

## Requirements Recap
- Generator bodies run imperatively and may contain the full Able statement set (loops, match, rescue/ensure, nested blocks). Every `gen.yield(expr)` must suspend execution, surface `expr` to the consumer, and resume from the following statement on the next `next()` call.
- `gen.stop()` terminates the iterator. Subsequent `next()` calls return the singleton `IteratorEnd`.
- Iterators participate in member dispatch:
  - `next()` must be a native method on `(Iterator T)` values.
  - `iterator()` on iterable structs/interfaces returns `(Iterator T)`.
- Iteration consumers (e.g. `for` loops) drive arrays, ranges, or iterator values lazily, mirroring the Go interpreter.
- Prevent re-entrancy: calling `next()` again while the generator body is mid-yield should produce a deterministic runtime error (same as Go’s “re-entered while suspended” guard).

## Implementation Options Considered
1. **Reuse the future scheduler.** Treat generator bodies as hidden spawned tasks and expose the yielded value via the cooperative scheduler (`FutureYieldSignal`). This would let generators “pause” without new infrastructure, but resuming would re-run the body from the beginning unless we encoded manual stage variables (what future tests currently do). That fails the spec requirement to resume automatically after `yield`.
2. **Compile generator literals to explicit state machines (desugaring).** Translate the AST into a switch-based state machine stored on a closure. This quickly becomes intractable: we would need to model every statement form—including nested loops, `match`, error handling, and pattern assignments—to preserve semantics, effectively re-implementing the interpreter as a compiler pass.
3. **Introduce interpreter continuations for generators.** Instrument the evaluator so that, when running under a generator context, it records enough frame state (current node, environment, loop indices, etc.) to resume after a `YieldStatement`. Suspension is signalled via a dedicated error, and resumption walks the saved frame stack. This keeps semantics close to the existing interpreter and mirrors the Go coroutine runner.

**Decision:** adopt option (3). It requires targeted changes across the evaluator but keeps Able semantics authoritative, preserves host parity, and avoids a one-off lowering pass that would diverge from future interpreter features.

## Design Overview

### Runtime Values
- Extend `RuntimeValue` with:
  - `IteratorValue` carrying `state`, `next`, and `close` handlers.
  - `IteratorEndValue` singleton (`{ kind: "iterator_end" }`) used as the sentinel result.
- Expose `iterator_end` via globals (mirroring Go’s `IteratorEnd`) and ensure `valueToString` renders it as `"IteratorEnd"`.
- `IteratorValue.next()` drives the generator context; it returns either the yielded Able value or `IteratorEndValue`, and propagates runtime errors.

### Generator Context & Frames
- Add `generatorStack: GeneratorContext[]` to `Interpreter`. Each context tracks:
  - Captured lexical environment.
  - Generator controller object exposed as `gen`.
  - A frame stack (`GeneratorFrame[]`) describing the suspended execution.
  - Status flags (`started`, `done`, `busy`, `closed`, `pendingYield`).
- Introduce `GeneratorYieldSignal` and `GeneratorStopSignal` to unwind the JS stack without losing metadata.
- Augment evaluation helpers (`evaluateBlockExpression`, `evaluateIfExpression`, `evaluateForLoop`, `evaluateWhileLoop`, `evaluateMatchExpression`, `evaluateFunctionCall`, etc.) so that when a generator context is active they:
  - Register a frame describing their execution state (e.g., current statement index for blocks, iterator position for loops, clause index for `match`).
  - On normal completion, pop the frame.
  - When a `GeneratorYieldSignal` bubbles up, capture the current position on the frame, rethrow the signal, and leave the frame on the stack for resumption.
- `GeneratorContext.resume()` reconstructs the evaluation by re-invoking the appropriate helper with the saved frame positions, allowing execution to proceed after the suspended statement.

### Yield & Stop Hooks
- During iterator literal evaluation we synthesise the controller `gen` struct with native bound methods:
  - `gen.yield(value)`:
    - Validates re-entrancy (`busy` flag).
    - Stores the yielded value on the context and raises `GeneratorYieldSignal`.
  - `gen.stop()`:
    - Marks the generator as done, clears pending frames, and raises `GeneratorStopSignal`.
- `YieldStatement` evaluation delegates to `gen.yield` so generator authors can also write plain `yield expr` once the parser supports it; initial implementation will focus on `gen.yield`.

### Iterator Literal Evaluation
- `IteratorLiteral` creates a `GeneratorContext` with a cloned environment, installs the controller (`gen`) into a child scope, and returns an `IteratorValue` with:
  - `next()` => calls `context.next()`, which:
    1. Rejects re-entrancy.
    2. If first run, pushes the context on `generatorStack` and evaluates the body until it yields or completes.
    3. If resuming, reinstalls the frame stack, swaps in the controller, and continues from the saved state.
    4. Translates completion (`done`/`stop`) into `IteratorEndValue`.
  - `close()` => marks the context closed and clears outstanding frames (idempotent).

### For-Loop Integration
- Expand `evaluateForLoop`:
  - Arrays and ranges keep the eager path.
  - If the iterable is `iterator`, call `next()` repeatedly until `IteratorEnd`.
  - Otherwise, resolve `iterator()` via member dispatch (struct/interface) and consume lazily.
- Ensure loop bodies run under the loop’s environment on each iteration, and that generator errors propagate without being swallowed.

### Member Dispatch & Stringify
- Update `members.ts` so `next` on `IteratorValue` and UFCS fallbacks map to the runtime helper.
- Extend `stringify.ts` with cases for iterator values/end sentinels to match Go parity.

## Incremental Implementation Plan
1. **Scaffolding**
   - Define `IteratorValue` / `IteratorEndValue` in `values.ts`.
   - Add `GeneratorYieldSignal` / `GeneratorStopSignal`.
   - Introduce `GeneratorContext` scaffolding with no frame-saving yet; wire `IteratorLiteral` to return a stub that errors on `next()` (behind a feature flag in tests).
2. **Frame Recording for Blocks & Yield**
   - Handle block bodies and bare statements so simple generators (`yield` in straight-line code) work.
   - Add Jest tests mirroring Go’s simplest iterator cases.
3. **Control Flow Coverage**
   - Extend frame support to `if/or`, `while`, `for`, and pattern assignments.
   - Port Go iterator regression tests (laziness, error propagation, for-loop integration).
4. **Member Dispatch & Stdlib Hooks**
   - Surface `IteratorEnd`, ensure `iterator()` dispatch works, and update stdlib/channel helpers once runtime is stable.

## Testing Strategy
- Mirror the Go tests in `interpreter-go/pkg/interpreter/interpreter_iterators_test.go` with Bun tests under `v11/interpreters/ts/test/runtime/iterators.test.ts`.
- Add fixture coverage (`fixtures/ast/*`) once the runtime semantics are solid so both interpreters stay in lockstep.
- Run `./run_all_tests.sh` before landing to keep TypeScript + Go parity green.

## Open Questions
- Should we expose `Iterator::close` to Able code now or keep it internal until the spec calls for it?
- How does `ensure`/`rescue` inside generators interact with suspension? (Plan: bubbles through the same as in Go; document expectations.)
- Do we need cross-interpreter fixture updates immediately, or can TypeScript land first with local unit tests before regenerating fixtures?
