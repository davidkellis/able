# External Benchmark Comparison

- Generated: `2026-07-15T06:26:41.620956Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `monte_carlo_pi, pidigits, mandelbrot`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `monte_carlo_pi` | `compiled` | ok (1) | verified (1) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.1900 | 0.1890 | 1.01x |
| `pidigits` | `compiled` | ok (1) | verified (1) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.2700 | 1.0658 | 1.19x |
| `mandelbrot` | `compiled` | ok (1) | verified (1) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.1800 | 0.0466 | 3.86x |
