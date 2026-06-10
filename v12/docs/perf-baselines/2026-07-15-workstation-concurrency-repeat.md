# External Benchmark Comparison

- Generated: `2026-07-15T19:47:01.033995Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/tmp/able-workstation-concurrency-interpreters.json`
- Fresh Go reference rows: `/tmp/able-workstation-concurrency-go.json`
- Suite: `custom`
- Able benchmarks: `future_await_race, await_channel_mux, mutex_await_journal`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `future_await_race` | `compiled` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.0700 | 0.0030 | 23.33x | 0.0490 | 1.43x | 0.0301 | 2.33x |
| `future_await_race` | `bytecode` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1100 | 0.0030 | 36.67x | 0.0490 | 2.24x | 0.0301 | 3.65x |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3140 | 0.0038 | 82.63x | 0.0908 | 3.46x | 0.1167 | 2.69x |
| `await_channel_mux` | `bytecode` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.1920 | 0.0038 | 50.53x | 0.0908 | 2.11x | 0.1167 | 1.65x |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.7000 | 0.0030 | 233.33x | 0.0392 | 17.86x | 0.0177 | 39.55x |
| `mutex_await_journal` | `bytecode` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.1520 | 0.0030 | 50.67x | 0.0392 | 3.88x | 0.0177 | 8.59x |
