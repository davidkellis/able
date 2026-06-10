# External Benchmark Comparison

- Generated: `2026-07-15T08:57:35.897503Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `fixed_width_128, rational_series, word_frequency`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `fixed_width_128` | `compiled` | ok (3) | verified (3) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 7.4033 | 0.0053 | 1396.85x |
| `rational_series` | `compiled` | ok (3) | verified (3) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 2.1633 | 0.0123 | 175.88x |
| `word_frequency` | `compiled` | ok (3) | verified (3) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.2133 | 0.0049 | 43.53x |
