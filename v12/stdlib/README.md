# Able v12 Standard Library (Draft)

The v12 stdlib is being rebuilt in place. Active modules live under `v12/stdlib/src`; the pre-refactor copy sits in `v12/stdlib/quarantine` until each module is restored.

Current restored surface:
- core/errors, core/interfaces, core/options, core/iteration, core/numeric
- collections/array, enumerable, range, list, linked_list, vector, deque, queue, lazy_seq, hash_map, set, hash_set, bit_set, heap, tree_map, tree_set
- text/string (string iteration yields bytes; import before calling methods/iterators), text/regex (literal-only compile/match/find_all; metacharacters still raise RegexUnsupportedFeature)

Package name: `able`. Import modules with paths like `able.collections.hash_set`.

Testing and usage notes:
- Smoke suites live in `v12/stdlib/tests`.
- Run `go test ./...` in `v12/interpreters/go` to keep both interpreters aligned as modules are restored.
- Module loaders discover the stdlib via `collectModuleSearchPaths` (auto-detected bundled roots or search-path overrides) when wiring bespoke runners. The bundled scan now covers both `stdlib/src` and `v12/stdlib/src` alongside the new `v12/kernel/src` bootstrap package.
- Manifest-driven runs default to the bundled boot packages: `able deps install` writes stdlib + kernel entries into `package.lock` when they are not specified, and the CLIs load lockfile roots before falling back to bundled detection.

Next steps: keep restoring modules from quarantine one at a time, add/refresh tests as they land, and continue trimming native runtime surfaces per `PLAN.md` items 7–10.
