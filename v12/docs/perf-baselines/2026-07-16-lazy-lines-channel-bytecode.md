# External Benchmark Comparison

- Generated: `2026-07-16T08:10:50.960659Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-b-async-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `channel_rollup`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.6580 | 0.0389 | 16.92x | 0.0508 | 12.95x |
