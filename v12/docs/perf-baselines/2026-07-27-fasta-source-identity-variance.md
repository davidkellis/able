# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
Every Able and fresh-reference component retained exactly 5 successful verifier-backed runs.

| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `fasta_generation` | `bytecode` | 10 | median=1.8600, mean=1.9280, range=0.6000, CV=9.54% | python: median=0.2032, mean=0.2031, range=0.0241, CV=4.52%; ruby: median=0.3027, mean=0.3046, range=0.0243, CV=3.29% | python: median=9.4929, mean=9.4929, range=0.2954, CV=2.20%; ruby: median=6.3296, mean=6.3296, range=0.1970, CV=2.20% |
| `fasta_generation` | `compiled` | 10 | median=0.0400, mean=0.0500, range=0.0800, CV=49.89% | go: median=0.0166, mean=0.0166, range=0.0003, CV=0.60% | go: median=3.0120, mean=3.0120, range=0.9639, CV=22.63% |

## Inputs

- `v12/docs/perf-baselines/2026-07-27-fasta-source-identity-refresh-a.json` — generated `2026-07-27T17:03:58.802879Z`, CPU `4`
- `v12/docs/perf-baselines/2026-07-27-fasta-source-identity-refresh-b.json` — generated `2026-07-27T17:06:52.131607Z`, CPU `4`
