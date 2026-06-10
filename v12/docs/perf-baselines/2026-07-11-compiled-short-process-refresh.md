# Compiled short-process refresh and ReverseComplement attribution (2026-07-11)

## Purpose

Validate the only large ratios remaining in the fair pinned Able-versus-Go
scorecard before treating either text/byte or float/control work as a compiler
candidate. Three-run values were insufficient because the Go processes take
only 17--69 ms.

## Matched 30-process result

Both sides used the unchanged external inputs, CPU `2`, `GOMEMLIMIT=1GiB`,
`GOGC=50`, `GOMAXPROCS=1`, a 45-second guard, and the suite verifier after
every normal process. Every row is 30/30 verified.

| Application | Able compiled | Pinned Go | Able/Go |
| --- | ---: | ---: | ---: |
| I-Before-E | 0.1290 s | 0.0669 s | 1.93x |
| Mandelbrot | 0.1467 s | 0.0511 s | 2.87x |
| ReverseComplement | 0.1227 s | 0.0161 s | 7.62x |

The ratios are therefore persistent process-level signals, not three-run timer
noise. They do not, by themselves, identify an implementation boundary.

## ReverseComplement phase attribution

The current full ReverseComplement binary was built once, then launched 30
times from the real FASTA suite directory with
`ABLE_GO_PHASE_CPU_PROFILE_DIR`; every output passed the public verifier. The
merged `main.cpu.pprof` has zero CPU samples. Its generated `main` work is too
short for Go's CPU sampling interval, while the ordinary verified process still
takes 122.7 ms.

This rules out the apparent byte-complement loop as a profile-backed candidate:
the measured gap is overwhelmingly outside the profiled generated `main`
phase. It cannot be compared to I-Before-E's material `read_lines`, string
validation/search, and environment-switch descendants or to Base64/JSON's
codec/parser controls. Mandelbrot's existing current-source profile remains
its independent generated float pixel loop.

## Decision

Keep no compiler, runtime, or `able-stdlib` performance change. In particular,
do not add a ReverseComplement, FASTA, byte-complement, string/byte, or
float-expression fast path. The high-repeat numbers establish which
applications deserve investigation; the phase result establishes that no
common generated-main leaf has yet been found.

## Next recommendation

Audit generated-binary bootstrap and initialization across I-Before-E,
ReverseComplement, and the Go-competitive JSON control before looking for
another compiler candidate. Why: ReverseComplement's full-process gap lies
outside `main`, so optimizing its source body would optimize noise. The work
entails bounded cold-start/init attribution and generated-import/registration
comparison under the same verifier lane, then a change only if one concrete
initialization boundary repeats materially in both misses and remains neutral
on JSON.
