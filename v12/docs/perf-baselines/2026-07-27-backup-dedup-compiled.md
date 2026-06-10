# External Benchmark Comparison

- Generated: `2026-07-27T18:06:49.003590Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `v12/docs/perf-baselines/2026-07-27-backup-dedup-go-reference.json`
- Suite: `custom`
- Able benchmarks: `backup_dedup`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `4` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `backup_dedup` | `compiled` | ok (5) | verified (5) | bf4d5c89239bd78c6dcb9d755b8df4e90bc092c2a64bf15e45786e815918504e | 0.0600 | 0.0092 | 6.52x |
