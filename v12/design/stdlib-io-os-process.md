# Stdlib IO/OS/Process Design (v11)

Status: Draft

## Goal
Provide a complete, cross-platform IO/FS/OS/process standard library built on
minimal extern primitives declared in stdlib. Keep the kernel unchanged and
preserve correctness across Linux/macOS/BSD/Windows.

## Principles
- Opaque handles: no file descriptors are exposed to Able.
- Byte-first boundary: externs operate on bytes; stdlib owns text decoding.
- Consistent errors: structured `IOError` with portable `kind` values.
- Handle identity is preserved across extern calls; no registry is exposed.

## Core Types
`IoHandle` and `ProcHandle` are primitive, opaque handle types (see spec §4.2.1).
```able
struct TermSize { rows: i32, cols: i32 }

union IOErrorKind =
  NotFound | PermissionDenied | AlreadyExists | InvalidInput |
  TimedOut | BrokenPipe | Closed | EOF | Unsupported | Other

struct IOError { kind: IOErrorKind, message: String, path: ?String }
```

## Package Layout
- `able.io`: byte readers/writers, buffered IO, `read_line`, `write`, `flush`.
- `able.term`: raw mode, terminal size, key parsing, line editor.
- `able.fs`: file/directory operations built on IO handles.
- `able.os`: environment, cwd, temp/home dirs, args, exit.
- `able.process`: spawn/wait/kill, pipe wiring, stdio handles.

## Extern Surface (Declared in Stdlib)
### able.io
```able
extern go fn io_stdin() -> IoHandle { ... }
extern go fn io_stdout() -> IoHandle { ... }
extern go fn io_stderr() -> IoHandle { ... }
extern go fn io_read(handle: IoHandle, max_bytes: i32) -> Result (?Array u8) IOError { ... }
extern go fn io_write(handle: IoHandle, bytes: Array u8) -> Result i32 IOError { ... }
extern go fn io_flush(handle: IoHandle) -> Result void IOError { ... }
extern go fn io_close(handle: IoHandle) -> Result void IOError { ... }
```
(Parallel TypeScript externs are defined in the same modules.)

### able.term
```able
extern go fn term_is_tty(handle: IoHandle) -> bool { ... }
extern go fn term_set_raw_mode(handle: IoHandle, enabled: bool) -> Result void IOError { ... }
extern go fn term_get_size(handle: IoHandle) -> Result TermSize IOError { ... }
```

### able.fs
```able
extern go fn fs_open(path: String, flags: FsOpenFlags, mode: ?FsMode) -> Result IoHandle IOError { ... }
extern go fn fs_stat(path: String) -> Result FsMetadata IOError { ... }
extern go fn fs_read_dir(path: String) -> Result Array DirEntry IOError { ... }
extern go fn fs_mkdir(path: String, recursive: bool) -> Result void IOError { ... }
extern go fn fs_remove(path: String, recursive: bool) -> Result void IOError { ... }
extern go fn fs_rename(src: String, dst: String) -> Result void IOError { ... }
```

### able.os
```able
extern go fn os_env(name: String) -> ?String { ... }
extern go fn os_set_env(name: String, value: String) -> Result void IOError { ... }
extern go fn os_cwd() -> String { ... }
extern go fn os_chdir(path: String) -> Result void IOError { ... }
extern go fn os_home_dir() -> ?String { ... }
extern go fn os_temp_dir() -> String { ... }
extern go fn os_args() -> Array String { ... }
extern go fn os_exit(code: i32) -> void { ... }
```

### able.process
```able
extern go fn process_spawn(spec: ProcessSpec) -> Result ProcHandle IOError { ... }
extern go fn process_wait(handle: ProcHandle) -> Result ProcessStatus IOError { ... }
extern go fn process_kill(handle: ProcHandle, signal: ProcessSignal) -> Result void IOError { ... }
extern go fn process_stdin(handle: ProcHandle) -> IoHandle { ... }
extern go fn process_stdout(handle: ProcHandle) -> IoHandle { ... }
extern go fn process_stderr(handle: ProcHandle) -> IoHandle { ... }
```

## Text and Line Handling
- `able.io.TextReader` decodes UTF-8 and normalizes `\r\n` to `\n`.
- `read_line` returns `?String` (nil on EOF).
- `able.term` line editor uses raw bytes to parse key sequences; key parsing and
  history live in stdlib.

## Error Semantics
- Externs return `Result` with `IOError`; stdlib raises on `Error` by default,
  but also exposes `try_*` helpers for explicit handling.
- `IOError.kind` is mapped from platform-specific errors (e.g., `ENOENT` ->
  `NotFound`, `EPIPE` -> `BrokenPipe`).

## Handle Identity
Handles are passed through extern calls as opaque host objects. Identity is
preserved without a registry; externs must treat handle values as opaque and
avoid introspection beyond their target-language APIs.

## Host Mapping (Go/Bun)
Go:
- `IoHandle` is `*os.File` (stdin/stdout/stderr and files opened via `os.Open`).
- `ProcHandle` is `*exec.Cmd` (with stdio wired to `*os.File` handles).
- `term_is_tty` uses `golang.org/x/term` or `syscall` as needed.

TypeScript (Bun):
- `IoHandle` is a small object wrapping a file descriptor, e.g. `{ fd: number }`.
- Externs use `node:fs` async APIs where possible (`read`, `write`, `close`) and
  resume the suspended task when complete.
- `ProcHandle` wraps the spawned process (e.g., `{ proc: ChildProcess }`).
  `process_wait` listens for exit asynchronously and resumes the suspended task
  when the process exits.
- `term_is_tty` uses `node:tty.isatty(handle.fd)` and `setRawMode` when present.

## Async Host APIs (TypeScript)
Extern bodies may complete asynchronously, but they must integrate with the
interpreter's suspension mechanism so the current task yields while the host
operation is in flight. This keeps the Able API synchronous while preserving
Go-like progress for other procs.

## Performance
- Byte buffers use `Array u8` and allow caller-provided reuse.
- `BufferedReader` and `BufferedWriter` provide amortized IO performance.
- Process piping relies on IO handles to avoid extra copies.
