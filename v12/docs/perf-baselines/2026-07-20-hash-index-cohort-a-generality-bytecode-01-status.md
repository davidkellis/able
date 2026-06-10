# External Benchmark Comparison

- Generated: `2026-07-20T13:18:14.031954Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-interpreter-reference-status.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1800 | n/a | n/a | 44.1456 | 0.00x |
| `binarytrees` | `bytecode` | timeout (1) | not run | n/a | n/a | n/a | n/a | n/a | n/a |
| `matrixmultiply` | `bytecode` | ok (1) | verified (1) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 4.7700 | 48.9848 | 0.10x | 46.2509 | 0.10x |
