# Concurrent Tree Folds scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_tree_folds` in both the compiled and bytecode selections.
Retain its source-equivalent six-lane application, exact verifier, two
five-process cohorts per lane, profiles, coverage evidence, and closure
records. Retain no compiler, VM, tree-walker, canonical-stdlib, language,
dependency, named-container, non-primitive nominal, or WASM change.

## Measurement and scorecard

All 50 timed processes verify. Pooled arithmetic means are `0.376s` compiled
Able versus `0.003670908s` Go (`102.427x`), and `0.385s` bytecode Able versus
`0.056498857s` Python (`6.814x`) and `0.054968845s` Ruby (`7.004x`).

The promoted scorecard has 58 applications, 116 full-status rows, and 109
selected rows: 58 compiled and 51 bytecode. Every selected row has five
successful Able samples and five successful reference samples. The selection
manifest SHA-256 is
`faa54b0b27a55f7a9dc0ea4bbc10e856fc1d534279bc8b453d369f405e9c3b76`.

The regenerated performance frontier has eight target meets, 101 misses, five
established guards, zero actionable local groups, and
`179.2777894736842` seconds of summed target excess. The weighted feature
interaction frontier has no zero-depth or depth-one triple and minimum depth
eight. `concurrent_tree_folds` raises both former minimum-depth interactions:

- concurrency × functions/closures × interface dispatch;
- functions/closures × inherent methods × interface dispatch.

## Profile and candidate gate

Three compiled profiles put `bridge.currentGID`/`runtime.Stack` at 97.92%
cumulative, repeating the closed compiled-concurrency owner. Three bytecode
profiles average `188354435 ns/op`, `23205968 B/op`, and `381976`
allocations/op. The bytecode cost is diffuse across call and member dispatch,
binary operations, allocation, synchronization bookkeeping, return
completion, and cache lookup. The trace shows the hot application methods are
already inline or fast after four cold interface resolutions.

No exact new owner is both dominant and separable across unlike applications.
The generality gate therefore admits no production change and does not reopen
the currentGID, cache/lock, call-frame, return, raw-integer, or type-match
lanes.

## Closure and architecture reconciliation

Only `compiled-concurrency` and `bytecode-concurrency` changed membership and
evidence. Both were reviewed and advanced together. The ledger contains 21
current closures and zero invalidations; all production and canonical-stdlib
scope fingerprints are unchanged.

The deterministic architecture/ABI dependency chain was regenerated in
order. Its decisions remain unchanged: no current local cross-engine
mechanism, semantic-region tier, native-tier prototype, portable foreign
backend, shared-runtime semantic-ABI migration, or closed-region production
cutover is admitted.

## Verification

- exact output parity across tree-walker, bytecode, compiled Able, Go, Python,
  and Ruby;
- two complete verifier-backed five-process cohorts per runtime lane;
- three compiled and three bytecode profiles;
- catalog, selection, coverage, operation-depth, matrix, triple, scorecard,
  frontier, closure-ledger, and architecture dependency checks;
- JSON, source-identity, source-line, formatting, and whitespace checks.

## Next recommendation

Add a ninth materially different portable application using independently
rendered fixed-point scene tiles.

Why: the same two high-value interactions remain the shallowest frontier rows
at depth eight. A tiled geometry workload can raise both without repeating
trees, graphs, queues, pipelines, state machines, signals, or callback
batches.

What it entails: four Futures each rasterize an independent integer grid
tile; nominal point and tile-accumulator types expose inherent methods; a
user-defined interface selects signed-distance fields; and captured shading
callbacks cross that interface boundary. Implement exact Able, Go, Python,
and Ruby versions plus a schedule-independent verifier, establish six-lane
parity, run two five-process cohorts, and profile only after correctness.
Admit a production change only if one generic owner repeats across unlike
applications. Update canonical `able-stdlib` only for a reusable API or
correctness defect, and do not begin WASM work.
