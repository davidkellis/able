# Per-run scorecard sample retention — 2026-07-15

## Scope

The five-run workstation default must be auditable, not only averaged. The
benchmark harnesses now retain every attempted process in their JSON output:

- `bench_perf` records mode, run number, status, wall/user/system time, GC
  count, bytecode-runtime metrics when available, and output hash;
- Go, Python, and Ruby reference refreshers record run number, status, wall
  time, and verified-output hash; and
- `bench_compare_external` carries the Able samples and any fresh-reference
  samples into its comparison report.

`bench_variance_report` accepts one modern comparison report when a row has at
least two successful embedded Able samples. It reports mean, median, range,
and coefficient of variation for Able and fresh-reference timings. Ratio
variation deliberately remains report-level: independently launched Able and
reference samples are not a paired experiment and must not be paired by array
position.

## Verification

Fresh five-run, canonical-verifier-backed self-checks used the ordinary
`option_result_config` portable application in separate compiled and bytecode
commands, matching the scorecard's bounded per-mode grouping. Both rows had
five successful Able samples and five successful relevant foreign-reference
samples; each calculated Able mean exactly matched the retained row average.

The records were written only under `/tmp` for harness validation and are not
promoted as a new performance baseline or before/after claim. The compiled
and bytecode runtimes, canonical stdlib, workload, target, and timeout did not
change.

## Decision

Keep no runtime or benchmark optimization. Future five-run scorecards can now
show their timing spread without rerunning a workload, while genuine ratio
spread continues to require independent matched source reports. This preserves
the distinction between workstation noise, a profile-attribution result, and
a broadly admitted performance candidate.

The variance reader can also consume two complete scorecard cohorts directly.
It expands their recorded comparison sources and requires identical coverage
and no reused source reports before calculating ratio spread. This keeps the
full 33-application suite auditable as a cohort rather than a hand-selected
set of convenient rows. Its `--require-runs 5` option further distinguishes
fresh candidate-selection evidence from older aggregate reports by requiring
five retained verifier-backed Able and reference samples for every component.
