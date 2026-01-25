# Package Management and CLI Blueprint

## Goals

- Treat the Able standard library just like any other dependency: no hidden built-ins.
- Provide a manifest/lock file pair (`package.yml` / `package.lock`) inspired by Crystal’s `shard.yml` but with Cargo-style reproducibility guarantees.
- Expose a CLI surface that feels familiar to Crystal (`able build`, `able run`, `able deps`, …) while delivering Cargo-grade dependency management.
- Cache packages globally under a configurable `ABLE_HOME` (default `$HOME/.able`) with deterministic layout for registry mirrors and tooling.

## Manifests and Lock Files

### Manifest (`package.yml`)

Top-level keys:

- `name`, `version`, `license`, `authors` (strings/arrays)
- `targets`: map of short target name → entrypoint Able source file (relative to the manifest directory). Every target currently builds as an executable, and all dependencies are shared across targets.
- `dependencies`, `dev_dependencies`, `build_dependencies`: map of dependency name → descriptor
- `workspace`: reserved for future multi-package coordination

Dependency descriptor fields (values optional depending on source):

- `version`: semantic constraint, e.g. `"~> 1.2.0"`
- `git`: repository URL, plus optional `rev`, `tag`, `branch`
- `path`: local override for development
- `registry`: alternate registry name/URL
- `features`: list of enabled feature flags
- `optional`: boolean for optional deps

### Lock File (`package.lock`)

A generated file capturing resolved dependency graph:

- Header metadata: `root`, `generated` timestamp, CLI/tool version
- For each package:
  - `name`
  - `version`
  - `source`: registry URL, git URL + commit, or local path id
  - `checksum`: integrity hash of source archive
  - `dependencies`: list of `{ name, version }` pairs actually used

Lock updates only happen via explicit commands (e.g., `able deps update`). Normal builds reuse the recorded graph for reproducibility.

## CLI Surface

- `able build [target]`: compile the selected target (default first executable target)
- `able run [target] [-- args]`: build then execute an entrypoint
- `able deps install`: resolve manifest, update lock if missing, download and cache dependencies
- `able deps update [package]`: re-resolve constraints and refresh lock entries
- `able check`: run parser/typechecker without producing binaries
- `able test [target]`: execute test targets (depends on test harness)
- `able fmt`: apply formatter (future)
- `able env`: print environment (paths, cache directory)
- `able init`: scaffold new package structure (`src/`, `spec/`, manifest)

Subcommand behavior mirrors Crystal where possible but integrates Cargo-like dependency resolution semantics under the hood.

## Global Cache Layout

`ABLE_HOME` (defaults to `$HOME/.able`) contains:

- `bin/`: installed executables (`able install target`)
- `pkg/`
  - `cache/registry/<registry-id>/index`: registry metadata clone (git/http)
  - `cache/downloads/<sha>.tar.gz`: immutable source archives per version
  - `src/<namespace>/<package>/<version>/`: unpacked sources for resolution (Go-style deterministic paths)
- `registry/`: optional curated registry manifests
- `artifacts/`: build outputs (optional)
- `log/`, `tmp/`: diagnostics and transient data

Project layout:

- `lib/`: symlink or hardlink into cached `pkg/src/...` directories for editor visibility
- `package.lock`: pinned versions and checksums used by CLI and loader

## Package Resolution Flow

1. Loader indexes local project sources per Able spec (package paths from directories + optional declarations).
2. For imports that do not belong to the current root, consult `package.lock` for resolved package/version.
3. Locate cached sources under `$ABLE_HOME/pkg/src/...`. If missing, instruct user to run `able deps install`.
4. Evaluate stdlib modules via the same mechanism (stdlib published as a versioned package).
5. Support override paths in manifest (e.g., local dev builds) before hitting global cache.

## Roadmap

- Manifest parser + schema validation (YAML → Go structs, semver constraints)
- Lock file writer/reader with checksum enforcement
- Cache manager: download, verify, and unpack archives into deterministic paths
- CLI commands: implement `deps install/update`, integrate with existing `run` flow
- Loader enhancements: search project `lib/` and `$ABLE_HOME/pkg/src` for external packages
- Package stdlib with its own manifest; seed cache during setup
- Integration tests covering dependency resolution and lock reproducibility
- ✅ Transitive dependency resolution across path/registry/git manifests with dependency edges captured in `package.lock`
