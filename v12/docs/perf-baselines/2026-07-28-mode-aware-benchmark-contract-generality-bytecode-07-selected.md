# External Benchmark Comparison

- Generated: `2026-07-28T19:54:23.937604Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `distance_field, rms_norm`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `distance_field` | `bytecode` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 5.7600 | 0.5807 | 9.92x | 0.3396 | 16.96x |
| `rms_norm` | `bytecode` | ok (5) | verified (5) | 255c3e1c7ae7f523918e96244a6ac395b58699c4d2220549b097702faaa1037b | 4.6820 | 0.7947 | 5.89x | 0.5263 | 8.90x |
