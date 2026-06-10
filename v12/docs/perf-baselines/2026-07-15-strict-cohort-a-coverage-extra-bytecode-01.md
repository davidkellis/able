# External Benchmark Comparison

- Generated: `2026-07-16T04:27:31.338568Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `fixed_width_128, rational_series, word_frequency`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fixed_width_128` | `bytecode` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 8.1580 | 0.3117 | 26.17x | 0.5768 | 14.14x |
| `rational_series` | `bytecode` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 4.1040 | 0.0918 | 44.71x | 0.1194 | 34.37x |
| `word_frequency` | `bytecode` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.5340 | 0.0174 | 88.16x | 0.0449 | 34.16x |
