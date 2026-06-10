# External Benchmark Comparison

- Generated: `2026-07-20T19:40:14.692162Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-interpreter-reference-status.json`
- Suite: `custom`
- Able benchmarks: `quicksort, sudoku_masks`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `bytecode` | timeout (1) | not run | n/a | n/a | 25.3076 | n/a | 15.2717 | n/a |
| `sudoku_masks` | `bytecode` | timeout (1) | not run | n/a | n/a | 17.8622 | n/a | 21.2515 | n/a |
