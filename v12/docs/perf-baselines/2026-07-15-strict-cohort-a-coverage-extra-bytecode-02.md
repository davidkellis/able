# External Benchmark Comparison

- Generated: `2026-07-16T04:35:12.369962Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `document_audit, lexical_rollup, regex_suffix_audit`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `bytecode` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3380 | 0.0120 | 28.17x | 0.0366 | 9.23x |
| `lexical_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5180 | 0.0145 | 35.72x | 0.0418 | 12.39x |
| `regex_suffix_audit` | `bytecode` | timeout (5) | not run | n/a | n/a | 0.0342 | n/a | 0.0662 | n/a |
