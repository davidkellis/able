# Array `slice` semantics audit

**Status:** accepted and implemented 2026-07-14

## Current state

Section 6.8 formerly showed the proposed surface:

```able
arr.slice(start: u64, end: u64) -> Array T
```

but explicitly left copy-versus-view behavior unresolved. That gap is now
closed by the canonical external `able-stdlib` method; it is an ordinary Able
implementation layered over kernel Array primitives, not a tree-walker,
bytecode, compiler, or kernel special case.

The nearby private text helper `slice_bytes` allocates a new `Array u8` and
copies the selected bytes. `Array.clone_shallow()` also copies the Array
backing while preserving element values. Those are useful compatibility
evidence, but neither silently defines public `Array.slice` semantics.

## Existing representation constraints

An Able `Array` is currently a whole-storage container:

- Interpreter `runtime.ArrayValue` has elements, a storage handle, tracked
  state, and a lease, but no view offset or visible-length field.
- Dynamic and every mono primitive `ArrayStore` state own one full backing
  sequence plus its capacity and revision.
- Compiled static Array carriers expose `length`, `capacity`, and
  `storage_handle`; their native backing is a whole Go slice.
- Assignment, argument passing, returns, structs, interfaces, and futures
  preserve aliasing of the same Array storage. The handle-lifetime contract
  requires all such aliases to observe mutation and capacity changes.

Consequently, a borrowed view cannot be added as a small stdlib method: it
would require an offset/length-aware representation, read/write/index/length/
capacity rules for every store kind, revision and lease propagation, compiled
carrier conversion, and a defined interaction with `push`, `pop`, `reserve`,
and `clear`.

## Options

| Option | Observable behavior | Cost and compatibility |
| --- | --- | --- |
| Independent shallow copy | Returns a new Array containing the selected element values. Element replacement, length, capacity, and growth do not affect the source; referenced mutable element values retain their ordinary value semantics. | `O(end - start)` time and new backing storage. Fits existing ArrayStore, compiler native carriers, `clone_shallow`, and the text helper without a new representation. |
| Borrowed Array view | Returns an Array whose element replacement aliases a selected source range. | Potentially `O(1)` creation, but needs new language-wide offset, capacity, reallocation, lifetime, and alias rules. Existing whole-storage assumptions make it a new representation project, not a normal API addition. |
| Separate `ArrayView T` type | Keeps `Array.slice` as a copying operation or omits it; exposes borrowing through a distinct non-owning type. | Makes aliasing explicit, but needs an independent protocol/API design and is not described by the present proposed `Array.slice` surface. |

## Accepted contract and implementation

`Array.slice` has **independent shallow-copy** semantics. `end` is exclusive,
and exactly `0 <= start <= end <= arr.size()` is valid. Invalid bounds return
`IndexError` through `!Array T`. A success result has fresh Array storage with
both length and capacity equal to `end - start`; it copies the selected element
values in order. Consequently, slot replacement, length/capacity changes, and
growth do not alias the source, while referenced mutable elements retain their
ordinary value semantics.

`able-stdlib/src/collections/array.able` implements this with a checked
u64-indexed loop over the existing `Array.with_capacity`, `read_slot`, and
`push` methods. No runtime representation, ArrayStore lease rule, compiler
lowering, opcode, or named-container branch is added. The bounds check occurs
before narrowing indices to the kernel's i32 slot ABI.

`exec/06_12_29_stdlib_array_slice` covers non-empty and empty selections,
reversed and out-of-range bounds, exact capacity, source/result slot
independence, and post-slice source growth. It passes in tree-walker, bytecode,
and strict no-fallback compiled execution. The canonical stdlib Array smoke
test covers the same API contract.

## Non-goals

This audit does not add a compiler optimization, VM opcode, `HashMap` or other
named-container exception, benchmark, or WASM work. It also does not choose
semantics for persistent `Vector.slice`, `String`/`Grapheme` views, or a future
generic `Sliceable` protocol.
