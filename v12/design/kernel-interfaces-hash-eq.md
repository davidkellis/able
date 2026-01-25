# Kernel Interfaces + Primitive Hash/Eq Plan

Status: Draft
Last updated: 2026-01-15

## Context

The v12 spec says primitive types ship with always-available implementations of
`Display`, `Clone`, `Default`, `Eq`/`Ord`, and `Hash`. Today those interfaces live
in `able.core.interfaces` (stdlib), but the implementations are implicit in the
interpreters. This yields three problems:

- The stdlib cannot implement or override primitive `Eq`/`Hash` directly.
- The kernel `HashMap` only accepts primitive keys and hashes them internally,
  so user-defined `Hash`/`Eq` impls do not affect `HashMap`/`HashSet`.
- The TS typechecker has to special-case primitive `Hash`/`Eq` satisfaction.

We want the most correct solution: keep interpreter internals minimal and move
as much as possible into Able code, while preserving the spec guarantee that
primitive `Eq`/`Hash`/etc are always available.

## Goals

- Make `Eq`/`Hash` (and related always-available interfaces) defined in a kernel
  library that is loaded ahead of user code.
- Provide explicit primitive impls in the kernel library so the behaviour is
  still "always available" but not hardcoded in interpreters.
- Ensure `HashMap`/`HashSet` honour the public `Hash`/`Eq` contract for all key
  types, not just primitives.
- Keep TS/Go interpreters thin: only host helpers that cannot be expressed in
  Able should remain native.
- Update the spec to describe the kernel interface location and hashing rules.

## Non-Goals (for this tranche)

- Rewriting HashMap as a fully Able-level data structure.
- Generalizing hashing for persistent collections beyond the minimal changes
  needed to honour `Hash`/`Eq`.

## Current State (Summary)

- `able.core.interfaces` defines `Eq`, `Hash`, `Hasher`, `Display`, etc.
- Kernel `HashMap` uses a host-only hashing path (TS `keyLabel`, Go equivalent),
  which rejects non-primitive keys and does not call `Hash.hash` or `Eq.eq`.
- `persistent_map` uses a `Hasher` bridge but hashes primitives by string
  conversion rather than `Hash.hash`.
- Spec section 14 states primitive impls are provided by runtimes.

## Proposed Design

### 1) Kernel-Resident Interfaces

Move the "always available" interfaces into the kernel library so they load
before any stdlib modules. The kernel package (`able.kernel`) already exists and
is loaded ahead of workspace code, so it is the natural home for these
definitions.

Interfaces to define in the kernel library:
- `Display`, `Clone`, `Default`
- `PartialEq`, `Eq`
- `PartialOrd`, `Ord`
- `Hash`, `Hasher`

The stdlib `able.core.interfaces` becomes a shim that re-exports the kernel
interfaces (to keep import paths stable). Do not re-define them in stdlib to
avoid `(Interface, Type)` duplication.

### 2) Kernel Primitive Implementations

Add explicit impls for all primitive types in the kernel library:
- `bool`
- all integer widths (`i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`)
- float widths (`f32`, `f64`)
- `char`
- `String`

These implementations must avoid recursive calls. In particular, `Eq.eq` cannot
be implemented in terms of `==` if `==` itself dispatches to `Eq.eq`. The kernel
should use dedicated helpers for raw equality and hashing.

### 3) Kernel Hash/Eq Helpers (Host Bridges)

Implement `Hasher` in Able code (kernel library) while adding minimal host
bridges for the missing low-level bit primitives. This keeps hashing logic in
Able but still allows raw-bit hashing for floats and deterministic byte order.

Kernel `Hasher` implementation:
- `struct Hasher { state: u64 }` (FNV-1a in Able code).
- `Hasher.write_u8`, `write_u16`, `write_u32`, `write_u64`, `write_i*`, `write_bool`,
  `write_char`, `write_bytes`, `finish`.
- `write_*` routines emit bytes in a canonical big-endian order and update
  `state = (state ^ byte) * FNV_PRIME`.
- `Hash.hash` for primitives calls these `write_*` helpers and never hashes
  stringified values.

Host bridges required to support Able-level hashing:
- `__able_f32_bits(value: f32) -> u32`
- `__able_f64_bits(value: f64) -> u64`
- `__able_u64_mul(lhs: u64, rhs: u64) -> u64` (wrapping multiply for FNV-1a)
- (reuse) `__able_String_from_builtin(value: String) -> Array u8`
- (reuse) `__able_char_to_codepoint(value: char) -> i32`

