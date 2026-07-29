# External Benchmark Comparison

- Generated: `2026-07-28T19:45:00.133335Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `quicksort, sudoku_masks`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `bytecode` | ok (5) | verified (5) | 88148e21399796b608b9762acf30ecb3a1d938a57a60945c20653dc74c6b3e60 | 12.1780 | 0.6732 | 18.09x | 0.6615 | 18.41x |
| `sudoku_masks` | `bytecode` | ok (5) | verified (5) | 9354bc257cae59f24fce2f106308db1c36a10976f52089fa54a6d50b7e50b506 | 24.3940 | 1.7962 | 13.58x | 2.1343 | 11.43x |
