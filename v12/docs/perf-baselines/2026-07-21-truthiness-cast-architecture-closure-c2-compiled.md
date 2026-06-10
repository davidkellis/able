# External Benchmark Comparison

- Generated: `2026-07-22T02:13:35.452902Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-architecture-closure-c2-go-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, channel_rollup, concurrent_event_routing, concurrent_text_index, dependency_wave_validation, future_await_race, future_pipeline, mutex_await_journal, mutex_ledger, mutex_work_queue, validated_job_pipeline`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3720 | 0.0060 | 62.00x |
| `channel_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5800 | 0.0079 | 73.42x |
| `concurrent_event_routing` | `compiled` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 2.9440 | 0.0067 | 439.40x |
| `concurrent_text_index` | `compiled` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 1.0060 | 0.0074 | 135.95x |
| `dependency_wave_validation` | `compiled` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 1.3160 | 0.0065 | 202.46x |
| `future_await_race` | `compiled` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.0860 | 0.0047 | 18.30x |
| `future_pipeline` | `compiled` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.3700 | 0.0069 | 53.62x |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.8040 | 0.0042 | 191.43x |
| `mutex_ledger` | `compiled` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.7040 | 0.0047 | 149.79x |
| `mutex_work_queue` | `compiled` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 1.9580 | 0.0043 | 455.35x |
| `validated_job_pipeline` | `compiled` | ok (5) | verified (5) | 96cca38f1e5b45bea159f191a7a49507fc3cc26613c759617a30a27af27db9e2 | 2.9960 | 0.0065 | 460.92x |
