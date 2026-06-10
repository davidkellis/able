# External Benchmark Comparison

- Generated: `2026-07-15T08:55:04.755425Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-async-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `channel_rollup, future_pipeline, future_await_race`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `bytecode` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5600 | 0.0365 | 15.34x | 0.0471 | 11.89x |
| `future_pipeline` | `bytecode` | ok (3) | verified (3) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.4100 | 0.0542 | 7.56x | 0.0650 | 6.31x |
| `future_await_race` | `bytecode` | ok (3) | verified (3) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1467 | 0.0280 | 5.24x | 0.0489 | 3.00x |
