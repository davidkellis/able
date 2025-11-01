# Able v10 Standard Library Vision

Last updated: 2025-02-14
Status: Draft — for review with project leads

## 1. Purpose & Scope

- Capture a cohesive vision for the Able v10 standard library that aligns with the language specification and ongoing interpreter work.
- Provide a target inventory of packages, protocols, and core data types that every runtime must support.
- Document the design inspirations (Scala, Crystal, Ruby, Rust, Go) and how their strengths inform Able.
- Record open decisions so we can converge quickly as implementation work begins.

## 2. Guiding Principles

**Spec Alignment**
- Treat `spec/full_spec_v10.md` as authoritative; stdlib features must extend, not contradict, the spec.
- Respect the AST contract — stdlib types must be representable without host-specific leakage.
- Keep observable behaviour identical across the Go and TypeScript interpreters.

**Ergonomics & Familiarity**
- Bias toward Crystal’s standard library structure: clear separation of eager `Collection` APIs and lazy `Iterator` pipelines, explicit module mixins, and predictable return types.
- Offer Scala-inspired parity between immutable and mutable collections with identical method names (`push`, `pop`, `map`, `flat_map`, etc.).
- Adopt Crystal/Go-inspired `String` primitives: UTF-8 by default, indexed by byte with codepoint iterators, and efficient slicing via views.

**Performance Contracts**
- Publish Big-O guarantees for core operations on each collection (Scala approach).
- Provide predictable allocation behaviour (Crystal/Go) via `reserve`, `with_capacity`, and builder APIs.
- Default to persistent (immutable) data in pure contexts, with mutable variants available when callers opt in.

**Runtime & Concurrency Integration**
- Collections and protocols must cooperate with Able’s cooperative scheduler and Go’s goroutines.
- Respect cancellation semantics for iterators and async constructs (e.g., `Sequence#each_proc` must yield).

## 3. Package Layout & Module Tiers

| Tier | Package path | Summary |
| --- | --- | --- |
| Prelude | `able` | Minimal surface automatically imported; re-exports Option/Result, core interfaces. |
| Core | `able.core.*` | Interfaces, Option/Result, errors, range helpers, tuples. |
| Text | `able.text.*` | String, Substring views, builders, Unicode utilities. |
| Collections (immutable) | `able.collections.immutable.*` | Persistent Vector/List/Set/Map/Range, lazy views. |
| Collections (mutable) | `able.collections.mutable.*` | Array, Deque, LinkedList, HashMap, HashSet, TreeMap/TreeSet, Heap. |
| Functional utilities | `able.fn.*` | Higher-order helpers, partial application utilities, memoization. |
| Concurrency | `able.concurrent.*` | Proc/Future/Channel/Mutex wrappers, scheduler helpers, concurrent collections. |
| Numerics | `able.math.*` | Int/Float helpers, BigInt/Decimal, random. |
| IO & Time | `able.io.*`, `able.time.*` | Streams, file system, networking stubs (host-provided), time/duration types. |

## 4. Foundational Interfaces & Protocols

| Interface | Purpose | Required methods |
| --- | --- | --- |
| `Equatable` | Value equality | `equals(self, other: Self) -> bool` |
| `Hashable` | Hash participation | `hash(self, hasher: Hasher) -> void`; requires `Equatable`. |
| `Comparable` | Total ordering | `compare(self, other: Self) -> Ordering` (`Ordering` union: `Less`, `Equal`, `Greater`). |
| `Clone` | Deep copy | `clone(self) -> Self`; `Persistent` structs can implement cheap clone. |
| `Default` | Identity element | `default() -> Self`. |
| `Display` | Human-readable string | `to_string(self) -> string`. |
| `Debug` | Developer formatting | `inspect(self) -> string`; optional formatting params later. |
| `Iterable` | Produces iterators | `iter(self) -> Iterator Item`. |
| `Iterator` | Lazy pull-based pipeline | `next(self) -> Option Item`; optional `size_hint`, `rewind?`. |
| `IteratorPipeline` | Default combinators | `map`, `filter`, `flat_map`, `zip`, `chunk`, `tap`, all returning `Iterator`. (Implemented as default methods on `Iterator`). |
| `Collection` | Sized aggregate with eager ops | `len`, `is_empty`, `contains`, `count`, `each`. Provides eager `map`/`select` returning `Self`. |
| `MutableCollection` | In-place mutation | `clear`, `reserve`, `retain`. Extends `Collection`. |
| `Indexable` | Random access | `get`, `get_mut?`, `set`; safe optional return + checked variant raising errors. |
| `Sliceable` | Views | `slice(range) -> Self`, `split_at(idx) -> (Self, Self)`. |
| `Collectable<C>` | Realize iterators | `collect(source: Iterator Item) -> C`. Implemented by concrete collection builders. |
| `Builder<C>` | Efficient construction | `push`, `extend`, `finish -> C`. Use for persistent collections. |
| `Serializable` | Structured persistence | `serialize`, `deserialize`. (Phase 2 goal). |
| `ProcLike` | Callable values | Align with `Apply` interface already in spec; used for passing closures/procs. |

