# External Benchmark Comparison

- Generated: `2026-07-29T23:37:20.785926Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `dependency_plan, discrete_event_simulation, inventory_reconciliation, option_result_config, unicode_scalar_pipeline`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `12-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_plan` | `compiled` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.0200 | 0.0038 | 5.26x |
| `discrete_event_simulation` | `compiled` | ok (5) | verified (5) | 6aebca9b31a78441438d2321290a7b66dc831ddbc7671d783e4a725aed6e7405 | 0.0420 | 0.0138 | 3.04x |
| `inventory_reconciliation` | `compiled` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 0.1400 | 0.0095 | 14.74x |
| `option_result_config` | `compiled` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.0460 | 0.0038 | 12.11x |
| `unicode_scalar_pipeline` | `compiled` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 0.1200 | 0.0107 | 11.21x |
