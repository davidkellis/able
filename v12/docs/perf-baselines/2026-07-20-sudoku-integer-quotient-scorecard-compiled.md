# External Benchmark Comparison

- Generated: `2026-07-20T07:45:28.565892Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-sudoku-integer-quotient-go-reference.json`
- Suite: `custom`
- Able benchmarks: `quicksort, sudoku_masks`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `compiled` | ok (5) | verified (5) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.6880 | 2.3475 | 0.72x |
| `sudoku_masks` | `compiled` | ok (5) | verified (5) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 1.7720 | 0.5334 | 3.32x |
