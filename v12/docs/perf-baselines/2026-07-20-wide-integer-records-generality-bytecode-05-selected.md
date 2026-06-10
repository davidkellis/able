# External Benchmark Comparison

- Generated: `2026-07-20T18:04:18.719393Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `reverse_complement, k_nucleotide, fasta_generation`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `reverse_complement` | `bytecode` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 3.3100 | 0.0241 | 137.34x | 0.0681 | 48.60x |
| `k_nucleotide` | `bytecode` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 40.3900 | 1.2220 | 33.05x | 1.2127 | 33.31x |
| `fasta_generation` | `bytecode` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 1.8380 | 0.1998 | 9.20x | 0.2910 | 6.32x |
