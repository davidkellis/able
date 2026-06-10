# Able Test CLI and Framework Protocol: Historical Draft Record

## Status

The live command and protocol are in
[testing-cli-protocol.md](./testing-cli-protocol.md). This record preserves
the pre-implementation protocol draft and its proposed extensions; it is not a
testing roadmap.

## Retained design decisions

The early design correctly separated user-authored tests from interpreter
fixtures and required co-located `.test.able`/`.spec.able` modules. It proposed
ordinary Able framework registration, a shared descriptor/event protocol,
central command invocation, and human/TAP/JSON output. Those principles are
now implemented through the Go CLI and external canonical stdlib.

The framework shape was intentionally language-level: a framework identifies
itself, discovers descriptors, and runs a plan while emitting events through a
reporter. The current canonical definitions supersede the former inline
pseudocode; use `../able-stdlib/src/test/protocol.able`, `harness.able`,
`registry.able`, reporters, and `spec.able` for field names and behavior.

## Deferred proposals

The draft also described features that are not active commitments:

- manifest `tests: false` opt-outs or custom discovery globs;
- framework configuration schemas and `--framework-opt` flags;
- artifact attachments, tag conventions, or additional CLI shortcuts;
- remote/distributed execution and serialized plan transport; and
- additional hierarchy fields beyond the existing descriptor metadata.

The old discussion of future runner caches and test-artifact directories is
also not a current compiler/build contract. Normal source commands keep their
existing `--with-tests` behavior, while `able test` performs its own discovery
and load flow.

## Use

Use this record to evaluate a specific testing/tooling requirement. A feature
must first have a stable language or user workflow need, then update the
canonical stdlib protocol and Go CLI together with interpreted and compiled
coverage. Do not infer a performance, WASM, or runtime optimization assignment
from user-test infrastructure.
