# External Benchmark Comparison

- Generated: `2026-07-15T08:53:42.961251Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-async-go-reference.json`
- Suite: `custom`
- Able benchmarks: `channel_rollup, future_pipeline, future_await_race`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `compiled` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 1.1367 | 0.0052 | 218.60x |
| `future_pipeline` | `compiled` | ok (3) | verified (3) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.6167 | 0.0065 | 94.88x |
| `future_await_race` | `compiled` | ok (3) | verified (3) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1200 | 0.0039 | 30.77x |
