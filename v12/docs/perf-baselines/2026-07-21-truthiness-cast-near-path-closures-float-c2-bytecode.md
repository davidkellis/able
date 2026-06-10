# External Benchmark Comparison

- Generated: `2026-07-21T22:39:29.843231Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-near-path-closures-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `monte_carlo_pi, rms_norm`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `monte_carlo_pi` | `bytecode` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 3.2740 | 2.1717 | 1.51x | 2.1583 | 1.52x |
| `rms_norm` | `bytecode` | ok (5) | verified (5) | 255c3e1c7ae7f523918e96244a6ac395b58699c4d2220549b097702faaa1037b | 5.2460 | 1.0289 | 5.10x | 0.6665 | 7.87x |
