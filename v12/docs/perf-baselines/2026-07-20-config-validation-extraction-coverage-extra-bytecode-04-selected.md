# External Benchmark Comparison

- Generated: `2026-07-20T21:56:33.338724Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `dependency_plan, inventory_reconciliation, option_result_config, unicode_scalar_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_plan` | `bytecode` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.5720 | 0.0172 | 33.26x | 0.0526 | 10.87x |
| `inventory_reconciliation` | `bytecode` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 3.0940 | 0.0721 | 42.91x | 0.1146 | 27.00x |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.9060 | 0.0219 | 41.37x | 0.0504 | 17.98x |
| `unicode_scalar_pipeline` | `bytecode` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 4.1180 | 0.2431 | 16.94x | 0.3123 | 13.19x |
