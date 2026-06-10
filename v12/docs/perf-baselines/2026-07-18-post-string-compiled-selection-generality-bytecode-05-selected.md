# External Benchmark Comparison

- Generated: `2026-07-18T16:49:32.550368Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `reverse_complement, k_nucleotide`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `reverse_complement` | `bytecode` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 7.5020 | 0.0269 | 278.88x | 0.0808 | 92.85x |
| `k_nucleotide` | `bytecode` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 44.9340 | 1.3879 | 32.38x | 1.4065 | 31.95x |