These helpers avoid exposing host internals while giving Able code the exact bit
patterns needed for hashing.

### 4) HashMap/HashSet Contract

Update the kernel `HashMap` host implementation to:
- Accept any key type that implements `Hash` + `Eq`.
- Compute a hash via `Hash.hash` and use `Eq.eq` when resolving collisions.
- Preserve existing iteration order behaviour.

This requires interpreter changes so `__able_hash_map_*` can call Able methods
(`hash` / `eq`) on the provided keys. This keeps the hash map implementation
hosted but respects the user-defined interface contract.

`HashSet` builds on `HashMap`, so once `HashMap` uses `Hash`/`Eq`, `HashSet`
inherits the correct behaviour.

### 5) Stdlib Re-exports + Persistent Map Alignment

- `able.core.interfaces` should re-export the kernel definitions.
- Remove or convert any existing stdlib impls that would conflict with kernel
  primitive impls.
- Update `persistent_map` to use `Hash.hash` consistently for all key types
  (and remove string-based primitive hashing).

### 6) Typechecker + Loader Alignment

Typecheckers should treat kernel interface definitions as always in scope.
Remove built-in special-casing for primitive `Eq`/`Hash` once the kernel impls
exist. Update prelude resolution so `able.kernel` interfaces are always
available even when the stdlib is absent.

### 7) Spec Updates Required

The v12 spec should be updated to:
- Describe the kernel library and which interfaces live there.
- Define the canonical `Hash`/`Hasher` signatures (resolve the current mismatch
  between spec and stdlib).
- Clarify that built-in primitive implementations are provided by the kernel
  library, not ad-hoc runtime behaviour.
- Specify `HashMap`/`HashSet` semantics as driven by `Hash` + `Eq`.
- Codify float equality semantics (IEEE `==`) and note that floats do not
  implement `Eq`/`Hash` by default.

### 8) Fixtures + Tests

Exec fixtures should cover:
- Primitive key hashing in `HashMap` and `HashSet` (u8, i32, u64, bool, char,
  String, f32/f64).
- Custom `Hash`/`Eq` implementations (ensure the map uses them; add counters).
- Intentional hash collisions (Hash returns a constant; `Eq` disambiguates).
- Float edge cases: NaN/Inf behaviour once specified.
- Kernel interface availability without importing `able.core.interfaces`.

Interpreter tests should cover:
- TS/Go host hash map calling into `Hash.hash`/`Eq.eq`.
- New hasher helper APIs for primitives.

## Migration Strategy (Incremental)

1) Land spec/TODO + design doc updates.
2) Add kernel interface definitions and basic re-export shims in stdlib.
3) Introduce host hashing helpers and kernel primitive impls.
4) Update `HashMap` host kernels to call `Hash`/`Eq`.
5) Align `persistent_map` hashing and remove primitive hashing hacks.
6) Remove typechecker special-casing for primitive `Hash`/`Eq`.
7) Add exec fixtures + tests; update parity harness.

## Open Decisions

- Whether `HashMap` must reject keys that do not implement `Hash`/`Eq`, or
  whether this is a typechecker-only constraint.

## Decisions (2026-01-15)

- `Hash.hash` remains sink-style: `fn hash(self, hasher: Hasher) -> void`, with
  `Hasher.finish()` producing the final `u64`. This keeps hashing incremental,
  composable, and aligned with `persistent_map`/kernel helpers.
- The default `Hasher` implementation lives in the kernel library and is
  implemented in Able (FNV-1a with big-endian byte emission). Host runtimes only
  supply bitcast/byte helpers (`__able_f32_bits`, `__able_f64_bits`, and existing
  string/char bridges) so Able code can hash raw bit patterns deterministically.
- Float equality uses IEEE semantics for `==`/`!=` (so `NaN != NaN` and
  `+0 == -0`). Floats implement `PartialEq`/`PartialOrd` only; they do not
  implement `Eq`, `Ord`, or `Hash` in the kernel. Hash-based containers therefore
  reject raw `f32`/`f64` keys unless wrapped in a user-defined type that provides
  `Eq`/`Hash` (e.g., `NotNaN`, `FloatKey`).
- Primitive interface impls remain kernel-owned and non-overridable, so
  alternate float equality/hashing must be expressed via wrapper types or named
  impls used explicitly (never via `==` or `HashMap` keys).

## Risks

- Changing `HashMap` internals affects hash stability and iteration order;
  parity tests must lock in the expected behaviour.
- Introducing new kernel helpers requires updates in both interpreters and
  typecheckers; missing one will break parity or loading.
