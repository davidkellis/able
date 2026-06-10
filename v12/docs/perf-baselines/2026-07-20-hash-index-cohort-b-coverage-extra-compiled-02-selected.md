# External Benchmark Comparison

- Generated: `2026-07-20T14:47:07.906031Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `document_audit, lexical_rollup, regex_suffix_audit`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `compiled` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.1000 | 0.0038 | 26.32x |
| `lexical_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.1180 | 0.0041 | 28.78x |
| `regex_suffix_audit` | `compiled` | ok (5) | verified (5) | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f | 0.1640 | 0.0047 | 34.89x |
