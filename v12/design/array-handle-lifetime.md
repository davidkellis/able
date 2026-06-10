# Array-handle Lifetime: Active ArrayStore Contract

## Status

Generic ArrayStore last-owner reclamation is live. An Array storage handle is
an implementation capability, not a language resource that an arbitrary
wrapper may free. Backing state remains available while any tracked,
handle-bearing owner can expose the Array and is removed only when the final
lease releases it.

This applies uniformly to dynamic and mono Array storage, canonical kernel
Array structs, `ArrayValue` views, bridge conversion, owned extern results,
and generated compiler Array carriers. It is a language/kernel boundary rule,
not a Base64, buffer-pool, source-shape, or named-container optimization.

## Required behavior

1. Every live Able alias observes the same mutations, capacity changes, and
   mono/dynamic representation transitions.
2. A handle-bearing owner records its lease before becoming observable and
   releases or transfers that lease exactly once when it is replaced or lost.
3. The final release removes backing state, kind, revision, and relevant hot
   metadata under the ArrayStore lock. A released handle cannot silently
   recreate state through an old view or cache entry.
4. Raw-handle replacement is atomic at the visible owner boundary; compiler
   carrier moves have no transient zero-owner interval.
5. Temporary kernel-hook borrows and weak interpreter/cache entries are not
   owners. They must validate a live handle before use.
6. Tree-walker, bytecode, compiled carriers, bridge conversion, and Futures
   preserve the same alias and lifetime semantics.

The contract preserves automatic memory management. Able exposes no manual
free operation, and wrapper collection is not an alternate semantic rule for
backing-state reclamation.

## Implemented ownership boundary

`ArrayStoreTrackArrayValueLease` and the generic
`ArrayStoreTrackLeaseOwner` attach a lease to pointer-owned Array wrappers and
compiler carriers. `ArrayStoreUpdateLease(..., 0)` and
`ArrayStoreReleaseArrayValueLease` release the owner; final release deletes
the storage and metadata. `ArrayStoreMoveLease` transfers a compiler-created
temporary without an observable zero-owner gap.

`ArrayStoreEnsureHandle` only validates an existing state. The explicit host
boundary for a never-published handle is `ArrayStoreAdoptHandle`; no ordinary
runtime path may revive a released handle. Token-only cleanup uses an
address-plus-generation key, so it does not retain an owner or let a delayed
callback affect a later wrapper at the same address.

Canonical runtime Array structs use an Array-only native sidecar rather than
growing every struct. Generated compiler carrier clone and move paths manage
independent and transferred leases respectively. The interpreter's
`arraysByHandle` tracker is weak, and bytecode handle caches are borrowed
metadata: both must miss after a release.

## Evidence and verification

Focused runtime coverage proves final release for every dynamic and mono
storage kind, alias survival until final release, explicit host adoption,
stale-handle rejection, cache/revision invalidation, token-generation
isolation, concurrent updates, and atomic moves. Interpreter and compiler
coverage proves canonical `storage_handle` assignment and the observable
copy/argument/return/mutation alias result `10 30 20` across all Go execution
modes.

The bounded retention probe also reached zero handles, revisions, states, and
backing bytes after Base64, string split/join, iterator collect, and numeric
Array-map program scopes completed. Those measurements validate this generic
lifetime correction; they are not a new optimization assignment. Exact dated
measurements and the rollout history are retained in
[the historical record](./array-handle-lifetime-historical.md).

## Separate frame-local observer

An Array wrapper can remain unreachable until its token-only cleanup while a
bytecode `main` still runs. The VM's opt-in observer can attribute some such
wrappers as frame-local, but no frame-local release path is enabled. Its
sidecar experiment regressed an unlike workload and lacks the required
three-application shared-leaf evidence. See the active
[frame-local diagnostic boundary](./bytecode-frame-array-ownership.md).

Do not interpret this live ArrayStore contract as permission to revive that
experiment. Any future VM release candidate must satisfy the performance
policy's shared-leaf, alias-proof, and broad-regression gates.

## Non-goals

This contract does not add manual memory management, alter Array copy or
mutation semantics, create a container-specific compiler rule, promise a
collector-cycle deadline, or solve cycles reachable only through their own
ArrayStore state. The latter remains a possible future cycle-collection
question, not a justification for premature release.
