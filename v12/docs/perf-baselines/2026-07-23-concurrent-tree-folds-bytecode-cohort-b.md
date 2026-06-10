# External Benchmark Comparison

- Generated: `2026-07-23T20:11:18.261616Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-tree-folds-interpreter-b.json`
- Suite: `custom`
- Able benchmarks: `concurrent_tree_folds`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_tree_folds` | `bytecode` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.3900 | 0.0568 | 6.87x | 0.0544 | 7.17x |
