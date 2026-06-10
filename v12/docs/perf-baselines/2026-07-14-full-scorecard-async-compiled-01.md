# External Benchmark Comparison

- Generated: `2026-07-14T17:01:46.912855Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-full-scorecard-async-go-reference.json`
- Suite: `custom`
- Able benchmarks: `channel_rollup, future_pipeline, future_await_race`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `compiled` | ok (1) | verified (1) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 1.3400 | 0.0052 | 257.69x |
| `future_pipeline` | `compiled` | ok (1) | verified (1) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.7100 | 0.0049 | 144.90x |
| `future_await_race` | `compiled` | ok (1) | verified (1) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1200 | 0.0038 | 31.58x |
