# Able WASM JS Host ABI (Draft)

This document defines the JavaScript host interface required by the Able v12 Go
runtime when compiled to WebAssembly. The goal is to keep the host surface
small, deterministic, and easy to polyfill in both Node and browser runtimes.

## Scope

The ABI covers:
- Stdout/stderr forwarding.
- Time + timers for the cooperative scheduler.
- Filesystem access used by module loading and stdlib IO.
- Module search roots (replacing `ABLE_MODULE_PATHS` in WASM).

The kernel bridges described in `spec/full_spec_v12.md` remain implemented inside
the runtime; this ABI only provides the host-facing pieces the WASM runtime
cannot implement on its own.

## Conventions

- **Import module name:** `able_host`.
- **Memory:** the runtime exports a linear `memory`; the host reads/writes UTF-8
  data directly into that buffer.
- **Strings:** UTF-8 bytes; lengths are byte counts (`i32`).
- **Handles:** host-returned handles are `i32` values; `0` is never a valid
  handle. Callers must close handles once consumed.
- **Errors:** functions return `0`/positive values on success and `-1` on error.
  On error, the host must populate a thread-local error message retrievable via
  `last_error_*`.
- **Variable-size outputs:** use a two-step `*_len` + `*_copy` pattern so the
  runtime can allocate the destination buffer.

## Required host imports

### Stdout / stderr

- `write_stdout(ptr: i32, len: i32) -> void`
- `write_stderr(ptr: i32, len: i32) -> void`

The host must decode the UTF-8 bytes at `ptr..ptr+len` and forward them to the
appropriate output stream.

### Time + timers

- `now_unix_nanos() -> i64`
- `set_timeout(ms: i64, token: i32) -> void`

`set_timeout` schedules a wakeup after `ms` milliseconds. When the timer fires,
the host must call the runtime export `__able_host_wake(token)` to resume the
awaiting task.

### Filesystem

Stat:
- `fs_stat(path_ptr: i32, path_len: i32) -> i32`

Return value is a bitmask on success:
- `1` = exists
- `2` = is_dir
- `4` = is_file

Read file:
- `fs_read_file(path_ptr: i32, path_len: i32) -> i32` (returns handle)
- `fs_read_len(handle: i32) -> i32`
- `fs_read_copy(handle: i32, out_ptr: i32) -> void`
- `fs_read_close(handle: i32) -> void`

Read directory:
- `fs_read_dir(path_ptr: i32, path_len: i32) -> i32` (returns handle)
- `fs_read_dir_len(handle: i32) -> i32`
- `fs_read_dir_copy(handle: i32, out_ptr: i32) -> void`
- `fs_read_dir_close(handle: i32) -> void`

`fs_read_dir_copy` writes a UTF-8 JSON array of directory entries (filenames
only). Example: `["mod.able","util.able","subdir"]`.

Write file:
- `fs_write_file(path_ptr: i32, path_len: i32, data_ptr: i32, data_len: i32, flags: i32) -> i32`

`flags` bitmask:
- `1` = append (otherwise truncate/overwrite)
- `2` = create_only (error if the path exists)

### Module search roots

- `module_roots_len() -> i32`
- `module_roots_copy(out_ptr: i32) -> void`

`module_roots_copy` writes a UTF-8 JSON array of objects:

```json
[
  {"path": "/able/stdlib", "kind": "stdlib"},
  {"path": "/workspace", "kind": "user"}
]
```

Allowed kinds are `stdlib` and `user`. The runtime will scan these roots in
order when resolving `import`/`dynimport`.

### Error reporting

- `last_error_len() -> i32`
- `last_error_copy(out_ptr: i32) -> void`

The host should overwrite the stored error message on every failure. The
runtime will read the message and surface it via `host_error(...)` or `Error`.

## Required WASM exports (runtime -> host callbacks)

- `__able_host_wake(token: i32) -> void`

The host must call this export to resume tasks scheduled via `set_timeout`.

## Notes

- This ABI intentionally avoids WASI to keep the browser target viable. A WASI
  adapter may implement the same surface in the future.
- Additional host hooks should be introduced only when required by the v12 spec
  or stdlib; keep the surface minimal and deterministic.
