# External Benchmark Comparison

- Generated: `2026-07-16T09:24:48.428029Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-async-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `channel_rollup, future_pipeline, future_await_race`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5720 | 0.0368 | 15.54x | 0.0477 | 11.99x |
| `future_pipeline` | `bytecode` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.4960 | 0.0593 | 8.36x | 0.0689 | 7.20x |
| `future_await_race` | `bytecode` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1580 | 0.0294 | 5.37x | 0.0499 | 3.17x |
