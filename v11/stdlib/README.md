# Able v11 Standard Library (Draft)

The v11 stdlib is being rebuilt in place. Active modules live under `v11/stdlib/src`; the pre-refactor copy sits in `v11/stdlib/quarantine` until each module is restored.

Current restored surface:
- core/errors, core/interfaces, core/options, core/iteration, core/numeric
- collections/array, enumerable, range, list, linked_list, vector, deque, queue, lazy_seq, hash_map, set, hash_set, bit_set, heap, tree_map, tree_set
- text/string (string iteration yields bytes; import before calling methods/iterators), text/regex (stubbed; operations currently return RegexUnsupportedFeature)

Package name: `able`. Import modules with paths like `able.collections.hash_set`.

Testing and usage notes:
- Smoke suites live in `v11/stdlib/tests` and the TS ModuleLoader integration tests under `v11/interpreters/ts/test/stdlib`.
- Run targeted `bun test v11/interpreters/ts/test/stdlib/...` and quick `go test ./...` in `v11/interpreters/go` to keep both runtimes aligned as modules are restored.
- Module loaders discover the stdlib via `collectModuleSearchPaths` (auto-detected bundled roots or search-path overrides) when wiring bespoke runners. The bundled scan now covers both `stdlib/src` and `v11/stdlib/src` alongside the new `v11/kernel/src` bootstrap package.
- Manifest-driven runs default to the bundled boot packages: `able deps install` writes stdlib + kernel entries into `package.lock` when they are not specified, and the CLIs load lockfile roots before falling back to bundled detection.

Next steps: keep restoring modules from quarantine one at a time, add/refresh tests as they land, and continue trimming native runtime surfaces per `PLAN.md` items 7–10.
