# External Benchmark Comparison

- Generated: `2026-07-16T08:10:32.019061Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `lexical_rollup, regex_suffix_audit, regex_set_audit, regex_stream_audit`
- Able modes: `compiled`
- Reference languages: `go`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `lexical_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.1080 | 0.0035 | 30.86x |
| `regex_suffix_audit` | `compiled` | ok (5) | verified (5) | 48835ea1a1741c659d1b6b215a56e6611e525366596e08e9a10ec985106f598a | 2.6240 | 0.0322 | 81.49x |
| `regex_set_audit` | `compiled` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 0.1760 | 0.0041 | 42.93x |
| `regex_stream_audit` | `compiled` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 0.1700 | 0.0041 | 41.46x |
