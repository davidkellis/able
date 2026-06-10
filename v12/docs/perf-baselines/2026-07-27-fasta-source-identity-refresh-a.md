# External Benchmark Comparison

- Generated: `2026-07-27T17:03:58.802879Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-27-fasta-source-identity-interpreter-reference-cpu4.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-27-fasta-source-identity-go-reference.json`
- Suite: `custom`
- Able benchmarks: `fasta_generation`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `4` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fasta_generation` | `compiled` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 0.0580 | 0.0166 | 3.49x | n/a | n/a | n/a | n/a |
| `fasta_generation` | `bytecode` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 1.8980 | n/a | n/a | 0.2031 | 9.35x | 0.3046 | 6.23x |
