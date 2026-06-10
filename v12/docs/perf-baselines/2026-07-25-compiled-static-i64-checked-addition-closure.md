# Compiled static-i64 checked-addition closure

Date: 2026-07-25

## Decision

Retain no compiler or runtime change, and close arithmetic-helper
microbranches.

The current main-phase profiles identified checked signed addition as the
last exact compiler arithmetic leaf material in three unlike strict
applications. A general primitive-only candidate lowered statically known
`i64` addition to one native Go addition plus a sign-bit overflow test, while
leaving dynamic widths and every narrower integer path unchanged.

The candidate was semantically correct and reached all three admitted
applications. It improved Concurrent Scene Tiles and Array Slice Window in
both independent cohorts, but Sensor Calibration reversed from a 2.974%
improvement to a 2.550% regression and was neutral when pooled. The explicit
gate rejects the candidate if any admitted row is neutral or regresses, so the
production and test prototype was completely removed.

Machine-readable results are in
`2026-07-25-compiled-static-i64-checked-addition-closure.json`.

## Scope and controls

The material cohort and controls were frozen before implementation:

| application | role | static-i64 checked-add calls | principal coverage |
| --- | --- | ---: | --- |
| Concurrent Scene Tiles | admitted | 53 | concurrency, nominal values, interfaces, callbacks |
| Array Slice Window | admitted | 5 | native Array copies and integer loops |
| Sensor Calibration | admitted | 29 | file/text parsing, nominal values, numeric methods |
| Option/Result Config | negative/regression control | 5 | generic unions, errors, callbacks |
| Monte Carlo Pi | zero-reach, Go-competitive control | 0 | hot `i32` arithmetic and loops |

Monte Carlo Pi was selected as the current Go-competitive arithmetic control:
its compiled Able implementation is faster than its equivalent Go
implementation, but its hot additions are statically `i32`. The corrected
candidate therefore could not affect its generated or linked code through the
new static-i64 path.

One current compiler binary generated all baseline modules:

```text
8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e
```

The authoritative corrected candidate compiler was:

```text
a177a698dc84e0bbf14deba5b70c23cd6565ea69a5b554b5885e28c53f895162
```

Every application was generated with `--no-fallbacks`. Every baseline and
candidate dependency graph contains 96 packages and omits
`able/interpreter-go/pkg/interpreter`. Equivalent Go 1.26 applications came
from the public benchmark suites.

## Candidate

For statically known signed `i64` addition only, the candidate emitted:

```go
result := int64(uint64(a) + uint64(b))
if ((a ^ result) & (b ^ result)) < 0 {
    return 0, __able_raise_overflow(node)
}
return result, nil
```

This is a general primitive lowering rule. It did not name an application,
stdlib container, nominal type, or benchmark. The existing
`__able_checked_add_signed(a, b, width, node)` implementation remained
byte-for-byte unchanged for dynamic-width calls, and the i8/i16/i32,
subtraction, multiplication, diagnostic, and control-transfer paths were
unchanged.

Generated assembly for the new helper's successful path was a native `LEAQ`,
two `XORQ` instructions, `TESTQ`, and a conditional jump. Only the overflow
edge called `__able_raise_overflow`; it contained no dynamic-width branch.

The candidate source guard required the static call, preserved the old dynamic
helper, and rejected delegation from the dynamic helper. Exact execution
covered both `MaxInt64` and `MinInt64`, and the existing overflow fixture
verified the error edge.

## Excluded exploratory candidate

The first exploratory build also delegated the 64-bit branch inside the
dynamic-width helper to the new helper. That contaminated the Monte Carlo
control because its statically `i32` calls still traversed the reshaped
dynamic helper. Its timing cohorts are excluded from the retention decision.

The candidate was corrected and rebuilt before authoritative measurement.
Normalized Monte Carlo disassembly then showed no difference in
`__able_checked_add_signed`, no linked `__able_checked_add_i64` symbol, and
zero static-i64 calls. This makes the final control causal rather than merely
nominal.

## Repeated baseline/candidate/Go result

