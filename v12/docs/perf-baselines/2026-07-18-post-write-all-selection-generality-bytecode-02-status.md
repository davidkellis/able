# External Benchmark Comparison

- Generated: `2026-07-19T02:34:24.173149Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-interpreter-reference-status.json`
- Suite: `custom`
- Able benchmarks: `quicksort, sudoku_masks`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `bytecode` | timeout (1) | not run | n/a | n/a | 25.6492 | n/a | 15.8552 | n/a |
| `sudoku_masks` | `bytecode` | timeout (1) | not run | n/a | n/a | 18.6339 | n/a | 23.1227 | n/a |
