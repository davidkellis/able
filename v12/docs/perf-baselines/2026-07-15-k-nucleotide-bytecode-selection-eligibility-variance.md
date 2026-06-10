# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
Every Able and fresh-reference component retained exactly 5 successful verifier-backed runs.

| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `k_nucleotide` | `bytecode` | 5 | median=39.2400, mean=39.1680, range=1.5700, CV=1.69% | python: median=1.1680, mean=1.1685, range=0.0294, CV=1.04%; ruby: median=1.2174, mean=1.2047, range=0.0954, CV=3.00% | python: median=33.5199, mean=33.5199, range=0.0000, CV=0.00%; ruby: median=32.5127, mean=32.5127, range=0.0000, CV=0.00% |

## Inputs

- `v12/docs/perf-baselines/2026-07-15-k-nucleotide-bytecode-selection-eligibility.json` — generated `2026-07-16T02:17:48.504702Z`, CPU `None`
