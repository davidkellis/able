# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `document_audit` | `compiled` | 5 | median=0.0500, mean=0.0620, range=0.0600, CV=43.28% | n/a | n/a |
| `k_nucleotide` | `compiled` | 5 | median=2.0500, mean=2.0520, range=0.0500, CV=0.94% | n/a | n/a |
| `word_frequency` | `compiled` | 5 | median=0.1500, mean=0.1620, range=0.0600, CV=16.56% | n/a | n/a |

## Inputs

- `v12/docs/perf-baselines/2026-07-16-primitive-kernel-helper-controls.json` — generated `2026-07-16T11:18:39.294277Z`, CPU `None`
