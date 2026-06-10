# External Benchmark Comparison

- Generated: `2026-07-23T04:55:27.757991Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-concurrent-stencil-reduction-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-concurrent-stencil-reduction-go-reference.json`
- Suite: `custom`
- Able benchmarks: `concurrent_stencil_reduction`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_stencil_reduction` | `compiled` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 0.2140 | 0.0051 | 41.96x | n/a | n/a | n/a | n/a |
| `concurrent_stencil_reduction` | `bytecode` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 1.8620 | n/a | n/a | 0.1240 | 15.02x | 0.0963 | 19.34x |
