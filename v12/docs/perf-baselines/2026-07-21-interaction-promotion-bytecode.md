# External Benchmark Comparison

- Generated: `2026-07-21T05:23:23.328393Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-interaction-promotion-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `concurrent_text_index, validated_job_pipeline, dependency_wave_validation, concurrent_event_routing`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_text_index` | `bytecode` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 0.6340 | 0.0844 | 7.51x | 0.1024 | 6.19x |
| `validated_job_pipeline` | `bytecode` | ok (5) | verified (5) | 96cca38f1e5b45bea159f191a7a49507fc3cc26613c759617a30a27af27db9e2 | 0.8340 | 0.1056 | 7.90x | 0.0838 | 9.95x |
| `dependency_wave_validation` | `bytecode` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 0.4200 | 0.0480 | 8.75x | 0.0514 | 8.17x |
| `concurrent_event_routing` | `bytecode` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 3.0320 | 0.0297 | 102.09x | 0.0488 | 62.13x |
