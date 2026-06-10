# External Benchmark Comparison

- Generated: `2026-07-15T08:30:56.124259Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `monte_carlo_pi, pidigits, mandelbrot`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `monte_carlo_pi` | `compiled` | ok (3) | verified (3) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.1967 | 0.1837 | 1.07x |
| `pidigits` | `compiled` | ok (3) | verified (3) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.2067 | 1.0615 | 1.14x |
| `mandelbrot` | `compiled` | ok (3) | verified (3) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.1400 | 0.0461 | 3.04x |
