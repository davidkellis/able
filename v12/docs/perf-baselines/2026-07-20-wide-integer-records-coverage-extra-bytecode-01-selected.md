# External Benchmark Comparison

- Generated: `2026-07-20T18:26:57.609039Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `fixed_width_128, rational_series, wide_integer_records, word_frequency`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fixed_width_128` | `bytecode` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 7.2940 | 0.3580 | 20.37x | 0.6022 | 12.11x |
| `rational_series` | `bytecode` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 3.8500 | 0.0963 | 39.98x | 0.1297 | 29.68x |
| `wide_integer_records` | `bytecode` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 5.1160 | 0.0632 | 80.95x | 0.1384 | 36.97x |
| `word_frequency` | `bytecode` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.4080 | 0.0186 | 75.70x | 0.0488 | 28.85x |
