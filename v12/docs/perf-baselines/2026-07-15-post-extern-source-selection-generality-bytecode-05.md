# External Benchmark Comparison

- Generated: `2026-07-15T06:34:24.940379Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `reverse_complement, k_nucleotide`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `reverse_complement` | `bytecode` | ok (1) | verified (1) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 5.9600 | 0.0270 | 220.74x | 0.0766 | 77.81x |
| `k_nucleotide` | `bytecode` | ok (1) | verified (1) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 36.1800 | 1.1866 | 30.49x | 1.1236 | 32.20x |
