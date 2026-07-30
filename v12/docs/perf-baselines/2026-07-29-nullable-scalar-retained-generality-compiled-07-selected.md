# External Benchmark Comparison

- Generated: `2026-07-29T23:04:28.176580Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `distance_field, rms_norm`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `12-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `distance_field` | `compiled` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 0.0380 | 0.0147 | 2.59x |
| `rms_norm` | `compiled` | ok (5) | verified (5) | 255c3e1c7ae7f523918e96244a6ac395b58699c4d2220549b097702faaa1037b | 0.0300 | 0.0140 | 2.14x |
