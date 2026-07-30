# External Benchmark Comparison

- Generated: `2026-07-30T13:40:26.758225Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/var/tmp/able-versioned-telemetry-20260730/measurements/interpreter-reference-final.json`
- Suite: `custom`
- Able benchmarks: `versioned_telemetry_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `12` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `versioned_telemetry_pipeline` | `bytecode` | ok (5) | verified (5) | cd6312c6a89d2107854648f463d6bd7de9dc25b74d32fb5cd5ab65354df2418e | 3.3040 | 0.2023 | 16.33x | 0.1258 | 26.26x |
