# Array-handle Lifetime: Historical Rollout

## Status

The active source-backed contract is
[array-handle-lifetime.md](./array-handle-lifetime.md). This record retains
the 2026-07-11 investigation, completed delivery stages, and dated retention
measurements. It does not leave a lifecycle implementation queue.

## Problem that motivated the work

An externally verified bytecode Base64 run retained 2.13 GiB after forced
final collection, almost entirely primitive-byte outputs from extern-call
boundaries. Repeated I-Before-E String-array results showed a distinct
returned-Array shape. The root issue was process-wide ArrayStore backing state
without a complete owner-release protocol, not either application or its
algorithm.

Raw handles could be carried by canonical Array `storage_handle` fields,
`ArrayValue` views, bridge materialization, generated compiler carriers, and
owned extern results. A cleanup on one wrapper could not prove it was the last
observable alias; a tracker cache also could not be allowed to retain wrappers
indefinitely. The original profile is in
`docs/perf-baselines/2026-07-11-external-bytecode-miss-profiles.md`.

## Completed delivery stages

1. `ArrayStoreStatsSnapshot` supplied isolated, lock-consistent dynamic and
   mono backing measurements, with test-only scoped registry setup.
2. `ArrayStoreLease` introduced explicit owner counts for Array values, views,
   bridge conversion, owned u8 extern results, and generated Array carriers.
3. Canonical runtime Array structs gained an Array-only sidecar; shared
   named/positional `storage_handle` assignment transfers its token.
4. Array views and compiler carriers gained non-retaining, generation-safe
   token cleanup, while the interpreter tracker became weak.
5. A raw-handle audit classified every production use as an owner, temporary
   opaque borrow, or metadata. It fixed adoption in interpreter stringification,
   generated Array extraction, and generated formatting.
6. Final-owner release now removes every backing-map, kind, revision, and hot
   entry under the ArrayStore lock. Explicit adoption, stale-handle rejection,
   and atomic move close the remaining release gaps.

Compiler `Array` clone creates a distinct lease for an alias-preserving copy;
move consumes only compiler-created bridge/writeback temporaries and transfers
the token. This is a canonical language/kernel boundary, not special lowering
for a nominal stdlib collection.

## Historical proof matrix

The completed suite covered aliasing through copy/argument/return and
mutation, direct views, owned extern u8 results, raw-handle replacement,
nesting through structs/interfaces/Futures, mono representation changes,
concurrent lease updates, deterministic cleanup generation checks, and final
release for all storage kinds. It also verified compiled bridge/writeback and
tree-walker/bytecode parity.

One-program-per-process retention probes later recorded zero handles,
revisions, states, and direct backing bytes after final collection for Base64,
string split/join, iterator collect, and numeric Array map. The detailed
before/after readings are in
`docs/perf-baselines/2026-07-11-array-store-reclamation.md`.

## Boundary with the rejected VM experiment

The generic last-owner protocol did not imply that a VM may release a wrapper
when an inline frame returns. A release-disabled observer explored the
additional provenance and escape proof that would require. Its later
explicit-release follow-up produced a small Base64 improvement and an
I-Before-E regression, so the code was removed. Its active status and the
full sidecar model are in
[bytecode-frame-array-ownership.md](./bytecode-frame-array-ownership.md) and
its historical companion.

Use this historical evidence to interpret lifecycle regressions, never to
reopen either implementation without the active performance-selection gate.
