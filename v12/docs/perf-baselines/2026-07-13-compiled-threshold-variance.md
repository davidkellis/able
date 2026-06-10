# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.

| Benchmark | Mode | Verified samples | Able seconds | Reference ratios |
| --- | --- | ---: | --- | --- |
| `base64` | `compiled` | 3 | median=2.4800, mean=2.5400, range=0.2000, CV=4.44% | go: median=0.9936, mean=1.0181, range=0.0839, CV=4.62% |
| `k_nucleotide` | `compiled` | 3 | median=3.7100, mean=4.2500, range=1.7400, CV=23.24% | go: median=66.0036, mean=73.2331, range=38.1193, CV=27.39% |
| `nbody` | `compiled` | 3 | median=0.4900, mean=0.4700, range=0.0800, CV=9.27% | go: median=14.2045, mean=13.8841, range=2.0702, CV=7.72% |

## Inputs

- `v12/tmp/perf/2026-07-13-scoreboard-variance/compiled-r1.json` — generated `2026-07-13T20:33:22.773678Z`, CPU `15`
- `v12/tmp/perf/2026-07-13-scoreboard-variance/compiled-r2.json` — generated `2026-07-13T20:35:22.192572Z`, CPU `15`
- `v12/tmp/perf/2026-07-13-scoreboard-variance/compiled-r3.json` — generated `2026-07-13T20:37:21.210756Z`, CPU `15`
