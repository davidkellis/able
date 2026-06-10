# External Benchmark Comparison

- Generated: `2026-07-18T18:19:31.505151Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-async-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `channel_rollup, future_pipeline, future_await_race, await_channel_mux, mutex_ledger, mutex_await_journal`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `channel_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5660 | 0.0431 | 13.13x | 0.0581 | 9.74x |
| `future_pipeline` | `bytecode` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.5240 | 0.0641 | 8.17x | 0.0837 | 6.26x |
| `future_await_race` | `bytecode` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1880 | 0.0436 | 4.31x | 0.0661 | 2.84x |
| `await_channel_mux` | `bytecode` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.2520 | 0.1247 | 2.02x | 0.1048 | 2.40x |
| `mutex_ledger` | `bytecode` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.3880 | 0.0416 | 9.33x | 0.0528 | 7.35x |
| `mutex_await_journal` | `bytecode` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.2240 | 0.0261 | 8.58x | 0.0626 | 3.58x |
