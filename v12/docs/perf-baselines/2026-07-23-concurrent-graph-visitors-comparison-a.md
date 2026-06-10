# External Benchmark Comparison

- Generated: `2026-07-23T19:00:37.167936Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-graph-visitors-interpreter-reference-a.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-graph-visitors-go-reference-a.json`
- Suite: `custom`
- Able benchmarks: `concurrent_graph_visitors`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_graph_visitors` | `compiled` | ok (5) | verified (5) | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee | 0.2580 | 0.0040 | 64.50x | n/a | n/a | n/a | n/a |
| `concurrent_graph_visitors` | `bytecode` | ok (5) | verified (5) | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee | 0.9540 | n/a | n/a | 0.0539 | 17.70x | 0.0564 | 16.91x |
