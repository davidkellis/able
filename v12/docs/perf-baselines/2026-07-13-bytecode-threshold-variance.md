# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.

| Benchmark | Mode | Verified samples | Able seconds | Reference ratios |
| --- | --- | ---: | --- | --- |
| `future_pipeline` | `bytecode` | 3 | median=0.5000, mean=0.4933, range=0.0200, CV=2.34% | python: median=8.4175, mean=8.3616, range=0.6939, CV=4.19%; ruby: median=7.3314, mean=7.2176, range=0.8761, CV=6.22% |
| `json` | `bytecode` | 3 | median=0.8500, mean=0.8533, range=0.0100, CV=0.68% | python: median=0.3253, mean=0.3157, range=0.0448, CV=7.58%; ruby: median=0.4942, mean=0.4952, range=0.0116, CV=1.19% |
| `monte_carlo_pi` | `bytecode` | 3 | median=2.6600, mean=2.6533, range=0.0600, CV=1.15% | python: median=1.7445, mean=1.7251, range=0.1015, CV=3.10%; ruby: median=1.6392, mean=1.6419, range=0.0219, CV=0.68% |

## Inputs

- `v12/tmp/perf/2026-07-13-scoreboard-variance/bytecode-r1.json` — generated `2026-07-13T20:38:34.514447Z`, CPU `15`
- `v12/tmp/perf/2026-07-13-scoreboard-variance/bytecode-r2.json` — generated `2026-07-13T20:38:47.025532Z`, CPU `15`
- `v12/tmp/perf/2026-07-13-scoreboard-variance/bytecode-r3.json` — generated `2026-07-13T20:38:59.729977Z`, CPU `15`
