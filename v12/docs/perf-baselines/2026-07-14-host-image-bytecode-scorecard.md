# External Benchmark Comparison

- Generated: `2026-07-14T07:02:24.779312Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-bytecode-generality-interpreter-refresh.json`
- Suite: `core`
- Able benchmarks: `i_before_e, json, channel_rollup`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `i_before_e` | `bytecode` | ok (3) | verified (3) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.5200 | 0.0923 | 5.63x | 0.1216 | 4.28x |
| `json` | `bytecode` | ok (3) | verified (3) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.8300 | 2.7039 | 0.31x | 1.7460 | 0.48x |
| `channel_rollup` | `bytecode` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5767 | n/a | n/a | n/a | n/a |
