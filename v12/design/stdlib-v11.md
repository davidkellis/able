# Able v11 Standard Library Plan

Last updated: 2025-11-01

This document consolidates the earlier “stdlib notes” and “stdlib vision” drafts into a single plan that stays aligned with `spec/full_spec_v11.md`. The specification remains authoritative; everything below either satisfies an explicit spec requirement or is flagged as future work that must never contradict the spec.

## 1. Purpose & Scope

- Define the canonical Able v11 standard library packaged under `stdlib/`.
- Capture a cohesive vision that aligns with the language specification and ongoing interpreter work.
- Provide a target inventory of packages, protocols, and core data types that every runtime must support.
- Ensure spec-described syntax (`[]`, `..`, `spawn`, etc.) has a concrete stdlib implementation.
- Offer a reference for Go and TypeScript runtime contributors so both implementations converge on identical semantics.
- Document design inspirations (Scala, Crystal, Ruby, Rust, Go) and how their strengths inform Able.
- Record open decisions so we can converge quickly as implementation work begins.

### 1.1 Guiding Principles

**Spec Alignment**
- Treat `spec/full_spec_v11.md` as authoritative; stdlib features must extend, not contradict, the spec.
- Respect the AST contract — stdlib types must be representable without host-specific leakage.
- Keep observable behaviour identical across the Go and TypeScript interpreters.

**Ergonomics & Familiarity**
- Bias toward Crystal’s standard library structure: clear separation of eager `Collection` APIs and lazy `Iterator` pipelines, explicit module mixins, and predictable return types.
- Offer Scala-inspired parity between immutable and mutable collections with identical method names (`push`, `pop`, `map`, `flat_map`, etc.).
- Adopt Crystal/Go-inspired `String` primitives: UTF-8 by default, grapheme-aware (`char`) iteration by default with explicit helpers for byte-level access, and efficient slicing via views.

**Performance Contracts**
- Publish Big-O guarantees for core operations on each collection.
- Provide predictable allocation behaviour via `reserve`, `with_capacity`, and builder APIs.
- Default to persistent (immutable) data in pure contexts, with mutable variants available when callers opt in.

**Runtime & Concurrency Integration**
- Collections and protocols must cooperate with Able’s cooperative scheduler and Go’s goroutines.
- Respect cancellation semantics for iterators and async constructs (e.g., `Sequence#each_future` must yield).

## 2. Alignment With the v11 Specification

| Spec section | Requirement | Stdlib obligation |
| --- | --- | --- |
| 6.8 Arrays | Mutable `Array T` with literals, indexing, `size() -> u64`, `get`, `set`, `slice`, `push`, `pop`; indexing raises `IndexError`. | `able.collections.mutable.array` must provide these APIs, raising the right errors and exposing `Index`/`IndexMut`. |
| 6.10 Dynamic metaprogramming | Host helpers drive dynamic packages. | Expose bridge modules under `able.core.host` without diverging semantics. |
| 11.2 Option/Result | `Option T = nil | T`, `Result T = Error | T`, `!` propagation helpers. | `able.core.option_result` supplies unions and helper methods consistent with the spec. |
| 11.3 Errors | `DivisionByZeroError`, `OverflowError`, `ShiftOutOfRangeError`, `IndexError`, `FutureError` (plus message contracts). | Core error structs live in `able.core.errors`; runtimes raise them as described. |
| 12.2 Future | `Future T` interface (`status`, `value`, `cancel`) and `FutureStatus` union. | `able.concurrent.future` defines the interface/structs and extern hooks per spec semantics. |
| 12.3 Future | Transparent, memoised evaluation of `Future T` on demand. | `able.concurrent.future` wraps host handles and enforces implicit blocking semantics. |
| 12.5 Synchronisation | `Channel T` API (`new`, `send`, `receive`, `try_*`, `close`, `is_closed`) and `Mutex` with `lock`/`unlock`/`with_lock`. Errors: `ClosedChannelError`, `SendOnClosedChannelError`, `NilChannelError`. | `able.concurrent.channel` and `.mutex` provide these types, forwarding to native helpers with the exact spec names/behaviour. |
| 13 Imports | Package layout dictates module paths (hyphen → underscore). | Standard library directory structure and manifest must respect the loader rules. |
| 14 Core interfaces | `Iterator`, `Iterable`, `Range`, operator traits (`Add`, `Sub`, etc.), comparison (`PartialEq`, `Eq`, `PartialOrd`, `Ord`), `Display`, `Clone`, `Default`, `Hash`, `Apply`, `Index`, `IndexMut`. | `able.core.interfaces` exports these definitions verbatim, including default method bodies where the spec provides them (e.g., `Iterable.iterator`). See https://itsfoxstudio.substack.com/p/comparison-traits-understanding-equality for reference. |
| 16 Tooling | Canonical type mapping, error mapping. | Stdlib docs track these correspondences so host runtimes stay in sync. |

Everything else in this document is layered on top of that baseline and must never invalidate an item in the table.

## 3. Package Layout & Module Tiers

| Tier | Package path | Summary |
| --- | --- | --- |
| Prelude | `able` | Thin façade automatically imported by tooling; re-exports Option/Result helpers, core interfaces, and commonly used aliases. |
| Core | `able.core.*` | Spec-mandated interfaces, Option/Result, error types, range helpers, tuple utilities, host bridges. |
| Concurrency | `able.concurrent.*` | `Future`, `Channel`, `Mutex`, scheduler helpers, concurrency collections. |
| Collections (mutable) | `able.collections.mutable.*` | Array, Deque, LinkedList, HashMap, HashSet, TreeMap, TreeSet, Heap, BitSet, SmallVec. |
| Collections (persistent) | `able.collections.immutable.*` | Persistent Vector, List, Set, Map, SortedSet, LazySeq, Queue with structural sharing. |
| Text | `able.text.*` | String, SubString, builders, Unicode helpers, ByteString. |
| Numerics | `able.math.*` | Extended numeric helpers, algebraic interfaces, random utilities, BigInt/Decimal. |
| Functional | `able.fn.*` | Higher-order utilities, partial application helpers, memoisation, functional optics. |
| IO & Time | `able.io.*`, `able.time.*` | Streams, filesystem/network shims (host-provided), `Duration`, `Instant`. |

