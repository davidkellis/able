# Channels & Mutexes – Host-Backed Concurrency Primitives

Status: Draft – interpreter implementation pending  
Owners: Able v10 interpreter team

## Overview
- Provide first-class `Channel<T>` and `Mutex` types with Crystal-style APIs.
- Keep Able surface syntax unchanged; expose functionality through standard library structs.
- Implement the semantics natively inside each interpreter (TypeScript + Go) via extern/prelude hooks.

## Design Shape
1. **Stdlib API**
   - Define `struct Channel<T>` / `struct Mutex` in Able source under `stdlib/v10/concurrency`.
   - Public API mirrors Crystal:
     - `Channel.new(capacity = 0)`, `send`, `recv`, `close`, `receive?`, etc.
     - `Mutex.new`, `lock`, `unlock`, `with_lock`.
   - Methods that need host semantics use `extern <target>` bodies; higher-level helpers (e.g., `with_lock`) stay in Able code.

2. **Interpreter Responsibilities**
   - Extend `V10Value` with `channel` and `mutex` variants containing scheduler metadata (queues, owners, buffer, closed flag).
   - Register native helpers when the interpreter boots:
     - `__able_channel_new`, `__able_channel_send`, `__able_channel_recv`, `__able_channel_close`.
     - `__able_mutex_new`, `__able_mutex_lock`, `__able_mutex_unlock`.
   - Implement helpers using the cooperative scheduler:
     - Channel send/recv block and resume procs via `ProcYieldSignal`, respecting FIFO semantics and buffer capacity.
     - Mutex lock/unlock manage ownership and waiting proc queue.
   - Expose helpers to extern bodies via `prelude` initialisation.
   - Mirror behaviour in the Go interpreter (channels map to Go `chan`, mutexes to `sync.Mutex`, while preserving Able semantics).

3. **Spec Alignment**
   - No new AST nodes or syntax required; all operations remain standard method/struct usage.
   - Update v10 spec concurrency section to document Channel/Mutex as host-backed primitives available via stdlib + extern.

4. **Testing & Fixtures**
   - After runtime support lands, add AST fixtures covering:
     - Channel creation, buffered/unbuffered send & receive across procs, close semantics.
     - Mutex lock/unlock across multiple procs, `with_lock` behaviour, error cases.
   - Update `design/parser-ast-coverage.md` rows 121–122 once fixtures exist.

## Open Questions / Follow-ups
- Decide on optional helpers (`receive?`, nonblocking send) and error reporting strategy (exception vs result).
- Ensure Go + TS interpreters expose a uniform extern helper API so stdlib stays target-agnostic.
- Document manifest expectations for concurrency fixtures (e.g. result vs stdout) when they land.

## TypeScript Runtime Alignment (2025-10-23)

With the executor contract shared across Go and TypeScript, the next milestone is to mirror Go-style blocking semantics for channels/mutexes inside the TypeScript interpreter. Key decisions/principles:

1. **State tracking**
   - Extend `channelStates` entries to include `sendWaiters` and `receiveWaiters`, each storing cooperative task handles plus any buffered payload.
   - Add optional fields on `ProcHandleValue` (e.g., `pendingChannelSend`, `pendingChannelReceive`) so a proc reactivated after `proc_yield` can resume without re-evaluating payload expressions. The pending payload captured on the first attempt is reused on subsequent retries.
2. **Send path**
   - Fast-path when buffered capacity or waiting receivers allow immediate delivery (mirrors Go rendezvous behaviour).
   - When the channel is full (or unbuffered with no receivers), enqueue the current proc + payload, mark it pending, and yield. The wake-up signal is triggered by either a consumer draining buffered data or a new receiver arriving.
3. **Receive path**
   - Return immediately if buffered data exists.
   - If a sender is waiting, atomically transfer its stored payload to the receiver, reschedule the sender proc, and complete without buffering.
   - Otherwise, enqueue the receiver proc, mark it pending, and yield until a sender arrives or the channel is closed.
4. **Nil channel semantics**
   - Preserve Able’s rule that nil handles block forever unless cancellation is requested. The pending receive/send bookkeeping should treat handle `0` as an infinite wait without enqueuing real state.
5. **Determinism & fairness**
   - Wake suspended procs in FIFO order (matching Go channel semantics) to keep fixtures deterministic under the cooperative scheduler.
   - Ensure `proc_flush` drains any runnable continuations so fixtures remain reliable.
6. **Error handling**
   - Surface `"send on closed channel"` / `"receive on closed channel"` errors consistently with the Go runtime.
   - Guard against double-enqueue by clearing pending state when a proc is cancelled; cancellation should remove the waiter entry and propagate an appropriate error when the proc resumes.

Implementation landed in `interpreter10/src/interpreter/channels_mutex.ts` (with supporting `ProcHandle` metadata) and is covered by dedicated Bun tests plus the existing AST fixtures. Remaining follow-ups:

- Mirror any future Go enhancements (e.g., select/timeouts) once spec language settles.
- Audit nil-channel cancellation paths under heavier load; consider property tests around cancellation race behaviour.
- Keep spec prose in sync as we document the helper guarantees called out in the TODO.

## 2025-11-05 Wiring Audit (Phase α wrap-up)

- **Runtime helpers registered in both interpreters.** Verified that `InterpreterV10.ensureChannelMutexBuiltins` and the Go `initChannelMutexBuiltins` install the full helper set (`__able_channel_new/send/receive/try_send/try_receive/close/is_closed`, `__able_mutex_new/lock/unlock`) and seed per-handle state (TS: cooperative queues; Go: buffered `chan`/`sync.Mutex`).
- **Typechecker/fixture coverage.** Confirmed the helper signatures are declared in `interpreter10/src/typechecker/checker.ts` and `interpreter10-go/pkg/typechecker/decls.go`, and that the shared AST fixture suite already exercises channel/mutex semantics (`fixtures/ast/concurrency/*`, `fixtures/ast/stdlib/channel_mutex_helpers`). No schema drift detected between TS and Go decoders.
- **New interpreter smoke tests.** Added Bun tests for nil-channel cancellation and mutex re-entry errors (`interpreter10/test/concurrency/channel_mutex.test.ts`) to mirror the Go parity suite behaviour.
- **Outstanding parity TODO.** Native helpers still raise generic runtime errors (`"send on closed channel"`, `"channel already closed"`) instead of materialising the stdlib `ChannelClosed/ChannelNil/ChannelSendOnClosed` structs. Capture conversion logic on both runtimes before we graduate Phase β; tracked in `PLAN.md` and `spec/todo.md`.
