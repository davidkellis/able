# Bytecode Frame-local Array Ownership: Diagnostic Boundary

## Status

There is no production bytecode frame-local Array reclamation path and no
selected implementation candidate. The package-private observer in
`bytecode_vm_array_ownership.go` is opt-in test/profile instrumentation: it
tracks pointer identities and provenance, but never calls
`ArrayStoreReleaseArrayValueLease`.

Ordinary bytecode execution has neither an observer nor ownership contexts.
The lowering metadata and observer contexts activate only in the bounded
diagnostic harness after a program reaches a canonical Array-creation
boundary. They do not change Array lifetime, allocation, or release behavior.

The generic backing-store contract is separate and live: ArrayStore reclaims
state after its proven last owner releases its lease. See
[array-handle-lifetime.md](./array-handle-lifetime.md). This record concerns a
possible earlier release of a wrapper proven local to an active VM frame; it
must never weaken ArrayStore's alias semantics.

## Why release is not enabled

Able assignment, argument passing, returns, aggregates, closures, Futures,
extern results, and unknown calls can preserve an Array alias. Clearing a VM
slot, or observing that one wrapper has left a callee, does not prove that it
is the last observable owner. Releasing from either fact can invalidate an
Array still reachable elsewhere.

The observer confirms that generic nested Array and Matrix temporary graphs
can be frame-local, while flat `Array.map` results conservatively cross a
closure boundary. That is useful attribution evidence, not a release proof.
An opt-in explicit-release experiment preserved correctness in its controls
but was only 1.9% faster for `base64` and 8.9% slower for `i_before_e`.
It was removed. Two unrelated rows, and especially a regression in one, do
not meet the performance policy's three-application shared-leaf gate.

No production sidecar, cleanup call, ArrayStore lease change, benchmark
exception, or named-container rule is authorized from this evidence.

## Diagnostic contract

When an eligible test/profile run enables it, the observer:

- records only Arrays created at verified canonical boundaries: Array
  literals, canonical `Array.new`, and canonical kernel Array materialization;
- follows a returned tracked pointer through inline-frame and detached-return
  boundaries, including an otherwise empty caller context;
- records conservative escapes through public returns, environments,
  aggregates, closures, Futures, unknown calls, and writes into a borrowed
  Array; and
- classifies a non-returned, non-escaped pointer as frame-local or
  error-unwound without releasing its lease.

Program finalization marks canonical creation and capture/dynamic/spawn/
aggregate barriers only to select this diagnostic observer. The metadata is
not an ownership proof and must not add work to ordinary non-profile VMs.
Pointer identity is intentional: aliases share an `ArrayValue` wrapper in the
interpreter representation, while ArrayStore remains the only backing-state
owner mechanism.

Focused coverage proves canonical creation detection, returned-value adoption,
empty-caller transfer, public and explicit returns, escape classification,
and error unwind without release. The normal-loader hook is bounded and
opt-in through `ABLE_BENCH_RUNTIME_ARRAY_OWNERSHIP_OUT`; it is not a CLI or
language API.

## Future selection rule

Do not resume the former sidecar design merely because a local graph is
visible. It may be reconsidered only after a material language/application
change produces the same concrete, non-nominal leaf in at least three unlike
verifier-backed applications, with a safe end-to-end alias/escape proof and
no broad regression. That is the mandatory gate in
[performance-competitiveness-vision.md](./performance-competitiveness-vision.md).

The prior sidecar shape, transfer/barrier model, profile counts, and rejected
release experiment remain in
[the historical record](./bytecode-frame-array-ownership-historical.md). They
are evidence for such a future investigation, not an assignment.