## 4. Baseline Deliverables (Spec Parity)

### 4.1 Core Interfaces (able.core.interfaces)

- Mirror spec names exactly: `Display`, `Clone`, `Default`, `Hash`, `Hasher`, `PartialEq`, `Eq`, `PartialOrd`, `Ord`, `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg`, `Not`, `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`, `Apply`, `Index`, `IndexMut`.
- `Iterator T` returns `T | IteratorEnd` from `next`; `IteratorEnd` is the singleton sentinel described in the spec.
- `Iterable T` provides default implementations:
  - `each` defined in terms of `iterator`.
  - `iterator` defined in terms of `each` using the generator literal showcased in Section 6.8 of the spec.
- `Range Start End Out` exposes `inclusive_range` and `exclusive_range`; any type that uses `..`/`...` must provide implementations.

### 4.2 Core Data Types

- `Option T` and `Result T` unions ship with helper methods (`unwrap`, `map`, `ok_or`, etc.) that follow the semantics in Section 11.2, raising `OptionUnwrapError` / `ResultUnwrapError` which implement `Error`.
- `Array T` must provide runtime backing for `size() -> u64`, `capacity()`, `is_empty()`, `push`, `pop -> ?T`, `get -> ?T`, `set -> !nil`, `slice -> Array T`. `arr[i]` / `arr[i] = value` raise `IndexError` on out-of-bounds.
- `Map K V` exists as an interface; `HashMap K V` is the first concrete implementation. Minimal surface: `new`, `get`, `set`, `remove`, `contains`, `size`, `is_empty`, obeying `Hash` + `PartialEq`.
- `Range` helpers at least cover integer stepping for inclusive/exclusive operators; future work generalises via numeric interfaces.
- Spec-mandated error types (`DivisionByZeroError`, `OverflowError`, `ShiftOutOfRangeError`, `IndexError`) and channel errors (`ClosedChannelError`, `SendOnClosedChannelError`, `NilChannelError`) must be present.
- All user-defined errors conform to the `Error` interface; no parallel hierarchy is introduced.

### 4.3 Concurrency Surface

- `FutureStatus` union (`Pending`, `Resolved`, `Cancelled`, `Failed { error: FutureError }`) and `FutureError` struct exactly match Section 12.2.
- `Future T` implements `status`, `value`, and `cancel` with the blocking semantics described in the spec.
- `Future T` behaves transparently: evaluating it in a `T` context blocks until completion and memoises the result (Section 12.3). No explicit `poll` API is required at the Able level.
- `Channel T` and `Mutex` wrappers expose the precise methods spelled out in Section 12.5, forwarding to runtime helpers while preserving Go-compatible semantics. Channel iteration blocks until closed and drained.

### 4.4 Distribution & Loader Expectations

- `package.yml` declares `name: able`; directory names map to package paths using the loader rules (hyphen → underscore).
- `src/lib.able` centralises exports without aggressive wildcard re-exports until module contents stabilise.
- Tooling (Go CLI, Bun harness) automatically injects the stdlib path so `import able.core.interfaces` works without manual configuration.

### 4.5 Testing & Fixtures

- Shared fixtures under `fixtures/ast` must cover every spec-mandated stdlib surface (Array operations, channel send/receive, Future semantics).
- Runtimes continue running `bun run scripts/run-fixtures.ts` and `go test ./pkg/interpreter`.
- Whenever stdlib behaviour changes, update `spec/todo.md` if wording adjustments are required and extend both interpreters’ test suites.

### 4.6 Kernel vs stdlib split (current → target)

**Current kernel (native, baked into interpreters)**
- Scheduler primitives: `future_yield`, `future_cancelled`, `future_flush`, `future_pending_tasks`.
- Concurrency bridges: channel helpers (`__able_channel_new/send/receive/try_send/try_receive/await_try_send/await_try_recv/close/is_closed`), mutex helpers (`__able_mutex_new/lock/unlock`), await wakers.
- String bridges: `__able_string_from_builtin`, `__able_string_to_builtin`, `__able_char_from_codepoint`.
- **Native helpers that should migrate to stdlib:** array methods (`size`, `push`, `pop`, `get`, `set`, `clear`, `iterator`) and string helpers (`len_*`, `substring`, `split`, `replace`, `starts_with`, `ends_with`, `chars`/`graphemes` iterators) are currently implemented as host natives inside both interpreters.

**Target minimal kernel (host-only)**
- Keep only scheduler primitives, channel/mutex bridges, string/char encoding bridges, and low-level array buffer management hooks needed by the stdlib `Array` implementation (`new`, `with_capacity`, slot read/write).
- Move user-facing array/string helpers into Able code in `stdlib/` (`able.collections.mutable.array`, `able.text.string`) that call the minimal bridges. Runtimes should no longer expose overlapping native methods once the Able implementations are in place.

**Actions**
- Document the minimal kernel contract in the spec and keep this section updated.
- Port array/string helpers to Able stdlib and adjust interpreters/typecheckers/fixtures to rely on the stdlib surface.
- Remove or shim redundant native array/string methods after the stdlib layer owns the behaviour; keep parity harness green during the transition.
- Plan: move "always available" interfaces (`Eq`, `Hash`, `Display`, etc.) and
  their primitive impls into the kernel library so they load before the stdlib;
  `able.core.interfaces` will re-export the kernel definitions. See
  `v11/design/kernel-interfaces-hash-eq.md` for the detailed plan.

