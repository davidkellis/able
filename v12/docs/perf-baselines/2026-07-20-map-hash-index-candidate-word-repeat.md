# External Benchmark Comparison

- Generated: `2026-07-20T12:42:17.169385Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-map-three-app-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-map-three-app-go-reference.json`
- Suite: `custom`
- Able benchmarks: `word_frequency`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `word_frequency` | `compiled` | ok (10) | verified (10) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.1510 | 0.0064 | 23.59x | n/a | n/a | n/a | n/a |
| `word_frequency` | `bytecode` | ok (10) | verified (10) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.4230 | n/a | n/a | 0.0217 | 65.58x | 0.0537 | 26.50x |
