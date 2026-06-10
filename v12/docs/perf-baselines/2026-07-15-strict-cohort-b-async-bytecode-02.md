# External Benchmark Comparison

- Generated: `2026-07-16T06:28:02.653579Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-b-async-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, mutex_ledger, mutex_await_journal`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `bytecode` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.2640 | 0.1235 | 2.14x | 0.0979 | 2.70x |
| `mutex_ledger` | `bytecode` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.7180 | 0.0317 | 22.65x | 0.0519 | 13.83x |
| `mutex_await_journal` | `bytecode` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.2540 | 0.0195 | 13.03x | 0.0468 | 5.43x |
