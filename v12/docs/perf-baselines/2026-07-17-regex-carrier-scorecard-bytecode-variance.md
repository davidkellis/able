# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
Every Able and fresh-reference component retained exactly 5 successful verifier-backed runs.

| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `array_slice_window` | `bytecode` | 15 | median=0.7100, mean=0.7080, range=0.0800, CV=3.51% | python: median=0.0280, mean=0.0302, range=0.0144, CV=13.15%; ruby: median=0.0607, mean=0.0666, range=0.0682, CV=26.07% | python: median=24.2953, mean=23.6227, range=4.7924, CV=10.44%; ruby: median=11.8771, mean=10.8529, range=3.2127, CV=16.91% |
| `regex_set_audit` | `bytecode` | 15 | median=5.2200, mean=5.3000, range=1.4600, CV=9.51% | python: median=0.0221, mean=0.0235, range=0.0147, CV=22.05%; ruby: median=0.0496, mean=0.0498, range=0.0353, CV=20.95% | python: median=210.1695, mean=231.7366, range=105.1136, CV=24.07%; ruby: median=93.4414, mean=109.5860, range=50.2418, CV=26.23% |
| `regex_stream_audit` | `bytecode` | 15 | median=4.4700, mean=4.4293, range=1.0200, CV=6.63% | python: median=0.0186, mean=0.0204, range=0.0091, CV=15.40%; ruby: median=0.0459, mean=0.0463, range=0.0217, CV=12.42% | python: median=211.0101, mean=220.6098, range=67.5989, CV=15.78%; ruby: median=93.6232, mean=96.5865, range=27.6686, CV=14.57% |

## Inputs

- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-bytecode-03-selected.json` — generated `2026-07-17T15:12:52.624552Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-bytecode-03-selected.json` — generated `2026-07-17T16:32:44.279961Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-17-regex-carrier-scorecard-bytecode.json` — generated `2026-07-17T19:54:54.103001Z`, CPU `0-15`
