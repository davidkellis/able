# Cross-mode primitive Array capacity/backing-growth gate

Date: 2026-07-18

## Decision

Complete the cross-mode Base64, Reverse Complement, FASTA Generation,
Lexical Rollup, and Array Slice Window gate and retain no compiler, generated
runtime, bytecode VM, canonical-stdlib, benchmark, fixture, or language
change. Fresh generated-main and warmed-bytecode CPU/allocation evidence does
not identify one capacity or backing-growth operation that is material in
three unlike applications.

The result confirms the earlier compiled-only rejection on the current source
after the retained `write_all` change and after FASTA joined the portable
suite. No Array, codec, bioinformatics, iterator, slicing, or named-container
lowering rule is justified. WASM remains deferred.

## Source and execution contract

All five Able source hashes exactly match the promoted post-`write_all`
scorecard. That scorecard supplies five verifier-backed normal processes in
each selected mode:

| Application | Compiled mean | Bytecode mean | Runs / status |
| --- | ---: | ---: | --- |
| Base64 | 2.544 s | 2.950 s | 5 + 5 / verified |
| Reverse Complement | 0.102 s | 4.774 s | 5 + 5 / verified |
| FASTA Generation | 0.108 s | 2.102 s | 5 + 5 / verified |
| Lexical Rollup | 0.110 s | 0.560 s | 5 + 5 / verified |
| Array Slice Window | 0.124 s | 0.684 s | 5 + 5 / verified |

The diagnostic processes used canonical `../able-stdlib`, CPU 0,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, catalog working directories and
arguments, and a 55-second per-process cap. Generated binaries used the
default product compiler and monomorphic-Array lowering. Every compiled
profile output passed its public verifier and reproduced its promoted stdout
SHA-256.

Bytecode steady-state measurements loaded and typechecked once per process,
warmed `main()` once, and then measured one call. Five independent processes
per application were averaged; all 25 completed without timeout or runtime
failure. Their exact benchmark counters were:

| Application | Mean runtime | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: |
| Base64 | 2.543 s | 2,201,653,518 | 561 |
| Reverse Complement | 3.892 s | 463,983,102 | 7,876,340 |
| FASTA Generation | 1.768 s | 70,480,446 | 1,817,172 |
| Lexical Rollup | 0.114 s | 2,466,360 | 15,132 |
| Array Slice Window | 0.532 s | 14,192,624 | 422,311 |

These runtime-only processes intentionally discard program output, so their
success is not substituted for verifier evidence. The normal-process table
above remains the correctness authority on the identical source and input
fingerprints.

## Generated-main profiles

CPU-only phase profiling excluded launcher/bootstrap work. Short programs
used repeated independent launches and merged only `main.cpu.pprof`: three
Base64, 30 Reverse Complement, 30 FASTA, 100 Lexical Rollup, and 100 Array
Slice Window launches. Separate allocation-only processes supplied exact main
phase counters without CPU-profiler sampling.

| Application | CPU samples | Main bytes / allocs | Concrete backing owner |
| --- | ---: | ---: | --- |
| Base64 | 7.50 s | 2,201,553,528 / 128 | exact host codec/MD5 result slices; `makeslice` 4.27% cumulative; no `growslice` sample |
| Reverse Complement | 610 ms | 9,314,944 / 64 | generated transform/wrapping append; `growslice` 49.18% cumulative |
| FASTA Generation | 460 ms | 1,058,528 / 445 | pre-sized output driven by `append_random`; no `growslice` sample |
| Lexical Rollup | 1.52 s | 2,392,960 / 30,980 | ordinary `Iterator.collect`; `growslice` 11.84% cumulative |
| Array Slice Window | 670 ms | 1,441,352 / 24,012 | `Array.slice` 40.30% and exact `makeslice` 26.87%; no `growslice` sample |

Only Reverse Complement and Lexical Rollup repeat geometric backing growth.
Base64 allocates exact large codec results. FASTA pre-sizes its output. Array
Slice Window creates 24,002 independently backed shallow copies, as required
by the v12 specification; changing those allocations into views or shared
backing would change mutable Array semantics.

