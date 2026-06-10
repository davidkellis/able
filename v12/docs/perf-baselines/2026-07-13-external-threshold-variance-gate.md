# External application threshold-variance gate — 2026-07-13

## Decision

Do not add a timing pass/fail threshold yet. The first independent paired-run
screen establishes the measurement spread needed to make that decision, and it
shows that the current one-process launch protocol is not stable enough near a
95% boundary. This is a measurement conclusion, not an optimization candidate:
keep no VM, compiler, bridge, runtime, canonical-stdlib, or benchmark-source
change.

## Method

Each table row is three independent one-process repeats, not a `--runs 3`
average. Every point used CPU 15, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, a 45-second per-process cap, and the benchmark's canonical Ruby
verifier. Each repeat rebuilt the relevant Able binary before timing it, then
used a fresh matching reference result from the same pinned lane.

The applications were selected for unlike execution families rather than their
scoreboard rank:

| Runtime | Applications | Families |
| --- | --- | --- |
| Compiled | N-body, Base64, K-Nucleotide | primitive numeric, byte/host codec, text/map boxing |
| Bytecode | JSON, Monte Carlo Pi, Future Pipeline | host codec/parser, numeric control flow, Future/concurrency |

The report artifacts retain the individual verified samples and derived sample
standard deviation/range:

- `2026-07-13-compiled-threshold-variance.{json,md}`
- `2026-07-13-bytecode-threshold-variance.{json,md}`

The temporary per-repeat source reports remain cleanup-eligible under
`v12/tmp/perf/2026-07-13-scoreboard-variance`; all values used for this gate
are preserved in the checked-in aggregate JSON files.

## Results

| Runtime | Application | Able samples (s) | Ratio samples | Able CV | Ratio CV |
| --- | --- | --- | --- | ---: | ---: |
| Compiled | N-body | 0.50, 0.42, 0.49 | Go: 14.20x, 12.69x, 14.76x | 9.27% | 7.72% |
| Compiled | Base64 | 2.67, 2.47, 2.48 | Go: 1.07x, 0.99x, 0.99x | 4.44% | 4.62% |
| Compiled | K-Nucleotide | 3.71, 5.39, 3.65 | Go: 57.79x, 95.91x, 66.00x | 23.24% | 27.39% |
| Bytecode | JSON | 0.85, 0.85, 0.86 | Python: 0.29x, 0.33x, 0.33x; Ruby: 0.49x, 0.49x, 0.50x | 0.68% | 7.58% / 1.19% |
| Bytecode | Monte Carlo Pi | 2.68, 2.62, 2.66 | Python: 1.77x, 1.66x, 1.74x; Ruby: 1.63x, 1.64x, 1.65x | 1.15% | 3.10% / 0.68% |
| Bytecode | Future Pipeline | 0.50, 0.50, 0.48 | Python: 8.42x, 8.68x, 7.99x; Ruby: 7.33x, 7.60x, 6.72x | 2.34% | 4.19% / 6.22% |

The near-boundary compiled Base64 ratio straddles the 1.0526x target in this
same guarded protocol. A strict threshold would call the first sample a miss
and the latter two passes. K-Nucleotide is too variable for a baseline
threshold, while N-body remains plainly outside the target even at its best
point. The bytecode rows clearly classify JSON as inside and Monte Carlo Pi /
Future Pipeline as outside their targets, but that classification does not
justify a generic VM candidate: their current concrete profile descendants are
different or already rejected.

## Consequence

The refreshed external scoreboard remains a report-first release gate. Do not
promote it to a failing timing threshold from these three samples. A future
threshold proposal needs a declared five-or-more-repeat paired protocol and a
guard band wider than the observed near-boundary spread, then must be checked
on a deliberately mixed application set. It must not be derived from a single
benchmark, a timeout, or a synthesized ratio.

No canonical `able-stdlib` change is needed.
