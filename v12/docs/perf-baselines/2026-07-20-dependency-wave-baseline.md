# External Benchmark Comparison

- Generated: `2026-07-21T04:45:31.943284Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-dependency-wave-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-dependency-wave-go-reference.json`
- Suite: `custom`
- Able benchmarks: `dependency_wave_validation`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `dependency_wave_validation` | `compiled` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 1.2700 | 0.0043 | 295.35x | n/a | n/a | n/a | n/a |
| `dependency_wave_validation` | `bytecode` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 0.4260 | n/a | n/a | 0.0506 | 8.42x | 0.0346 | 12.31x |
