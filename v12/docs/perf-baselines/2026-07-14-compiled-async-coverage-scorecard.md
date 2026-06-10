# External Benchmark Comparison

- Generated: `2026-07-14T04:06:12.212768Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-async-coverage-go-refresh.json`
- Suite: `core`
- Able benchmarks: `channel_rollup, future_pipeline, future_await_race, await_channel_mux, mutex_ledger, mutex_await_journal`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `compiled` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 3.3500 | 0.0056 | 598.21x |
| `future_pipeline` | `compiled` | ok (3) | verified (3) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.8867 | 0.0053 | 167.30x |
| `future_await_race` | `compiled` | ok (3) | verified (3) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1200 | 0.0037 | 32.43x |
| `await_channel_mux` | `compiled` | ok (3) | verified (3) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.4067 | 0.0046 | 88.41x |
| `mutex_ledger` | `compiled` | ok (3) | verified (3) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.5733 | 0.0042 | 136.50x |
| `mutex_await_journal` | `compiled` | ok (3) | verified (3) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.4700 | 0.0040 | 117.50x |
