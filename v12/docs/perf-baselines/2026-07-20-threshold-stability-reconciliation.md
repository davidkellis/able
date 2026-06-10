# Target-threshold stability reconciliation

Date: 2026-07-20

## Decision

Classify compiled Base64 as an established target guard, bytecode Base64 as a
variance-sensitive miss, and compiled Monte Carlo Pi as a volatile target
crossing. Retain no compiler, bytecode VM, runtime, parser, canonical-stdlib,
benchmark, fixture, language, nominal-lowering, or WASM performance change.

The complete promoted scorecard remains the authoritative current snapshot:
5/41 compiled and 3/34 selected bytecode rows meet their product targets. This
partial three-row reconciliation does not replace it. Across independent
cohorts, the conservative established guard set remains four compiled rows
(Binary Trees, QuickSort, Base64, and JSON) and two bytecode rows (JSON and
PiDigits).

## Measurement contract

The promoted five-run scorecard supplies cohort 1. Two new cohorts used fresh
matched Go, Python, and Ruby references and fresh current Able executables.
The order was reversed between cohorts: cohort 2 measured Go then Python/Ruby
references before compiled Base64, bytecode Base64, and compiled Monte Carlo;
cohort 3 measured Ruby/Python then reversed Go and Able application order.

Every lane used the catalog's serial CPU-0 contract from pool `0-15`, a
55-second process cap, `GOMEMLIMIT=1GiB`, `GOGC=50`, canonical external
stdlib, benchmark working directory, arguments, source-root policy, and public
Ruby verifier. Builds and reference preparation remained outside timed
launches. All 45 Able samples and all 60 required reference samples completed
and verified. Arithmetic means retain every sample.

The focused cohort-1 JSON files are mechanical single-row projections of the
promoted comparison reports; they do not duplicate measurements. The variance
artifact validates exactly five successful verifier-backed Able/reference
runs in every source report and pools fifteen samples per lane.

## Results

The product ratio budget is `1.052632`. Pooled ratios divide the fifteen-sample
Able mean by the corresponding fifteen-sample reference mean.

| Application/mode | Cohort ratios | Able mean | Limiting reference mean | Pooled ratio | Classification |
| --- | --- | ---: | ---: | ---: | --- |
| Base64 compiled | 1.028, 1.000, 0.974x Go | 2.5587 s | 2.5565 s Go | 1.001x | established meet |
| Base64 bytecode | 1.048, 1.206, 1.136x Ruby | 2.8680 s | 2.5443 s Ruby | 1.127x | variance-sensitive miss |
| Monte Carlo Pi compiled | 0.927, 1.158, 0.950x Go | 0.2173 s | 0.2163 s Go | 1.005x | volatile; not an established guard |

Compiled Base64 meets in every cohort and has only 2.69% cohort-ratio CV, so
it is durable enough to guard future compiler candidates. Bytecode Base64
beats Python in all three cohorts, but the product target requires both
interpreters; it misses Ruby in two cohorts and by 12.72% in the pooled ratio.
Monte Carlo's pooled mean meets, but its cohort ratios span 0.927-1.158 with
12.57% CV. That flip reproduces its earlier threshold volatility.

## Frontier reconciliation

No implementation group reopens:

- compiled Base64 remains in `compiled-target-guards`;
- bytecode Base64 remains in the closed `bytecode-byte-output` group because
  the retained general raw-u8 path already removed the shared primitive-array
  leaf and the residual Base64/FASTA/Reverse owners diverge; and
- compiled Monte Carlo remains in the closed `compiled-float-numeric` group,
  where prior raw-float/carrier candidates failed broad wall-time guards.

The generated frontier intentionally continues to report the complete current
snapshot's 8 meets and 67 misses. The cross-cohort established classification
is a candidate-admission guard layered over that snapshot, not a partial
scorecard rewrite.

## Artifacts and checks

- `2026-07-20-threshold-stability-c{1,2,3}-*.json` retain every Able and
  reference sample used here;
- `2026-07-20-threshold-stability-variance.{json,md}` records pooled mean,
  median, range, standard deviation, and coefficient of variation;
- all nine comparison inputs pass `bench_variance_report --require-runs 5`;
- the regenerated cross-mode frontier passes its exact check;
- the promoted full scoreboard remains structurally valid; and
- `git diff --check` passes.

## Next recommendation

Completed by
`2026-07-20-performance-stability-manifest.md`. The schema-2 frontier now
validates a versioned stability manifest, reports eight snapshot meets versus
six established guards, and retains the complete cohort/source evidence for
every selected crossing without changing raw scorecard status.
