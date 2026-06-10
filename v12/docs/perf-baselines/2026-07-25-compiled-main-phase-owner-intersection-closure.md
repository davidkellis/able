# Compiled main-phase owner intersection closure

Date: 2026-07-25

## Decision

Retain no compiler or runtime change.

Fresh main-only CPU profiles across Concurrent Scene Tiles, Array Slice
Window, Option/Result Config, and Sensor Calibration found one exact generated
helper material in three unlike applications:
`__able_checked_mul_signed`. A general prototype replaced the nonnegative
64-bit division-based overflow check with `bits.Mul64` while preserving the
existing sub-64-bit and signed paths.

The prototype made the selected helper cheaper, but two independent
60-triplet baseline/candidate/Go cohorts were neutral or mixed end to end.
Pooled candidate changes ranged from a 3.053% improvement to a 0.241%
regression. Concurrent Scene Tiles was neutral at -0.251%, and Array Slice
Window regressed. The production and test prototype was therefore removed.

Machine-readable results are in
`2026-07-25-compiled-main-phase-owner-intersection-closure.json`.

## Reproducible scope

One current compiler binary generated all four baseline modules:

```text
8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e
```

Every application was generated with `--no-fallbacks`. Baseline and candidate
dependency graphs each contain 96 packages and omit
`able/interpreter-go/pkg/interpreter`. Equivalent Go 1.26 applications were
built from the public benchmark suites.

| application | CPU / executor contract | principal coverage |
| --- | --- | --- |
| Concurrent Scene Tiles | CPUs 5,10,15,11; `GOMAXPROCS=4`; goroutine executor | concurrency, nominal values, interfaces, callbacks |
| Array Slice Window | CPU 5; `GOMAXPROCS=1` | native Array copies and integer loops |
| Option/Result Config | CPU 5; `GOMAXPROCS=1` | generic unions, errors, callbacks |
| Sensor Calibration | CPU 5; `GOMAXPROCS=1`; `readings.txt` | file/text parsing, nominal values, numeric methods |

All processes used `GOMEMLIMIT=1GiB`. Initial Able, candidate, and Go smoke
processes passed the public Ruby verifiers. Every later process matched the
corresponding verifier-proven stdout SHA-256.

## Baseline main-only CPU intersection

The generated phase profiler starts immediately before `RunRegisteredMain`,
excluding initialization and registration. Because the applications run for
only milliseconds, the profiles merge many independent verified processes:

| application | processes | merged samples | checked multiply flat | checked add flat | leading other owner |
| --- | ---: | ---: | ---: | ---: | --- |
| Concurrent Scene Tiles | 200 | 730 ms | 19.18% | 5.48% | render/record/interface work and allocation |
| Array Slice Window | 120 | 720 ms | 22.22% | 12.50% | `Array.slice` and rolling checksum |
| Option/Result Config | 80 | 1.18 s | 0.85% | below reporting floor | bridge integer/type/union allocation |
| Sensor Calibration | 120 | 940 ms | 8.51% | 6.38% | parsing, String split, conversion, allocation |

Checked multiplication is the only exact compiler-owned flat leaf material in
three unlike rows. Shared Go allocator and collector functions are aggregate
parents over different application allocation graphs and do not qualify.

## Exact allocation intersection

Three separate exact allocation processes per application produced stable
main-phase counters:

| application | mean bytes | mean allocations | mean GCs | leading application-owned allocation |
| --- | ---: | ---: | ---: | --- |
| Concurrent Scene Tiles | 1,192,453 | 37,057.33 | 1.00 | render tile, record update, interface sampling |
| Array Slice Window | 1,441,720 | 24,016.00 | 0.00 | 24,002 `Array.slice` objects |
| Option/Result Config | 9,936,648 | 182,896.00 | 3.33 | type-expression metadata, errors, union/struct values |
| Sensor Calibration | 3,609,488 | 57,536.33 | 1.00 | integer conversion, String split, parsing, union/struct values |

No exact application allocation leaf occurs materially in three rows.
Combining these under `mallocgc`, GC scanning, “nominal allocation,” or
“Array work” would erase the distinct source mechanisms. Named Array,
Option/Result, or SensorReading rules remain prohibited.

Allocation-profile subtraction also contains allocations made while
serializing the exact pprof boundary snapshot. The lightweight `MemStats`
deltas above exclude that observer work and are authoritative for totals.

## Candidate

