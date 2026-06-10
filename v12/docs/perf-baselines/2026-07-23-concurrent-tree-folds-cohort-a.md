# External Benchmark Comparison

- Generated: `2026-07-23T20:08:27.700058Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-tree-folds-interpreter-a.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-tree-folds-go-a.json`
- Suite: `custom`
- Able benchmarks: `concurrent_tree_folds`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_tree_folds` | `compiled` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.3840 | 0.0037 | 103.78x | n/a | n/a | n/a | n/a |
| `concurrent_tree_folds` | `bytecode` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.3800 | n/a | n/a | 0.0562 | 6.76x | 0.0555 | 6.85x |
