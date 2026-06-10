# Compiled post-registration scorecard refresh

## Decision

Keep no further compiler, interpreter, canonical-stdlib, or benchmark-source
performance change. The retained diagnostic-origin reservation is a small,
shared bootstrap allocation improvement, but the refreshed application results
do not expose another material leaf shared by two independently shaped compiled
misses. In particular, do not add an I-Before-E text path, a Mandelbrot float
path, a ReverseComplement byte path, or a named-container lowering rule.

## Method

The measurement used current source, the canonical external stdlib, CPU `2`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second guard per
process. Every Able and Go output was checked with the benchmark suite's
`verify.rb`.

- `v12/bench_refresh_go_refs` rebuilt the Go 1.26 reference outside timing and
  ran five independent pinned processes for each application.
- `v12/bench_compare_external` ran five independent pinned current Able
  processes in compiled and bytecode modes for each application. Its displayed
  Go columns use the stored corpus and are not used below; the fresh rebuilt Go
  rows are the comparison baseline.
- CPU-only generated-binary phase profiling rebuilt current I-Before-E and
  Mandelbrot binaries once, then merged `main.cpu.pprof` across 30 and 20
  separately verifier-checked launches, respectively. This mode avoids the
  allocation-profiler/forced-GC distortion of the exact allocation mode.

## Fresh verified scorecard

| Application | Compiled Able | Fresh Go 1.26 | Able/Go | Bytecode Able |
| --- | ---: | ---: | ---: | ---: |
| I-Before-E | 0.1440 s | 0.0625 s | 2.30x | 0.5680 s |
| JSON | 0.7520 s | 1.4738 s | 0.51x | 0.9700 s |
| Mandelbrot | 0.1480 s | 0.0498 s | 2.97x | 6.7000 s |
| ReverseComplement | 0.1120 s | 0.0162 s | 6.91x | 6.7740 s |

All 40 Able processes (five for each of the eight application-mode rows) and
all 20 Go rows verified. The bytecode columns are a correctness/non-regression control,
not support for a compiler-only exception. They also show that the large
Mandelbrot and ReverseComplement VM costs must be investigated through the
broader bytecode suite rather than hidden by compiled lowering.

## Current-source phase evidence

| Application | Launches | Main CPU evidence | Consequence |
| --- | ---: | --- | --- |
| I-Before-E | 30/30 verified | 1.05 s merged samples; generated `main` is 84.8% cumulative. Its material descendants are allocation/GC work, `String.len_bytes` (15.2% cumulative), `String.contains` (21.9%), UTF-8 validation, and string search. | Matches the prior text profile; it is not a generic numeric-loop leaf. |
| Mandelbrot | 20/20 verified | 970 ms merged samples; generated `pixel_byte` is 96.9% flat and checked signed shift is 3.1%. | A self-contained float/escape-loop kernel, not a bridge, registration, or string boundary. |
| ReverseComplement | prior 30/30 verified phase runs | Generated `main` had zero samples while the full short process remained slow; the measured cost is before that phase. | Its only sampled common startup boundary was registration, whose diagnostic-node map growth has already been reduced. |

The short bootstrap profiles captured while collecting the two current-source
profiles still place `RegisterIn` above package decode/registration, but have
only 80 ms (I-Before-E) and 40 ms (Mandelbrot) of samples. The earlier exact
allocation audit is the appropriate evidence for that boundary and has already
retained the capacity reservation. These short CPU profiles do not justify a
second registration micro-change.

## Why no candidate is selected

I-Before-E's string/allocation path and Mandelbrot's primitive numeric kernel
share no material helper. ReverseComplement is outside generated `main`, and
the JSON control remains substantially faster than fresh Go. Optimizing any
one observed leaf would therefore make one benchmark faster without evidence
that it improves a broad class of Able applications, contrary to the suite's
selection rule. No `able-stdlib` change is needed.

The transient generated source, binaries, profiles, captured outputs, and
fresh JSON reports consumed about 1.4 GB during analysis and were removed
after this evidence was recorded.

## Next recommendation

Refresh the verifier-backed bytecode comparison against freshly rebuilt Python
and Ruby reference programs for the full, feature-diverse completed suite,
then profile only a concrete leaf that repeats in at least two real bytecode
misses and stays neutral on a nontrivial control. The present five-run bytecode
control confirms correctness but its stored Python/Ruby rows are not a fair
current baseline, and Mandelbrot/ReverseComplement have no usable external
interpreter row. A fresh reference runner plus a broader feature mix will tell
us whether a VM-wide dispatch, allocation, or primitive-value boundary truly
recurs before any runtime work is attempted.
