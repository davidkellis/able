# Snapshot Store Design (Draft)

## Goals

- Persist snapshot expectations to disk so `match_snapshot` can diff real files instead of relying on the in-memory defaults.
- Keep the API flexible: projects can inject custom stores, and the CLI can toggle update mode via stdlib helpers.
- Support deterministic layout so snapshots round-trip cleanly in version control and tooling.
- Provide room for richer diffs later without complicating the initial implementation.

## File Layout

- Store snapshots alongside tests under a predictable directory:
  - For a test module `spec/foo/bar.test.able`, default snapshot path: `spec/foo/__snapshots__/bar.snap`.
  - Each `.snap` file contains one or more snapshot entries (see serialization).
- Nested suites use breadcrumb paths to build names (`Array::push`). To prevent path explosion, we flatten to file name + identifier:
  - `bar.snap` contains entries: `Array::push`, `Array::pop`, etc.
  - Alternative: one file per snapshot; but grouping reduces filesystem clutter.
- CLI calculates snapshot directory by replacing the filename with `__snapshots__/<basename>.snap`. Configurable via CLI flag or manifest (future).

## Serialization Format

Human-readable, friendly to diffs. Candidate formats:
- **TOML or YAML**: simple key/value with heredoc blocks.
- **Custom plain text**: minimal overhead, closer to Jest's format.

Proposed TOML variant:

```toml
[[snapshot]]
name = "Array::push"
value = """
<rendered output here>
"""

[[snapshot]]
name = "Array::pop"
value = """
...
"""
```

- Consecutive `[[snapshot]]` tables, each with `name` and `value` string.
- Escaped multi-line content via `"""` to preserve newlines. Future metadata (version, updated_at) can be added per entry.

## Update Mode Semantics

- CLI sets update mode prior to module load using `snapshot_set_update_mode(true)` when `--update-snapshots` is passed.
- During `match_snapshot`:
  - If snapshot doesn't exist: create file/directory and write entry (update mode only). Without update mode, fail with “missing snapshot”.
  - If snapshot exists and differs: overwrite when update mode is on; otherwise, fail.
- After tests finish, reset update mode to `false`.
- Consider printing summary like “Updated 3 snapshots” via CLI (extend harness or store API to expose counts).

## Store Implementation

Add file-backed store in stdlib interacting with `SnapshotStore` API:

```able
fn file_snapshot_store(root: string) -> SnapshotStore

struct FileSnapshotStore {
  root: string,
  update: bool
}

impl SnapshotStore for FileSnapshotStore { ... }
```

- `read(name)` → map snapshot name to `<root>/<module>/__snapshots__/<file>.snap`, load TOML, return matching entry.
- `write(name, value)` → ensure directory exists, merge entry into TOML (preserve sort order by name).
- `update_mode()` returns global toggle provided by CLI (mirrors `snapshot_set_update_mode`).
- Provide helper `snapshot_set_store(store: SnapshotStore)` so CLI swaps default store before module load.

### Name Mapping

- Snapshot names currently use `build_example_id` (`Array::push`). We need a stable file path for each module:
  - Module path recorded in `ExampleDefinition.module_path` -> use to derive snapshot file. Example: `spec/foo/bar` → `spec/foo/__snapshots__/bar.snap`.
  - When module path absent, fallback to descriptor ID hashed (avoid collisions).
- `match_snapshot` should accept optional explicit path override later: `match_snapshot_in(name, path)`.

## Diff Strategy

- Initially rely on string diff produced by CLI/reporter when `details` = `snapshot diff (expected vs actual)`. CLI can call external diff (e.g., `diff` command) or use built-in string diff once available.
- Later enhancements:
  - Provide utility to generate unified diff and include in JSON/TAP outputs.
  - Add snapshot metadata (e.g., `updated_at`, `file`) to help reporters locate the snapshot file.

## CLI Integration Steps

1. Compute snapshot root (default: project root). Allow override via `--snapshot-dir` or config file in future.
2. Before module evaluation:
   - Create file-backed store: `snapshot_set_store(file_snapshot_store(root))` (requires new setter in stdlib).
   - Toggle update mode if `--update-snapshots` is set.
3. After run, optionally report counts (store tracks writes/updates).

## Stdlib Changes Required

- Extend `able.testing.snapshots` with:
  - `snapshot_set_store(store: SnapshotStore)` to override global store.
  - `snapshot_get_store()` for introspection/tests.
  - `file_snapshot_store(root: string)` implementation.
- Update `match_snapshot` to call `snapshot_current_store()` (instead of default store) so CLI injection works.
- Store should normalize newlines (e.g., convert Windows `\r\n` to `\n`) for consistency.
- Ensure thread-safety if harness ever runs tests in parallel (guard with simple mutex or restrict to single writer and note limitation).

## Open Questions

- How to handle binary snapshots? For now, scope to text; consider base64 encoding or separate binary support later.
- Should we support snapshot pruning (delete entries not referenced)? Possibly via CLI flag `--prune-snapshots` comparing recorded names vs. stored file.
- Do we need per-snapshot metadata (e.g., custom serializer)? Maybe bring in when we add structured diff support.

## Next Steps

1. Implement `snapshot_set_store` and in-memory defaults, ensuring backwards compatibility.
2. Implement `file_snapshot_store` with TOML serialization (Bun interpreter may need `toml` helper or simple manual parser if dependencies are heavy).
3. Update `match_snapshot` to use current store.
4. Extend CLI (future) to inject store and manage update mode.
5. Add stdlib tests that write to a temp directory to ensure file store works, using Bun FS helpers.

