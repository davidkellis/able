# Compiled Generality Refresh Selection — 2026-07-13

## Decision

Keep no compiler, bridge, runtime, bytecode VM, canonical-stdlib, or benchmark
source change. This fresh full-application selection screen confirms the
current compiled gaps but identifies no new cross-family profile pair. Its
largest misses are the already-profiled, divergent families: K-Nucleotide
text/map conversion, Sudoku recursive search allocation, and N-body numeric
package calls.

This is a one-run-per-application selection report, not a performance-release
claim and not a replacement for the retained scorecard or threshold controls.
It must not select a source change by ratio alone.

## Method

The sixteen current external generality applications were rebuilt and measured
on CPU 15 with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second
cap. Go 1.26 references were rebuilt outside timed processes, then every Able
compiled binary was built from current source and run once. Every successful
Go and Able output passed its benchmark-specific Ruby verifier.

| Benchmark | Fresh Go | Able compiled | Able / Go | Result |
| --- | ---: | ---: | ---: | --- |
| Fib | 3.1776 s | 3.3600 s | 1.06x | verified |
| BinaryTrees | 33.5949 s | 30.7400 s | 0.92x | verified |
| MatrixMultiply | 0.9794 s | 1.2100 s | 1.24x | verified |
| QuickSort | 2.4884 s | 1.9200 s | 0.77x | verified |
| Sudoku | 0.1483 s | cap | n/a | timeout at 45 s |
| Sudoku Masks | 0.5708 s | 11.7800 s | 20.64x | verified |
| I-Before-E | 0.0619 s | 0.1900 s | 3.07x | verified |
| Base64 | 2.5016 s | 2.4700 s | 0.99x | verified |
| JSON | 1.4863 s | 0.8000 s | 0.54x | verified |
| Monte Carlo Pi | 0.2066 s | 0.2200 s | 1.06x | verified |
| PiDigits | 1.1905 s | 1.4400 s | 1.21x | verified |
| Mandelbrot | 0.0494 s | 0.1300 s | 2.63x | verified |
| Reverse Complement | 0.0149 s | 0.1100 s | 7.38x | verified |
| K-Nucleotide | 0.0638 s | 3.7100 s | 58.15x | verified |
| N-body | 0.0332 s | 0.5000 s | 15.06x | verified |
| TapeLang Alphabet | 1.9015 s | 3.8100 s | 2.00x | verified |

The raw JSON and Markdown reports, rebuilt references, generated sources,
binaries, captured outputs, and timing files are under
`v12/tmp/perf/2026-07-13-compiled-generality-refresh/` and its transient
benchmark workdirs. They are cleanup-eligible.

## Selection Analysis

The fresh order matches existing profile evidence rather than exposing a new
shared leaf:

- K-Nucleotide remains the largest completed miss. Its retained profile is
  text/counting map allocation and `bridge.ToInt`/`bridge.ToUint` conversion.
- Sudoku and Sudoku Masks are recursive search/allocation workloads. The prior
  cross-family profile attributes Sudoku Masks almost entirely to generated
  `find_best_empty`, not a general compiler boundary.
- N-body remains numeric generated `sqrt`/`abs` plus a small package
  environment-swap share. The immediately preceding N-body/I-Before-E refresh
  showed that the only overlap is an already-rejected execution-context path.
- Reverse Complement, I-Before-E, Mandelbrot, and TapeLang exercise distinct
  byte/text, file/String, float-escape, and named nominal-method paths. Their
  one-run ratios cannot establish a new common compiler leaf.

The previously retained cross-family profile refresh already tested
K-Nucleotide, N-body, Sudoku Masks, and TapeLang together and found no
material common source descendant. Reprofiling those same known-divergent rows
solely because their one-run ratios are large would optimize ranking noise, not
general Able programs.

## Next Recommendation

Run an equally fresh, verifier-backed bytecode generality selection refresh
against rebuilt Python and Ruby references, then profile only a concrete leaf
that repeats across unlike bytecode misses and has not failed a broad guard.
The compiled selection lane is now current and supplies no eligible candidate;
the independent bytecode target needs current multi-reference evidence before
another VM change. The next tranche should retain raw reports separately, keep
the existing deterministic scoreboard unchanged until a repeatable protocol is
available, and preserve timeouts as unranked rather than treating them as
ratios.
