# Compiled allocation-light loop/lowering gate

Date: 2026-07-16

## Decision

Keep no compiler, runtime, canonical-stdlib, or benchmark optimization from
this gate. TapeLang Alphabet, Matrix Multiply, NBody, and the recursive
QuickSort guard all spend most of their measured time in generated code, but
they do not share one material concrete lowering cost across three unlike
applications.

The generated primitive Array accesses are already slice-backed, their length
helper is inlined, and Go eliminates the secondary native bounds checks behind
the emitted Able guards. Removing the remaining Able checks would require
relational length/range facts which the compiler does not have. No Tape,
matrix, NBody, QuickSort, named-container, or source-pattern rule was added.

## Cohort and method

The four current sources were rebuilt against canonical `../able-stdlib` with
`--no-fallbacks`. They deliberately cover different static shapes:

- TapeLang: mutable flat `Array i32`, direct methods, and an opcode loop;
- Matrix Multiply: nested `Array f64` and a cubic dot-product loop;
- NBody: seven flat `Array f64` values, updates, and imported math calls; and
- QuickSort: recursive control flow and mutable `Array i32` indexing.

Normal runs used `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, a 55-second
process timeout, and each catalog working directory, arguments, and public
Ruby verifier. Volatile rows received additional independent processes and all
samples were retained in their arithmetic mean.

| Application | Processes | Mean | Median | CV | Existing Go mean | Able / Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| TapeLang Alphabet | 15 | 4.668 s | 4.283 s | 23.72% | 1.8406 s | 2.536x |
| Matrix Multiply | 10 | 1.279 s | 1.267 s | 5.85% | 0.9282 s | 1.377x |
| NBody | 5 | 0.421 s | 0.420 s | 1.89% | 0.0306 s | 13.749x |
| QuickSort guard | 5 | 2.017 s | 1.997 s | 1.39% | 2.3469 s | 0.859x |

All 35 normal outputs passed their verifiers and each application produced one
stable SHA-256. TapeLang's first batch coincided with substantial workstation
load; five additional samples reduced but did not remove the variance, so the
full 15-process mean is reported rather than selecting the later faster band.
The existing Go rows are not same-session A/B controls and are used only to
show current product distance.

## Merged CPU evidence

Short programs used multiple independent CPU-only profile processes so their
merged profile had useful sample depth: three TapeLang, three Matrix, five
NBody, and three QuickSort processes. All 14 profile outputs passed the public
verifiers.

| Application | Merged samples | Material current owner |
| --- | ---: | --- |
| TapeLang | 10.10 s | generated `execute` 99.90% cumulative; `Tape.inc` 26.44%, inline `Tape.get` 6.83%, and `Tape.move` 0.99% |
| Matrix Multiply | 3.18 s | generated `matmul` 96.54% cumulative; the inner direct `f64` load/multiply/add loop owns nearly all flat time |
| NBody | 1.67 s | generated stdlib `sqrt` 64.67% cumulative, inline `abs` 17.96%, imported `sqrt` entry 87.43%, and `advance` 100% |
| QuickSort | 5.14 s | recursive `quicksort` 74.32% cumulative, `swap` 12.65%, and checked signed multiply in decimal parsing 14.98% |

Allocation/GC does not form a material repeated wall: it is absent from the
TapeLang and NBody top sets, about 2.5% cumulative in Matrix, and about 1.4% in
QuickSort. The cohort therefore does test generated execution rather than
merely rediscovering the allocation-heavy branch closed by the preceding gate.

## Inlining and bounds-check diagnostics

Each generated package was rebuilt separately with Go inlining/escape and SSA
bounds-check diagnostics.

- Every primitive specialization of `__able_slice_len` is inlineable at Go
  cost 3, and every inspected hot call was inlined.
- Tape's `get` method is inlineable; `inc` is not because its required checked
  arithmetic, write growth, and control propagation exceed the Go budget.
- Matrix's hot loop is already one direct generated body. Failure to inline
  the large `matmul` function does not place a call in its inner loop.
- NBody's small `abs` body inlines. Its Newton-iteration `sqrt` body and the
  cross-package entry wrapper do not; the entry wrapper must preserve package
  environment semantics.
- Recursive QuickSort and its checked, mutating `swap` body do not inline.

Go's `ssa/check_bce` output contains no remaining `IsInBounds` diagnostic at
the profiled Array-access lines in any of the four applications. The emitted
condition, for example `index >= 0 && index < len(elements)`, proves the
following Go slice access safe. Replacing `__able_slice_len` with spelling
`len` directly would therefore not remove a call or a native bounds check.

The explicit Able guards cannot generally be dropped from these sources:

- TapeLang bounds `ip` by `program.kinds` but also indexes independently stored
  `values` and `jumps` arrays;
- Matrix rows are dynamically obtained arrays whose length equality with `n`
  is a program invariant;
- NBody obtains `n` from one of seven independently supplied arrays; and
- QuickSort's moving indices remain safe because of algorithm/pivot
  invariants, not a simple counted-loop fact.

A legal removal needs interprocedural relational length/range facts and error
semantics, not a local source-shape shortcut. The current compiler has no such
proof, so no range candidate was built.

## Generality reconciliation

The concrete residuals divide below the shared generated-code parent:

- checked integer mutation and direct methods in TapeLang;
- nested `f64` memory traffic and arithmetic in Matrix;
- a pure-Able iterative math function plus cross-package entry in NBody; and
- recursion, swaps, and parser-only checked multiplication in QuickSort.

QuickSort already beats its matched Go row, and its parser multiply does not
repeat in the numeric applications. NBody's `sqrt` is the largest isolated
language-level lead, but no other application in this cohort uses it. The
cross-package wrapper is likewise not present in the Matrix/Tape/Quick hot
paths. Neither meets this gate's three-unlike-program selection rule.

## Verification and cleanup

- Four strict no-fallback binaries built successfully.
- 35/35 normal processes passed their external verifiers.
- 14/14 CPU-profile processes passed their external verifiers.
- Go inlining/escape and bounds-check diagnostic builds completed for all four
  generated packages inside the one-minute limit.
- Focused compiler Array/static-loop checks and both compiler CLI builds pass
  with no production candidate present.
- Generated packages, binaries, outputs, profiles, and diagnostic logs are
  temporary and removed after this record.

## Next recommendation

Expand primitive-math benchmark coverage before changing `able.math.sqrt`.
Add two independent verifier-backed applications that use square root in
different algorithms, with matched Go, Python, and Ruby references, then
evaluate the canonical stdlib implementation in both compiled and bytecode
execution.

Why: NBody's `sqrt`/`abs` path owns about 82.6% of its current CPU samples and
is the clearest remaining operation-level gap, but NBody is presently the only
full application exercising `math.sqrt`. Optimizing it immediately would fail
the project's cross-program evidence bar. Adding distinct distance/geometry
and norm/statistics workloads closes a real feature-coverage hole and can show
whether the pure-Able Newton algorithm, cross-package entry ABI, or both are
generally material. A generic primitive-math improvement could benefit the
compiler and both interpreters rather than one benchmark.

What it entails: add Able and matched reference programs plus public verifiers;
record repeated compiled, bytecode, Go, Python, and Ruby baselines; capture
bounded warmed bytecode and generated-main profiles; and test any candidate
through `sqrt`, `hypot`, zero, sub-unit, very large, domain-error, NaN, and
infinity semantics. Prefer an improvement in canonical `../able-stdlib` that
reduces iterations for every runtime. Add or change a kernel primitive bridge
only if the broader evidence shows that pure Able cannot approach the product
targets, and then update the v12 spec and all runtime implementations together.
Continue to defer WASM.
