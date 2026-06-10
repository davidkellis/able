# External Benchmark Comparison

- Generated: `2026-07-17T14:27:47.951550Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.2620 | 2.9887 | 1.09x |
| `binarytrees` | `compiled` | ok (5) | verified (5) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 9.7260 | 10.1090 | 0.96x |
| `matrixmultiply` | `compiled` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.4640 | 1.0017 | 1.46x |
