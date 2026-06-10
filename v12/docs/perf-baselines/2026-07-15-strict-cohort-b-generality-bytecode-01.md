# External Benchmark Comparison

- Generated: `2026-07-16T05:48:00.217787Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1540 | 57.5715 | 0.00x | 45.5928 | 0.00x |
| `binarytrees` | `bytecode` | timeout (5) | not run | n/a | n/a | 12.7937 | n/a | 54.8039 | n/a |
| `matrixmultiply` | `bytecode` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 4.7120 | 49.2453 | 0.10x | 51.8666 | 0.09x |
