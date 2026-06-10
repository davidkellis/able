# External generality scorecard refresh and selection — 2026-07-11

## Method

The retained [machine-readable scorecard](2026-07-11-external-generality-validation-refresh.json)
and [rendered report](2026-07-11-external-generality-validation-refresh.md)
ran all 15 `generality` workloads once in compiled and bytecode mode against
the checked-in Go, Ruby, and Python results. Able processes ran on CPU `2`
with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second per-process
guard. Public-suite `verify.rb` scripts validated successful stdout outside the
timed process; workloads without a verifier retain their stdout hash but are
explicitly marked `unavailable`.

This is a one-run selection pass, not a claim of sub-percent precision. Timed
out or verifier-unavailable rows are status/correctness evidence only and do
not select an optimization target.

## Scorecard state

| Mode | Completed | Timed out | Failures | Verified successful rows | Within the 95%-speed floor |
| --- | ---: | ---: | ---: | ---: | ---: |
| compiled | 14/15 | 1 | 0 | 11 | 2/12 Go-reference rows |
| bytecode | 10/15 | 5 | 0 | 8 | 3/7 Ruby-reference rows; 3/6 Python-reference rows |

Compiled QuickSort was verified and Go-competitive (`0.98x` Able/Go). Verified
Fib and Monte Carlo Pi were both moderately outside the compiled Go target
(`1.17x` each). The larger compiled gaps differ in workload shape: I-Before-E
is file/text scanning (`4.00x`), PiDigits is bigint arithmetic (`1.92x`),
Mandelbrot is float-heavy control flow (`5.00x`), and ReverseComplement is a
large text/byte path (`19.00x`). BinaryTrees is also constrained by this
single-CPU pass despite its goroutine executor, so its `8.04x` row is not a
fair unconstrained parallel-Go comparison.

For bytecode, successful text/codec (`i_before_e`, `base64`, `json`) and
scalar-RNG (`monte_carlo_pi`) rows miss the Ruby/Python floor, while Fib,
MatrixMultiply, and PiDigits clear the available reference rows. The five
timeouts are distinct workload/status evidence, not ratios. The prior VM
profiling already showed that the text, iterator, and numeric fixture controls
do not share one safe type-match, map, or raw-carrier candidate.

## Decision

Keep no code. The scorecard rules out selecting a text/codec, nominal-union,
Array, float-loop, or named-stdlib optimization from a single large deficit.
Neither `able-stdlib` nor benchmark source changed.

The next bounded investigation is compiled CPU/allocation attribution for the
two *verified*, independently shaped moderate misses: Fib and Monte Carlo Pi.
Use Go-competitive QuickSort as the primary control and Mandelbrot as an
unrelated float/control guard. A compiler candidate is authorized only when a
concrete generated helper or lowering boundary is material in both Fib and
Monte Carlo Pi, preserves the language's generic/nominal lowering rules, and
does not regress either control. MatrixMultiply's unverified external row may
be a secondary profile observation, never the sole selection proof.
