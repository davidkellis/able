# External Benchmark Comparison

- Generated: `2026-07-28T20:20:04.924888Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `backup_dedup, fixed_width_128, rational_series, wide_integer_records, binary_event_log, word_frequency`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `backup_dedup` | `bytecode` | ok (5) | verified (5) | bf4d5c89239bd78c6dcb9d755b8df4e90bc092c2a64bf15e45786e815918504e | 1.9140 | 0.2661 | 7.19x | 0.1320 | 14.50x |
| `fixed_width_128` | `bytecode` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 8.0300 | 0.3702 | 21.69x | 0.6500 | 12.35x |
| `rational_series` | `bytecode` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 4.1460 | 0.1082 | 38.32x | 0.1302 | 31.84x |
| `wide_integer_records` | `bytecode` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 5.5700 | 0.0650 | 85.69x | 0.1352 | 41.20x |
| `binary_event_log` | `bytecode` | ok (5) | verified (5) | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 | 5.8360 | 0.1818 | 32.10x | 0.2483 | 23.50x |
| `word_frequency` | `bytecode` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.5020 | 0.0204 | 73.63x | 0.0551 | 27.26x |
