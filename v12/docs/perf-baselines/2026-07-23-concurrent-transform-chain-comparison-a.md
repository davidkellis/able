# External Benchmark Comparison

- Generated: `2026-07-23T14:25:00.143928Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-transform-chain-interpreter-reference-a.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-transform-chain-go-reference-a.json`
- Suite: `custom`
- Able benchmarks: `concurrent_transform_chain`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_transform_chain` | `compiled` | ok (5) | verified (5) | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 | 8.1760 | 0.0065 | 1257.85x | n/a | n/a | n/a | n/a |
| `concurrent_transform_chain` | `bytecode` | ok (5) | verified (5) | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 | 2.8300 | n/a | n/a | 0.2095 | 13.51x | 0.1628 | 17.38x |
