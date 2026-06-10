# Compiled Array Range-Proof Admission Closure

Date: 2026-07-25

Decision: retain no compiler or runtime code.

## Question

The preceding compute/control cohort found explicit native Array guards in
Tapelang Alphabet and Matrix Multiply, but not in Fib. This tranche asked
whether one exact, generally provable Array range fact was material in at
least three unlike current strict compiled target misses.

The admission rule deliberately required both:

1. the same proof obligation in at least three unlike applications; and
2. repeated CPU evidence that the exact guard, rather than merely its
   surrounding Array access or loop, was material in all three.

Static guard counts alone were not sufficient.

## Coverage-Wide Census

The retained compiler at
`8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e`
generated all 46 current complete strict target misses successfully with
`-no-fallbacks -main`.

The census considered only application-owned compiled functions and methods;
stdlib package bodies were excluded. It found 278 explicit native Array
guards:

- 157 strict guards of the form `index < 0 || index >= len`;
- 121 safe/nullable guards of the form `index >= 0 && index < len`;
- 30 applications with at least one application-owned guard;
- 16 applications with no application-owned guard.

The largest static counts were:

| Application | Safe | Strict | Total |
| --- | ---: | ---: | ---: |
| NBody | 0 | 52 | 52 |
| Sudoku Masks | 0 | 35 | 35 |
| Binary Event Log | 24 | 2 | 26 |
| Concurrent Policy Callbacks | 16 | 2 | 18 |
| Concurrent Transform Chain | 12 | 2 | 14 |
| Tapelang Alphabet | 12 | 1 | 13 |
| Concurrent Stencil Reduction | 10 | 2 | 12 |
| Dependency Plan | 0 | 10 | 10 |
| K-Nucleotide | 9 | 2 | 11 |
| Matrix Multiply | 0 | 8 | 8 |

The 16 zero-guard applications were Await Channel Mux, Dependency Wave
Validation, Distance Field, Fib, Fixed Width 128, Future Await Race, Future
Pipeline, Inventory Reconciliation, Mandelbrot, Mutex Await Journal, Mutex
Work Queue, Option/Result Config, Rational Series, RMS Norm, Unicode Scalar
Pipeline, and Wide Integer Records.

The complete per-application counts are in the companion JSON record.

## Dynamic Protocol

Fresh strict executables were built for the four static leaders relevant to
the native numeric cohort and for two simpler same-Array controls:

- NBody;
- Sudoku Masks;
- Tapelang Alphabet;
- Matrix Multiply;
- Array Slice Window;
- Reverse Complement.

Every executable passed public-output verification. Every dependency graph
contained 96 packages and omitted `able/interpreter-go/pkg/interpreter`.
Matrix Multiply and Tapelang Alphabet reproduced the exact `compiled.go`
hashes from the preceding compute/control record, so their generated
application source did not drift between tranches.

Runs used:

- CPU 5 via `taskset`;
- `GOMAXPROCS=1`;
- `GOMEMLIMIT=1GiB`;
- `GOGC=50`;
- phase CPU profiles via `ABLE_GO_PHASE_CPU_PROFILE_DIR`;
- a 60-second per-process timeout.

The four long applications received ten verified profile runs each. The two
short same-Array controls received thirty each to compensate for profiler
sampling granularity.

| Application | Verified runs | Profiled wall mean | Range |
| --- | ---: | ---: | ---: |
| NBody | 10 | 115.600 ms | 105-125 ms |
| Sudoku Masks | 10 | 1,901.500 ms | 1,806-2,026 ms |
| Tapelang Alphabet | 10 | 4,155.100 ms | 3,784-4,834 ms |
| Matrix Multiply | 10 | 1,153.200 ms | 1,105-1,231 ms |
| Array Slice Window | 30 | 53.800 ms | 51-62 ms |
| Reverse Complement | 30 | 62.067 ms | 55-77 ms |

All 100 profiled processes verified.

## Exact Proof Classification

### NBody

The merged profile contained 610 ms of samples. `advance` accounted for
510 ms flat and 610 ms cumulative. A `mass[j]` guard alone accounted for
120 ms, so range machinery is material.

