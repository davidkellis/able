# External Benchmark Comparison

- Generated: `2026-07-20T12:39:47.204578Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-map-three-app-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-map-three-app-go-reference.json`
- Suite: `custom`
- Able benchmarks: `word_frequency, k_nucleotide, inventory_reconciliation`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `word_frequency` | `compiled` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.1820 | 0.0064 | 28.44x | n/a | n/a | n/a | n/a |
| `word_frequency` | `bytecode` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.6040 | n/a | n/a | 0.0217 | 73.92x | 0.0537 | 29.87x |
| `k_nucleotide` | `compiled` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 3.5720 | 0.0794 | 44.99x | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `bytecode` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 43.4480 | n/a | n/a | 1.3160 | 33.02x | 1.2162 | 35.72x |
| `inventory_reconciliation` | `compiled` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 0.3120 | 0.0097 | 32.16x | n/a | n/a | n/a | n/a |
| `inventory_reconciliation` | `bytecode` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 2.6080 | n/a | n/a | 0.0643 | 40.56x | 0.0936 | 27.86x |
