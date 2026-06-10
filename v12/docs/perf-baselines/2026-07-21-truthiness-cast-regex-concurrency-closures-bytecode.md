# External Benchmark Comparison

- Generated: `2026-07-22T01:31:58.978520Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-regex-concurrency-closures-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, channel_rollup, concurrent_event_routing, concurrent_text_index, dependency_wave_validation, future_await_race, future_pipeline, mutex_await_journal, mutex_ledger, mutex_work_queue, validated_job_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `bytecode` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.2760 | 0.1167 | 2.37x | 0.1024 | 2.70x |
| `channel_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.4240 | 0.0636 | 6.67x | 0.0676 | 6.27x |
| `concurrent_event_routing` | `bytecode` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 2.8140 | 0.0338 | 83.25x | 0.0591 | 47.61x |
| `concurrent_text_index` | `bytecode` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 0.5680 | 0.0716 | 7.93x | 0.0871 | 6.52x |
| `dependency_wave_validation` | `bytecode` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 0.5380 | 0.0353 | 15.24x | 0.0559 | 9.62x |
| `future_await_race` | `bytecode` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1400 | 0.0334 | 4.19x | 0.0647 | 2.16x |
| `future_pipeline` | `bytecode` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.4100 | 0.0625 | 6.56x | 0.0820 | 5.00x |
| `mutex_await_journal` | `bytecode` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.1840 | 0.0262 | 7.02x | 0.0565 | 3.26x |
| `mutex_ledger` | `bytecode` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.3260 | 0.0434 | 7.51x | 0.0624 | 5.22x |
| `mutex_work_queue` | `bytecode` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 0.3220 | 0.0295 | 10.92x | 0.0508 | 6.34x |
| `validated_job_pipeline` | `bytecode` | ok (5) | verified (5) | 96cca38f1e5b45bea159f191a7a49507fc3cc26613c759617a30a27af27db9e2 | 0.7580 | 0.0625 | 12.13x | 0.0610 | 12.43x |
