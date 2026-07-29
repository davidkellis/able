# External Benchmark Comparison

- Generated: `2026-07-28T20:23:02.540091Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `config_validation_extraction`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `config_validation_extraction` | `bytecode` | ok (5) | verified (5) | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 | 1.3640 | 0.0187 | 72.94x | 0.0467 | 29.21x |
