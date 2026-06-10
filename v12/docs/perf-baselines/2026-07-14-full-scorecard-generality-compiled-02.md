# External Benchmark Comparison

- Generated: `2026-07-14T16:39:20.904267Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `quicksort, sudoku, sudoku_masks`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `compiled` | ok (1) | verified (1) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.9300 | 2.6809 | 0.72x |
| `sudoku` | `compiled` | timeout (1) | not run | n/a | n/a | 0.1534 | n/a |
| `sudoku_masks` | `compiled` | ok (1) | verified (1) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 9.1100 | 0.5779 | 15.76x |
