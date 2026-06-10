# External Benchmark Comparison

- Generated: `2026-07-20T20:36:40.268745Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-config-validation-extraction-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-config-validation-extraction-go-reference.json`
- Suite: `custom`
- Able benchmarks: `config_validation_extraction`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `config_validation_extraction` | `compiled` | ok (5) | verified (5) | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 | 0.1000 | 0.0048 | 20.83x | n/a | n/a | n/a | n/a |
| `config_validation_extraction` | `bytecode` | ok (5) | verified (5) | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 | 1.2960 | n/a | n/a | 0.0198 | 65.45x | 0.0410 | 31.61x |
