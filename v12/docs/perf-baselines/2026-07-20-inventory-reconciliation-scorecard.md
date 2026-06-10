# External Benchmark Comparison

- Generated: `2026-07-20T12:19:39.439290Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-inventory-reconciliation-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-inventory-reconciliation-go-reference.json`
- Suite: `custom`
- Able benchmarks: `inventory_reconciliation`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `inventory_reconciliation` | `compiled` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 2.8240 | 0.0081 | 348.64x | n/a | n/a | n/a | n/a |
| `inventory_reconciliation` | `bytecode` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 4.6380 | n/a | n/a | 0.0617 | 75.17x | 0.0793 | 58.49x |
