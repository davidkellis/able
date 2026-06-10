# External Benchmark Comparison

- Generated: `2026-07-14T07:21:19.424229Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-compiled-coverage-go-refs.json`
- Suite: `custom`
- Able benchmarks: `word_frequency, document_audit, lexical_rollup`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `word_frequency` | `compiled` | ok (3) | verified (3) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.2100 | 0.0054 | 38.89x |
| `document_audit` | `compiled` | ok (3) | verified (3) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.0800 | 0.0040 | 20.00x |
| `lexical_rollup` | `compiled` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.1200 | 0.0044 | 27.27x |
