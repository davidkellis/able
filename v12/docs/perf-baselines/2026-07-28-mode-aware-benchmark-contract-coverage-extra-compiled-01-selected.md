# External Benchmark Comparison

- Generated: `2026-07-28T19:59:21.735272Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `backup_dedup, fixed_width_128, rational_series, wide_integer_records, binary_event_log, word_frequency`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `backup_dedup` | `compiled` | ok (5) | verified (5) | bf4d5c89239bd78c6dcb9d755b8df4e90bc092c2a64bf15e45786e815918504e | 0.0620 | 0.0095 | 6.53x |
| `fixed_width_128` | `compiled` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 0.0900 | 0.0060 | 15.00x |
| `rational_series` | `compiled` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 0.0520 | 0.0137 | 3.80x |
| `wide_integer_records` | `compiled` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 0.0780 | 0.0377 | 2.07x |
| `binary_event_log` | `compiled` | ok (5) | verified (5) | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 | 0.1520 | 0.0093 | 16.34x |
| `word_frequency` | `compiled` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.0500 | 0.0060 | 8.33x |
