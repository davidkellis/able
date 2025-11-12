# Able v10 Standard Library (Draft)

This directory houses the versioned Able v10 standard library.  The canonical
package name is `able`; all modules live under this namespace.

The initial cut focuses on providing the foundational interfaces and data
structures that the language specification expects to exist in library space:

- Core protocols (`Iterator`, `Iterable`, `Index`, `Add`, `Eq`, `Ord`, etc.)
- Error contracts and the well-known error structs raised by the runtime
- Collection scaffolding (`Array`, `HashMap`) that future work will connect to
  host-runtime intrinsics

The implementation bodies in this draft are intentionally skeletal and will be
fleshed out alongside the runtime.  Each method includes a `TODO` comment
indicating the behaviour required by the specification so downstream work can
fill in the details on a per-target basis.

## Layout

```
stdlib/v10/
├── package.yml        # Able manifest (package name `able`)
└── src/
    ├── lib.able       # Convenience entry point
    ├── core/
    │   ├── errors.able
    │   ├── interfaces.able
    │   ├── iteration.able
    │   └── options.able
    ├── collections/
    │   ├── array.able
    │   ├── hash_map.able
    │   └── range.able
    └── concurrency/
        ├── channel.able
        ├── future.able
        ├── mutex.able
        └── proc.able
```

Each `.able` file uses directory layout to establish its package path
(`able.core`, `able.collections`, ...).  Modules may freely import from the
same root when additional factoring makes sense.

## Next Steps

- Back Array/HashMap methods with host-runtime implementations (Go + TS).
- Flesh out additional collection protocols (Range iterators, Set, Queue).
- Provide concrete implementations for `Hasher` plus default hash functions.
- Wire concurrency primitives (`Proc`, `Future`, `Channel`, `Mutex`) into each runtime and add fixtures.
- Implement range constructors and higher-level helpers (additional Option/Result utilities, with-lock helpers).

## Loader Integration

The Go CLI now discovers `stdlib/v10/src` automatically. To point the toolchain
at an alternate location, set `ABLE_STD_LIB` to a path (or OS-specific
path-list) that contains the standard library sources. This augments the
existing `ABLE_PATH`/`ABLE_MODULE_PATHS` mechanism used for project-level overrides.
