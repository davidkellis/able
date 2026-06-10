# External Benchmark Comparison

- Generated: `2026-07-22T01:33:50.429062Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-regex-concurrency-closures-c2-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, channel_rollup, concurrent_event_routing, concurrent_text_index, dependency_wave_validation, future_await_race, future_pipeline, mutex_await_journal, mutex_ledger, mutex_work_queue, validated_job_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `bytecode` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.2040 | 0.1176 | 1.73x | 0.1030 | 1.98x |
| `channel_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.4360 | 0.0464 | 9.40x | 0.0536 | 8.13x |
| `concurrent_event_routing` | `bytecode` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 2.8560 | 0.0309 | 92.43x | 0.0554 | 51.55x |
| `concurrent_text_index` | `bytecode` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 0.6200 | 0.0632 | 9.81x | 0.0852 | 7.28x |
| `dependency_wave_validation` | `bytecode` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 0.4640 | 0.0340 | 13.65x | 0.0523 | 8.87x |
| `future_await_race` | `bytecode` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1400 | 0.0318 | 4.40x | 0.0564 | 2.48x |
| `future_pipeline` | `bytecode` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.4080 | 0.0656 | 6.22x | 0.0769 | 5.31x |
| `mutex_await_journal` | `bytecode` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.1940 | 0.0253 | 7.67x | 0.0460 | 4.22x |
| `mutex_ledger` | `bytecode` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.3240 | 0.0350 | 9.26x | 0.0598 | 5.42x |
| `mutex_work_queue` | `bytecode` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 0.3240 | 0.0327 | 9.91x | 0.0590 | 5.49x |
| `validated_job_pipeline` | `bytecode` | ok (5) | verified (5) | 96cca38f1e5b45bea159f191a7a49507fc3cc26613c759617a30a27af27db9e2 | 0.7820 | 0.0638 | 12.26x | 0.0561 | 13.94x |
