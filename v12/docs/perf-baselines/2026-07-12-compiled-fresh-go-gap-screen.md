# Compiled fresh-Go gap screen (2026-07-12)

## Method

This bounded screen refreshed three previously material-looking,
single-threaded compiler rows against rebuilt Go 1.26.4 references. Every Go
and compiled Able row used ten CPU-2-pinned, verifier-backed processes with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second cap.

| Benchmark | Fresh Go mean (s) | Compiled Able mean (s) | Able / fresh Go | Valid runs |
| --- | ---: | ---: | ---: | --- |
| Base64 | 2.9549 | 2.8590 | 0.97x | 10/10 each |
| Monte Carlo Pi | 0.2389 | 0.2600 | 1.09x | 10/10 each |
| PiDigits | 1.3861 | 1.5230 | 1.10x | 10/10 each |

Monte Carlo Pi's reference output is statistically verified and therefore
nondeterministic across runs; its suite verifier, not stdout equality, is the
correctness contract. The other rows have stable verified output hashes.

## Decision

Make no compiler, VM, or `able-stdlib` change. Base64 is already faster than
its fresh Go reference. Monte Carlo Pi and PiDigits narrowly miss the 95%
compiler target, but they are different float/integer and arbitrary-precision
workloads and expose no repeated material gap. Profiling either would invite a
single-benchmark lowering.

The ordinary `bench_compare_external` Markdown still displays older stored Go
rows (`2.2000s`, `0.1800s`, and `0.7400s` respectively). Those values were not
used for this decision; the fresh Go-reference report is authoritative here.

## Next recommendation

Add generic support for consuming the existing fresh Go-reference JSON in
`bench_compare_external`, alongside its Python/Ruby reference input. Why:
fair current Able/Go scorecards should not require manual ratio calculation or
accidentally present stored-reference ratios as current. The work entails a
schema adapter, report tests, and a fresh-reference end-to-end guard; it is
measurement infrastructure only, not a runtime optimization.