Notes:
- Interfaces above should live in `able.core.interfaces` with Crystal-inspired module naming (`Iterator(T)`, `Iterable(T)`, `Collection(T)`).
- Provide blanket implementations where possible (e.g., any `Iterator` automatically mixes in `IteratorPipeline` to gain lazy adapters).
- `Hasher` should expose streaming API with deterministic seed (FNV-1a baseline, customizable later).
- Maintain extensibility: avoid baking host-specific behaviour into interface contracts so future specialized libraries (numerics, text processing, etc.) can implement them without compromise.

### 4.1 Lazy vs Eager Defaults (Crystal Alignment)

- Any `Collection` must expose `each` that returns an `Iterator`. Chaining off `each` (or `iter`) produces **lazy** pipelines. Example: `array.each.map(&.succ).select(&.even?)` returns an `Iterator`.
- Direct calls on the collection (`array.map`, `set.filter`) are **eager** and return a concrete collection of the same type by default.
- Methods ending with `_iter` (e.g., `map_iter`, `filter_iter`) exist on collections to shortcut to laziness without an explicit `each`.
- Cross-collection conversions use `collect_into(TargetType)` or `to_array`, `to_vector`, etc., to make the target explicit.
- `Iterator#to(Type)` and `Iterator#collect` are the primary realization points; equivalently `Iterator#to_a` matches Crystal.

### 4.2 Pipeline Realization Hooks

- Keep the lazy story focused on `Iterator` pipelines; no first-class transducers in v10.
- `Iterator#collect_into(collectable: Collectable)` materializes the sequence without exposing transducer primitives.
- Persistent collections reuse iterator combinators internally; we can revisit transducers in a future spec once primitives settle.

## 5. String & Text Plan

- `String`: UTF-8 owned buffer, copy-on-write semantics using reference-counted handles managed by host runtimes. Methods: `len_bytes`, `len_chars`, `empty`, `push`, `pop`, `insert`, `remove`, `replace`, `to_upper`, `to_lower`, `strip`, `split`, `lines`, `bytes`, `chars`, `graphemes`. Provide `with_capacity`, `reserve`, `clear`.
- `SubString`: view into a `String` (start byte + length). Cheap slicing without copying; read-only, `Clone` shallow copies.
- `StringBuilder`: mutable buffer specialized for concatenation. Implements `Builder<String>`.
- `ByteString`: explicit binary sequence for non-text payloads.
- Unicode helpers: `Codepoint`, `GraphemeCluster`, normalization utilities (Phase 2).
- Interop: `String` implements `Iterable` over grapheme clusters by default; `bytes` view returns `Iterable<u8>`.
- Inspired by Go’s string immutability and Crystal’s `Slice` types while offering Ruby-like convenience API.

### 5.1 Host Runtime Backing (Go & TypeScript)

**Strings**
- **Go**: Able `String` should wrap the native `string` type for storage efficiency and fast slices. Because Go strings are immutable byte slices, we can store them directly while tracking UTF-8 length metadata in an auxiliary struct (`struct { raw: string, grapheme_len: i32 }`). Reference counting is unnecessary in Go, but we expose copy-on-write semantics so persistent collections can share handles.
- **TypeScript**: JavaScript strings are UTF-16; to deliver UTF-8 semantics we store the original JS string plus a lazily materialized `Uint8Array` cache representing UTF-8 bytes. Core operations (`len_bytes`, `slice`) operate on the byte cache; we update it lazily. This keeps interop with host APIs simple while honoring Able semantics.
- **Bridging APIs**: Provide `String::from_host_string(host_value)` / `to_host_string()` in a `able.core.host` module so extern packages can interoperate without violating invariants.

**Numbers**
- Able distinguishes signed/unsigned integers (`i8`..`i128`, `u8`..`u128`) and floats (`f32`, `f64`).
- **Go**: Map each Able numeric directly onto the corresponding Go primitive (`int64`, `uint64`, `float64`, `big.Int` for wider widths). We keep helper constructors in the interpreter to validate range and convert to canonical `AbleNumber` structs carrying type tags.
- **TypeScript**: JavaScript only offers IEEE-754 doubles and `BigInt`. We represent fixed-width integers via `BigInt` internally, tagging them with their Able width. Floats remain JS numbers. The interpreter already handles boxing; stdlib APIs should use interfaces (see §6.1) rather than host specifics.

**Booleans & Nil**
- Map to host-native `bool`/`boolean` and `nil`/`null` while exposing Able semantics through tiny wrapper structs (`AbleBool`, `AbleNil`) so we can attach methods (e.g., `Bool#then`, `Bool#to_string`). Bridging helpers in `able.core.host` convert seamlessly.

**Collections**
- Mutable collections (`Array`, `HashMap`) rely on host-managed buffers (Go slices/maps, JS arrays/Maps) behind opaque handles. The stdlib exposes manipulators while preserving copy-on-write semantics for persistent views.
- Provide `HostArrayHandle`, `HostMapHandle` abstractions in the interpreters with `Clone`, `Drop`, and `len` hooks. Able-level collections never access raw host containers directly; they interact via extern shims exposed in `able.core.host.collections`.

