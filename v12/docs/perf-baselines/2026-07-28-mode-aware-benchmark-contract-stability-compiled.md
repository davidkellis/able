# External Benchmark Comparison

- Generated: `2026-07-28T20:53:22.020498Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/var/tmp/able-mode-aware-stability-20260728-JbVhIT/go.json`
- Suite: `custom`
- Able benchmarks: `binarytrees, quicksort, base64, json, monte_carlo_pi, pidigits`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `2,8,13,0` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `binarytrees` | `compiled` | ok (5) | verified (5) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 11.9720 | 11.8174 | 1.01x |
| `quicksort` | `compiled` | ok (5) | verified (5) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.7940 | 2.5462 | 0.70x |
| `base64` | `compiled` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.3540 | 2.5190 | 0.93x |
| `json` | `compiled` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.6780 | 1.5051 | 0.45x |
| `monte_carlo_pi` | `compiled` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.1960 | 0.2001 | 0.98x |
| `pidigits` | `compiled` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.1780 | 1.2357 | 0.95x |
