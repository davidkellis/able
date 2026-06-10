# External Benchmark Comparison

- Generated: `2026-07-22T02:56:26.829505Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-sudoku-quotient-go-reference.json`
- Suite: `custom`
- Able benchmarks: `sudoku_masks`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `sudoku_masks` | `compiled` | ok (5) | verified (5) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 2.1660 | 0.6342 | 3.42x |
