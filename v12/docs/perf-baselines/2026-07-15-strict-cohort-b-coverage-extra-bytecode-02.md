# External Benchmark Comparison

- Generated: `2026-07-16T06:52:52.331210Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `document_audit, lexical_rollup, regex_suffix_audit`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `bytecode` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3300 | 0.0131 | 25.19x | 0.0436 | 7.57x |
| `lexical_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5100 | 0.0163 | 31.29x | 0.0505 | 10.10x |
| `regex_suffix_audit` | `bytecode` | timeout (5) | not run | n/a | n/a | 0.0396 | n/a | 0.0748 | n/a |
