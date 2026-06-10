# External Benchmark Comparison

- Generated: `2026-07-14T16:34:39.676550Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.5500 | 3.1392 | 1.13x |
| `binarytrees` | `compiled` | ok (1) | verified (1) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 30.2700 | 33.3653 | 0.91x |
| `matrixmultiply` | `compiled` | ok (1) | verified (1) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.2100 | 0.9940 | 1.22x |
