# Generality verifier closure (2026-07-12)

## Benchmark-harness changes

The sibling benchmark repository now supplies semantic `verify.rb` scripts for
the two previously unverified generality lanes:

- MatrixMultiply requires exactly one finite numeric result within `1e-6` of
  the shared 1000x1000 calculation, `-95.58358333329998`.
- TapeLang Alphabet compares bytes exactly with the reverse alphabet and its
  trailing newline.

The external Dockerfiles have always run MatrixMultiply with the `1000`
argument, while the fresh local Go reference runner previously supplied no
argument and accidentally used its source default of `100`. The generic
benchmark catalog now passes `1000` for MatrixMultiply to both fresh Go and
Able processes, restoring the same workload without changing an algorithm or
compiler lowering. Positive Go-output and negative bad/truncated-output
preflights pass, and both verifier scripts pass Ruby syntax checks.

## Fresh verified comparison

Go 1.26.4 and compiled Able each ran three CPU-2-pinned processes with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second cap.

| Benchmark | Fresh Go (s) | Compiled Able (s) | Able/Go | Status |
| --- | ---: | ---: | ---: | --- |
| MatrixMultiply | 0.9745 | 1.2000 | 1.23x | verified 3/3 |
| TapeLang Alphabet | 1.8630 | 3.7633 | 2.02x | verified 3/3 |

MatrixMultiply's Go and Able output hashes differ only because Go prints six
fractional digits while Able prints more; the shared numeric verifier accepts
both. TapeLang output hashes are identical.

## Decision

Keep no compiler, VM, runtime, `able-stdlib`, or benchmark-algorithm change.
These rows close the two coverage gaps in the compiled generality ledger, but
they do not create a reusable optimization pair. MatrixMultiply remains the
typed f64 Array triple-loop path already separated from Monte Carlo,
Mandelbrot, and N-body in the numeric profiles. TapeLang is a distinct
program-defined tape interpreter. Do not add a matrix, float, tape, or
source-shape-specific path.

## Next recommendation

Publish one versioned fresh compiled-generalities report using the now
verifier-complete suite, with every runnable row measured under the same
settings and timeouts represented as status. Why: prior current rows were
collected in several bounded scorecards; a single provenance artifact will
make the 95%-of-Go target auditable before the next profile-selection pass. The
work entails fresh Go and compiled Able runs, explicit one-core parallel-lane
annotation, and no implementation candidate unless two material verified rows
share one non-nominal helper.
