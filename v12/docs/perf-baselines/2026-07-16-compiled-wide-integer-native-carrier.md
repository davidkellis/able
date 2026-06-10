# Compiled Wide-Integer Native Carrier (2026-07-16)

## Decision

Retain a compiler/runtime native carrier for the primitive Able `i128` and
`u128` types. The candidate is primitive-wide and does not recognize the
stdlib `Int128`, `UInt128`, `Rational`, or any benchmark/package name.

The previous compiler mapped both primitives to `runtime.Value`. Static
arithmetic consequently allocated `runtime.IntegerValue`, `big.Int`, type-AST,
and interpreter-operation objects. The retained mapping uses two unsigned
64-bit words for each primitive and boxes only at an explicit runtime, host,
interface, or fallback boundary.

## Implementation

- `runtime.Int128` and `runtime.Uint128` preserve one two-word bit pattern.
- Direct lowering covers literals, checked add/subtract/multiply, Euclidean
  integer division/remainder, real division, comparison, bitwise operations,
  signed/unsigned shifts, exponentiation, negation, complement, and primitive
  casts.
- Runtime boundary conversion preserves the Able integer suffix and supports
  value, pooled pointer, and interface-wrapped integer carriers.
- No package-global `big.Int` is initialized. Negative conversion uses the
  two-word magnitude/two's-complement representation, so programs that never
  use `i128/u128` gain no new carrier initialization work.
- Wide lowering is isolated in `generator_wide_integer.go`; it is a primitive
  lowering rule, not a non-primitive nominal exception.

## Selection evidence

All timings are complete-process wall times from independent launches on the
workstation with `GOMAXPROCS=16`. Volatile programs were repeated and averaged.
Every output in a row had one stable SHA-256; Fixed Width also passed its
external Ruby verifier on retained and candidate binaries.

| Program | Preserved baseline | Final candidate | Improvement |
| --- | ---: | ---: | ---: |
| `int128_accumulate_small` | 3.032 s (5) | 0.438 s (5) | 6.92x |
| `uint128_accumulate_small` | 3.112 s (5) | 0.232 s (5) | 13.40x |
| Fixed Width 128 | 6.773 s (3) | 0.359 s (5) | 18.88x |

The original candidate appeared to regress to 16.21 seconds. CPU profiling
showed that 88.05% of the run was interpreter evaluation: the benchmark's
`main` had become uncompilable because wide primitive return types lacked a
generated zero value for control exits. Adding the same generic zero-value
support as every other primitive made the strict `--no-fallbacks` build pass.
The result above therefore measures generated Go, not interpreter fallback.

The full Rational Series external application is an additional positive guard:
its five-run verified mean is 0.375 s versus the preceding scorecard's 2.626 s
(7.00x faster), with the same output hash. Rational benefits through its
ordinary use of primitive `i128`; it has no compiler special case.

## Broad guards

- Full Binary Trees: three verified candidate runs, 12.728 s mean,
  12.685-12.769 s range, and the established output hash. This is comfortably
  within the prior workstation cohorts. Its generated code has no wide
  primitive path.
- `binarytrees_small`: ten runs, 0.0491 s mean and one output hash.
- `sum_u32_small`: thirty complete-process runs, 0.0543 s mean,
  0.0514-0.0686 s range, and one output hash. The candidate has no wide global
  initialization.
- The direct compiled primitive test requires zero fallbacks and executes
  signed/unsigned arithmetic, Euclidean division/remainder, wide shifts,
  arithmetic right shift, and signed-to-unsigned bit-pattern casts.
- The `Int128` and `UInt128` no-bootstrap stdlib fixtures compile with zero
  fallbacks and pass their expected output/error behavior.
- Randomized runtime checks compare 10,000 unsigned divisions and 5,000
  signed/unsigned checked arithmetic pairs with `math/big`.

Focused runtime/compiler, cast, nullable, callback-boundary, extern-boundary,
and whole-module compile-only gates pass. The aggregate short compiler package
run is not green independently of this tranche: it reports the existing Sudoku
source audit's `runtime.Value` index-assignment error carrier and then exceeds
the five-minute package cap. The failing generated function contains no
`i128/u128` carrier operation; it is recorded rather than attributed to this
candidate.

## Next recommendation

Refresh bounded post-carrier profiles for Fixed Width 128, Rational Series,
and the signed/unsigned accumulation fixtures, then refresh the affected
compiled scorecard rows.

Why: the retained carrier removes 85-95% of wall time from its selection
programs, so all pre-carrier profiles are obsolete. Fixed Width remains about
73x the prior Go reference (0.359 s versus 0.0049 s), and Rational remains
about 32x (0.375 s versus 0.0118 s). The next useful work is to identify the
new shared residual rather than optimize the dynamic wall that no longer
exists.

What it entails: capture separate bounded CPU and allocation profiles from
strict no-fallback binaries, reconcile exact generated/runtime descendants
across both signednesses and the unlike Rational/Fixed Width applications, and
advance only a generic primitive-carrier, nominal-construction, or boundary
cost repeated materially across the cohort. Keep unrelated small-integer,
Binary Trees, and startup guards, and do not begin WASM work.