## Warmed bytecode profiles

The final process in each five-process steady-state cohort supplied a bounded
CPU profile: 2.76 s Base64, 3.91 s Reverse Complement, 1.69 s FASTA, 110 ms
Lexical Rollup, and 520 ms Array Slice Window. A separate allocation profile
used rate 1 for Base64, Lexical Rollup, and Array Slice Window, and bounded
rate 4096 sampling for the multi-million-allocation Reverse and FASTA rows.
The five-process benchmark counters above, not profiler-instrumented wall
time, are the allocation totals.

The exact capacity leaf again fails the breadth gate:

- Reverse Complement reaches `ArrayEnsureCapacity` at 0.51% cumulative CPU
  and 9.69% of sampled allocation bytes.
- Lexical Rollup has no material CPU sample in the leaf and attributes only
  1.49% of sampled allocation bytes to it.
- Array Slice Window attributes 16.34% of sampled allocation bytes to
  `ArrayEnsureCapacity`, but this is the one exact backing allocation for each
  pre-reserved independent slice result, not repeated geometric growth.
- Base64 and FASTA have no material `ArrayEnsureCapacity` leaf.

`execCallMemberArraySlot` and Array push machinery recur at broader levels,
but their semantic descendants differ: codec bridge results, primitive reads
and transforms, random-output construction, iterator collection, and required
slice copies. A common VM-dispatch parent cannot authorize a capacity policy
change.

## Reconciliation

No candidate reached the predeclared three-unlike-application admission rule,
so the JSON, Monte Carlo Pi, Word Frequency, and Binary Trees guard gate was
not invoked. Those controls are required for an admitted A/B candidate, not a
reason to trial an unsupported one.

The profiles do expose a different, narrower lead. Reverse Complement and
FASTA attribute 26.97% and 77.23% of sampled allocation bytes respectively to
the exact chain
`invokeExternHostFunction -> toHostValue -> coerceRuntimeToHost ->
toArrayValue -> deoptTypedArrayToDynamic`. Both call the multi-argument
`io_write(IoHandle, Array u8)` extern. Base64 avoids this chain because its
single-argument primitive-byte externs use the existing borrowed/owned fast
invoker. Two bulk-output applications are not enough to admit a change, but
this is a concrete next lead rather than an Array-growth hypothesis.

## Verification and cleanup

- 25/25 warmed bytecode measurement processes completed.
- 263/263 compiled CPU-profile processes passed their public verifiers.
- 5/5 compiled allocation-only processes passed their public verifiers.
- Five bounded bytecode allocation-profile processes completed.
- The promoted five-run normal-process evidence remains verified for all ten
  application/mode rows.
- Raw generated trees, binaries, and profiles were temporary and are removed
  after this record.

## Next recommendation

Run a bytecode multi-argument primitive-byte extern-conversion gate across
Reverse Complement, FASTA Generation, and Mandelbrot, with Base64 as the
existing one-argument borrowed-byte control.

Why: Reverse and FASTA now reproduce the same exact host-coercion and
mono-`u8` deoptimization chain, while Mandelbrot is an unlike numeric/image
producer that naturally reaches the same `write_all` boundary. If Mandelbrot
repeats the leaf materially, the three applications establish a general host
ABI problem rather than a bioinformatics or benchmark pattern.

What it entails: first profile Mandelbrot under the same bounded warmed
contract. If the leaf repeats, inventory every extern signature with a
primitive byte Array in any parameter position and prototype one generic
argument-conversion path that honors the existing explicit
`able_borrowed_bytes(...)` opt-in for multi-argument externs. It must borrow
mono `u8` storage only for the synchronous call lifetime, preserve the copied
fallback for dynamic/aliased values, and remain independent of `io_write` and
application names. Then use repeated verifier-backed A/B cohorts for the
three consumers and guard Base64, JSON, Word Frequency, and Array Slice
Window. This direction is worth testing because it can remove per-byte boxing
for every eligible multi-argument host boundary without changing Array
semantics or nominal lowering.
