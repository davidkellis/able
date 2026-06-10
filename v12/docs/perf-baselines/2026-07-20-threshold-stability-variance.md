# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
Every Able and fresh-reference component retained exactly 5 successful verifier-backed runs.

| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `base64` | `bytecode` | 15 | median=2.8600, mean=2.8680, range=0.4100, CV=3.82% | python: median=3.9369, mean=3.9837, range=0.4061, CV=3.37%; ruby: median=2.4661, mean=2.5443, range=0.6947, CV=7.87% | python: median=0.7250, mean=0.7200, range=0.0341, CV=2.45%; ruby: median=1.1364, mean=1.1299, range=0.1580, CV=7.01% |
| `base64` | `compiled` | 15 | median=2.5300, mean=2.5587, range=0.5700, CV=5.99% | go: median=2.5523, mean=2.5565, range=0.3156, CV=3.38% | go: median=0.9998, mean=1.0006, range=0.0538, CV=2.69% |
| `monte_carlo_pi` | `compiled` | 15 | median=0.2100, mean=0.2173, range=0.0900, CV=11.33% | go: median=0.2048, mean=0.2163, range=0.0893, CV=11.96% | go: median=0.9498, mean=1.0117, range=0.2307, CV=12.57% |

## Inputs

- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c1-base64-compiled.json` — generated `2026-07-21T01:50:04.611061Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c2-base64-compiled.json` — generated `2026-07-21T02:54:12.147455Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c3-base64-compiled.json` — generated `2026-07-21T02:59:03.794006Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c1-base64-bytecode.json` — generated `2026-07-21T02:03:24.657931Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c2-base64-bytecode.json` — generated `2026-07-21T02:55:14.635554Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c3-base64-bytecode.json` — generated `2026-07-21T02:58:13.505603Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c1-monte-carlo-compiled.json` — generated `2026-07-21T01:52:33.908646Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c2-monte-carlo-compiled.json` — generated `2026-07-21T02:55:51.221924Z`, CPU `0-15`
- `v12/docs/perf-baselines/2026-07-20-threshold-stability-c3-monte-carlo-compiled.json` — generated `2026-07-21T02:57:48.454910Z`, CPU `0-15`