**Guidelines**
- Shared goal: reuse host performance characteristics where they align with Able semantics, but gate access behind well-defined extern helpers so behaviour stays consistent cross-runtime.
- Any direct conversion functions live in an explicit `able.core.host` namespace to keep host-specific dependencies contained.

## 6. Core Types Beyond Collections

- `Option<T>`, `Result<T, E>`, `Either<L, R>`, `Try<T>` alias for `Result<T, Error>`.
- `Ordering` union for comparisons.
- `Range<T>` generic over numeric/Comparable types with inclusive/exclusive endpoints; integrates with `Iterable` and `Sequence`.
- `Duration`, `Instant` inside `able.time`.
- `Error` hierarchy aligning with spec §11/§12, plus extension points for IO/network errors.
- `Pair`, `Triple`, `TupleN` helpers (until we have variadic generics).
- `Lazy<T>` for deferred computation with memoization semantics.

### 6.1 Core Primitive Behaviour Interfaces

To keep stdlib code agnostic of host representations, we define the following interfaces in `able.core.interfaces`:

| Interface | Purpose | Key methods |
| --- | --- | --- |
| `StringLike` | Shared string behaviour | `len_bytes()`, `len_graphemes()`, `is_empty()`, `byte_at(index) -> u8?`, `slice(range) -> Self`, `concat(other: Self) -> Self`, `to_string() -> string` |
| `StringBuffer` | Mutable string builders | `push_str(segment)`, `insert(idx, segment)`, `remove(range)`, `finish() -> string` |
| `Numeric` | Common numeric API | `zero()`, `one()`, `abs()`, `sign()`, `compare(self, other)`, arithmetic ops (via operator interfaces) |
| `IntegerLike` | Integer-specific helpers | `bit_length()`, `checked_add/sub/mul`, `div_mod(divisor) -> (Self, Self)`, `to_bigint()` |
| `FloatLike` | Floating-point behaviour | `is_nan()`, `is_infinite()`, `floor()`, `ceil()`, `round()`, `fract()` |
| `BooleanLike` | Boolean helpers | `and_then(fn)`, `or_else(fn)`, `xor(other)` |
| `CollectionLike<T>` | Shared collection expectations | `len()`, `is_empty()`, `contains(value)`, `iter() -> Iterator<T>`; extends `Iterable<T>` |
| `MutableCollectionLike<T>` | Mutation contract | `clear()`, `reserve(capacity)`, `push(value)`, `extend(iter: Iterator<T>)` |

Notes:
- Concrete types (`String`, `StringBuilder`, `Array<T>`, `Vector<T>`) implement these interfaces so extensions in stdlib can target the interfaces rather than the underlying type.
- Operator interfaces from the spec (`Add`, `Sub`, `Mul`, etc.) cover arithmetic dispatch; `Numeric` ties them together with constants and comparisons.
- Host interop helpers (`String::from_host_string`, `Int::from_host_number`) ensure conversions respect these interfaces (e.g., `from_host_number` returns something implementing `Numeric`).

### 6.2 Numeric & Ordering Hierarchy (Scala-inspired)

We mirror Scala’s `math` hierarchy so built-in numerics share common behaviour while remaining extensible:

| Interface | Extends | Responsibilities |
| --- | --- | --- |
| `Equatable` | – | `equals`, `hash_code` (already defined in §4). |
| `PartialOrdering<T>` | – | `compare_partial(self, other) -> Option<Ordering>`; used for floats (`NaN` cases). |
| `TotalOrdering<T>` | `PartialOrdering<T>` | Guarantees total order; exposes `compare(self, other) -> Ordering`, `min`, `max`, `clamp`. |
| `OrderingOps<T>` | `TotalOrdering<T>` | Default mix-in adding `<=`, `>=`, `between?`, `sort_pair`. |
| `Semigroup<T>` | – | `combine(self, other) -> T`. |
| `Monoid<T>` | `Semigroup<T>` | Adds `empty() -> T`. |
| `Group<T>` | `Monoid<T>` | Adds `inverse(self) -> T`. |
| `AdditiveMonoid<T>` | `Monoid<T>` | Aliases `combine` to `+`, `empty` to `zero`. |
| `AdditiveGroup<T>` | `Group<T>` | Adds `negate`. |
| `MultiplicativeMonoid<T>` | `Monoid<T>` | Provides `one()`, `mul`. |
| `Numeric<T>` | `AdditiveGroup<T>`, `MultiplicativeMonoid<T>`, `TotalOrdering<T>` | Core numeric surface: `from_i64`, `to_i64`, `abs`, `signum`, `pow`, `clamp`, `to_string(radix)`. |
| `Integral<T>` | `Numeric<T>` | Integer-specific: `div_mod`, `quot`, `rem`, `bit_length`, `bit_count`. |
| `Fractional<T>` | `Numeric<T>` | Adds `reciprocal`, `div`, `fract`, `round(mode)`. |
| `Bitwise<T>` | – | `and`, `or`, `xor`, `not`, `shift_left`, `shift_right`. |
| `Signed<T>` | `Numeric<T>` | `is_negative?`, `is_positive?`, `abs`. |
| `Unsigned<T>` | `Numeric<T>` | `leading_zeros`, `trailing_zeros`. |
| `NumericConversions<T>` | – | `to_i32`, `to_u32`, `to_u64`, `to_f64`, `to_bigint`, `to_decimal`. |
| `ToHostNumeric` | – | `to_host_number(runtime) -> host_number`. Bridge for externs. |