However, `n` is derived from `x.len()` while `mass`, `y`, `z`, `vx`, `vy`,
and `vz` are independently supplied parameters. Eliminating the material
guards requires an interprocedural equality or minimum-length contract across
sibling Arrays. It is not a same-Array local length proof.

### Sudoku Masks

The merged profile contained 18.35 seconds of samples.
`find_best_empty` accounted for 5.27 seconds flat and 16.90 seconds
cumulative. Array length/index work is distributed through the routine, while
the explicit condition lines are mostly individually cold.

The proof requires fixed minimum lengths for recursively supplied board and
mask parameters, minimum lengths for nested board rows, and coordinate facts
preserved across recursive calls and `square_index`. This is an
interprocedural fixed-shape proof, distinct from NBody's sibling-array length
equality.

### Tapelang Alphabet

The merged profile contained 40.86 seconds of samples. `execute` accounted for
25.47 seconds flat and 40.84 seconds cumulative; `Tape.inc` accounted for
11.38 seconds flat.

The safe guard around `program.Kinds[ip]` accounted for 6.38 seconds flat.
The loop has a local upper fact, `ip < program_len`, where `program_len` came
from `Kinds.len()`. A complete proof still needs `ip >= 0`, because `ip` is
assigned from `program.Jumps`. The other material accesses require equality
among the parallel `Kinds`, `Values`, and `Jumps` Arrays or the mutable
`Tape.pos`/`Tape.cells` invariant.

This is neither NBody's parameter-array contract nor Sudoku's recursive
fixed-shape contract.

### Matrix Multiply

The merged profile contained 10.99 seconds of samples; `matmul` accounted for
10.88 seconds flat. All six application `matmul` guard-condition lines had
zero samples even though the surrounding native loads and tight arithmetic
loop were hot.

`n` comes from `a.len()`, while the inner work needs `b` and `c` outer lengths
and dynamically selected row lengths to equal `n`. The only simple local
same-Array outer access was cold. The material relation is square nested
Arrays plus sibling-matrix equality, not a local loop fact.

### Same-Array Controls

Array Slice Window has a direct `index < batch.len()` safe-access shape, but
the exact guard and load received zero samples across 30 profiles.

Reverse Complement has same-Array reverse traversal, but its same-Array guard
lines received only 20 ms in aggregate across 30 profiles. Its table lookup
also depends on a separate byte/table-size construction relation.

These controls show that the locally provable shape exists, but not that it is
broadly material.

## Admission Decision

No candidate passed.

| Candidate proof | Material applications | Admission |
| --- | --- | --- |
| Same-Array local loop bound | Tapelang upper bound; controls cold or negligible | reject: fewer than three, and Tapelang still needs a nonnegative jump invariant |
| Sibling parameter Arrays have equal/minimum length | NBody | reject: one exact owner |
| Recursive fixed-shape Arrays and rows | Sudoku | reject: one exact owner |
| Parallel program Arrays plus jump/Tape invariants | Tapelang | reject: one exact owner |
| Square nested Arrays and sibling matrices | Matrix | reject: one exact owner |

Calling all five “Array bounds checks” would erase the exact semantic
distinctions that make them safe. Their common parent would be a broad
interprocedural relational, alias, mutation, and effect proof system. The
current evidence does not establish one repeated material relation that would
justify that architectural change.

Therefore:

- no compiler or runtime code was changed;
- no named-container, application, benchmark, or observed-input rule was
  introduced;
- no semantic-test or 20-cohort A/B gate was warranted;
- no stdlib, interpreter, VM, language, dependency, or WASM change was made.

## Next

Run a post-interpreter-package-cut fixed-cost decomposition across at least six
unlike short strict compiled target misses and their equivalent Go programs.
Use repeated alternating verified launches to separate OS/process startup, Go
package initialization, generated registration/bootstrap, and application
work. Advance a candidate only when the exact fixed-cost owner repeats in at
least three unlike applications.

This is next because current short rows still take roughly 24-52 ms in Able
versus 3.8-5.3 ms in Go, while the phase CPU profiles for short programs
capture little application work. The earlier startup/registration work
predates the retained interpreter-package cut, so the residual fixed-cost
shape must be re-measured rather than inferred. Closing it is important
because native hot-loop parity cannot make ordinary short compiled tools match
Go while a shared launch cost dominates their total runtime.
