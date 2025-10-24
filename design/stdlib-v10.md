# Able v10 Standard Library Notes

_Last updated: 2025-10-27_

## Purpose

Record decisions and outstanding questions while we stand up the v10 standard
library. The goal is to keep the language surface consistent across runtimes
by treating the stdlib as a regular Able package that happens to ship with the
toolchain.

## High-Level Shape

- **Versioned package**: `stdlib/v10/` is the canonical source for the v10
  standard library. The manifest is a normal `package.yml` with `name: able`
  so the loader resolves imports like `import able.core`.
- **Entry point**: `src/lib.able` currently just imports the core modules;
  we are not re-exporting aggressively yet because we expect tooling to
  reference the more specific modules (e.g. `able.core.interfaces`).
- **Compilation/Distribution**: At build/install time the package manager will
  place `able` into `$ABLE_HOME/pkg/src/...` just like any other dependency.
  Projects can pin the stdlib version through `package.lock`.
- **Tooling discovery**: The Go CLI now auto-adds `stdlib/v10/src` to the
  loader search path (deduced via `ABLE_STD_LIB`, `$PWD`, or the executable
  location). This keeps `import able.*` working without manual path tweaks.

## Implemented Modules (Skeleton)

- `core/interfaces.able`: foundational protocols (Display, Clone, Default,
  Hash/Hasher, Error, PartialEq/Eq, PartialOrd/Ord, arithmetic/bitwise operators,
  Apply, Index/IndexMut, Not).
- `core/iteration.able`: iteration protocol (`IteratorEnd`, `Iterator`, `Iterable`,
  `Range` constructor interface).
- `core/errors.able`: runtime error structs required by the spec
  (`DivisionByZeroError`, `OverflowError`, `ShiftOutOfRangeError`, `IndexError`).
- `core/options.able`: canonical `Option T` / `Result T` unions, helper
  predicates, and unwrap/map/or-else/ok-or combinators aligned with §11.
- `collections/array.able`: placeholder `Array<T>` struct with the API shape the
  spec expects (constructors, push/pop, len/capacity, Iterable + Index traits).
- `collections/hash_map.able`: `Map` interface plus skeletal `HashMap<K,V>`
  struct using open addressing semantics.
- `collections/range.able`: integer range implementation providing
  `Iterator`/`Iterable` support and a `Range` factory for `i32`.
- Go runtime now exposes native `Array` helpers (`Array.new`, `push`, `pop`,
  `clone`, `size`, `get`, `set`) and `HashMap` helpers (`new`, `set`, `get`,
  `remove`, `contains`, `size`, `clear`) built on a deterministic FNV-1a based
  key hash supporting strings, numbers, bools, chars, and array composites. The
  stdlib wrappers can forward to these while we flesh out the Able APIs.
- `concurrency/proc.able`: `ProcStatus`, `ProcError`, and the `Proc T` interface.
- `concurrency/future.able`: opaque `Future T` shell (await semantics TBD).
- `concurrency/channel.able`: `Channel T` scaffolding with error types aligned to
  spec (`ChannelClosed`, `ChannelNil`, `ChannelSendOnClosed`).
- `concurrency/mutex.able`: `Mutex` wrapper and `with_lock` helper stub.

All behavioural bodies are `...` for now; host runtimes will provide concrete
implementations once we settle on the FFI surface.

## Design Decisions

- **Opaque storage handles**: `Array` and `HashMap` carry integer handles instead
  of embedding slices/maps directly. This lets Go/TS runtimes manage memory and
  keep the Able AST contract independent from host representations.
- **Protocols first**: We codified interfaces even if implementations are stubbed
  so other projects (interpreters, typechecker) can reference the canonical types
  immediately.
- **Error structs mirror spec**: We covered the arithmetic/indexing errors from
  §11 and added concurrency errors surfaced in §12 (channel misuse).
- **No wild-card re-exports yet**: To reduce churn we are not providing `pub use`
  surfaces until we understand how tooling wants to consume the modules.

## Outstanding Work

- Flesh out the `Array`/`HashMap` methods in Go + TypeScript runtimes and add
  conformance fixtures.
- Provide a concrete `Hasher` implementation and default hashing utilities (Go runtime now exposes FNV-backed hashers; need Able wrappers/tests).
- Implement range constructors and canonical Range/Iterator types. (i32 version
  stubbed; need generic numeric support and integration with parser range
  literals.)
- Back the concurrency shells (`Proc`, `Future`, `Channel`, `Mutex`) with real
  host integrations and scheduling hooks.
- Define formatting helpers (`Display` derivations) and string utilities.
- Publish documentation/tests alongside each module once bodies are real.

## Open Questions

- How do we expose host-managed buffers safely to Able code? (Copy-on-write?
  shared references?)
- Should `Array` expose mutating iterators, or do we rely on immutable snapshots?
- How do we version stdlib breaking changes (semver vs spec minor revs)?
- What is the minimal surface we must stabilize before external packages can
  meaningfully target the stdlib?
