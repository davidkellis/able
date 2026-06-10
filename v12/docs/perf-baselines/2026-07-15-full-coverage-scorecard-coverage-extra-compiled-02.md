# External Benchmark Comparison

- Generated: `2026-07-15T09:02:32.110648Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `document_audit, lexical_rollup, regex_suffix_audit`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `compiled` | ok (3) | verified (3) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.0933 | 0.0038 | 24.55x |
| `lexical_rollup` | `compiled` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.1100 | 0.0045 | 24.44x |
| `regex_suffix_audit` | `compiled` | ok (3) | verified (3) | 48835ea1a1741c659d1b6b215a56e6611e525366596e08e9a10ec985106f598a | 2.4000 | 0.0320 | 75.00x |
