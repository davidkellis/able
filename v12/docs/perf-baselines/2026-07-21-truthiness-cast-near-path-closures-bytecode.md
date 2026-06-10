# External Benchmark Comparison

- Generated: `2026-07-21T22:31:53.345857Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-near-path-closures-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `distance_field, mandelbrot, monte_carlo_pi, rms_norm`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `distance_field` | `bytecode` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 6.7260 | 0.6968 | 9.65x | 0.4783 | 14.06x |
| `mandelbrot` | `bytecode` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 7.2000 | 1.3980 | 5.15x | 2.1426 | 3.36x |
| `monte_carlo_pi` | `bytecode` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 3.4200 | 2.1717 | 1.57x | 2.1583 | 1.58x |
| `rms_norm` | `bytecode` | ok (5) | verified (5) | 255c3e1c7ae7f523918e96244a6ac395b58699c4d2220549b097702faaa1037b | 7.7000 | 1.0289 | 7.48x | 0.6665 | 11.55x |
