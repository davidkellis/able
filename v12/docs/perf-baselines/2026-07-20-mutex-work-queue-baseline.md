# External Benchmark Comparison

- Generated: `2026-07-20T15:41:51.177102Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-mutex-work-queue-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-mutex-work-queue-go-reference.json`
- Suite: `custom`
- Able benchmarks: `mutex_work_queue`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `mutex_work_queue` | `compiled` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 1.3560 | 0.0047 | 288.51x | n/a | n/a | n/a | n/a |
| `mutex_work_queue` | `bytecode` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 0.4120 | n/a | n/a | 0.0259 | 15.91x | 0.0474 | 8.69x |
