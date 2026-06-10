# External Benchmark Comparison

- Generated: `2026-07-20T15:04:53.792232Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `dependency_plan`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_plan` | `bytecode` | ok (10) | verified (10) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.4740 | 0.0222 | 21.35x | 0.0536 | 8.84x |
