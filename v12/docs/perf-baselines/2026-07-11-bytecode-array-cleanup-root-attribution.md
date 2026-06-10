# Bytecode Array Cleanup Root Attribution

## Decision

Keep no bytecode-runtime, compiler, canonical-stdlib, or benchmark-source
change. The surviving ArrayStore handles are not retained by the inspected VM
frame, lookup, scope, iterator, or primitive-cache roots. They are unreachable
leased wrappers waiting for Go's asynchronous
`runtime.AddCleanup` callbacks. That is a real generic lifetime problem for
allocation-heavy nested Array programs, but it does not yet make an eager
release change safe: Able arrays can alias, escape through calls, be returned,
or be captured.

## Method

A temporary opt-in test-only probe ran the three existing diagnostic controls,
pinned to CPU 2 with `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. At each
explicit print marker it forced three collections, selected every still-leased
ArrayStore handle, and followed it through:

- current and suspended VM slot frames and stacks;
- iterator, loop, ensure, and transient-scope state;
- global, active, and hot lexical lookup values plus their environments;
- call-name cache callees/receivers; and
- the `f64` primitive-array and matrix-row caches.

The harness was removed after the one-program-per-process runs. It did not
change CLI behavior. The normal output-checked controls and their hashes remain
recorded in the preceding live-ownership attribution. Root JSON artifacts are
retained under `v12/tmp/live-array-ownership-2026-07-11/` as
`nested-roots.json`, `matrix-roots.json`, and `array-map-roots.json`.

## Results

| Control / marker | Leased states | Handles reachable from VM roots | Unreachable leased handles |
| --- | ---: | ---: | ---: |
| Nested `i32`, rounds 1--6 and 8 | 25 | 0 | 25 |
| Nested `i32`, round 7 | 29 | 0 | 29 |
| Matrix `f64`, round 1 | 100 | 75 | 25 |
| Matrix `f64`, round 2 | 179 | 75 | 104 |
| Matrix `f64`, round 3 | 175 | 75 | 100 |
| Flat Array map, round 1 | 2 | 2 | 0 |
| Flat Array map, rounds 2--6 | 3 | 2 | 1 |

The three Matrix slot roots are the current iteration's `a`, `b`, and `d`
matrices: each reaches one dynamic outer Array and 24 `f64` rows, explaining
the 75 expected live handles. No `f64` cache entry held a handle. The remaining
25--104 Matrix handles are therefore stale rather than cache-owned. The nested
`round(...)` driver has no active Array slot after its call returns, yet every
one of its 25 leased handles remains. The flat guard has only one stale mapped
Array after its first round.

This rules out the previously plausible VM lookup/frame/cache reference path.
The ArrayStore lease ledger intentionally uses token-only `runtime.AddCleanup`
callbacks, whose contract permits callbacks to run arbitrarily later than the
object becoming unreachable. Three forced GCs do not form a deterministic
release boundary, and nested graphs amplify that delay. The old cache-clear
candidate could not address this class of retention, consistent with its
rejected readings.

## Next recommendation

Map Array ownership escape boundaries before testing eager release. Add a
temporary allocation-to-escape diagnostic for bytecode Array values that
classifies each creation as returned, captured, passed to an unknown call,
stored outside its frame, or proven frame-local; include nested aggregates so a
row placed in a local outer Array is tracked with that outer value. Why: exact
root attribution shows cleanup latency is shared and material, but explicit
release without conservative ownership/escape evidence could free an aliased
or returned Array and violate language semantics. The work should use the two
nested controls plus the flat guard, preserve normal outputs, and authorize an
implementation only if a meaningful, shared set is proven frame-local without
new named-container or benchmark-specific rules.
