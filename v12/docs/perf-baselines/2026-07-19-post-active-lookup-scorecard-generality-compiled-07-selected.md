# External Benchmark Comparison

- Generated: `2026-07-19T07:15:44.917809Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `distance_field, rms_norm`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `distance_field` | `compiled` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 0.1340 | 0.0117 | 11.45x |
| `rms_norm` | `compiled` | ok (5) | verified (5) | 255c3e1c7ae7f523918e96244a6ac395b58699c4d2220549b097702faaa1037b | 0.1360 | 0.0106 | 12.83x |
