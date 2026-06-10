# External Benchmark Comparison

- Generated: `2026-07-20T21:25:35.972612Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `reverse_complement, k_nucleotide, fasta_generation`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `reverse_complement` | `bytecode` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 3.5200 | 0.0293 | 120.14x | 0.0819 | 42.98x |
| `k_nucleotide` | `bytecode` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 43.3140 | 1.3555 | 31.95x | 1.2526 | 34.58x |
| `fasta_generation` | `bytecode` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 1.7940 | 0.1978 | 9.07x | 0.2889 | 6.21x |
