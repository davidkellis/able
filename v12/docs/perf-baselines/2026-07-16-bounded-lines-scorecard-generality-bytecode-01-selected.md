# External Benchmark Comparison

- Generated: `2026-07-16T09:07:56.521681Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `fib, matrixmultiply`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1480 | 54.3191 | 0.00x | 43.2450 | 0.00x |
| `matrixmultiply` | `bytecode` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 4.6460 | 48.4717 | 0.10x | 44.2432 | 0.11x |
