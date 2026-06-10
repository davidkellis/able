# External Benchmark Comparison

- Generated: `2026-07-21T05:05:36.364011Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-concurrent-event-routing-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-concurrent-event-routing-go-reference.json`
- Suite: `custom`
- Able benchmarks: `concurrent_event_routing`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_event_routing` | `compiled` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 2.8780 | 0.0045 | 639.56x | n/a | n/a | n/a | n/a |
| `concurrent_event_routing` | `bytecode` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 2.7460 | n/a | n/a | 0.0502 | 54.70x | 0.0310 | 88.58x |
