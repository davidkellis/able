# External Benchmark Comparison

- Generated: `2026-07-16T03:29:05.713237Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1460 | 55.3957 | 0.00x | 41.6611 | 0.00x |
| `binarytrees` | `bytecode` | timeout (5) | not run | n/a | n/a | 10.9425 | n/a | 50.2250 | n/a |
| `matrixmultiply` | `bytecode` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 4.3380 | 46.8189 | 0.09x | 43.8530 | 0.10x |
