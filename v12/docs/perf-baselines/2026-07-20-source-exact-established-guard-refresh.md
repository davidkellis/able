# Source-exact established-guard refresh

Date: 2026-07-20

## Decision

Retain compiled Binary Trees, QuickSort, and JSON plus bytecode JSON and
PiDigits as established target guards against the current canonical-stdlib
source tree. Retain no compiler, bytecode VM, runtime, parser,
canonical-stdlib, benchmark, fixture, language, nominal-lowering, or WASM
performance change.

The promoted scorecard remains the current-snapshot authority and was not
rewritten. This tranche replaces only the five guards' older compatible-source
durability evidence with two independent cohorts measured against current
stdlib tree
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
Compiled Base64 was already established against that tree, so all six durable
guards are now source-exact.

## Measurement contract

The promoted current-product scorecard reports supply cohort 1. One fresh
five-run cohort reversed application order: Go references and compiled Able
ran JSON, QuickSort, then Binary Trees; interpreter references and bytecode Able
ran PiDigits then JSON. Each Able lane used fresh matched references, canonical
external stdlib, catalog run directory/arguments/verifier, `GOMEMLIMIT=1GiB`,
`GOGC=50`, a 55-second child-process cap, and CPU pool `0-15` resolved to each
catalog budget. Binary Trees used CPUs `0,1,2,3` with the goroutine executor;
all other lanes used serial CPU 0. Builds remained outside timed launches.

All 25 Able processes and all 35 reference processes completed and verified:
15 compiled Able and 15 Go, plus 10 bytecode Able, 10 Python, and 10 Ruby.
Arithmetic means retain every sample. Mechanical single-row projections of
the promoted reports avoid duplicating cohort-1 measurements. The strict
variance report accepts exactly five successful verifier-backed Able/reference
runs in all seven source reports and pools ten samples per implementation.

## Results

The product ratio budget is `1.052632`. Pooled ratios divide the ten-sample Able
mean by the ten-sample limiting-reference mean.

| Application/mode | Cohort ratios | Able mean | Limiting reference mean | Pooled ratio | Result |
| --- | --- | ---: | ---: | ---: | --- |
| Binary Trees compiled | 0.934, 0.975x Go | 10.5980 s | 11.0975 s Go | 0.955x | established meet |
| QuickSort compiled | 0.757, 0.722x Go | 1.9910 s | 2.6902 s Go | 0.740x | established meet |
| JSON compiled | 0.588, 0.533x Go | 0.8160 s | 1.4551 s Go | 0.561x | established meet |
| JSON bytecode | 0.474, 0.593x Ruby | 0.9070 s | 1.7013 s Ruby | 0.533x | established meet |
| PiDigits bytecode | 0.563, 0.620x Python | 2.4370 s | 4.1240 s Python | 0.591x | established meet |

Every cohort and pooled ratio meets. Bytecode JSON remains below both Ruby and
Python in both cohorts; PiDigits remains below both Python and Ruby. Workstation
spread is retained in the variance artifact and does not alter a classification.

## Frontier reconciliation

`bench-performance-stability.json` now points these five entries to the
source-exact variance artifact, records their current pooled/cohort ratios, and
records the current evidence stdlib tree. The schema-2 frontier remains:

- 8 current snapshot meets and 67 misses;
- 6 established guards: 4 compiled and 2 bytecode;
- 2 non-established snapshot crossings;
- 121.969 seconds aggregate target excess; and
- zero actionable ownership groups.

No ownership disposition reopens. The new measurements strengthen regression
admission evidence; they do not select an application-specific candidate.

## Artifacts and verification

- `2026-07-20-source-exact-guards-stdlib-source-state.json` pins the current
  canonical stdlib source identity;
- `2026-07-20-source-exact-guards-c2-{go-reference,interpreter-reference,
  compiled,bytecode}.{json,md}` retain all fresh measurements;
- `2026-07-20-source-exact-guards-c1-*.json` are mechanical focused projections
  of promoted cohort 1;
- `2026-07-20-source-exact-guards-variance.{json,md}` retains the pooled sample
  statistics and passes `--require-runs 5`;
- the exact frontier and promoted-scoreboard checks pass; and
- JSON validation and `git diff --check` pass.

## Next recommendation

Reconcile the zero-actionable frontier with the still-large product gap through
an architecture-level cross-family ownership audit rather than another local
hot-path experiment.

Why: 67 selected rows still miss by 121.969 aggregate seconds, but every known
family-level leaf is protected or closed by insufficient breadth, divergent
owners, or failed broad candidates. Repeating a closed family profile is
unlikely to produce a general win. The remaining plausible gains are shared
representation and execution boundaries that sit above individual map, regex,
numeric, concurrency, or output algorithms.

What it entails: use the latest exact CPU/allocation evidence from the largest
unlike bytecode families and representative compiled misses to build one
cross-family census of semantic encoding, boxing, dispatch, call, and runtime
initialization costs; compare those costs with emitted Go/reference structure;
identify an architectural candidate only if the same concrete boundary is
material in at least three unlike applications; and define its broad guard set
before implementation. This is a feasibility/admission tranche, not permission
for benchmark-, nominal-container-, or family-specific lowering, and it does
not begin WASM work.
