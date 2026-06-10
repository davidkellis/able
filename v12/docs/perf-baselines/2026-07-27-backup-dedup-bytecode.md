# External Benchmark Comparison

- Generated: `2026-07-27T18:07:10.076448Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `v12/docs/perf-baselines/2026-07-27-backup-dedup-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `backup_dedup`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `4` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `backup_dedup` | `bytecode` | ok (5) | verified (5) | bf4d5c89239bd78c6dcb9d755b8df4e90bc092c2a64bf15e45786e815918504e | 1.9160 | 0.2794 | 6.86x | 0.1168 | 16.40x |
