# External Benchmark Comparison

- Generated: `2026-07-16T04:17:30.394180Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `document_audit, lexical_rollup, regex_suffix_audit`
- Able modes: `compiled`
- Reference languages: `go`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `compiled` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.1060 | 0.0032 | 33.12x |
| `lexical_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.1240 | 0.0034 | 36.47x |
| `regex_suffix_audit` | `compiled` | ok (5) | verified (5) | 48835ea1a1741c659d1b6b215a56e6611e525366596e08e9a10ec985106f598a | 2.7320 | 0.0303 | 90.17x |
