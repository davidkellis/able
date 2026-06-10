# External Benchmark Comparison

- Generated: `2026-07-22T00:12:12.987199Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-text-byte-next-closures-go-reference.json`
- Suite: `custom`
- Able benchmarks: `i_before_e, inventory_reconciliation, k_nucleotide, unicode_scalar_pipeline, word_frequency`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `i_before_e` | `compiled` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.1540 | 0.0822 | 1.87x |
| `inventory_reconciliation` | `compiled` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 0.1960 | 0.0091 | 21.54x |
| `k_nucleotide` | `compiled` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 2.6700 | 0.0684 | 39.04x |
| `unicode_scalar_pipeline` | `compiled` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 0.2620 | 0.0131 | 20.00x |
| `word_frequency` | `compiled` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.1460 | 0.0064 | 22.81x |
