# External Benchmark Comparison

- Generated: `2026-07-21T02:02:45.268187Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-current-product-scorecard-generality-interpreter-reference-status.json`
- Suite: `custom`
- Able benchmarks: `quicksort, sudoku_masks`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `bytecode` | timeout (1) | not run | n/a | n/a | 24.5204 | n/a | 15.7502 | n/a |
| `sudoku_masks` | `bytecode` | timeout (1) | not run | n/a | n/a | 18.6741 | n/a | 22.5619 | n/a |
