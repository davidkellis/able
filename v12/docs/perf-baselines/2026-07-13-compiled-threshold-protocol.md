# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.

| Benchmark | Mode | Verified samples | Able seconds | Reference ratios |
| --- | --- | ---: | --- | --- |
| `base64` | `compiled` | 5 | median=2.5400, mean=2.6160, range=0.4600, CV=7.33% | go: median=1.0096, mean=1.0199, range=0.1070, CV=4.28% |
| `k_nucleotide` | `compiled` | 5 | median=3.7300, mean=3.8320, range=0.7300, CV=7.79% | go: median=59.4896, mean=59.8486, range=13.2966, CV=9.08% |
| `nbody` | `compiled` | 5 | median=0.4900, mean=0.5160, range=0.2500, CV=18.52% | go: median=13.9535, mean=14.1328, range=3.0620, CV=8.61% |
| `quicksort` | `compiled` | 5 | median=1.8600, mean=1.9000, range=0.2300, CV=5.30% | go: median=0.7412, mean=0.7198, range=0.1478, CV=7.90% |

## Inputs

- `v12/tmp/perf/2026-07-13-compiled-threshold-protocol/compiled-r1.json` — generated `2026-07-13T20:47:09.319479Z`, CPU `15`
- `v12/tmp/perf/2026-07-13-compiled-threshold-protocol/compiled-r2.json` — generated `2026-07-13T20:49:55.306445Z`, CPU `15`
- `v12/tmp/perf/2026-07-13-compiled-threshold-protocol/compiled-r3.json` — generated `2026-07-13T20:52:52.949653Z`, CPU `15`
- `v12/tmp/perf/2026-07-13-compiled-threshold-protocol/compiled-r4.json` — generated `2026-07-13T20:55:35.097976Z`, CPU `15`
- `v12/tmp/perf/2026-07-13-compiled-threshold-protocol/compiled-r5.json` — generated `2026-07-13T20:58:14.413543Z`, CPU `15`
