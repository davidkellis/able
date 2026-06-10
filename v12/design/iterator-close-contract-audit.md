# Public Iterator `close()` contract audit

**Status:** public contract and lazy `able.fs.lines` implemented and verified

## Problem

The v12 specification requires the public iterator protocol to expose both
`next() -> T | IteratorEnd` and `close()`. The canonical
`able.core.iteration.Iterator T` interface now declares both methods.

This is not merely a documentation omission. The Go runtimes already retain
internal iterator finalizers (`runtime.IteratorValue.Close()`), and `for`
loops close their runtime iterator adapters. Able source, however, cannot
reliably close an iterator early through the public interface. That makes a
lazy file-line iterator unsafe to add: a caller that stops early needs a
portable way to release its file handle.

## Implemented public contract

`able.core.iteration.Iterator T` provides this default method:

```able
fn close(self: Self) -> void {}
```

The default preserves every existing iterator implementation. Resource-owning
iterators override it. The contract is:

- `close()` is idempotent and may be called before the first `next()`, after
  `IteratorEnd`, or more than once.
- After a successful close, subsequent `next()` calls return `IteratorEnd` and
  must not resume work or read another resource value.
- A resource close failure raises the existing domain error on the first call;
  the iterator is still considered closed so retries do not repeat an
  externally visible close attempt.
- Exhaustion closes resource-owning iterators automatically. Explicit close is
  needed for early termination.
- Runtime-generated iterator values retain their current once-only finalizer;
  the public method must dispatch to that same finalizer rather than creating
  a second lifecycle.

## Implemented propagation

Default `Iterable.each`, `Iterator.filter`, `map`, `filter_map`, and `collect`
close their input on normal completion, early generator cancellation, and
error exit. A wrapper iterator propagates its own close to the captured
upstream iterator exactly once. The tree-walker generator now stops at a
closed yield boundary so its `ensure` cleanup runs; compiled loops close at
loop exit as well as on function exit.

`for` closes both runtime and statically lowered iterator carriers in every Go
execution mode, and agrees with direct source-level `iterator.close()` calls.

## Verified core matrix

`exec/06_12_27_stdlib_iterator_close` passes in tree-walker, bytecode, and
compiled execution. It proves typed custom iterator idempotence, direct
generator close, `for`-break cleanup, and normal-completion propagation through
`filter`, `map`, `filter_map`, `collect`, and `each`. The focused Go generator
test also proves that early close runs `ensure` cleanup without resuming the
generator body.

## `able.fs.lines` API

The public close contract now enables this lazy line iterator:

```able
fn lines(path: Path | String) -> FileLines
```

`FileLines` owns an `IoHandle` plus `BufferedReader`, implements
`Iterator String`, normalizes LF/CRLF with the existing reader semantics, and
closes its handle on EOF or explicit `close()`. Its `close` method calls the
qualified `io.close` operation so the resource close cannot resolve to the
iterator method itself. A read error marks the iterator closed, attempts
best-effort handle cleanup without hiding the original error, and is then
re-raised. Opening a missing path raises the existing `IOError { kind:
NotFound }`. `read_lines` remains the eager convenience API; it is neither
removed nor silently rewritten to a streaming algorithm.

## Verified `FileLines` fixture matrix

`exec/06_12_28_stdlib_fs_lines` passes in tree-walker, bytecode, and compiled
execution. The compiler fixture uses its normal `RequireNoFallbacks` gate and
is in the default compiler-fixture set. It proves:

1. A temporary-file `FileLines` iterator returns LF and CRLF lines, preserves a
   final unterminated line, auto-closes at EOF, and supports an early explicit
   close followed by cleanup/rename.
2. `for`-break closes the iterator before cleanup/rename. Missing-file and an
   induced closed-handle read preserve `IOError` kind, and the failed iterator
   subsequently returns `IteratorEnd` without another read.
3. Parsed shorthand field lowering uses the bytecode named-struct opcode, so
   `FileLines { handle, reader, closed: false }` evaluates its local values in
   the VM rather than falling back to tree evaluation through a closure.

No benchmark is changed in this stage. Any later streaming performance
application must have fair Go, Python, and Ruby counterparts and must be used
only as one member of the broad performance suite.
