# Bytecode Array Escape Attribution

## Decision

Keep no bytecode-runtime, compiler, canonical-stdlib, or benchmark-source
change. The temporary probe identifies a material shared set of nested Array
graphs that is frame-local at bytecode return, but it does not make an
unconditional eager release safe. A runtime Array pointer may still be aliased
through a return, closure, environment assignment, or arbitrary callable;
today's lease is attached to the wrapper rather than every language-level
reference.

## Method

The temporary one-process test hook observed only two VM boundaries:

- every inline-call return, before its callee slot frame was cleared; and
- operands to generic call paths, conservatively marking those handles as
  `passed_unknown_call`.

At return it excluded parameter-slot graphs as borrowed caller values and
classified non-parameter Array graphs as returned, stored outside the frame,
captured, passed through a generic call, or frame-local. Dynamic outer Arrays
were followed through their contained values so a nested graph is classified as
a whole. Canonical Array get/new/slot-member operations were treated as local
kernel operations; generic dispatch was conservatively retained as unknown.

The three existing output-checked controls ran pinned to CPU 2 with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second guard. The probe
and its temporary VM hooks were removed after collection. JSON artifacts remain
under `v12/tmp/live-array-ownership-2026-07-11/` as `*-escape.json`.

## Results

| Control | Return events | Frame-local graphs | Returned graphs | Stored/captured/unknown local graphs |
| --- | ---: | ---: | ---: | ---: |
| Nested `i32` `round` | 8 | 8 × 25 handles | 0 | 0 |
| Matrix `f64` `build_matrix` | 6 | 0 | 6 × 25 handles | 0 |
| Matrix `f64` `matmul` | 3 | 3 × 25 handles (`c`) | 3 × 25 handles (`d`) | 0 |
| Flat Array-map `build` / `map` | 7 | 0 | 7 × 1 handle | 0 |

The probe saw generic call operands only in callers: the Matrix `main` passed
the pre-existing `a` and `b` graphs into `matmul`, and flat `main` passed its
source Array through `map` and `sum`. Those are parameter borrows, not local
graphs in the receiving frames. Neither nested local graph was returned,
captured, stored outside its frame, nor passed through an unknown call.

Together with the preceding root attribution, this explains the asymmetric
retention: the generic nested driver leaves a full dead frame-local graph every
round, Matrix leaves its temporary transpose graph in addition to current and
returned matrices, while flat map's result graph intentionally returns to
`main`. It also shows why clearing VM caches cannot solve the problem.

## Next recommendation

Design a conservative bytecode escape/liveness plan for an explicit
frame-owned Array graph class, then prototype it behind focused semantic tests.
The class must be defined by language-level flow rather than a source name:
created in the frame, reachable only from local slots/local aggregates, never
returned or captured, never stored in an environment/struct/global, and never
passed to an unknown callable. It must transfer ownership when a graph crosses
one of those boundaries. Why: the controls prove a repeated, generic candidate
set, but a return-time scan alone cannot establish ownership for arbitrary Able
programs. Only after that plan is represented in bytecode metadata should a
release candidate be tested against nested, flat, alias/return/capture, and
external benchmark guards.
