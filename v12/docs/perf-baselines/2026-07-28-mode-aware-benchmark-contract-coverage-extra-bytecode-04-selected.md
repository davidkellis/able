# External Benchmark Comparison

- Generated: `2026-07-28T20:22:53.054504Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `dependency_plan, discrete_event_simulation, inventory_reconciliation, option_result_config, unicode_scalar_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_plan` | `bytecode` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.4940 | 0.0180 | 27.44x | 0.0492 | 10.04x |
| `discrete_event_simulation` | `bytecode` | ok (5) | verified (5) | 6aebca9b31a78441438d2321290a7b66dc831ddbc7671d783e4a725aed6e7405 | 4.7040 | 0.1737 | 27.08x | 0.2210 | 21.29x |
| `inventory_reconciliation` | `bytecode` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 2.4900 | 0.0698 | 35.67x | 0.0898 | 27.73x |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.7400 | 0.0179 | 41.34x | 0.0495 | 14.95x |
| `unicode_scalar_pipeline` | `bytecode` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 3.8120 | 0.2359 | 16.16x | 0.3176 | 12.00x |
