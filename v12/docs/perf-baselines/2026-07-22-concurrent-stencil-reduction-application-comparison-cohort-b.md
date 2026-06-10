# External Benchmark Comparison

- Generated: `2026-07-23T04:56:52.658658Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-concurrent-stencil-reduction-interpreter-reference-cohort-b.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-concurrent-stencil-reduction-go-reference-cohort-b.json`
- Suite: `custom`
- Able benchmarks: `concurrent_stencil_reduction`
- Able modes: `bytecode, compiled`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_stencil_reduction` | `bytecode` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 1.8020 | n/a | n/a | 0.1050 | 17.16x | 0.1406 | 12.82x |
| `concurrent_stencil_reduction` | `compiled` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 0.2500 | 0.0047 | 53.19x | n/a | n/a | n/a | n/a |
