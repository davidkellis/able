# External Benchmark Comparison

- Generated: `2026-07-20T10:00:17.378592Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.9760 | 3.2577 | 1.22x |
| `binarytrees` | `compiled` | ok (5) | verified (5) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 9.6600 | 10.6970 | 0.90x |
| `matrixmultiply` | `compiled` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.1040 | 1.0892 | 1.01x |
