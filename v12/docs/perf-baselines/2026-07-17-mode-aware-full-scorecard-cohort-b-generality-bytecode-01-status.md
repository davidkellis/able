# External Benchmark Comparison

- Generated: `2026-07-17T16:01:05.352068Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-interpreter-reference-status.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1500 | n/a | n/a | 45.9127 | 0.00x |
| `binarytrees` | `bytecode` | timeout (1) | not run | n/a | n/a | n/a | n/a | 55.1481 | n/a |