**Interface relationships**
- All concrete integers (`i8` … `i128`, `u8` … `u128`) implement: `Integral`, `Bitwise`, `NumericConversions`, plus `Signed` or `Unsigned`.
- Floats (`f32`, `f64`) implement: `Fractional`, `NumericConversions`, `PartialOrdering` (not `TotalOrdering` due to NaN). We provide `TotalOrdering` via wrapper `TotalFloat` that defines NaN ordering (like Scala’s `Ordering.Double.TotalOrdering`).
- `BigInt`, `BigDecimal` (future) adopt same interfaces; persistent numeric types can implement `Numeric`.
- `Bool` implements `Equatable`, `OrderingOps`, `Monoid<bool>` (with `true` as identity for `&&`, or provide separate `AllMonoid`, `AnyMonoid`).

**Collection alignment**
- `Range` depends on `Numeric<T>` + `TotalOrdering<T>` to step generically over numeric types.
- `Vector<T>.sum` requires elements implementing `AdditiveMonoid<T>`.
- `Sequence#sorted` accepts a `TotalOrdering<T>` instance (default provided for builtin comparable types).

**Conversion helpers**
- Provide implicit (or explicit) functions: `numeric_from_literal<T: Numeric>(Literal)`, `integral_literal<T: Integral>`.
- `NumericConversions` ensures stdlib functions (like `math.abs`, `math.sin`) accept any type implementing proper interface, while runtime conversions use `to_host_number`.

**Documentation**: Each built-in type module should state which interfaces it implements (similar to Scala’s scaladoc). We can generate a matrix:

| Type | Interfaces |
| --- | --- |
| `i32` | `Integral`, `Bitwise`, `Signed`, `NumericConversions`, `TotalOrdering`, `AdditiveGroup`, `MultiplicativeMonoid` |
| `u32` | `Integral`, `Bitwise`, `Unsigned`, `NumericConversions`, `TotalOrdering`, `AdditiveMonoid`, `MultiplicativeMonoid` |
| `f64` | `Fractional`, `NumericConversions`, `PartialOrdering`, `AdditiveGroup`, `MultiplicativeMonoid` |
| `bool` | `Equatable`, `OrderingOps`, `AnyMonoid`, `AllMonoid` (implemented via newtype wrappers) |
| `string` | `Sequence<char>`, `TotalOrdering`, `StringLike` |

We’ll codify this matrix in the stdlib docs once implementations exist.

### 6.3 Compatibility with Future Numerics Libraries

To remain flexible for an eventual Able numerics ecosystem (akin to Typelevel Spire or Go’s Gonum), our interfaces must satisfy the following principles:

- **Lawfulness & Extensibility**: Each algebraic interface (`Semigroup`, `Monoid`, `Group`, `Ring`, `Field`) must document required laws (associativity, identities, distributivity). Future numerics libraries can depend on these laws for optimizations. We should consider adding `Ring<T>` (`AdditiveGroup` + `MultiplicativeMonoid`) and `Field<T>` (`Ring` + multiplicative inverses) once fractional types settle.
- **Hierarchy Stability**: Keep the inheritance graph minimal and acyclic beyond algebraic specialization, so third-party types (matrices, complex numbers, rationals) can implement interfaces without ambiguous diamonds. Avoid tying numeric interfaces to concrete host types.
- **Type Class Style**: Allow implementations either on the type itself or via orphan-able `impl`s so packages can retrofit interfaces onto externally defined numerics (mirroring Spire’s type classes). Ensure coherence rules from the spec still hold.
- **Precision Awareness**: Provide optional capability markers like `HasExactEquals`, `HasExactZero`, `HasInfinity` so libraries can branch on numeric capabilities (Spire differentiates exact vs approximate fields).
- **Conversion Strategy**: `NumericConversions` must be loss-aware. We should add methods returning `Result` (`try_to_i32`) to signal overflow, enabling high-precision libraries to avoid silent truncation.
- **Generic Algorithms**: Core stdlib algorithms (e.g., statistics, interpolation) should target interfaces, not concrete types, to stay compatible with future advanced numerics.
- **Host Independence**: Bridges (`ToHostNumeric`) must remain optional; high-performance numeric libraries may implement values entirely in Able without host interop. Interfaces must not assume the presence of a host primitive.

We will revisit these interfaces when designing the dedicated `able.math.numeric` package, ensuring our base contracts align with Spire/Gonum concepts like `VectorSpace`, `NormedVectorSpace`, `MetricSpace`, etc.

