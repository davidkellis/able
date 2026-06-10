# Shared value/heap conformance contract

Date: 2026-07-22

Status: retained; test-only Go binding conformance next

## Decision

**`retain-shared-value-heap-conformance-contract`**.

The isolated semantic ABI now has a generated, backend-neutral ownership and
tracing contract for every current reference-backed value. A deterministic
pure-Go model proves the contract without importing it into the tree-walker,
bytecode VM, compiler, or live runtime.

This is a semantic feasibility result, not a performance result. No benchmark
was timed because no production path executes the model.

## Generated layout authority

The manifest generator reads the canonical runtime `Kind` declaration and
emits three exhaustive sets:

- seven immediate kinds, carried entirely in a 16-byte pointer-free cell;
- 20 shared-heap kinds, each with exactly one object layout;
- four host-registry kinds, each with exactly one lifetime descriptor.

Object fields use only six storage classes: scalar, bytes, one cell, one object
reference, a cell vector, or an object-reference vector. Cell and object fields
are traced; scalar and byte fields are not. This makes ownership a property of
the generated descriptor rather than a switch in an executor.

| Mutability | Shared runtime layouts |
| --- | --- |
| Mutable | Array, HashMap, Hasher, StructInstance, Package, Iterator |
| Immutable | String, Function, FunctionOverload, StructDefinition, TypeRef, InterfaceDefinition, InterfaceValue, UnionDefinition, DynPackage, DynRef, Error, BoundMethod, ImplementationNamespace, PartialFunction |

The six internal layouts are Environment, BindingCell, SequenceStorage,
MapStorage, IteratorState, and ErrorContext. Environment and BindingCell are
separate so captured bindings can change without changing closure identity.
The storage layouts are generic; they do not authorize Array-, HashMap-, or
other nominal-specific compiler lowering.

NativeFunction, HostHandle, NativeBoundMethod, and Future live in a separate
generation-checked registry. Descriptors state mutability, cancellation, and
whether the entry may retain semantic cells. Future is mutable and cancelable;
release, rather than cancellation, ends its retained-result lifetime.

## Model rules

- Object and host indices and generations are nonzero.
- Allocation reuses the lowest free index and increments its generation.
- Resolution distinguishes invalid, stale-generation, and released identities.
- A tagged cell must agree with its referenced object or host descriptor.
- Immutable layouts reject mutation. Mutable updates replace validated fields
  atomically and increment an object revision.
- Root frames are explicit, replaceable, and closeable.
- Every active host-registry entry is a host-held root for its retained cells.
- Marking follows tagged cells and internal object references, with an identity
  visited set for cycles. Marking completes before any sweep begins.
- A stale or malformed traced edge produces a deterministic error and no
  partial sweep.
- Collection order and free-slot reuse are deterministic.

The model contains no case keyed to a container or nominal type. Allocation,
mutation, tracing, and collection consult only generated layout metadata.

## Conformance evidence

Eight vectors run twice in-process and produce identical reports:

1. two StructInstance cells observe one mutation and revision;
2. String mutation is rejected by its immutable layout;
3. an unrooted Environment self-cycle collects;
4. a recursive Function/Environment/BindingCell cycle survives a function root
   and all three objects collect after that root closes;
5. deterministic slot reuse increments the generation and rejects the old ID;
6. a canceled Future retains an Array result until registry release;
7. one Iterator root retains an iterator-state, Error, ErrorContext,
   InterfaceValue, InterfaceDefinition, and Array graph;
8. replacing a root-frame cell releases the old identity and retains the new.

Additional corruption tests reject stale traced edges, host tag disagreement,
immutable host updates, and unsupported cancellation. Generated-manifest tests
prove all 20 shared and four host kinds occur exactly once.

Checked evidence:
`v12/docs/perf-baselines/2026-07-22-shared-value-heap-conformance-contract.json`.

## Production boundary

This model does not convert, execute, or replace `runtime.Value`. It does not
claim that Go closures, maps, AST definitions, iterator callbacks, native
functions, or Futures can already cross the ABI without loss. It provides the
oracle against which such a binding can now be tested.

The manifest identity changes because layout and host-lifetime declarations are
part of ABI compatibility. The existing shadow-flow evidence was regenerated;
its graph counts and image sizes remain unchanged.

## Next recommendation

Complete **`shared-value-heap-go-binding-conformance`** as a test-only adapter.

Snapshot and reconstruct representative current Go runtime graphs through the
generated layouts, then compare both interpreters' ordinary results with the
round-tripped graph. Cover identity/alias mutation, recursive closures and
binding updates, definitions and nominal wrappers, errors, iterators, package
bindings, native boundaries, and Future-held results. Adapter dispatch must be
generic by runtime kind/layout and must not add named-container or nominal
lowering rules.

Why: the abstract contract is now internally consistent, but the remaining
risk is semantic impedance between concrete Go pointer/map/function shapes and
the contract. A test-only binding exposes missing fields or ownership rules
before production values depend on them. Production migration should be
considered only if unlike graphs round-trip without identity loss, graph
conversion ambiguity, or interpreter disagreement.

Exclusions remain: no production migration, foreign heap, cgo runtime,
JIT/backend dependency, executable memory, benchmark branch, named-container
or non-primitive nominal special case, or WASM.

## Reproduction

```sh
just bench-semantic-abi-heap-contract-check
just bench-architecture-budget-check
just bench-evidence-ledger --check
```
