# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
Every Able and fresh-reference component retained exactly 5 successful verifier-backed runs.

| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `array_slice_window` | `compiled` | 15 | median=0.0900, mean=0.1127, range=0.0800, CV=31.01% | go: median=0.0045, mean=0.0048, range=0.0030, CV=18.35% | go: median=21.4634, mean=23.6992, range=15.6719, CV=34.06% |
| `regex_set_audit` | `compiled` | 15 | median=0.1300, mean=0.1287, range=0.1000, CV=24.90% | go: median=0.0052, mean=0.0051, range=0.0017, CV=8.37% | go: median=23.3333, mean=25.5464, range=11.0418, CV=22.88% |
| `regex_stream_audit` | `compiled` | 15 | median=0.1300, mean=0.1400, range=0.0900, CV=19.65% | go: median=0.0046, mean=0.0049, range=0.0016, CV=10.12% | go: median=29.3617, mean=29.0661, range=3.4969, CV=6.08% |

## Inputs

- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-compiled-03-selected.json` — generated `2026-07-17T15:08:30.539598Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-compiled-03-selected.json` — generated `2026-07-17T16:28:22.801910Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-17-regex-carrier-scorecard-compiled.json` — generated `2026-07-17T19:53:50.959584Z`, CPU `0-15`
