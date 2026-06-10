# External Benchmark Comparison

- Generated: `2026-07-23T20:10:44.354281Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-tree-folds-interpreter-b.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-tree-folds-go-b.json`
- Suite: `custom`
- Able benchmarks: `concurrent_tree_folds`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_tree_folds` | `compiled` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.3820 | 0.0037 | 103.24x | n/a | n/a | n/a | n/a |
| `concurrent_tree_folds` | `bytecode` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.3780 | n/a | n/a | 0.0568 | 6.65x | 0.0544 | 6.95x |
