# External Benchmark Comparison

- Generated: `2026-07-23T17:43:21.352309Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-state-machines-interpreter-reference-a.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-state-machines-go-reference-a.json`
- Suite: `custom`
- Able benchmarks: `concurrent_state_machines`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_state_machines` | `compiled` | ok (5) | verified (5) | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 | 0.2640 | 0.0037 | 71.35x | n/a | n/a | n/a | n/a |
| `concurrent_state_machines` | `bytecode` | ok (5) | verified (5) | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 | 0.3020 | n/a | n/a | 0.0548 | 5.51x | 0.0602 | 5.02x |
