# External Benchmark Comparison

- Generated: `2026-07-13T15:50:49.909920Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-13-compiled-sudoku-masks-go-refresh.json`
- Suite: `core`
- Able benchmarks: `sudoku_masks`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `sudoku_masks` | `compiled` | ok (1) | verified (1) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 9.1500 | 0.5746 | 15.92x |
