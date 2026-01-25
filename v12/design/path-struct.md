# Path Struct Design (Draft)

## Goals

- Provide a first-class `Path` type similar to Crystal’s, offering chainable, immutable path manipulation.
- Keep the representation platform-neutral (canonical `/` separators), converting to native separators only when interacting with the host.
- Support both POSIX (`/usr/bin`) and Windows (`C:\Users\...`, UNC paths) semantics without forcing callers to branch.
- Integrate smoothly with `able.io.fs` so functions accept either `Path` or `string`.

## Representation

```able
struct Path {
  root: string   ## "", "/", "C:", "//server/share", etc.
  segments: Array string  ## normalized, no empty strings, no '.'
}
```

- Internally we canonicalize all separators to `/` and resolve `.` / `..` where safe.
- `root` captures drive/UNC prefixes so we can reconstruct native paths, detect absolutes, and produce correct relative paths.
- Methods return new `Path` values; the struct is immutable from a caller’s perspective.

## Constructors

- `path.new(parts: Path | string | Array string)`
- `path.parse(string)` (alias for `path.new` with normalization)
- `path.from_segments(Array string)` (assumes segments already normalized)
- `path.current()` / `path.home()` (requires runtime support via `fs`)

## Core Methods (aligned with Crystal’s API)

- `path.join(other: Path | string) -> Path`
- `path / other` (syntactic sugar via method)
- `path.parent -> Path`
- `path.basename -> string`
- `path.basename_without_extension -> string`
- `path.extension -> ?string`
- `path.with_extension(ext: string) -> Path`
- `path.segments -> Array string`
- `path.root -> string`
- `path.normalize -> Path` (re-run normalization)
- `path.clean -> Path` (alias)
- `path.relative?` / `path.absolute?`
- `path.relative_to(base: Path) -> Path`
- `path.absolute -> Path` (resolve against current directory via runtime helper)
- `path.expand_home -> Path` (resolve `~` using environment)
- `path.to_string -> string` (canonical, `/` separators)
- `path.to_native -> string` (uses platform separator)

## Filesystem Helpers (delegating to `fs`)

- `path.exists? -> bool` (calls `fs.exists`)
- `path.file? -> !bool`, `path.dir? -> !bool`
- `path.stat -> !FileStat`
- `path.read_text`, `path.write_text` convenience wrappers (internally call `fs`)

## Normalization Rules

- Treat both `/` and `\` as separators when parsing.
- Collapse empty components and `.` segments.
- Handle `..` cautiously:
  - When rooted, do not traverse above root.
  - When relative, allow `..` to accumulate at the beginning (e.g., `../../foo`).
- Preserve trailing separator intent via flag if necessary (future).

## Windows Considerations

- Recognize drive letters (`C:`) and UNC prefixes (`\\server\share`).
- `path.absolute?` returns true if root is non-empty or UNC.
- `path.to_native` inserts `\` separators and reinstates drive/UNC prefixes.
- Normalize incoming mixed separators (`C:/Users` ➜ `C:/Users`).

## Integration with `able.io.fs`

- Allow all `fs` functions to accept `Path | string` via overload or by calling `Path` constructor internally.
- Provide helper `fn fs_path(path: Path | string) -> Path` to standardize conversion.

## Implementation Plan

1. Implement parsing/normalization helpers (pure Able code) in `stdlib/src/io/path.able`.
2. Provide the `Path` struct with methods described above; defer filesystem-dependent methods (`exists?`, `absolute`) until `fs` package wiring is available.
3. Add tests covering POSIX paths, Windows drive paths, UNC paths, join/relative operations, normalization.
4. Update `design/io-package.md` to reference `Path` integration points (done).
5. Once `fs` package lands, add convenience wrappers and runtime-dependent methods.
