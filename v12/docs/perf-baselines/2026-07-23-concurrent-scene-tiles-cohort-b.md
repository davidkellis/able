# External Benchmark Comparison

- Generated: `2026-07-23T21:16:22.442755Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-scene-tiles-interpreter-b.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-scene-tiles-go-b.json`
- Suite: `custom`
- Able benchmarks: `concurrent_scene_tiles`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_scene_tiles` | `compiled` | ok (5) | verified (5) | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe | 0.3760 | 0.0039 | 96.41x | n/a | n/a | n/a | n/a |
| `concurrent_scene_tiles` | `bytecode` | ok (5) | verified (5) | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe | 0.6220 | n/a | n/a | 0.0730 | 8.52x | 0.0758 | 8.21x |
