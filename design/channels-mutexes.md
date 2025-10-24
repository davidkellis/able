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
