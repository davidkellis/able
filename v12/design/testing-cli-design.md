# `able test` UX Draft: Historical Record

## Status

The active user-facing command and framework contract is
[testing-cli-protocol.md](./testing-cli-protocol.md). This former design draft
predates the Go-first implementation and is retained only for its UX rationale.
It does not define flags, a runner roadmap, or a TypeScript workstream.

## Retained UX principles

- User tests should be low ceremony: co-located, explicitly importing a
  framework, discoverable by conventional test suffixes, and runnable through
  one command.
- Test-framework semantics belong in ordinary Able source and the canonical
  stdlib; the CLI orchestrates modules, requests, and presentation rather than
  embedding a parallel assertion or suite model.
- Machine-readable output and predictable exit behavior are first-class so
  local and automated use share one command.
- User-authored tests remain distinct from AST, exec, parity, compiler, and
  benchmark fixtures that validate the Able implementation itself.

Those principles are now realized by the Go `able test` command and external
`able.test`/`able.spec` source. The active record names the supported filters,
formats, modes, discovery, and exit behavior.

## Superseded assumptions

The former proposed command surface is not authoritative. In particular,
TAP/JSON output and compiled test execution are implemented; the TypeScript
CLI skeleton, quarantine stdlib layout, reporter implementation phases, and
"implement the command" next step are historical. The active CLI does not
support the proposed `--filter-file`, `--quiet`, `--color`, `--pattern`, or
per-test timeout flags, and no such feature is scheduled by this record.

Remote runners, artifact caches, alternate glob engines, and cross-framework
parallel scheduling need a concrete tooling requirement before they can change
the canonical protocol or CLI. They are not performance, WASM, compiler, or
runtime work.
