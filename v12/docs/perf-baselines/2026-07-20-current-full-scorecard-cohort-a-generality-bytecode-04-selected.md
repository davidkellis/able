# External Benchmark Comparison

- Generated: `2026-07-20T10:19:50.371998Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `monte_carlo_pi, pidigits, mandelbrot`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `monte_carlo_pi` | `bytecode` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 2.7740 | 1.5116 | 1.84x | 1.6105 | 1.72x |
| `pidigits` | `bytecode` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 2.8540 | 4.1766 | 0.68x | 10.4766 | 0.27x |
| `mandelbrot` | `bytecode` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 6.6800 | 1.2087 | 5.53x | 1.8993 | 3.52x |
