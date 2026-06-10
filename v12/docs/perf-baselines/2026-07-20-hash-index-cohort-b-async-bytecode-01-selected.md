# External Benchmark Comparison

- Generated: `2026-07-20T14:39:07.002951Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-async-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `channel_rollup, future_pipeline, future_await_race`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5440 | 0.0457 | 11.90x | 0.0720 | 7.56x |
| `future_pipeline` | `bytecode` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.4680 | 0.0619 | 7.56x | 0.0797 | 5.87x |
| `future_await_race` | `bytecode` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1640 | 0.0305 | 5.38x | 0.0606 | 2.71x |