## 5. Extended Roadmap (Vision, Spec-Consistent)

These items expand the stdlib once the baseline is green. They reuse material from the earlier vision draft and remain compatible with the spec.

### 5.1 Collections Roadmap

#### 5.1.1 Interface Expansion Targets

| Interface | Purpose | Required methods |
| --- | --- | --- |
| `Equatable` | Value equality | `equals(self, other: Self) -> bool` |
| `Hashable` | Hash participation | `hash(self, hasher: Hasher) -> void` (requires `Equatable`) |
| `Comparable` | Total ordering | `compare(self, other: Self) -> Ordering` (`Ordering` union: `Less`, `Equal`, `Greater`) |
| `Debug` | Developer formatting | `inspect(self: Self) -> string` (optional formatting params later) |
| `IteratorPipeline` | Default lazy adapters | `map`, `filter`, `flat_map`, `zip`, `chunk`, `tap`, returning `Iterator` |
| `Enumerable T` | Each-driven convenience mixin | Requires `each(self: Self, visit: T -> void) -> void`; supplies Crystal-style helpers when mixed in |
| `Collection` | Sized aggregate with eager ops | `len`, `is_empty`, `contains`, `count`, `each`, eager `map` returning `Self` |
| `MutableCollection` | In-place mutation | `clear`, `reserve`, `retain` (extends `Collection`) |
| `Indexable` | Random access | `get`, `get_mut?`, `set`, plus safe optional variants |
| `Sliceable` | Views | `slice(range) -> Self`, `split_at(idx) -> (Self, Self)` |
| `Collectable C` | Realise iterators | `collect(source: Iterator Item) -> C` |
| `Builder C` | Efficient construction | `push`, `extend`, `finish -> C` |
| `Serializable` | Structured persistence | `serialize`, `deserialize` (phase 2 goal) |
| `Callable` | Callable values | Aligns with `Apply`; used for passing callables/closures |

Notes:
- Interfaces live in `able.core.interfaces` with Crystal-style module naming (`Iterator T`, `Iterable T`, `Collection T`).
- Provide blanket implementations where practical (e.g., any `Iterator` mixes in `IteratorPipeline`).
- `Hasher` exposes a streaming API built on FNV-1a with a runtime-chosen random seed (per process) to balance reproducibility and DoS resistance; allow future alternatives.
- Persistent builders follow Scala/Clojure precedent: dedicated builder objects maintain internal node structures compatible with the persistent representation and flush to the final value without routing through general-purpose mutable collections.

#### 5.1.2 Lazy vs Eager Defaults (Crystal Alignment)

- Any `Collection` exposes `each` that yields an `Iterator`. Chaining off `each` keeps operations lazy until a realisation method runs.
- Direct collection methods (`array.map`, `set.filter`) are eager and return a collection of the same type by default.
- Lazy chaining is expressed through `iterator()` (returning an `Iterator`) or `each` via `Enumerable`; convenience methods live on those mixins instead of `_iter` suffixed collection methods.
- Cross-collection conversions use `collect_into(TargetType)` or `to_array`, `to_vector`, etc., to keep results explicit.
- `Iterator#to Type` and `Iterator#collect` are the canonical realisation points.
- Persistent collections reuse iterator combinators internally; more advanced transducer support is deferred.

#### 5.1.3 Pipeline Realization Hooks

- Keep the lazy story focused on `Iterator` pipelines; no first-class transducers in the baseline.
- `Iterator#collect_into(collectable: Collectable)` materialises the sequence without exposing transducer primitives.
- Persistent collections reuse iterator combinators internally; revisit transducers once primitives settle.

#### 5.1.4 Collections Inventory

**Immutable (Persistent) Collections**

| Type | Backing structure | Key operations | Performance |
| --- | --- | --- | --- |
| `Vector T` | Relaxed radix balanced tree (RRB) | `push`, `pop`, `update`, `slice`, `concat` | Amortised `O(1)` tail push/pop, `O(log n)` index/update, `O(log n)` concat |
| `List T` | Cons cell linked list | `cons`, `head`, `tail`, `map`, `fold` | `O(1)` cons/head/tail, `O(n)` index |
| `Set T` | Hash array mapped trie (HAMT) | `insert`, `remove`, `contains`, `union`, `intersect` | `O(1)` average membership, structural sharing |
| `Map K V` | Ordered HAMT with bitmap nodes | `get`, `set`, `remove`, `merge`, `map_values` | `O(1)` average lookup/update, `O(log n)` worst |
| `SortedSet T` | Finger tree or red-black tree | `first`, `last`, `range`, `to_array` | `O(log n)` membership, `O(log n)` range ops |
| `Queue T` | Bankers deque (two lists) | `enqueue`, `dequeue`, `peek` | Amortised `O(1)` enqueue/dequeue |
| `LazySeq T` | Iterator wrapper with caching | `map`, `filter`, `take`, `cycle` | Lazy `O(1)` per step, optional caching |

Builders expose `Vector.builder()`, `Set.builder()`, etc., returning dedicated accumulators that mirror the persistent layout (RRB nodes, HAMT bitmaps). They accumulate changes structurally, then flush directly to the persistent representation—similar to Scala’s `VectorBuilder` or Clojure’s transient workflow—without detouring through generic mutable collections.

**Vector implementation notes (landed)**
The v11 `Vector T` now follows the Scala/Clojure model: a persistent 32-ary tree with a dedicated tail chunk. The root stores fixed-size nodes (`32` slots, `5` bits per level) and the most recent elements live inside the tail until it fills, at which point the chunk is promoted into the tree. `push`/`pop` therefore run in amortised `O(1)` time, while `get`/`set` touch at most one node per depth (`O(log₃₂ n)`). Structural sharing is preserved by cloning only the nodes along the updated path, so historical vectors remain valid without copying. Iteration walks the logical index range and yields values in order.

