# Bytecode Live Array Ownership Attribution

## Decision

Keep no bytecode-runtime, compiler, canonical-stdlib, or benchmark-source
change. Repeated nested-Array allocation retains prior round graphs while the
same bytecode `main` is active, in both a generic construction/release driver
and an independently shaped scaled MatrixMultiply driver. A flat existing
Array-map control does not share that material shape. The one generic
return-cache candidate did not improve the live plateaus and made one control
worse, so it was removed.

## Method

A temporary test-harness-only probe intercepted explicit diagnostic `print`
markers after each completed round. At every marker it recorded ArrayStore
state, lease-owner counts, and Go heap statistics before and after three forced
collections. It did not alter normal VM or CLI execution and was removed after
collection.

The controls were output-checked through the normal bytecode CLI, pinned to
CPU 2 with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second guard:

| Control | Shape | Completed output | CLI time |
| --- | --- | --- | ---: |
| Generic nested release | eight rounds of 24 `Array i32` rows, each 96 values | 8 deterministic lines; SHA-256 `74cfcd11774a7d8c7e22200615263555d17a93a71410dab096c646b43e738629` | 0.27 s |
| Scaled MatrixMultiply | three rounds of the existing nested `Array f64` matrix construction/transpose/multiply path | 3 deterministic lines; SHA-256 `46a351ceaba4d87bd91ff3120699c9eb9b273047d9148048304cccf3522aa3ea` | 0.19 s |
| Existing flat Array map | six rounds of the existing primitive map path | diagnostic control only | n/a |

Artifacts and JSON snapshots are retained under
`v12/tmp/live-array-ownership-2026-07-11/`.

## Live ownership evidence

| Control | First post-GC marker | Later post-GC marker | After `main` + 3 GCs |
| --- | --- | --- | --- |
| Nested `i32` release | 25 states: 1 dynamic outer + 24 `i32` rows; 25 lease owners | normally returns to 25 after each subsequent 50-state pre-GC peak | 24 `i32` states / 24 owners remain (9,216 direct bytes) |
| Matrix `f64` | 100 states: 4 dynamic outers + 96 `f64` rows; 100 owners | 179 after round 2 and 175 after round 3 | zero states and owners |
| Flat Array map | 2 states after round 1; at most 3 after later rounds | 2--3 states, 150--276 KB direct backing | zero states and owners |

Both nested controls retain graph-sized values after their round calls have
returned, even after forced collections. The flat control retains only its
current source/mapped arrays and clears after `main`. This matches the external
Sudoku observation that a nested-Array workload can become allocation/GC
bound while one `main` is still running, but it does not identify a common
strong reference path precisely enough to change ownership behavior.

## Rejected generic candidate

The candidate cleared per-program lexical scope-lookup values whenever an
inline function returned. Those cache entries can hold local environments and
values, so the change was a plausible generic lifetime correction. It did not
reduce the nested driver's per-round 25-state plateau. It changed its final
post-`main` snapshot to zero, but MatrixMultiply worsened from 175 to 191
post-round states and from zero to 96 `f64` states after `main`. The candidate
was therefore reverted; neither its effect nor its cost is broadly safe.

No source-specific treatment of Sudoku, matrices, masks, nested arrays, or a
nominal container is authorized. No `able-stdlib` change is warranted.

## Next recommendation

Attribute one surviving nested handle to an exact bytecode root before another
candidate. Add a temporary identity-only diagnostic that follows selected
ArrayStore handles through VM slot frames, active lookup entries, and
environment bindings at a print marker in both nested controls. Why: the
ownership pattern is real and repeats, but the rejected cache cleanup proves
that a plausible root is not sufficient evidence. The work entails no runtime
behavior change, output-checked bounded runs, and a candidate only if the same
root category owns the surviving handles in both controls and a flat guard
does not.
