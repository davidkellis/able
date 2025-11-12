# Able Stdlib I/O Package (Draft)

## Goals

- Provide a cohesive standard library surface for filesystem operations that works consistently across Able runtimes.
- Support both high-level helpers (e.g., `read_text`) and lower-level streaming primitives (`File`) so advanced features (snapshot store, tooling) can build on the same package.
- Use idiomatic Able error handling (`!T`, domain-specific `Error` types) instead of exceptions for routine failures.
- Keep the package modular to allow future expansion (temporary files, networking) without overloading a single module.

## Non-Goals (Initial Cut)

- Asynchronous/non-blocking I/O abstractions.
- Network/socket APIs.
- Platform-specific metadata beyond cross-platform basics (e.g., extended ACLs, symlink management) in v1.

## Package Layout

```
able.io
├── fs        # Files/directories, File handle, convenience helpers
├── path      # Path manipulation helpers
└── temp      # Temporary files/directories (future)
```

Initial focus: `fs` and `path`; `temp` can arrive once the core surfaces are proven. The `fs` package hosts both minimal primitives and higher-level helpers so users only need a single import.

### `able.io.fs`

#### Types

- `enum FileMode = Read | Write | Append | ReadWrite`
- `struct OpenOptions { read: bool, write: bool, append: bool, create: bool, truncate: bool }`
- `struct File { handle: i64 }` (opaque runtime-managed handle)
  - Methods: `read`, `write`, `flush`, `close`, each returning `!T`.
- `struct FileStat { size: i64, is_dir: bool, is_file: bool, modified_at: ?DateTime }`
- `struct DirEntry { name: string, path: string, is_dir: bool, is_file: bool }`

#### Functions (All returning `!T` unless noted)

- `fn read_text(path: string, encoding?: string = "utf-8") -> !string`
- `fn write_text(path: string, contents: string, encoding?: string = "utf-8", create_dirs?: bool = false) -> !void`
- `fn read_bytes(path: string) -> !Array u8`
- `fn write_bytes(path: string, bytes: Array u8, create_dirs?: bool = false) -> !void`
- `fn open(path: string, options: OpenOptions) -> !File`
- `fn exists(path: string) -> bool`
- `fn stat(path: string) -> !FileStat`
- `fn remove_file(path: string) -> !void`
- `fn create_dir(path: string, recursive?: bool = false) -> !void`
- `fn remove_dir(path: string, recursive?: bool = false) -> !void`
- `fn read_dir(path: string) -> !Array DirEntry`

#### Errors

Define under `able.io.errors` (or within `fs`):

- `struct FileNotFoundError { path: string }`
- `struct PermissionDeniedError { path: string }`
- `struct AlreadyExistsError { path: string }`
- `struct InvalidInputError { message: string }`

All implement `Error` with clear messages. Consumers can narrow via `rescue` or `!` propagation.

#### Convenience Helpers (within `fs`)

To keep the package ergonomic, ship higher-level helpers built atop the primitives:

- `fn read_json<T>(path: string) -> !T` and `fn write_json<T>(path: string, value: T, pretty?: bool = false) -> !void` (depends on stdlib JSON encoder/decoder).
- `fn copy_file(src: string, dst: string, overwrite?: bool = false) -> !void`
- `fn copy_dir(src: string, dst: string, overwrite?: bool = false) -> !void`
- `fn touch(path: string) -> !void` (create if missing, update mtime otherwise).
- `fn read_lines(path: string) -> !Array string` / `fn write_lines(path: string, lines: Array string, newline?: string = "\n") -> !void`.
- `fn temp_dir(prefix?: string) -> !TempDir` re-exported from `able.io.temp` when that module lands.

These helpers should use the core API internally so the runtime only needs one set of host functions. Advanced users can ignore them; newcomers get convenience without extra packages.

### `able.io.path`

Pure helpers (no filesystem interaction):

- `fn join(parts: Array string) -> string`
- `fn dirname(path: string) -> string`
- `fn basename(path: string) -> string`
- `fn extension(path: string) -> ?string`
- `fn normalize(path: string) -> string` (resolve `.` / `..` components without touching disk)
- `fn is_absolute(path: string) -> bool`

These ensure consistent path behaviour across runtimes.

## Runtime Responsibilities

- **TypeScript (Bun)**: wrap Bun’s `fs` APIs. Methods can call synchronous variants or use async under the hood while presenting blocking semantics.
- **Go**: use `os`, `io`, `path/filepath`. `File.handle` can wrap `*os.File` stored in runtime tables.
- Methods like `File.read`/`write` become extern/native functions implemented per runtime.

## Snapshot Store Integration

- Snapshot store will rely on `read_text`, `write_text`, `exists`, `create_dir`, and `read_dir` to manage `__snapshots__` directories.
- File-backed store will call `path.join` and `path.normalize` to build deterministic paths.
- Update mode semantics (write vs. fail) remain inside snapshot matcher; the store only provides primitives.

## Error Handling Strategy

- Use `!T` return types for operations that can fail; `exists` remains `bool` to avoid error overhead.
- Provide descriptive errors with path context to ease debugging.
- Avoid raising exceptions for routine errors; reserve `raise` for unrecoverable situations (e.g., internal invariants).

## Future Extensions

- `able.io.temp`: `mkdtemp`, `TempDir`, `TempFile` with auto-cleanup.
- Buffered readers/writers (e.g., `BufReader(File)`), bridging to `Iterator` for line-by-line reading.
- File watching APIs when runtimes expose watchers.
- More detailed metadata (`FilePermissions`, `Ownership`) behind optional features.
- `able.io.net`: separate network primitives with similar “minimal core + convenience” layering.

## Implementation Plan

1. **Define Interfaces**: Add `.able` module definitions under `stdlib/v10/src/io/fs.able` and `io/path.able` with extern stubs.
2. **Runtime Binding**: Implement TypeScript and Go externs for file operations; keep behaviour consistent (e.g., text encoding defaults).
3. **Error Types**: Introduce `able.io.errors` with the structs listed above.
4. **Tests**: Create stdlib tests using temporary directories to verify read/write/list semantics in both runtimes.
5. **Docs**: Add manual page / update existing docs describing new APIs and recommended patterns (e.g., using `!` to propagate errors).
