# Shared value/heap contract reconciliation

Date: 2026-07-22

Status: complete; bounded production pilot admitted

## Decision

**`admit-bounded-production-shared-value-pilot`**.

The three gaps found by the current-Go graph binding are reconciled. The
generated contract now owns wide integer payloads, exposes the complete
language-visible Hasher state through a narrow runtime API, and separates
Iterator semantic roots from its intentionally opaque Go callback driver.

The checked matrix now records 27 exact runtime-owned kinds, zero conditional
or blocked kinds, and four intentionally opaque host-registry kinds across all
31 runtime kinds. This admits one controlled live migration pilot; it does not
itself move either interpreter, the compiler, or the stdlib onto the heap.

## Owned wide scalars

An immediate cell may set `CellAuxIndirect` and use its payload as a
generation-checked shared-heap identity. The only valid backing layout is the
immutable internal `WideScalar` layout, whose fields hold a format scalar and
owned bytes. Collection traces this edge and rejects stale IDs or a mismatched
layout before sweeping.

Small integers remain directly encoded in the 64-bit payload. Wide signed and
unsigned Integer values store sign plus magnitude in `WideScalar`; decoding
reconstructs an independent `big.Int` and preserves the Able integer suffix.
The representation is generic scalar ownership, not an integer benchmark or
named-container rule.

## Hasher state

`HasherValue` now has a reconstruction constructor and a read-only
`SemanticState` method. These expose the complete language-visible state while
leaving its storage private to `pkg/runtime`. The shared Hasher layout stores
that state as its existing scalar field and restores it exactly.

This is deliberately narrower than exporting Hasher internals. It gives the
shared contract one semantic authority without changing normal hash behavior.

## Iterator host-driver boundary

`IteratorHostDriver` explicitly groups the current `Next`, `NextRaw`, and
finalizer callbacks with the semantic values those opaque callbacks retain.
The shared Iterator object stores a host-registry driver cell, retained value
cells, closed state, and the existing optional semantic state identity.

Encoding registers callbacks as an opaque host entry and traces every declared
retained value. Decoding constructs the Iterator wrapper before following its
edges, restores the driver, and preserves shared identity between a captured
root and any other graph reference. A checked vector resumes at the expected
position, runs the finalizer, and keeps captured roots live through collection.

Existing iterator constructors retain their behavior. Call sites that migrate
semantic captures to the shared heap must use the explicit host-driver boundary
and enumerate those captures; callbacks that remain wholly Go-owned stay
opaque and under Go lifetime management.

## Conformance result

The report reruns the earlier alias, cycle, closure, nominal, Error, Future,
tree-walker, and bytecode vectors, then adds:

- a negative `i128` larger than 64 bits with exact magnitude and suffix;
- a nonzero evolving Hasher state;
- an Iterator driver with resume position, finalizer, and a retained aliased
  Array root;
- heap-level indirect-scalar tracing, reclamation, and wrong-layout rejection.

Production execution, canonical stdlib, application output, dependencies,
foreign heaps, executable memory, benchmarks, and WASM remain unchanged. No
named-container or non-primitive nominal compiler special case was added.

No timing was run. These changes establish representation correctness outside
production hot paths, so timing them would not measure Able application speed.

## Next recommendation

Complete **`runtime-contract-performance-evidence-reconciliation`**, then begin
**`shared-value-heap-production-pilot`** only if that evidence is current.

The production runtime scope changed intentionally, so the evidence ledger now
selects all 21 compiled, bytecode, and cross-family closures. Refresh the named
closures with verifier-backed repeated processes and arithmetic means, update
their evidence records, and advance a closure only after its governing rows and
guards have been reviewed. The existing zero-invalidated structural conclusion
must remain unavailable until this is done.

Why: the value contract is semantically ready, but prior performance decisions
were measured against an older runtime. `IteratorValue` gained retained-root
state, and even changes expected to be cold can alter allocation size or cache
behavior. Fresh evidence prevents a correct architecture change from being
mistaken for a broadly profitable one—or from hiding a regression.

After reconciliation, move one generic call/return value boundary used by both
Go interpreters behind a bounded opt-in path. Parameters, returned values,
aliases, errors, and host effects must use the same shared identities, while
ordinary execution falls back without graph conversion. Exercise unlike
fixture families before making it the default. Calls and returns are broadly
shared by real programs and expose identity, error, closure, nominal, and host-
effect interactions together.

## Reproduction

```sh
just bench-semantic-abi-go-binding-check
just bench-architecture-budget-check
just bench-evidence-ledger --check
```
