# External Benchmark Comparison

- Generated: `2026-07-23T16:45:29.451204Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-stateful-pipeline-interpreter-reference-b.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-stateful-pipeline-go-reference-b.json`
- Suite: `custom`
- Able benchmarks: `concurrent_stateful_pipeline`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_stateful_pipeline` | `compiled` | ok (5) | verified (5) | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 | 0.8180 | 0.0043 | 190.23x | n/a | n/a | n/a | n/a |
| `concurrent_stateful_pipeline` | `bytecode` | ok (5) | verified (5) | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 | 0.3420 | n/a | n/a | 0.0481 | 7.11x | 0.0596 | 5.74x |