### 6.4 Regular Expression Library Vision (Exploratory)

**Goals**
- Provide a high-level API suitable for day-to-day string matching (akin to Ruby’s `Regexp`, JavaScript’s `RegExp`), while exposing low-level automata representations (NFA/DFA/AST) for advanced analysis and manual tweaking.
- Support multi-pattern matching and streaming search similar to Rust’s `regex` crate (`RegexSet`, `RegexSetBuilder`, `regex-automata`).
- Offer predictable performance and avoid catastrophic backtracking (prefer RE2-style guarantees).
- Enable compilation to reusable bytecode/automata objects for repeated matching across threads/procs.

**Feature Set Inspirations**
- **Rust `regex`/`regex-automata`**: deterministic finite automata construction, multi-pattern, lazy DFAs, support for `RegexSet`, ability to export automata as bytecode or graph structures.
- **Oniguruma** (Ruby/Crystal): rich syntax (named captures, lookbehind, Unicode scripting), but backtracking-based. Evaluate which features we can support while staying efficient.
- **RE2** (Go): guarantees linear-time matching via automata; strong candidate for backend semantics. No backreferences, look-around limitations.
- **PCRE2**: extremely feature rich but allows catastrophic backtracking; likely not desirable as default semantics, but we can consider optional compatibility.
- **Hyperscan**: vectorized multi-pattern matching; advanced but heavy dependency.

**Proposed API Layers**
1. **`Regex` high-level interface**
   - `Regex.compile(pattern, options) -> Regex`
   - Methods: `match?(string)`, `match(string) -> MatchData?`, `scan(string) -> Iterator<MatchData>`, `replace(string, repl)`, `split(string)`
   - Options for case folding, multiline, Unicode classes, atomic groups (subject to backend support).
2. **`RegexSet` / `MultiRegex`**
   - Compile multiple patterns simultaneously; returns indices of matched patterns (`RegexSet.match_indices(string)`).
   - Should share underlying automata state for efficiency.
3. **Automata Access**
   - `Regex.to_nfa() -> RegexNFA`, `Regex.to_dfa() -> RegexDFA` (where possible; backtracking-only features disable DFA export).
   - `RegexNFA` / `RegexDFA` provide introspection (`states`, `transitions`, `accepting_states`) and serialization.
4. **Builder & Analysis Tools**
   - `RegexBuilder` allows programmatic construction/composition (union, intersection, complement).
   - Provide transformations: `RegexDFA.minimize`, `derive(residue)`, `to_graphviz()`.
5. **Streaming & Lazy Matching**
   - Support incremental search over streams (feeding chunks, resuming state).
   - Provide `RegexScanner` state machines to maintain context between calls.

**Implementation Considerations**
- **Backend choices**:
  - Start with a safe, linear-time engine (RE2-style). Evaluate reusing existing backends via externs (Go: `regexp`, TypeScript: integrate with [rust regex via WASM?], or implement in Able).
  - Provide toggleable compatibility layer for backtracking features, but keep linear engine default.
- **Automata storage**:
  - Represent NFA/DFA nodes with persistent data structures to enable analysis.
  - Provide ability to export to host representations for integration with specialized libraries (Hyperscan, Oniguruma).
- **Unicode support**:
  - Full Unicode (grapheme clusters, script properties) is desirable. Need to confirm runtime ability; may start with basic categories and expand.
- **Extensibility**:
  - Design interfaces so alternative engines (backtracking, derivative-based) can plug in. For example, `RegexEngine` trait with `compile`, `match`, `find_iter`, `to_automata`.

**Integration with Core Interfaces**
- `StringLike` compatibility: regex APIs should accept any `StringLike` to ensure future text types (e.g., rope, rope slices) can be matched without copying.
- `Iterator` integration: `Regex#captures_iter` returns lazy iterators; `RegexSet#matches_iter` returns match results lazily.
- `Collectable`: allow collecting matches into arbitrary collections via `Iterator` semantics.

**Open Questions**
1. Do we guarantee RE2-style linear behaviour (disallowing backreferences) in the canonical engine, or offer optional modules for richer but unsafe features?
2. How do we expose automata modifications? Do we provide builder APIs (`RegexDFA#with_transition(state, char, target)`) or expect users to reconstruct from AST?
3. Should `Regex` compilation be pure Able code (interpreter-friendly) or rely on host-specific regex engines via extern wrappers initially?
4. What serialization format should we use for automata export/import (JSON, custom binary)?
5. How do we version regex syntax/features relative to stdlib versions to avoid breaking compatibility?

We should create a dedicated design note once we start implementing `able.text.regex`, incorporating survey findings from the resources above.

## 7. Collections Inventory

### 7.1 Immutable (Persistent) Collections

