# External Benchmark Comparison

- Generated: `2026-07-20T10:47:37.834088Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `dependency_plan, option_result_config, unicode_scalar_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_plan` | `bytecode` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.4760 | 0.0210 | 22.67x | 0.0560 | 8.50x |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.8880 | 0.0197 | 45.08x | 0.0536 | 16.57x |
| `unicode_scalar_pipeline` | `bytecode` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 3.6440 | 0.2807 | 12.98x | 0.3169 | 11.50x |
