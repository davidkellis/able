# External Benchmark Comparison

- Generated: `2026-07-28T20:09:47.092077Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `config_validation_extraction`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `config_validation_extraction` | `compiled` | ok (5) | verified (5) | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 | 0.0440 | 0.0042 | 10.48x |
