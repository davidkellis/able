# External Benchmark Comparison

- Generated: `2026-07-14T17:04:08.025465Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-full-scorecard-async-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `channel_rollup, future_pipeline, future_await_race`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `bytecode` | ok (1) | verified (1) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.6500 | 0.0398 | 16.33x | 0.0556 | 11.69x |
| `future_pipeline` | `bytecode` | ok (1) | verified (1) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.4800 | 0.0660 | 7.27x | 0.0695 | 6.91x |
| `future_await_race` | `bytecode` | ok (1) | verified (1) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1600 | 0.0294 | 5.44x | 0.0549 | 2.91x |
