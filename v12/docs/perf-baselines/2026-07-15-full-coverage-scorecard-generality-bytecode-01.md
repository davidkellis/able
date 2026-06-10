# External Benchmark Comparison

- Generated: `2026-07-15T08:37:01.116353Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (3) | verified (3) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1500 | n/a | n/a | 40.1801 | 0.00x |
| `binarytrees` | `bytecode` | timeout (3) | not run | n/a | n/a | n/a | n/a | n/a | n/a |
| `matrixmultiply` | `bytecode` | ok (3) | verified (3) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 3.8100 | 44.0134 | 0.09x | 41.9293 | 0.09x |
