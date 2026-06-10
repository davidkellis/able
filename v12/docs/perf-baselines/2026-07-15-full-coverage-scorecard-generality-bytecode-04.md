# External Benchmark Comparison

- Generated: `2026-07-15T08:44:53.048451Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `monte_carlo_pi, pidigits, mandelbrot`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `monte_carlo_pi` | `bytecode` | ok (3) | verified (3) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 2.1433 | 1.2851 | 1.67x | 1.3522 | 1.59x |
| `pidigits` | `bytecode` | ok (3) | verified (3) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 2.2667 | 3.6258 | 0.63x | 8.8207 | 0.26x |
| `mandelbrot` | `bytecode` | ok (3) | verified (3) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 6.1400 | 1.0302 | 5.96x | 1.6353 | 3.75x |
