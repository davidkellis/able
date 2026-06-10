# External Benchmark Comparison

- Generated: `2026-07-28T04:57:00.874700Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-27-compiled-callable-context-future-go-reference.json`
- Suite: `custom`
- Able benchmarks: `future_await_race`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)
- Experimental execution context: `enabled for compiled mode`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `future_await_race` | `compiled` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.0360 | 0.0046 | 7.83x |