**List implementation notes (landed)**
`List T` is a classic persistent cons list (singly linked nodes with `{ value, next }`). `prepend/cons` and `head/tail` all run in `O(1)` time, matching the ergonomics from Scala’s `List` and Clojure’s `list`. Concatenation clones just the left spine and shares the right-hand list, and helpers such as `nth`, `reverse`, and `to_array` provide the expected `O(n)` traversals when required.

**Map/Set/Queue implementation notes (landed)**
Persistent maps and sets now use a bitmap-indexed HAMT identical to the Scala/Clojure layout (32-way fan-out, bitmap compaction, collision nodes). Inserts/updates walk at most `log₃₂ n` nodes and clone only the modified path so older versions remain available. `PersistentSet` is a thin wrapper on the map storing `void` values, adding `union`/`intersect` helpers. Builders exist for both map and set: they gather entries eagerly (via mutable buffers) and emit a frozen persistent value on `finish()` so callers can accumulate without repeated persistent updates. The persistent queue mirrors Clojure’s design: a pair of persistent lists (`front`, `back`) with lazy rebalancing when the front becomes empty, yielding amortised `O(1)` enqueue/dequeue. `LazySeq` wraps any iterator, caching elements the first time they are pulled so subsequent traversals replay the cached prefix without re-running the iterator; evaluation remains incremental, and the cache grows only as callers demand more values. On the concurrency front we now expose `ConcurrentQueue`, a light wrapper around `Channel` that gives idiomatic `enqueue`/`dequeue` helpers while preserving the existing blocking/cancellation semantics.

**SortedSet implementation notes (landed)**
Sorted sets currently use a persistent AVL tree (matching the spirit of Scala’s immutable `TreeSet`). Inserts/removals return new trees built from re-used subtrees, `contains`/`first`/`last` all run in `O(log n)`, and range queries simply walk the ordered values, so the API aligns with the spec’s expectations until the planned finger-tree variant lands.

**Mutable TreeMap/TreeSet implementation notes (landed)**
The mutable `TreeMap` mirrors Java/Scala style ordered maps: it uses an intrusive AVL tree with reference semantics, so updates rebalance in place and lookups stay `O(log n)`. `TreeSet` is layered on the map (storing `void` values), inheriting the ordering guarantees as well as efficient insert/remove/contains operations.

**Heap implementation notes (landed)**
`Heap T` is a standard binary min-heap backed by our mutable `Array`, using either the default `Ord` or a caller-provided comparator. `push` bubbles up, `pop` bubbles down, and `peek` exposes the root in `O(1)`; all structural updates happen in-place thanks to reference semantics.

**BitSet implementation notes (landed)**
`BitSet` stores bits inside an `Array u64`. Operations (`set`, `reset`, `flip`, `contains`) compute the word/bit index and mutate the relevant word in place; iteration walks the words and yields every set bit in ascending order. This matches the roadmap’s `O(1)` per-word operations while keeping the code host-agnostic.

**Mutable Collections**

| Type | Backing structure | Key operations | Performance guarantees |
| --- | --- | --- | --- |
| `Array T` | Contiguous resizable buffer | `push`, `pop`, `insert`, `remove`, `reserve` | `O(1)` amortised push/pop, `O(n)` insert/remove mid |
| `Deque T` | Ring buffer | `push_front`, `push_back`, `pop_front`, `pop_back` | `O(1)` amortised |
| `LinkedList T` | Doubly linked nodes | `push_front`, `insert_after`, `remove` | `O(1)` manipulations with node handle |
| `HashMap K V` | Open addressing (quadratic probing) | `get`, `set`, `remove`, `iter_pairs` | `O(1)` average lookup/update |
| `HashSet T` | HashMap wrapper | `insert`, `remove`, `contains` | `O(1)` average |
| `TreeMap K V` | Red-black tree | `floor`, `ceil`, `range`, `keys_sorted` | `O(log n)` |
| `TreeSet T` | Red-black tree | `insert`, `remove`, `lower_bound` | `O(log n)` |
| `Heap T` | Binary heap | `push`, `pop`, `peek`, `replace` | `O(log n)` |
| `BitSet` | Dynamic bitset | `set`, `reset`, `flip`, `contains` | `O(1)` per word |
| `SmallVec T const N` | Inline small vector | `push`, `pop` with inline storage | `O(1)` amortised |

Mutable collections implement `MutableCollection`, `Indexable`, and optionally `Sliceable` (e.g., `Array`, `Deque`). Conversion helpers (`Array#to_vector`, `HashMap#to_persistent_map`) bridge to persistent counterparts.

By convention `import able.collections.Map` resolves to the mutable `HashMap K V`; persistent map variants must be imported explicitly (e.g., `able.collections.immutable.Map`).

#### 5.1.5 Iterator & Collectable Utilities

- Iterator adapters: `MapIterator`, `FilterIterator`, `Enumerate`, `Zip`, `Chain`, `TakeWhile`, `DropWhile`, `Flatten`, `Chunk`, `Window`, `Partition`, `Sliding`.
- `Iterator#lazy` returns self but signals intent; pipelines remain lazy until realised via `to_*` / `collect`.
- `Collectable` implementations (`ArrayCollectable`, `VectorCollectable`, `SetCollectable`, `MapCollectable`) materialise iterators without coupling to constructors.
- `Enumerator` bridges yield elements through Able `spawn` while respecting scheduler semantics.

#### 5.1.6 Common Operations & Naming Conventions