| Type | Backing structure | Key operations | Performance |
| --- | --- | --- | --- |
| `Vector<T>` | Relaxed radix balanced tree (RRB) | `push`, `pop`, `update`, `slice`, `concat` | `O(1)` amortized push/pop tail, `O(log n)` index/update, `O(log n)` concat |
| `List<T>` | Cons cell linked list | `cons`, `head`, `tail`, `map`, `fold` | `O(1)` cons/head/tail, `O(n)` index |
| `Set<T>` | Hash array mapped trie (HAMT) | `insert`, `remove`, `contains`, `union`, `intersect` | `O(1)` average for membership, structural sharing |
| `Map<K,V>` | Ordered HAMT with bitmap nodes | `get`, `set`, `remove`, `merge`, `map_values` | `O(1)` average lookup/update, `O(log n)` worst |
| `SortedSet<T>` | Finger tree / red-black tree | `first`, `last`, `range`, `to_array` | `O(log n)` membership, `O(log n)` range ops |
| `Queue<T>` | Bankers deque (two lists) | `enqueue`, `dequeue`, `peek` | Amortized `O(1)` enqueue/dequeue |
| `LazySeq<T>` | Iterator wrapper with caching | `map`, `filter`, `take`, `cycle` | Lazy `O(1)` per step, caches optionally |

Builders: Provide `Vector.builder()`, `Set.builder()` returning mutable accumulators that flush to persistent structure (`Builder` interface).

### 7.2 Mutable Collections

| Type | Backing structure | Key operations | Performance guarantees |
| --- | --- | --- | --- |
| `Array<T>` | Contiguous resizable buffer | `push`, `pop`, `insert`, `remove`, `reserve` | `O(1)` amortized push/pop, `O(n)` insert/remove mid |
| `Deque<T>` | Ring buffer | `push_front`, `push_back`, `pop_front`, `pop_back` | `O(1)` amortized |
| `LinkedList<T>` | Doubly linked nodes | `push_front`, `insert_after`, `remove` | `O(1)` manipulations given node handle |
| `HashMap<K,V>` | Open-addressing + robin hood | `get`, `set`, `remove`, `iter_pairs` | `O(1)` average lookup/update |
| `HashSet<T>` | HashMap wrapper | `insert`, `remove`, `contains` | `O(1)` average |
| `TreeMap<K,V>` | Red-black tree | `floor`, `ceil`, `range`, `keys_sorted` | `O(log n)` operations |
| `TreeSet<T>` | Red-black tree | `insert`, `remove`, `lower_bound` | `O(log n)` |
| `Heap<T>` | Binary heap | `push`, `pop`, `peek`, `replace` | `O(log n)` |
| `BitSet` | Dynamic bitset | `set`, `reset`, `flip`, `contains` | `O(1)` word ops |
| `SmallVec<T, const N>` | Inline small vector | `push`, `pop` with inline storage | `O(1)` amortized |

Mutable collections implement `MutableCollection`, `Indexable`, optional `Sliceable` (e.g., `Array`, `Deque`). Provide conversion helpers to persistent counterparts (`Array#to_vector`, `HashMap#to_persistent_map`).

### 7.3 Iterator & Collectable Utilities

- `Iterator` adapters: `MapIterator`, `FilterIterator`, `Enumerate`, `Zip`, `Chain`, `TakeWhile`, `DropWhile`, `Flatten`, `Chunk`, `Window`, `Partition`, `Sliding`.
- `Iterator#lazy` returns self (Crystal-style) but signals intent; pipelines remain lazy until `to_*` / `collect`.
- `Collectable` implementations (`ArrayCollectable`, `VectorCollectable`, `SetCollectable`, `MapCollectable`) materialize iterators without coupling to concrete constructors.
- `Enumerator` bridge yields elements through Able `proc` while respecting scheduler semantics.

## 8. Common Operations & Naming Conventions

- `push`, `pop`, `push_front`, `push_back`, `shift`, `unshift` for sequences.
- `append`, `prepend`, `concat`, `join` unify string and collection operations.
- `map`, `flat_map`, `filter`, `reject`, `reduce`, `fold_left`, `fold_right`, `collect`, `group_by`, `partition`, `zip`, `zip_with_index`.
- `map_iter`, `filter_iter`, `flat_map_iter`, `zip_iter` expose lazy variants straight from collections (Crystal-inspired `each` pipelines).
- `first`, `last`, `nth`, `take`, `drop`, `split_at`, `chunks`, `windows`.
- `any?`, `all?`, `none?`, `one?` boolean combinators with lazy short-circuit semantics on iterators.
- `sum`, `product`, `min`, `max` for numeric sequences; prefer lazy reduction for iterators.
- `each_proc` yields each element through Able `proc` to integrate with concurrency testing.
- All collections implement `Display`, `Debug`, `Clone`, `Equatable`, `Hashable` (where element types allow).
- Methods ending with `?` return `bool`; methods raising exceptions provide `!` suffix (`fetch!`).
- Provide both total (`get`, returning `Option`) and partial (`get!`, raising `IndexError`) accessors.

## 9. Performance Contracts & Documentation

- Each collection module includes a performance table (Big-O, amortized notes, iterator invalidation rules).
- Document mutation semantics (copy-on-write vs persistent sharing).
- Provide explicit guidance on scheduler interaction: long-running `Sequence#each` must call `proc_yield` periodically when iterating large data sets in interpreter mode.
- Encourage deterministic iteration order for hash-based collections by default (seeded hasher) with opt-in randomization later.

