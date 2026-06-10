# External Benchmark Comparison

- Generated: `2026-07-23T14:27:01.222794Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-transform-chain-interpreter-reference-b.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-transform-chain-go-reference-b.json`
- Suite: `custom`
- Able benchmarks: `concurrent_transform_chain`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_transform_chain` | `compiled` | ok (5) | verified (5) | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 | 8.8660 | 0.0059 | 1502.71x | n/a | n/a | n/a | n/a |
| `concurrent_transform_chain` | `bytecode` | ok (5) | verified (5) | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 | 2.8580 | n/a | n/a | 0.1567 | 18.24x | 0.1284 | 22.26x |