- Shared terminology: `push`, `pop`, `push_front`, `push_back`, `shift`, `unshift`, `append`, `prepend`, `concat`, `join`.
- Functional helpers: `map`, `flat_map`, `filter`, `reject`, `reduce`, `fold_left`, `fold_right`, `collect`, `group_by`, `partition`, `zip`, `zip_with_index`.
- Lazy usage relies on `iterator`/`each`: collections expose `iterator()` returning an `Iterator`, and the `Iterator` interface supplies chainable convenience methods (`map`, `filter`, `take`, etc.). Types mixing in `Enumerable T` gain Crystal-style helpers built on `each`.
- Access helpers: `first`, `last`, `nth`, `take`, `drop`, `split_at`, `chunks`, `windows`.
- Boolean combinators: `any?`, `all?`, `none?`, `one?` with lazy short-circuit semantics.
- Numeric reductions: `sum`, `product`, `min`, `max`.
- Concurrency hook: `each_future` yields via Able `spawn` for deterministic scheduling tests.
- Collections implement `Display`, `Debug`, `Clone`, `Equatable`, `Hashable` when element types allow.
- Methods ending with `?` return `bool`; `!` suffix denotes raising variants (`fetch!`).
- Expose both total (`get`, returning `Option`) and partial (`get!`, raising `IndexError`) accessors.

#### 5.1.7 Allocation-Free Transformation Strategies (Exploratory)

**Option A — Iterator Fusion (Chain Collapsing)** *(Preferred for v11)*
Fuse successive iterator adapters into a single iterator object during pipeline construction. Requires combinator methods returning specialised fused iterators and potentially a `FusableIterator` marker. Pros: minimal conceptual overhead, works for mutable and persistent collections. Cons: still allocates fused iterator objects; cloning pipelines for multiple consumers may duplicate work.

**Option B — Stateful Builders (Reusable Accumulators)**
Introduce mutable builders (`ArrayBuilder`, `VectorBuilder`) that emit directly into target buffers, eliminating intermediate iterators.

```able
builder := ArrayBuilder i32
builder.reserve(array.len)
array.each_builder(builder) { |value, emit| new_val := value + 1; emit.call(new_val) }
result := builder.finish()
```

Interfaces: `CollectionBuilder T C`, collection-level `each_builder`, iterator `into_builder`. Works well for persistent collections; must honour cancellation.

**Option C — Clojure-style Transducers (Push-based, Destination-Agnostic)** *(Prototype after v11 launch)*

```able
module Transducer In Out
  fn init(self) -> State
  fn step(self, state: State, input: In, emit: fn(Out)) -> StateStep
  fn complete(self, state: State, emit: fn(Out)) -> void
end

pipeline := Transducers.map(fn(x) { x + 1 }).then(Transducers.filter(&.even?))
result := pipeline.into(array, collectable: VectorCollectable i32)
```

Requires `Transducer`, `Reducible`, and enhanced `Collectable` signatures. Pros: allocation-free composition, target agnostic. Cons: complex state management; must integrate with scheduler yielding.

**Option D — In-place Mutation Pipelines**
Mutable collections expose `transform!` style APIs that mutate in place while iterating, removing allocations. Only applies to mutable structures; semantics must document evaluation order and error behaviour.

**Option E — Compiler-guided Fusion (Future)**
Annotations such as `@inline_pipeline` hint optimisers to fuse eager operations without new runtime APIs. Dependent on advanced compiler tooling; likely out of scope until a static compiler exists.

Each option must coexist with the collection interfaces above. Decisions should balance ergonomics, implementation complexity, and scheduler integration. Iterator fusion (Option A) is slated to land first.

### 5.2 Strings & Text

- Provide a standard library `String` struct backed by `Array u8` plus cached metadata (`byte_len`, optional rope for shared slices). Bytes always store UTF-8 but no implicit normalisation is applied.
- Expose layering helpers:
  - `String::bytes()` → `(Iterator u8)` for raw access.
  - `String::chars()` → `(Iterator char)` – iterates Unicode scalar values (Able `char` is `u32`).
  - `String::graphemes()` → `(Iterator Grapheme)` – relies on Unicode segmentation rules; `Grapheme` is a lightweight view over a subsequence of the byte buffer.
- Ship conversion and normalisation utilities (`to_nfc`, `to_nfd`, `to_nfkc`, `to_nfkd`) that return new `String` values.
- Provide builders (`StringBuilder`), slices (`SubString`), and binary variants (`ByteString`) backed by the same `Array u8`.
- Interop helpers (`String::from_host_string`, `to_host_string`) live in `able.core.host`.
- Document that default indexing APIs operate on **byte offsets**; char/grapheme aware slicing lives on dedicated helpers to avoid ambiguity.

#### 5.2.1 Host Runtime Backing (Go & TypeScript)

**Strings**
Go stores native `string` values plus cached metadata; TypeScript keeps JS strings with lazily computed `Uint8Array` caches. The Able `char` type is a Unicode scalar (`u32`). Grapheme segmentation is layered on top via dedicated helpers that rely on host-provided break iterators. Provide bridging APIs for extern packages.

**Numbers**
Map Able integers/floats to host primitives (`int64`, `uint64`, `float64` in Go; `BigInt`/number in TypeScript) while maintaining Able type tags and range checks.

**Booleans & Nil**
Expose thin wrappers so Able methods (`Bool#then`, etc.) can be attached without leaking host details.

**Collections**
Mutable collections rely on host-managed buffers behind opaque handles (`HostArrayHandle`, `HostMapHandle`). Able code interacts via extern shims in `able.core.host.collections`.

Guidelines: reuse host performance characteristics when they align with Able semantics, but keep host-specific logic behind explicit bridge modules.

### 5.3 Numeric & Ordering Interfaces

- Extend numeric capabilities without renaming spec interfaces.
- Provide capability flags (`HasExactZero`, `HasInfinity`) and conversion helpers (`try_to_i32 -> Result i32`) to avoid silent overflow.
- Document which interfaces each builtin numeric type implements; build parity tables for Go and TypeScript.