## 10. Runtime & Host Integration

- Go runtime backs mutable collections with native slices/maps; persistent structures implemented in Able for now, optimized later via FFI once stable.
- TypeScript runtime uses host arrays/maps but enforces same API and performance expectations; consider bundling persistent data structure library written in Able for parity.
- Provide extern shims for heavy operations (e.g., `String#encode_utf8`, `HashMap#rehash`) to keep interpreters performant.
- Ensure `Serializable` implementations cooperate with runtime-specific IO backends (Go’s JSON, TypeScript’s JSON).
- Channel/Mutex externs from `design/channels-mutexes.md` integrate with `able.concurrent` module; provide helper collections like `ConcurrentQueue` using those primitives.

## 11. Implementation Phases

1. **Foundations**
   - Finalize interfaces in `able.core.interfaces`, mirroring Crystal’s module layering (`Iterator(T)`, `Enumerable(T)`, `Indexable(T)` equivalents).
   - Flesh out `Option`, `Result`, `Range`, and error structs.
   - Implement `String`, `StringBuilder`, `Array`, `HashMap` using host externs.
   - Add basic iterator adapters (`map`, `filter`, `collect`) plus eager/lazy method naming conventions.
2. **Persistent Collections & Builders**
   - Implement `Vector`, `List`, `Set`, `Map` persistent variants.
   - Provide builder APIs, `Collectable` instances, and conversions between mutable/immutable forms.
   - Document performance tables and add fixtures covering iterator pipelines vs eager paths.
3. **Advanced Collections & Utilities**
   - Sorted structures (`TreeMap`, `TreeSet`, `Heap`, `BitSet`).
   - Lazy sequences, streaming IO adapters.
   - Concurrency-aware collections (`ConcurrentQueue`, `AsyncStream`).
   - Serialization protocols and time/IO modules.
4. **Polish & Tooling**
   - Generate API reference from Able doc comments.
   - Provide cookbook examples; ensure `run_all_tests.sh` covers new modules.
   - Update spec §14 with stabilized interfaces and mark completed work in `PLAN.md`.

## 12. Brainstorm Notes

- Persistent collections can share backing nodes across runtimes if we model them in Able; explore writing them once in Able and compiling to both Go/TS via interpreters.
- Consider `Enumerator` objects like Ruby to bridge synchronous iteration with `proc`-based concurrency.
- Multi-arity `map`/`zip` similar to Scala for processing multiple sequences simultaneously.
- Provide structural pattern matching helpers (e.g., `Vector::Unapply` interfaces) once parser supports the sugar.
- Offer `Lens`-style utilities for immutable updates (inspiration from Scala/Functional Java).
- Evaluate optional `serde`-like derive macros once macro system exists; for now, design interfaces with future derivation in mind.

### 12.1 Allocation-Free Transformation Strategies (Exploratory Only)

**Context**: To minimize intermediate allocations while transforming collections, we can take several approaches. Each option integrates differently with Able’s collection interfaces (`Iterator`, `Collectable`, `Builder`, `Collection`) and with host runtimes.

#### Option A — Iterator Fusion (Chain Collapsing) *(Preferred for v10)*
- Leverage our lazy `Iterator` pipelines but fuse successive adapters into a single iterator object at runtime (Crystal’s optimizer, Rust’s `Iterator` trait common patterns).
- **Implementation sketch**: When chaining `iter.map(...).filter(...)`, the runtime builds a fused iterator struct capturing both closures; allocation is limited to this single iterator handle.
- **Interfaces needed**:
  - `Iterator` must expose combinators (`map`, `filter`, `flat_map`) returning specialized fused iterator types.
  - `Collectable<C>` remains the realization surface; no new protocol.
  - Optional `FusableIterator` marker to help the runtime collapse adapters.
- **Pros**: Low conceptual overhead; aligns with our current plan; works for both mutable and persistent collections by reusing `iter`.
- **Cons**: Still allocates the fused iterator object; transformations remain pull-based, so fabricating pipelines for multiple consumers may clone iterators.
- **Fit**: Eager collection methods simply call into `iter`, build fused iterator, then `collect_into(self)` to produce same-type result.

#### Option B — Stateful Builders (Reusable Accumulators)
- Provide mutable builder objects (`ArrayBuilder`, `VectorBuilder`) that accept transformation closures and push results directly into the target buffer without intermediate iterators.
- **Implementation sketch**:
  ```able
  builder := ArrayBuilder(i32).with_capacity(array.len)
  array.each_builder(builder) { |value, emit| new_val := value + 1; emit.call(new_val) }
  result := builder.finish
  ```
- **Interfaces needed**:
  - `CollectionBuilder<T, C>`: `emit(value: T) -> void`, `finish() -> C`.
  - Collections add `each_builder(builder, transform)` hooking into their internal traversal.
  - Iterators can provide `into_builder(builder)` to stream lazily.
