# Collections: Eager vs Lazy Split (v11)

## Goals
- Keep lazy pipelines available (`Iterator.map`/`filter`/etc.) without ambiguity.
- Provide eager, type-preserving transforms on collections.
- Use `Default` + `Extend` for collection builders (consistent `collect` story).
- Preserve existing iteration semantics (`for` loops, `each`, `iterator`).

## Problem
If both `Iterable` and `Enumerable` define `map`, then `collection.map(...)` is ambiguous when a type implements both interfaces. The current method-resolution rules do not prefer one interface over another, so we must avoid overlapping eager/lazy method names on the same receiver type.

## Proposal

### 1) Keep `Iterable` minimal
`Iterable` only establishes the iteration protocol:

- `each(self, visit)`
- `iterator(self) -> Iterator T`

`Iterable` does **not** include lazy adapters (`map`, `filter`, etc.).

### 2) Put lazy adapters on `Iterator`
`Iterator` owns lazy transforms and always returns another `Iterator`:

- `map`, `filter`, `filter_map`, `flat_map`, `take`, `skip`, `zip`, etc.
- `collect(self) -> C where C: Default + Extend T` to materialize iterators.

`Iterator T` continues to implement `Iterable T` so `for` loops work on iterators.

### 3) Add eager, type-preserving `Enumerable` (HKT)
Introduce an HKT interface for eager, type-preserving transforms:

```able
interface Enumerable A for C _ : Iterable A {
  fn lazy(self: C A) -> (Iterator A) { self.iterator() }

  fn map<B>(self: C A, f: A -> B) -> C B
    where C B: Default + Extend B {
    acc: C B := C.default()
    self.each { v => acc = acc.extend(f(v)) }
    acc
  }

  fn filter(self: C A, predicate: A -> bool) -> C A
    where C A: Default + Extend A {
    acc: C A := C.default()
    self.each { v => if predicate(v) { acc = acc.extend(v) } }
    acc
  }
}
```

Collections implement `Enumerable`, not `Iterator`, so `collection.map(...)` is eager and unambiguous.

### 4) Explicit lazy switch
To switch to lazy evaluation on a collection, call:

- `collection.lazy().map(...)` (preferred)
- or `collection.iterator().map(...)`

### 5) Type-specific overrides
If a collection needs specialized semantics (e.g., `HashMap` mapping values vs entries),
it can override the eager method on the concrete type while still implementing `Enumerable`.

Decision: `HashMap.map` transforms values while preserving keys; entry-level transforms
should use `for_each` or an entries view (TBD).

## Examples

```able
arr := [1, 2, 3]
eager := arr.map(fn(x: i32) -> i32 { x * 2 })        ## Array i32
lazy := arr.lazy().map(fn(x: i32) -> i32 { x * 2 })  ## Iterator i32

bytes := "AbC".bytes()
lazy2 := bytes.map(to_ascii_lower)                   ## Iterator u8
```

## Compatibility Notes
- Existing lazy pipelines should move from `Iterable` to `Iterator`.
- Eager `Enumerable` methods should default to `Default + Extend` builders.
- This design avoids method-resolution ambiguity and aligns with Rust (lazy iterators) + Scala (eager collections).

## Follow-ups
- Update stdlib interfaces (`core/iteration.able`, `collections/enumerable.able`) to reflect this split.
- Adjust stdlib call sites and tests (including `.examples/`) to use `lazy()` where needed.
- Document `HashMap`/`HashSet` `Enumerable` semantics (values vs entries).
- Keep `String`/`StringBuilder` on bespoke eager `map`/`filter` (char-only) for now; they are not HKTs, so `Enumerable` cannot target them until a suitable wrapper exists. Use `.chars()`/`.bytes()` for lazy pipelines.