#### 5.3.1 Core Types Beyond Collections

- `Option T`, `Result T`, `Either L R`, `Try T` aliasing `Result T Error`.
- Ordering helpers via the `Ordering` union.
- `Range T` generic over numeric/comparable types, integrating with `Iterable`.
- `Rational` numbers under `able.numbers.rational` for exact fraction arithmetic; implements the numeric/type-class stack so callers can bridge between integers and floats without precision loss.
- `Int128` helper in `able.numbers.int128` that stores signed 128-bit values as two `u64`s so deterministic arithmetic/serialization works even when runtimes lack native bigints; `UInt128` mirrors this surface for `[0, 2¹²⁸-1]`.
- Time primitives (`Duration`, `Instant`) under `able.time`.
- Error hierarchy aligned with Sections 11 and 12, plus extension points for IO/network errors.
- Tuple helpers (`Pair`, `Triple`, `TupleN`) until variadic generics exist.
- `Lazy T` for deferred, memoised computation.

#### 5.3.2 Core Primitive Behaviour Interfaces

| Interface | Purpose | Key methods |
| --- | --- | --- |
| `StringLike` | Shared string behaviour | `len_bytes`, `len_graphemes`, `is_empty`, `byte_at`, `slice`, `concat`, `to_string` |
| `StringBuffer` | Mutable string builders | `push_str`, `insert`, `remove`, `finish` |
| `Numeric` | Common numeric API | `zero`, `one`, `abs`, `sign`, comparisons, arithmetic via spec operator interfaces |
| `IntegerLike` | Integer helpers | `bit_length`, `checked_add`, `div_mod`, `to_bigint` |
| `FloatLike` | Floating-point helpers | `is_nan`, `is_infinite`, `floor`, `ceil`, `round`, `fract` |
| `BooleanLike` | Boolean helpers | `and_then`, `or_else`, `xor` |
| `CollectionLike T` | Shared collection expectations | `len`, `is_empty`, `contains`, `iter` (extends `Iterable T`) |
| `MutableCollectionLike T` | Mutation contract | `clear`, `reserve`, `push`, `extend` |

#### 5.3.3 Numeric & Ordering Hierarchy (Scala-Inspired)

| Interface | Extends | Responsibilities |
| --- | --- | --- |
| `PartialOrdering T` | – | `compare_partial(self, other) -> Option Ordering` (supports floats/NaN) |
| `TotalOrdering T` | `PartialOrdering T` | Guarantees total order; provides `compare`, `min`, `max`, `clamp` |
| `OrderingOps T` | `TotalOrdering T` | Default mix-in adding `<=`, `>=`, `between?`, `sort_pair` |
| `Semigroup T` | – | `combine(self, other) -> T` |
| `Monoid T` | `Semigroup T` | Adds `empty() -> T` |
| `Group T` | `Monoid T` | Adds `inverse(self) -> T` |
| `AdditiveMonoid T` | `Monoid T` | Maps to `+` with `zero` identity |
| `AdditiveGroup T` | `Group T` | Adds `negate` |
| `MultiplicativeMonoid T` | `Monoid T` | Provides `one()`, `mul` |
| `Numeric T` | `AdditiveGroup T`, `MultiplicativeMonoid T`, `TotalOrdering T` | Core numeric surface (`from_i64`, `pow`, `clamp`, etc.) |
| `Integral T` | `Numeric T` | Adds `div_mod`, `quot`, `rem`, `bit_length`, `bit_count` |
| `Fractional T` | `Numeric T` | Adds `reciprocal`, `div`, `fract`, rounding modes |
| `Bitwise T` | – | `and`, `or`, `xor`, `not`, shifts |
| `Signed T` | `Numeric T` | `is_negative?`, `is_positive?`, `abs` |
| `Unsigned T` | `Numeric T` | `leading_zeros`, `trailing_zeros` |
| `NumericConversions T` | – | `to_i32`, `to_u32`, `to_u64`, `to_f64`, `to_bigint`, `to_decimal`, with `Result` return when lossy |
| `ToHostNumeric` | – | `to_host_number(runtime) -> host_number` |

Lawfulness: document associativity/identity/inverses for each algebraic interface. Keep inheritance minimal to avoid diamond problems; allow orphan `impl`s where coherence permits.

#### 5.3.4 Compatibility with Future Numerics Libraries

- Emphasise lawfulness and stability so external numerics libraries (e.g., Spire-style) can rely on the interfaces.
- Maintain a minimal, acyclic inheritance graph to support retrofitting implementations.
- Distinguish exact vs approximate fields via capability markers.
- Provide loss-aware conversions (returning `Result`) to prevent silent truncation.
- Ensure host interop is optional; high-performance numerics can remain in Able code.

### 5.4 Concurrency Extensions

- Add higher-level helpers (`select`, timer channels, async iteration helpers) built atop spec-backed primitives.
- Provide concurrency-aware collections (`ConcurrentQueue`, `AsyncStream`) that integrate with `spawn`/`Future` tasks.
- Offer guidance (but no automatic behaviour) for cooperative scheduling, e.g., optional helper adapters that invoke `future_yield` for test harnesses in the TypeScript runtime.

### 5.5 Regular Expression & Pattern Matching Vision

Goals: ship an RE2-inspired engine offering predictable performance with layered APIs.

1. **High-level `Regex` interface** — `compile`, `match?`, `match`, `scan`, `replace`, `split`, with options for case folding, multiline, Unicode classes.
2. **`RegexSet` / multi-pattern support** — compile multiple patterns and return indices of matches.
3. **Automata access** — export NFA/DFA representations (`to_nfa`, `to_dfa`), introspection (`states`, `transitions`), serialization (`to_graphviz`).
4. **Builder & analysis tools** — programmatic construction/composition, automata minimisation.
5. **Streaming & lazy matching** — incremental search over streams using resumable scanner states.

