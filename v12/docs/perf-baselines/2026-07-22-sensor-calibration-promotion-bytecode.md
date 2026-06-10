# External Benchmark Comparison

- Generated: `2026-07-23T00:18:54.891829Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-sensor-calibration-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `sensor_calibration`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `sensor_calibration` | `bytecode` | ok (5) | verified (5) | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb | 3.7940 | 0.0331 | 114.62x | 0.0732 | 51.83x |
