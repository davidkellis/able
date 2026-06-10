# External Benchmark Comparison

- Generated: `2026-07-20T14:55:01.373836Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `dependency_plan, inventory_reconciliation, option_result_config, unicode_scalar_pipeline`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_plan` | `compiled` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.0900 | 0.0045 | 20.00x |
| `inventory_reconciliation` | `compiled` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 0.2580 | 0.0106 | 24.34x |
| `option_result_config` | `compiled` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.2000 | 0.0042 | 47.62x |
| `unicode_scalar_pipeline` | `compiled` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 0.2420 | 0.0123 | 19.67x |