Implementation considerations:
- Prefer RE2-style guarantees (no catastrophic backtracking). Evaluate Rust `regex` and Clojure inspirations.
- Support lookahead/lookbehind cautiously; backreferences likely out of scope initially.
- Provide deterministic performance backed by compiled automata; optional bytecode export for reuse.

### 5.6 Documentation & Tooling

- Publish Big-O characteristics, mutation semantics, and iterator invalidation rules for each collection.
- Generate API reference from doc comments; provide cookbook examples.
- Track coverage via `run_all_tests.sh` (Bun + Go) and integrate new modules into CI.
- Produce interface/type matrices (Go vs TypeScript) for parity audits.

## 6. Runtime Integration Guidelines

- Arrays, hash maps, strings, channels, and mutexes use opaque handles managed by the interpreters to guarantee deterministic behaviour.
- All host-specific shims live in `able.core.host.*`, keeping dependencies explicit for future targets.
- Default hashing uses FNV-1a seeded with a per-process random value to mitigate collision attacks; allow opt-in deterministic modes for tooling later.
- Long-running stdlib routines (iterator pipelines, blocking operations) should integrate with `future_cancelled` / `future_yield`.
- Go runtime backs mutable collections with slices/maps; TypeScript uses JS arrays/maps while enforcing identical semantics. Persistent structures can be written in Able and shared across runtimes.
- Extern shims exist for heavy operations (`String#encode_utf8`, `HashMap#rehash`) to keep interpreters performant.
- Ensure `Serializable` implementations cooperate with runtime-specific IO backends (Go JSON, TypeScript JSON).

### 6.1 Runtime vs Able Responsibilities

The goal is to keep the interpreter-provided surface area tiny and fast while implementing everything else in ordinary Able code. The runtime only hosts the bits that must cooperate with the host VM (Go goroutines, JS event loop, garbage collector, etc.) or that would be prohibitively slow to reimplement in the interpreter.

| Component | Runtime-provided primitives | Able-level implementation |
| --- | --- | --- |
| `Array T` | Allocate/free storage, grow/shrink capacity, load/store elements, clone backing buffer, expose length/capacity, ensure GC-visible references. | Public API (`push`, `pop`, `get`, `set`, `slice`), iteration helpers, higher-order methods, builders, persistent conversions. |
| `HashMap K V` / `HashSet T` | Allocate table, compute seeded hash, probe/update slots, resize/rehash, expose iterator over live entries efficiently. | Interface (`Map K V`), error handling, iteration adapters, convenience APIs, persistent wrappers. |
| `String` / `SubString` / `StringBuilder` | Own host string/byte buffers, convert to/from UTF-8, slice by grapheme, lazily materialise byte views, bridge to host APIs. | Public string API, builders, text algorithms, formatting helpers, Unicode utilities layered on the primitive hooks. |
| `Future` | Schedule work, memoise results, propagate cancellation, map exceptions into `FutureError`, integrate with host executor. | Interface definition, ergonomic helpers, fixture utilities; user code treats them as ordinary Able values. |
| `Channel T`, `Mutex` | Create/destroy handles, send/receive/close with correct blocking semantics, lock/unlock, interact with cancellation, wake waiters. | Type wrappers, iteration, error mapping, helper functions such as `with_lock`. |
| Dynamic packages / host bridges (`able.core.host.*`) | Lookup host packages, call extern functions, manage opaque handles. | Higher-level abstractions that expose those capabilities safely to Able code. |
| `Hasher` | Low-level FNV-1a implementation seeded per process, byte ingestion, finalisation. | Trait definition, convenience wrappers, hashing utilities that compose the primitive operations. |

Everything else—including persistent collections (`Vector`, `List`, `PersistentMap`), iterator adapters, numeric/type-class scaffolding, regex helpers, etc.—lives entirely in Able source. The plan is to progressively shrink the native surface to “just the primitives” so the shared stdlib remains mostly host-independent.

## 7. Array-Backed Data Structures (Initial Design)

### 7.1 Common Building Blocks

- Use `Array T` plus primitive scalar fields (`i32`, `i64`, `bool`) only.
- Randomised hashing via the `Hash` trait and a stdlib `Hasher` (FNV-1a seeded per process).
- Use `IndexError` and `Result`/`Option` for error signalling; no host exceptions.

#### 7.1.1 Host-Level Array Requirements

Interpreters only need to expose a tiny, predictable surface so higher-level collections can stay in Able code:

| Capability | Purpose |
| --- | --- |
| `Array.with_capacity(capacity: i32) -> Array T` | Allocate a buffer of exact capacity (filled with the element type’s default/zero value). Used for initial tables and resizes. |
| `size(self: Array T) -> i32` / `capacity(self: Array T) -> i32` | Report logical length and total slots; higher-level structures keep their own logical counts but still consult capacity. |
| `read_slot(self: Array T, idx: i32) -> T` | Fetch the value stored at a slot without reallocating (runtime may return a copy). |
| `write_slot(mut self: Array T, idx: i32, value: T) -> void` | Store a value at a slot without changing capacity. |
| `clone(self: Array T) -> Array T` (optional) | Duplicate the backing buffer when a shallow copy is required (persistent conversions, builders). |
| `fill(self: Array T, value: T) -> void` (optional) | Efficiently initialise or clear the buffer; can be emulated in Able code if absent. |
| `drop(self: Array T) -> void` | Release the handle when the structure is discarded (already implicit in runtime GC). |

Everything else—probe logic, load-factor bookkeeping, iterators, tombstone management—is implemented in Able using these primitives.

### 7.2 HashMap Design Overview

