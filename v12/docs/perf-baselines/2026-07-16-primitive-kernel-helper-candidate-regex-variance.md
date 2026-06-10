# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `regex_set_audit` | `compiled` | 5 | median=0.0800, mean=0.0920, range=0.0700, CV=32.06% | n/a | n/a |
| `regex_stream_audit` | `compiled` | 5 | median=0.0900, mean=0.0980, range=0.0700, CV=30.10% | n/a | n/a |
| `regex_suffix_audit` | `compiled` | 5 | median=0.7600, mean=0.7500, range=0.0500, CV=2.67% | n/a | n/a |

## Inputs

- `v12/docs/perf-baselines/2026-07-16-primitive-kernel-helper-candidate-regex.json` — generated `2026-07-16T11:16:10.633210Z`, CPU `None`