The admitted prototype changed only the general nonnegative 64-bit branch of
`__able_checked_mul_signed`:

- baseline: compare against `max / b`, emitting signed `IDIVQ`;
- candidate: use `bits.Mul64` and reject a nonzero high word or a low word
  above `MaxInt64`, emitting native `MULQ`;
- preserve zero handling, negative operands, `MinInt64`, overflow control,
  diagnostic nodes, and all i8/i16/i32 paths.

The source guard, Park-Miller recurrence execution test, and exact
`06_01_compiler_integer_overflow_mul` fixture passed before measurement.
Candidate strict builds and public verifiers also passed.

Candidate main-only profiles used the same process counts as baseline:

| application | baseline checked-multiply samples | candidate samples |
| --- | ---: | ---: |
| Concurrent Scene Tiles | 140 ms | 70 ms |
| Array Slice Window | 160 ms | 110 ms |
| Option/Result Config | 10 ms | 50 ms |
| Sensor Calibration | 80 ms | 10 ms |

The first, second, and fourth rows confirm that the intended helper became
cheaper. Option/Result's low reach is below stable sample-level attribution
and acts as a negative control.

## Repeated baseline/candidate/Go result

Each independent cohort contains 60 verified triplets after 12 warmup
triplets. All six baseline/candidate/Go execution orders repeat ten times per
cohort. The pooled result therefore contains 120 accepted processes per
binary and application.

| application | cohort 1 candidate change | cohort 2 change | pooled baseline ms | pooled candidate ms | pooled Go ms | pooled change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Scene Tiles | -0.462% | -0.085% | 4.248712 | 4.238061 | 1.188431 | -0.251% |
| Array Slice Window | -0.564% | +0.997% | 5.970702 | 5.985091 | 1.581672 | +0.241% |
| Option/Result Config | +0.980% | -3.269% | 16.832906 | 16.605655 | 1.252460 | -1.350% |
| Sensor Calibration | -2.796% | -3.305% | 10.486278 | 10.166153 | 2.620619 | -3.053% |

Only Sensor repeats a clearly material end-to-end improvement. Concurrent is
neutral, Array changes sign and regresses pooled, and Option/Result changes
sign. This fails the explicit rule to retain no code on mixed or neutral
unlike-program evidence.

The pooled candidate remains 3.566x Go for Concurrent, 3.784x for Array,
13.258x for Option/Result, and 3.879x for Sensor. A cheaper multiplication
overflow check is therefore not the missing general route to Go parity.

## Verification

- 8/8 baseline Able/Go smoke processes passed public verifiers.
- 4/4 candidate smoke processes passed public verifiers.
- 520/520 baseline main-only CPU-profile processes verified.
- 520/520 candidate main-only CPU-profile processes verified.
- 12/12 exact allocation-profile processes verified.
- 1,440/1,440 accepted order-balanced timing processes verified.
- 288/288 timing warmup processes verified.
- Focused signed-helper source and recurrence tests pass.
- Exact multiplication-overflow fixture passes before and after prototype
  removal.
- The prototype fragments are absent and `git diff --check` passes.
- Baseline and candidate strict graphs omit the interpreter package.

No compiler, runtime, interpreter, VM, stdlib, language, dependency, or WASM
change was retained.

## Recommendation

Next run one bounded static-width checked signed-addition gate across
Concurrent Scene Tiles, Array Slice Window, and Sensor Calibration, with
Option/Result plus a current Go-competitive arithmetic application as
negative/regression controls.

Why: `__able_checked_add_signed` is now the next exact compiler leaf material
in three unlike current profiles at 5.48%, 12.50%, and 6.38%. It is distinct
from the rejected multiplication branch and has not yet received this current
three-program end-to-end gate.

What it entails: compare the current dynamic-width helper with a general
static-i64 path using native addition plus a bitwise/sign overflow test;
preserve i8/i16/i32 overflow, error/control propagation, diagnostic nodes, and
dynamic-width calls; add exact min/max and overflow guards; then require at
least twenty perfectly order-balanced baseline/candidate/Go triplets per
application. Retain no code if any admitted row is neutral or regresses.

Why it is important: this is the last exact shared primitive helper exposed by
the current four-program intersection. If it also fails the broad wall-time
gate, close arithmetic-helper microbranches and broaden profiling to a new
feature cohort rather than continuing local instruction shaving. Do not begin
WASM work.
