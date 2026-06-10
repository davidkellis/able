# Compiled control-flow current-source invalidation gate

Date: 2026-07-20

## Decision

Keep no compiler, generated-runtime, canonical-stdlib, benchmark, or fixture
performance change from this gate. Fresh current-source profiles reproduce the
same split as the retained July 16 compiled loop and result-ABI gates: Fib,
Matrix Multiply, Sudoku Masks, and TapeLang Alphabet do not share one material
primitive lowering, call-boundary, control-flow, or allocation descendant.

The earlier recommendation overlooked that these exact shapes already had
broad profile coverage. Rebuilding them was still a useful invalidation check
because substantial compiler work has landed since those reports. Current
evidence does not invalidate their decisions, so no source candidate was
admitted and no application- or nominal-type-specific lowering was added.

## Current artifacts and timing

One current `ablec` binary was preserved and used to generate all four strict
compiled applications against canonical `../able-stdlib`. Each application
binary then supplied five verifier-backed normal processes under
`GOMAXPROCS=1`, `GOGC=50`, CPU 0 affinity, and a 55-second process cap.
References were independently refreshed with five verifier-backed Go
processes under the same affinity.

| Application | Able compiled mean | Go mean | Able / Go | Current observation |
| --- | ---: | ---: | ---: | --- |
| Fib | 3.3740 s | 3.2615 s | 1.0345x | inside the 95% target in this cohort |
| Matrix Multiply | 1.0380 s | 1.0453 s | 0.9930x | inside the 95% target in this cohort |
| Sudoku Masks | 9.5660 s | 0.5536 s | 17.2796x | large miss |
| TapeLang Alphabet | 3.7000 s | 1.9817 s | 1.8671x | material miss |

All 40 timed outputs passed their public verifiers and each application had
one stable output hash. The Fib and Matrix observations are not promoted as a
new scorecard classification from one targeted cohort; they establish that a
candidate must guard these current near-target results rather than assuming
all four programs are misses.

The preserved compiler SHA-256 was
`114ead1081b975893ed867343dc259db73860998adba7e3b66aadebf9d1b85e2`.
The application binary SHA-256 values were:

- Fib: `95cdad538d3e5772de36e2c360264f626f309f9361a193e269dfb5d87d2cd2f1`;
- Matrix Multiply: `19dd883713f1a2d88981488d6a1a8257c573630147f276062cb6e64fed406038`;
- Sudoku Masks: `3296864a52a1921bdb919401a37f83ee2b75aa228c7f630f61f7b5c550eb5a79`;
  and
- TapeLang Alphabet:
  `d330b68ea17c9d87cf824b056d773db20851936fd4015d1a3791d1f33c52231a`.

## Main-only CPU attribution

Each current binary supplied one verified CPU-only main-phase profile. The
profile process excludes bootstrap CPU and preserves the ordinary public
workload.

| Application | Main samples | Exact current owner |
| --- | ---: | --- |
| Fib | 3.26 s | generated `fib` is 99.69% flat and 100% cumulative |
| Matrix Multiply | 1.10 s | generated `matmul` is 97.27% flat and 99.09% cumulative |
| Sudoku Masks | 8.50 s | generated `find_best_empty` is 87.29% cumulative; `runtime.growslice` is 37.41% and `runtime.mallocgc` is 51.53% cumulative |
| TapeLang Alphabet | 3.53 s | generated `execute` is 70.25% flat; `Tape.inc` is 22.66%, `Tape.get` 5.67%, and `Tape.move` 1.42% flat |

Fib has no generated call wrapper or runtime helper below its direct recursive
body. Matrix's nested primitive `f64` loop is likewise already direct. Tape's
work is direct checked `i32` mutation, flat Array access, and dispatch in the
application body; allocation and GC do not appear in its CPU profile. Sudoku
is the only allocation-dominated member of the cohort.

## Allocation attribution

Exact main-phase allocation counters completed for the three bounded cases:

| Application | Main allocated bytes | Main allocations | Main GCs |
| --- | ---: | ---: | ---: |
| Fib | 144 | 6 | 0 |
| Matrix Multiply | 32,897,352 | 8,018 | 2 |
| TapeLang Alphabet | 282,552 | 4,274 | 0 |

Exact allocation profiling makes every allocation observable. Sudoku reached
the 55-second diagnostic cap before `main` completed, so that one workload used
Go's ordinary sampled allocation profile instead. The sampled run completed,
passed the public verifier, and reported 2.89 GB and 160,532,091 objects in the
whole process. Generated `find_best_empty` owns 2.85 GB (98.77%) and
160,019,938 objects (99.68%). Its generated body constructs and appends three
integers to a new `Array i32` for every empty cell examined, before testing
whether that position improves the current best.

This is a concrete source-algorithm allocation, not a generated call boundary
or a repeated runtime carrier. General Array semantics cannot be weakened or
special-cased for Sudoku, and Fib, Matrix, and Tape provide no corroborating
descendant.

## Reconciliation

The fresh profiles reproduce the retained evidence:

- the July 16 allocation-light loop gate already separated Tape dispatch,
  Matrix arithmetic, NBody math, and QuickSort recursion;
- the post-result-ABI gate already separated Sudoku search/allocation from
  Binary Trees, K-Nucleotide, and Tape; and
- the Fib attribution already found essentially all samples in its direct
  recursive generated body.

No current profile exposes a new common call wrapper, environment swap,
control object, return matcher, raw primitive conversion, Array helper, or GC
owner in at least three unlike programs. The only two current large misses are
also structurally different. A candidate would therefore be selected by an
application name or source shape rather than by a broadly applicable compiler
wall, which is outside the project rules.

## Verification and cleanup

- one current compiler and four strict compiled binaries built successfully;
- 20/20 Able timing processes and 20/20 Go reference processes passed their
  public verifiers;
- 4/4 CPU-profile processes passed their public verifiers;
- exact main allocation counters completed for Fib, Matrix, and Tape;
- the sampled Sudoku allocation process completed and passed its verifier;
- the exact Sudoku diagnostic stopped at its 55-second cap without being
  extended past the repository limit; and
- no WASM work was performed.

Generated packages, binaries, profiles, and logs were temporary and are
removed after extracting this record.

## Next recommendation

Perform a source- and algorithm-equivalence reconciliation for Sudoku Masks
and TapeLang Alphabet before using either ratio to select another compiler
change.

Why: these are the two clear current compiled misses, but their profiles end in
application bodies rather than a shared compiler/runtime leaf. Sudoku's Able
source allocates a three-element Array for every examined empty cell while the
Go reference returns four scalar values. Tape's Able source executes a flat
parallel-array jump program, while the Go reference builds recursive operation
trees. Equal output and input do not by themselves prove equal work. Until the
representations and operation counts are reconciled, treating either ratio as
pure compiler overhead risks optimizing a benchmark-source mismatch.

What it entails: compare Able, Go, Python, and Ruby sources for input passes,
solver/dispatch strategy, temporary-result representation, integer semantics,
and output buffering; add bounded counters where static comparison is
ambiguous; and document which differences are required by each language versus
accidental benchmark drift. If a fair common algorithm can be expressed in all
four languages, update every implementation and verifier contract together,
then collect two independent five-run averaged cohorts. Only a remaining
primitive or shared nominal-lowering cost repeated in at least three unlike
applications should advance to a compiler candidate. Do not tune either Able
program alone, add named nominal lowering, or begin WASM.
