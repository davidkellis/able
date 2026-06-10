# External Benchmark Comparison

- Generated: `2026-07-18T16:43:04.928407Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-interpreter-reference-status.json`
- Suite: `custom`
- Able benchmarks: `quicksort, sudoku_masks`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `bytecode` | timeout (1) | not run | n/a | n/a | 27.9447 | n/a | 15.9612 | n/a |
| `sudoku_masks` | `bytecode` | timeout (1) | not run | n/a | n/a | 27.4226 | n/a | 24.1037 | n/a |
