# Bytecode i32 Frame Storage Gate

**Date:** 2026-07-13
**Decision:** reject and revert the first frame-storage candidate

## Scope

The candidate used the retained typechecker-backed concrete-`i32` proof to
place proven leaf-frame locals in a frame-owned `[]int32` carrier. It routed
only local loads, stores, i32 arithmetic, branches, and index reads through
that carrier. Calls, returns, captures, collections, dynamic operations, and
errors stayed on the existing materialized `runtime.Value` boundary.

This was a language-wide representation experiment. It contained no
benchmark-name, source-shape, named-container, compiler-lowering, or stdlib
special case.

## Semantic guard

The focused proof, call/return, and typechecker tests passed. An initial
one-process 23-application bytecode status screen produced 17 verifier-backed
successful rows, six known timeout-cap rows, and no verifier failure. That
screen established compatibility only; it was not used as a performance
comparison because it did not share the later pinned control protocol.

## Performance guard

The candidate was confirmed with MatrixMultiply on CPU 15, three runs, a
45-second cap, and `GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`:

| Source state | Verified runs | Average real time |
| --- | ---: | ---: |
| Candidate | 3/3 | 5.2967 s |
| Reverted current source | 3/3 | 4.7867 s |

The candidate was about 10.7% slower. Both states produced the canonical
MatrixMultiply output (SHA-256
`bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c`).

## Decision and next selection gate

All executable typed-frame code was removed. The typechecker proof metadata
remains because it is inert and conservatively falls back when typechecking,
diagnostics, aliases, generics, or assignment facts do not qualify.

Do not retry this frame shape or add a typed call/return ABI. First collect
bounded profiles for MatrixMultiply and an unrelated application with typed
locals. A later candidate is admissible only if both expose the same concrete,
non-nominal descendant below the VM-frame parent; otherwise this representation
direction has no broad evidence and feature-driven application coverage should
take priority.
