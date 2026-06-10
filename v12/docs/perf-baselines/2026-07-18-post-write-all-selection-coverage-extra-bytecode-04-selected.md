# External Benchmark Comparison

- Generated: `2026-07-19T03:05:10.585553Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `dependency_plan, option_result_config, unicode_scalar_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_plan` | `bytecode` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.5320 | 0.0155 | 34.32x | 0.0567 | 9.38x |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 1.1260 | 0.0255 | 44.16x | 0.0534 | 21.09x |
| `unicode_scalar_pipeline` | `bytecode` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 3.6760 | 0.2315 | 15.88x | 0.3270 | 11.24x |
