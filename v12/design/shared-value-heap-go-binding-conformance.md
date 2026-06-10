# Shared value/heap Go binding conformance

Date: 2026-07-22

Status: complete; historical blocker record, resolved by
`shared-value-heap-contract-reconciliation.md`

## Decision

**`hold-production-migration-reconcile-go-binding-gaps`**.

A test-only generic adapter now snapshots current Go `runtime.Value` graphs
into the shared heap contract and reconstructs independent Go graphs. It proves
that identity, cycles, closures, nominal wrappers, structured errors, and host-
held Future results can cross the boundary. It also exposes three representation
gaps that make a production migration unsafe today.

No tree-walker, bytecode VM, compiler, live runtime value, stdlib, benchmark,
application output, dependency, foreign heap, executable memory, or WASM path
imports or executes the adapter.

## Binding shape

The adapter uses generated runtime-kind and layout descriptors. Heap objects
are allocated in two phases: reserve a stable identity, recursively encode
fields, then initialize exactly once. Uninitialized objects cannot resolve or
trace. This permits arbitrary cycles without weakening immutable-layout rules.

Reference identity is memoized for current pointer-backed values and lexical
environments. Environment bindings become explicit BindingCell objects.
Object reconstruction memoizes identities before decoding child edges, so
aliases and cycles remain aliases and cycles rather than copied trees.

Logical Array contents are read through the ordinary generic ArrayStore API
when a tracked handle exists; the adapter does not inspect mono-array kinds or
add a container-specific compiler rule. Host kinds stay in the generated host
registry. Future result/Error cells are registry-held roots and a restored
Future receives a new Go identity with the same visible status and payload.

AST declarations, type expressions, and other immutable program metadata use
test-snapshot metadata indices. Host functions and handles remain opaque host
registry entries by contract. These are not claimed as a portable serialized
format.

## Descriptor correction

The audit found that Package and DynPackage layouts could not reconstruct name,
path, privacy, or public binding names. The generated descriptors now encode
those fields explicitly. No Go-only package sidecar remains. The manifest
identity and the two earlier checked semantic-ABI reports were regenerated;
their graph counts and image sizes are unchanged.

## Passing vectors

- An Array with a shared child and a self-cycle restores both identities.
- Tree-walker and bytecode execution each produce an aliased Array graph that
  round-trips without splitting the alias.
- Tree-walker and bytecode closures each preserve a recursive self binding and
  a captured integer; restored calls return 41.
- One StructInstance identity remains shared through Package, Error, and
  Future roots after restoration.
- Struct definitions, interfaces, union definitions, overloads, type refs,
  bound/partial functions, implementation namespaces, hash maps, and dynamic
  package references have descriptor-driven binding implementations and
  focused compilation coverage.

The checked matrix records 24 exact kinds, one conditional kind, four opaque
host-registry kinds, and two blocked shared kinds across all 31 runtime kinds.

## Blocking gaps

### Wide integers

Small Integer values fit the 64-bit payload. Arbitrary `i128`, `u128`, and
unbounded intermediate values do not. Treating KindInteger as always immediate
provides no traced immutable arena or other owner for those bits. A Go pointer
or test metadata index would violate the backend-neutral cell contract.

### Hasher

Hasher is classified as shared heap and mutable, but its evolving semantic
state is private to `pkg/runtime`. The adapter cannot inspect or restore it.
Making it an opaque host object would move language semantics out of the shared
authority, so that is not an acceptable shortcut.

### Iterator

The current Iterator stores private Go driver and finalizer closures. Those are
behavior, not serializable semantic state. The generated IteratorState layout
can represent values, references, and position, but there is no continuation/
driver identity connecting it to the current closures.

Because these shapes can occur in ordinary graphs, exact fallback could require
lossy graph conversion. Production migration therefore remains closed.

## Next recommendation

Complete **`shared-value-heap-contract-reconciliation`** before any production
value migration.

Define one generic owned representation for wide scalar payloads and teach cell
validation/tracing about its lifetime. Define an inspectable semantic Hasher
state contract. Split Iterator into semantic continuation/state and an explicit
host driver/finalizer boundary, then prove iterator position, close/finalize,
and captured-root behavior. Rerun the same Go binding matrix and admit migration
only if all 31 kinds are exact or explicitly opaque host kinds.

Why: the adapter has reduced the uncertainty to three concrete ownership gaps.
Working on an executor or migrating the 28 currently representable kinds would
either duplicate values at fallback or strand ordinary graphs containing one
of the blocked kinds. Reconciliation fixes the shared authority first and keeps
future performance work broadly applicable.

No timing was run because the adapter is non-production and timing it would not
measure application performance.

## Reproduction

```sh
just bench-semantic-abi-go-binding-check
just bench-architecture-budget-check
just bench-evidence-ledger --check
```
