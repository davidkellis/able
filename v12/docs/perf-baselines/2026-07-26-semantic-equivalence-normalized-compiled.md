# External Benchmark Comparison

- Generated: `2026-07-26T13:03:39.879714Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `v12/docs/perf-baselines/2026-07-26-semantic-equivalence-normalized-go-reference.json`
- Suite: `custom`
- Able benchmarks: `tapelang_alphabet, sudoku_masks`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `tapelang_alphabet` | `compiled` | ok (5) | verified (5) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.7060 | 2.9649 | 1.25x |
| `sudoku_masks` | `compiled` | ok (5) | verified (5) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 1.5740 | 0.7027 | 2.24x |
