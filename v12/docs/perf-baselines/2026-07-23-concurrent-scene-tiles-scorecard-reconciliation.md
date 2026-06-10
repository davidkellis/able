# Concurrent Scene Tiles scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_scene_tiles` in both the compiled and bytecode selections.
Retain its source-equivalent six-lane application, exact verifier, two
five-process cohorts per lane, bounded profiles, coverage evidence, and
closure records. Retain no compiler, VM, tree-walker, canonical-stdlib,
language, dependency, named-container, non-primitive nominal, or WASM change.

## Measurement and scorecard

All 50 timed processes verify. Pooled arithmetic means are `0.406s` compiled
Able versus `0.004168457s` Go (`97.398x`), and `0.647s` bytecode Able versus
`0.0747510262s` Python (`8.655x`) and `0.0766163248s` Ruby (`8.445x`).
Compiled Able is the most volatile lane: its two cohort means are `0.436s`
and `0.376s`, a 14.8% difference. Both independent cohorts and their pooled
mean remain clear target misses, so the classification is not sensitive to
workstation noise.

The promoted scorecard has 59 applications, 118 full-status rows, and 111
selected rows: 59 compiled and 52 bytecode. Every selected row has five
successful Able samples and five successful reference samples. The selection
manifest SHA-256 is
`64e61e0689eb5e51ca1e95a778ab792b4ae38a45530135284268f5662ded397f`.

The regenerated performance frontier has eight target meets, 103 misses, five
established guards, zero actionable local groups, and
`180.30052631578948` seconds of summed target excess. The weighted feature
interaction frontier has no zero-depth or depth-one triple and minimum depth
nine. `concurrent_scene_tiles` raises both former minimum-depth interactions:

- concurrency × functions/closures × interface dispatch;
- functions/closures × inherent methods × interface dispatch.

## Profile and candidate gate

Three compiled profiles put `bridge.currentGID` at 96.57% and `runtime.Stack`
at 95.71% cumulative, repeating the closed compiled-concurrency owner. Three
bytecode runtime profiles average `378861306 ns/op`, `64543432 B/op`, and
`744133` allocations/op. The merged bytecode cost is diffuse across
call/member dispatch, binary work, callee-environment setup, allocation,
return completion, cache lookup, and field loading. The 215,064-call trace
shows the hot record, signature, and sample methods already resolving inline
after four cold generic interface resolutions.

No exact new owner is both dominant and separable across unlike applications.
The generality gate therefore admits no production or stdlib change and does
not reopen the currentGID, cache/lock, call-frame, return, raw-integer,
type-match, shared-runtime, foreign-backend, or WASM lanes.

## Closure and architecture reconciliation

Only `compiled-concurrency` and `bytecode-concurrency` changed membership and
evidence. Both were reviewed and advanced together. The ledger contains 21
current closures and zero invalidations; all production and canonical-stdlib
scope fingerprints are unchanged.

The deterministic architecture/ABI dependency chain was regenerated in
order. Its decisions remain unchanged: no current local cross-engine
mechanism, semantic-region tier, native-tier prototype, portable foreign
backend, shared-runtime production migration, or closed-region production
cutover is admitted. The bytecode native proxy now spans 52 applications and
still leaves 41 target misses, while the structural strategy retains zero
concrete admitted routes.

## Verification

- exact output parity across tree-walker, bytecode, compiled Able, Go, Python,
  and Ruby;
- two complete verifier-backed five-process cohorts per runtime lane;
- three compiled and three bytecode profiles plus a bytecode call trace;
- catalog, selection, coverage, operation-depth, matrix, triple, scorecard,
  frontier, closure-ledger, and architecture dependency checks;
- JSON, source-identity, source-line, formatting, and whitespace checks.
- the complete v12 suite through
  `GOMEMLIMIT=1GiB GOGC=50 ./run_all_tests.sh`.

## Next recommendation

Add a tenth materially different portable application using independently
synthesized fixed-point audio voices.

Why: the same two high-value interactions remain the shallowest frontier rows
at depth nine. A deterministic audio-mixing topology can raise both while
avoiding the geometry, tree, graph, queue, pipeline, state-machine, signal,
and callback-batch shapes already represented.

What it entails: four Futures each synthesize an independent integer
fixed-point voice; nominal phase and mix accumulators expose inherent methods;
a user-defined oscillator interface selects distinct waveforms; and captured
envelope callbacks cross that interface boundary per sample. Implement exact
Able, Go, Python, and Ruby versions plus a schedule-independent verifier,
establish six-lane parity, run two five-process cohorts, and profile only after
correctness. Admit a production change only if one generic owner repeats
across unlike applications. Update canonical `able-stdlib` only for a
reusable API or correctness defect, and do not begin WASM work.
