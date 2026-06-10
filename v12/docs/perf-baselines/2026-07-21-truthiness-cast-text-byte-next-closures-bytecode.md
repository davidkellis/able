# External Benchmark Comparison

- Generated: `2026-07-22T00:14:09.466998Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-text-byte-next-closures-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `base64, fasta_generation, reverse_complement`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `base64` | `bytecode` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.7900 | 4.1192 | 0.68x | 2.5784 | 1.08x |
| `fasta_generation` | `bytecode` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 1.8120 | 0.2071 | 8.75x | 0.3112 | 5.82x |
| `reverse_complement` | `bytecode` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 3.2500 | 0.0260 | 125.00x | 0.0719 | 45.20x |
