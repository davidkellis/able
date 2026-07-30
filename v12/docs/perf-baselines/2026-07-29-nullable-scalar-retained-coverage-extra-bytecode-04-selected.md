# External Benchmark Comparison

- Generated: `2026-07-29T23:54:53.654803Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `dependency_plan, discrete_event_simulation, inventory_reconciliation, option_result_config, unicode_scalar_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `12-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_plan` | `bytecode` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.4700 | 0.0182 | 25.82x | 0.0580 | 8.10x |
| `discrete_event_simulation` | `bytecode` | ok (5) | verified (5) | 6aebca9b31a78441438d2321290a7b66dc831ddbc7671d783e4a725aed6e7405 | 4.9920 | 0.2255 | 22.14x | 0.2424 | 20.59x |
| `inventory_reconciliation` | `bytecode` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 2.7020 | 0.0728 | 37.12x | 0.0972 | 27.80x |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 1.0760 | 0.0292 | 36.85x | 0.0697 | 15.44x |
| `unicode_scalar_pipeline` | `bytecode` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 3.9840 | 0.3700 | 10.77x | 0.4910 | 8.11x |