Each independent cohort contains 60 verified baseline/candidate/Go triplets
per application after 12 warmup triplets. All six execution orders repeat ten
times. The pooled result contains 120 accepted processes per binary and
application: 1,800 accepted timing processes total.

| application | cohort 1 change | cohort 2 change | pooled baseline ms | pooled candidate ms | pooled Go ms | pooled change | Able/Go baseline → candidate |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Scene Tiles | -1.009% | -1.150% | 4.488281 | 4.439375 | 1.137173 | -1.090% | 3.9469x → 3.9039x |
| Array Slice Window | -6.801% | -3.231% | 6.390354 | 6.072644 | 1.678722 | -4.972% | 3.8067x → 3.6174x |
| Option/Result Config | -1.116% | +0.877% | 16.314534 | 16.309221 | 1.218656 | -0.033% | 13.3873x → 13.3830x |
| Sensor Calibration | -2.974% | +2.550% | 10.347833 | 10.332102 | 2.624448 | -0.152% | 3.9429x → 3.9369x |
| Monte Carlo Pi | +0.114% | -0.769% | 145.844282 | 145.362775 | 233.006632 | -0.330% | 0.6259x → 0.6239x |

Concurrent and Array repeat improvements, but the gate is deliberately
stricter than “some rows improve.” Sensor is an admitted material row and
changes sign, with only a 0.152% pooled difference. Option/Result also changes
sign, and the exact zero-reach Monte Carlo control remains neutral. The
candidate therefore does not clear the broad unlike-program bar.

This result also closes further local checked-arithmetic instruction shaving.
Checked multiplication and checked addition have now both reduced their local
helper work without producing a repeatable improvement across every admitted
application. The remaining 3.6x-13.4x gaps are not plausibly explained by one
more arithmetic-helper microbranch.

## Verification

- 10/10 baseline Able/Go smoke processes passed public verifiers.
- 5/5 corrected-candidate smoke processes passed public verifiers.
- 1,800/1,800 authoritative timing processes verified after 360/360 warmups.
- The excluded exploratory candidate had a separate 1,800 accepted processes
  and 360 warmups; none contribute to the decision.
- All baseline and corrected-candidate strict dependency graphs contain 96
  packages and omit the interpreter.
- Generated-source and assembly checks proved the intended static-i64 path.
- The Monte Carlo control proved zero static and linked reach after correction.
- Signed-helper, Monte Carlo recurrence, exact addition-overflow fixture, and
  `go test ./cmd/ablec` pass after prototype removal.
- Candidate helper, lowering branch, and candidate-only tests/files are absent.
- `git diff --check` passes for restored compiler files.

No compiler, runtime, interpreter, VM, stdlib, language, dependency, or WASM
change was retained.

## Recommendation

Next run a broad compiled runtime-value boundary census across at least six
unlike current strict target misses, using the existing typed-boundary
telemetry together with main-only CPU and allocation profiles.

Why: the two final shared arithmetic leaves became cheaper locally but failed
the unlike-program wall-time gate. The project goal is native Go carriers and
minimal compiled/runtime crossing; the next evidence should therefore locate
an actual repeated `runtime.Value` conversion, boxing, or dispatch boundary
rather than another machine-instruction microbranch.

What it entails: select a feature-diverse cohort spanning concurrency and
interfaces, unions/errors, file/text parsing, graph/iterator work, nominal
records, and Array-heavy code; generate every row strictly with
`--no-fallbacks` and typed-boundary telemetry; correlate exact boundary event
counts and generated callers with repeated main-only CPU/allocation profiles;
then admit one general boundary only if the same exact owner is material in
at least three unlike applications. Preserve native primitive/Array carriers,
shared nominal encoding, semantics, and interpreter-free dependency graphs.

Why it is important: this directly tests where compiled Able still boxes or
enters runtime services despite fallback-free generation. Removing a shared,
measured boundary can improve whole feature families and move the remaining
compiled applications toward native Go performance without prohibited
application or nominal-type special cases. Do not begin WASM work.
