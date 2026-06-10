# Bytecode external threshold-control gate — 2026-07-13

## Decision

Keep bytecode performance report-only. Five interleaved JSON and PIDigits
pairs are clear inside controls against **both** fresh Python and Ruby
references, but they are evidence that the end-to-end interpreter can meet its
target, not a reason to specialize the VM for JSON or BigInt programs.

No bytecode VM, tree-walker, compiler, runtime, canonical-stdlib, or benchmark
source change is selected by this measurement work.

## Protocol

Each of five pairs first runs freshly verifier-backed Python 3.14 and Ruby 4.0
reference programs, then runs freshly launched Able bytecode applications for
the same JSON and PIDigits inputs. Every timed process is pinned to CPU 15,
uses `GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`, has a 45-second cap, and has its
Able output checked with the canonical Ruby verifier. CLI build/setup work is
outside the timed application process.

The two controls have deliberately different execution shapes:

| Control | Family |
| --- | --- |
| JSON | file/text decode, JSON field extraction, typed Array values |
| PIDigits | native BigInt arithmetic, integer conversion, formatted output |

The retained aggregate is
`2026-07-13-bytecode-json-pidigits-threshold-protocol.{json,md}`. Temporary
per-pair sources under
`v12/tmp/perf/2026-07-13-bytecode-json-pidigits-threshold-protocol` remain
cleanup-eligible.

## Guard-band classification

The raw target is `1 / 0.95 = 1.0526x` of **each** reference. The existing 21%
report-only guard remains valid: JSON/Ruby has the widest new half spread at
17.39%, below the measured 20.04% maximum that set the guard.

- inside requires every Python and Ruby pair at or below `0.8316x`;
- outside requires every pair for at least one reference at or above `1.2737x`;
- otherwise the result is boundary.

| Application | Python ratios | Ruby ratios | Classification |
| --- | --- | --- | --- |
| JSON | 0.36x, 0.28x, 0.36x, 0.30x, 0.33x | 0.53x, 0.42x, 0.53x, 0.37x, 0.49x | inside |
| PIDigits | 0.71x, 0.58x, 0.60x, 0.59x, 0.58x | 0.27x, 0.24x, 0.21x, 0.23x, 0.23x | inside |

The generic `bench_external_threshold_controls` evidence check now supports
multiple required references. It classifies a bytecode control inside only when
every retained pair clears both languages; it cannot hide a slow Ruby result
behind a fast Python result, or vice versa. The check validates retained report
provenance in the ordinary test lane but never runs timings or imposes a
commit-performance threshold.

## Consequence and next evidence

The evidence now has two unlike clear bytecode inside controls, so stop adding
pass controls merely to increase their count. The next step is profile-led:
collect bounded one-process bytecode CPU profiles for unlike target misses such
as Base64, Monte Carlo Pi, and I-Before-E, then compare concrete descendants.
Why: controls establish that broad bytecode passes are possible, but only a
repeated cost in independently slow applications can justify a generally
applicable VM optimization. If no leaf repeats, retain no code change.
