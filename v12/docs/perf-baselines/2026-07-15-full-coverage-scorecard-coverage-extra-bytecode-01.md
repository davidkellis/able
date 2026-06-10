# External Benchmark Comparison

- Generated: `2026-07-15T09:08:38.103786Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `fixed_width_128, rational_series, word_frequency`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fixed_width_128` | `bytecode` | ok (3) | verified (3) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 6.7000 | 0.3386 | 19.79x | 0.8400 | 7.98x |
| `rational_series` | `bytecode` | ok (3) | verified (3) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 3.3767 | 0.1033 | 32.69x | 0.1457 | 23.18x |
| `word_frequency` | `bytecode` | ok (3) | verified (3) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.3000 | 0.0189 | 68.78x | 0.0593 | 21.92x |
