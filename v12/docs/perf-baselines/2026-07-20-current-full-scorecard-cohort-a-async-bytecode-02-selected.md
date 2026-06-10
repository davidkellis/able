# External Benchmark Comparison

- Generated: `2026-07-20T10:29:28.588542Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-async-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, mutex_ledger, mutex_await_journal`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `bytecode` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3660 | 0.1148 | 3.19x | 0.0956 | 3.83x |
| `mutex_ledger` | `bytecode` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.4900 | 0.0326 | 15.03x | 0.0535 | 9.16x |
| `mutex_await_journal` | `bytecode` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.2400 | 0.0199 | 12.06x | 0.0454 | 5.29x |
