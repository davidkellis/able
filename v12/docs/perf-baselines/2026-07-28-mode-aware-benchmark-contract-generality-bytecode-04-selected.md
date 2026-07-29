# External Benchmark Comparison

- Generated: `2026-07-28T19:46:47.239929Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `monte_carlo_pi, pidigits, mandelbrot`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `monte_carlo_pi` | `bytecode` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 2.7140 | 1.4459 | 1.88x | 1.5447 | 1.76x |
| `pidigits` | `bytecode` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 2.4320 | 4.0342 | 0.60x | 10.0133 | 0.24x |
| `mandelbrot` | `bytecode` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 6.5320 | 1.3706 | 4.77x | 1.9661 | 3.32x |