- **Pros**: Eliminates intermediate iterator allocations entirely; works well for persistent collections by constructing nodes directly.
- **Cons**: API surface gets larger; requires every collection to supply optimized traversal hooks; more error-prone (must ensure builders honour cancellation, errors).
- **Fit**: Mutable collections can reuse their in-place mutation logic; persistent collections implement builder in terms of structural sharing. Works nicely with `Collectable` (builder is the concrete implementation).

#### Option C — Clojure-style Transducers (Push-based, Destination-Agnostic) *(Preferred for follow-up prototype)*
- Introduce `Transducer<In, Out>` modules composing reusable transformations independent of source/destination.
- **Implementation sketch**:
  ```able
  module Transducer(In, Out)
    fn init(self) -> State
    fn step(self, state: State, input: In, emit: fn(Out)) -> StateStep
    fn complete(self, state: State, emit: fn(Out)) -> void
  end

  pipeline := Transducers.map(fn(x) { x + 1 }).then(Transducers.filter(&.even?))
  result := pipeline.into(array, collectable: VectorCollectable(i32))
  ```
- **Interfaces needed**:
  - `Transducer` protocol plus combinators (`map`, `filter`, `take`, etc.).
  - `Reducible` trait for sources supporting `reduce(transducer, collectable)`.
  - Extend `Collectable` with `emit` callback signature to be used by transducers.
- **Pros**: True allocation-free composition; target-agnostic (streams, channels, collections share pipelines); easy to reuse on concurrency primitives.
- **Cons**: Highest complexity; requires new abstractions; more difficult to reason about step state; must integrate with scheduler yield semantics carefully.
- **Fit**: Works with both mutable and persistent collections; persistent structures can build nodes on the fly during reduction.

#### Option D — In-place Mutation Pipelines
- For mutable collections (`Array`, `Deque`), provide transforms that mutate in place while iterating, eliminating extra allocations (e.g., `Array#transform!`).
- **Implementation sketch**:
  ```able
  array.transform!(map: fn(x) { x + 1 }, select: fn(x) { x.even? })
  ```
- **Interfaces needed**:
  - `MutableCollection` extension: `transform!(ops: TransformSpec<T>) -> Self`.
  - `TransformSpec` is a plain struct bundling optional closures.
- **Pros**: Zero allocations; keeps data local; great for performance-critical paths.
- **Cons**: Only applicable to mutable collections; semantics must clearly state evaluation order and error handling; less functional/pure.
- **Fit**: Complements other strategies; persistent collections can use builder approach to achieve similar effect without mutation.

#### Option E — Compiler-guided Fusion (Future)
- Provide annotations (`@inline_pipeline`) hinting to the compiler/interpreter optimizers to fuse multiple eager operations into one pass without exposing new runtime APIs.
- **Interfaces needed**: None new at runtime; relies on compiler metadata.
- **Pros**: Minimal user-facing changes.
- **Cons**: Requires advanced optimization infrastructure; hard to ensure same behaviour across Go/TS interpreters; likely out of scope until we have a static compiler.

**Reconciliation with Collection Interfaces**
- Regardless of option, we need clear contracts:
  - `Iterator` remains the universal lazy interface.
  - `Collectable<C>` (or `CollectionBuilder`) realizes sequences into concrete containers.
  - `Collection` should expose both eager (`map`, returning `Self`) and lazy (`iter`, `map_iter`) forms.
  - For Options B/C, introduce `Reducible` (sources) and `Transducible`/`Builder` protocols for destinations.
  - Ensure cancellation (`proc_yield`) hooks exist in each path (iterators yield per element; builders/transducers must yield periodically).

**Next questions**
1. Which subset (A+B vs A+C) delivers the most value for the first stdlib release?
2. Do we formalize `CollectionBuilder` as the standard `Collectable` implementation (enabling both Options A and B)?
3. How do we document the performance guarantees so users understand when allocations happen?

## 13. Open Decisions (Need Product Guidance)

1. **Default `Map` alias**: Should `import able.collections.Map` resolve to the persistent HAMT or the mutable HashMap?
2. **String slicing semantics**: Do slices operate on bytes (Go-style) or grapheme clusters (Ruby-style) by default? Proposal: default to byte indices with safe helpers for graphemes; confirm.
3. **Persistent builder strategy**: Prefer dedicated builder objects or reuse mutable counterparts (`Array` -> `Vector`)?
4. **Iterator cancellation**: Require every long-lived iterator to call `proc_yield` automatically after N steps? Need specification.
5. **Hash seeding**: Keep deterministic seed (stable across runs) vs randomized seeds for DoS resistance?
6. **Error hierarchy exposure**: Should user-defined errors inherit from a shared `Error` struct or implement an `ErrorLike` interface?
7. **Lazy naming clarity**: Are `_iter` suffixes on collection methods acceptable, or should we rely exclusively on `collection.iter.map` Crystal-style?

Please review these points; decisions will unblock detailed specs and implementation tickets.

---

This document complements `design/stdlib-v10.md` (implementation notes). Update both when decisions land so future contributors have a consistent picture.
