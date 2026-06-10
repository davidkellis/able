# External Benchmark Comparison

- Generated: `2026-07-16T10:40:43.163653Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `custom`
- Able benchmarks: `document_audit, lexical_rollup, reverse_complement`
- Able modes: `compiled`
- Reference languages: `go`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `compiled` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.0620 | n/a | n/a |
| `lexical_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.0720 | n/a | n/a |
| `reverse_complement` | `compiled` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.0820 | 0.0100 | 8.20x |
