# External Benchmark Comparison

- Generated: `2026-07-16T01:51:15.650366Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-error-reconcile-a-generality-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1420 | 53.7552 | 0.00x | 41.5265 | 0.00x |
| `binarytrees` | `bytecode` | timeout (5) | not run | n/a | n/a | 10.9391 | n/a | 55.7194 | n/a |
| `matrixmultiply` | `bytecode` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 4.3920 | 47.4247 | 0.09x | 46.7062 | 0.09x |
