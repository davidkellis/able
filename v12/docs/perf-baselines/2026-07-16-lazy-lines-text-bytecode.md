# External Benchmark Comparison

- Generated: `2026-07-16T08:11:47.812372Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `lexical_rollup, regex_set_audit, regex_stream_audit`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `lexical_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.4340 | 0.0163 | 26.63x | 0.0505 | 8.59x |
| `regex_set_audit` | `bytecode` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 5.0920 | 0.0181 | 281.33x | 0.0416 | 122.40x |
| `regex_stream_audit` | `bytecode` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 4.5280 | 0.0181 | 250.17x | 0.0427 | 106.04x |
