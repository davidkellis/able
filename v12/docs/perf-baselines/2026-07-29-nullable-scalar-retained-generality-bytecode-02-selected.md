# External Benchmark Comparison

- Generated: `2026-07-29T23:09:39.636853Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `quicksort, sudoku_masks`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `12-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `bytecode` | ok (5) | verified (5) | 88148e21399796b608b9762acf30ecb3a1d938a57a60945c20653dc74c6b3e60 | 13.3140 | 1.1138 | 11.95x | 1.2566 | 10.60x |
| `sudoku_masks` | `bytecode` | ok (5) | verified (5) | 9354bc257cae59f24fce2f106308db1c36a10976f52089fa54a6d50b7e50b506 | 28.0120 | 2.8424 | 9.86x | 2.5671 | 10.91x |
