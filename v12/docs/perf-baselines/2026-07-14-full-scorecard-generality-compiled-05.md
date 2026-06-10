# External Benchmark Comparison

- Generated: `2026-07-14T16:46:30.814037Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `reverse_complement, k_nucleotide`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `reverse_complement` | `compiled` | ok (1) | verified (1) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.1800 | 0.0155 | 11.61x |
| `k_nucleotide` | `compiled` | ok (1) | verified (1) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 3.5800 | 0.0602 | 59.47x |