| Aspect | Decision |
| --- | --- |
| Storage | Single `Array Entry` with entries `{ state: u8, key: K, value: V, hash: u64 }` (`state` encodes Empty, Filled, Tombstone). |
| Hashing | `hash = Hasher::fnv1a(key)` followed by `hash & (capacity - 1)` using power-of-two tables. Keys require `Hash` + `PartialEq`. |
| Probing | Quadratic probing (`i = i + step`, `step += 1`) keeps clusters short with power-of-two capacities. |
| Load factor | Resize when `len >= capacity * 0.7`; allocate next power of two and reinsert live entries. |
| Inserts | Probe until Empty or Tombstone; reuse tombstones but continue to ensure uniqueness. |
| Lookups | Probe until Empty (not found) or a matching filled entry. Skip tombstones. |
| Deletes | Mark slot as Tombstone; decrement `len`; optionally rebuild if tombstones exceed 20% of capacity. |
| Iteration | Iterator skips Empty/Tombstone slots and yields `(K, V)` pairs. |
| Error handling | `set` overwrites existing values, returning `nil` on success (`Result` only for overflow). `get` returns `Option`. |

Implementation strategy:
1. Build `HashMap K V` backed by the array plus counters (`len`, `used`).
2. Add unit tests in both interpreters covering collisions, growth, deletion, iteration order.
3. Derive `HashSet T` as `HashMap T nil`.
4. Explore Robin Hood or SwissTable layouts later for locality improvements.

### 7.3 Other Structures

- `Deque T`: circular buffer with head/tail indices, amortised `O(1)` push/pop at both ends.
- `Vector T` (persistent): tree of array chunks (RRB) once mutable structures are solid.
- `Heap T`: binary heap stored in Array with comparator via `Ord`.

## 8. Implementation Phases

1. **Foundations**
   - Finalise interfaces in `able.core.interfaces`, adopting Crystal-style layering.
   - Implement `Option`, `Result`, `Range`, error structs; ship `String`, `StringBuilder`, `Array`, `HashMap` using extern helpers.
   - Add core iterator adapters and naming conventions (eager vs lazy).
2. **Persistent Collections & Builders**
   - Implement `Vector`, `List`, `Set`, `Map` persistent variants.
   - Provide builders, `Collectable` instances, and conversions between mutable and persistent forms.
   - Document performance tables and add fixtures covering iterator vs eager pipelines.
3. **Advanced Collections & Utilities**
   - Implement sorted structures (`TreeMap`, `TreeSet`, `Heap`, `BitSet`), lazy sequences, streaming IO adapters.
   - Add concurrency-aware collections (`ConcurrentQueue`, `AsyncStream`) and serialization/time/IO modules.
4. **Polish & Tooling**
   - Generate API references, cookbook examples, and integrate into `run_all_tests.sh`.
   - Update spec Section 14 when interfaces stabilise; tick items off in `PLAN.md`.

## 9. Outstanding Work & Open Decisions

### 9.1 Near-Term Tasks

- Wire spec-accurate default methods into `Iterable` in both interpreters.
- Replace placeholder channel errors (`ChannelClosed`, etc.) in code with spec names (`ClosedChannelError`, `SendOnClosedChannelError`, `NilChannelError`).
- Implement the Array-backed `HashMap` outlined above and add fixtures for parity testing.
- Document Big-O guarantees for `Array`, `HashMap`, and upcoming structures once implementations land.
- Finalise host bridge APIs for strings and arrays, ensuring behaviour parity between Go and TypeScript.
- Revisit advanced roadmap items (strings, regex, numerics) once the baseline passes all fixtures and parity tests.

### 9.3 String primitivisation status

- Minimal kernel bridges kept: `__able_string_from_builtin`, `__able_string_to_builtin`, `__able_char_from_codepoint`. All user-facing behaviour now lives in `able.text.string`.
- Stdlib now operates on primitive `string` for `len_*`, `substring`, `split`, `replace`, `starts_with`/`ends_with`, iterators, and builders. The `String` wrapper remains private scaffolding only; public helpers wrap/unwrap to primitives.
- Runtime/typechecker touchpoints: TS/Go interpreters continue to register only the three bridges; typecheckers source method sets from the stdlib and emit import diagnostics. Native string helpers are removed.
- Tests/fixtures: TS ModuleLoader suites (`test/stdlib/string_stdlib_integration.test.ts`, `test/stdlib/string_array_methods.test.ts`) exercise the primitive path; Go parity relies on shared fixtures and the stdlib runtime.

### 9.2 Open Decisions (Need Product Guidance)

**Resolved**
- `able.collections.Map` aliases the mutable `HashMap K V`.
- `String` slicing operates on grapheme (`char`) boundaries; byte-oriented slicing remains opt-in via explicit helpers.
- User-defined errors implement the `Error` interface (no separate `ErrorLike` hierarchy).
- Lazy naming follows the `each`/`iterator` conventions with `Enumerable`/`Iterator` convenience mixins; no `_iter` method suffixes.
- Allocation-reduction roadmap will land Iterator Fusion first (Section 5.1.7 Option A).
- Default hashing uses per-process random seeds (FNV-1a baseline) with an opt-in deterministic mode for tooling.
- Persistent builders mirror Scala/Clojure designs with dedicated accumulators that flush directly to the persistent representation.
- Stdlib iterators do not auto-call `future_yield`; cooperative scheduling remains an explicit concern for callers/tests (primarily in the TypeScript runtime).

**Pending**
- None at present.

## 10. Brainstorm Backlog

- Implement persistent collections once in Able and share across runtimes.
- Explore Ruby-style `Enumerator` objects to bridge synchronous iteration with `spawn`-based concurrency.
- Support multi-arity `map`/`zip` similar to Scala.
- Provide structural pattern-matching helpers (e.g., `Vector::Unapply`) after parser support is available.
- Offer `Lens`-style utilities for immutable updates (inspiration from Scala and Functional Java).
- Evaluate serde-like derive mechanisms once macros exist; design interfaces with derivation in mind.
