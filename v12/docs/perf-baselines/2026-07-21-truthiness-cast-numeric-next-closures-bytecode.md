# External Benchmark Comparison

- Generated: `2026-07-21T23:06:56.138171Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-numeric-next-closures-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `fixed_width_128, rational_series, wide_integer_records`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fixed_width_128` | `bytecode` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 9.7040 | 0.5909 | 16.42x | 1.1344 | 8.55x |
| `rational_series` | `bytecode` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 4.6360 | 0.1815 | 25.54x | 0.1776 | 26.10x |
| `wide_integer_records` | `bytecode` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 5.8140 | 0.0745 | 78.04x | 0.2103 | 27.65x |
