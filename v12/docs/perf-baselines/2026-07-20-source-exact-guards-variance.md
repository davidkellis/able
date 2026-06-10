# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
Every Able and fresh-reference component retained exactly 5 successful verifier-backed runs.

| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `binarytrees` | `compiled` | 10 | median=10.4950, mean=10.5980, range=1.0700, CV=4.10% | go: median=11.0573, mean=11.0975, range=0.7765, CV=2.45% | go: median=0.9547, mean=0.9547, range=0.0416, CV=3.08% |
| `json` | `bytecode` | 10 | median=0.8500, mean=0.9070, range=0.6500, CV=21.09% | python: median=2.7217, mean=2.7501, range=0.3784, CV=4.37%; ruby: median=1.6928, mean=1.7013, range=0.2130, CV=4.01% | python: median=0.3293, mean=0.3293, range=0.0536, CV=11.51%; ruby: median=0.5338, mean=0.5338, range=0.1193, CV=15.80% |
| `json` | `compiled` | 10 | median=0.7700, mean=0.8160, range=0.4000, CV=16.27% | go: median=1.4621, mean=1.4551, range=0.1447, CV=3.46% | go: median=0.5607, mean=0.5607, range=0.0549, CV=6.92% |
| `pidigits` | `bytecode` | 10 | median=2.4650, mean=2.4370, range=0.4800, CV=6.67% | python: median=4.0836, mean=4.1240, range=0.5161, CV=3.65%; ruby: median=10.3941, mean=10.4326, range=1.3681, CV=4.38% | python: median=0.5914, mean=0.5914, range=0.0572, CV=6.84%; ruby: median=0.2335, mean=0.2335, range=0.0037, CV=1.12% |
| `quicksort` | `compiled` | 10 | median=1.9350, mean=1.9910, range=0.7000, CV=10.34% | go: median=2.6438, mean=2.6902, range=0.5794, CV=5.90% | go: median=0.7398, mean=0.7398, range=0.0348, CV=3.33% |

## Inputs

- `v12/docs/perf-baselines/2026-07-20-source-exact-guards-c1-binarytrees-compiled.json` — generated `2026-07-21T01:44:22.703692Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-source-exact-guards-c1-quicksort-compiled.json` — generated `2026-07-21T01:46:53.850542Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-source-exact-guards-c1-json-compiled.json` — generated `2026-07-21T01:50:04.611061Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-source-exact-guards-c1-json-bytecode.json` — generated `2026-07-21T02:03:24.657931Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-source-exact-guards-c1-pidigits-bytecode.json` — generated `2026-07-21T02:04:27.347812Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-source-exact-guards-c2-compiled.json` — generated `2026-07-21T03:27:24.852057Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-source-exact-guards-c2-bytecode.json` — generated `2026-07-21T03:28:52.043981Z`, CPU `0-15`
