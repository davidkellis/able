# External Benchmark Comparison

- Generated: `2026-07-15T06:39:12.591807Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-async-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, mutex_ledger, mutex_await_journal`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `bytecode` | ok (1) | verified (1) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.2300 | 0.1094 | 2.10x | 0.0904 | 2.54x |
| `mutex_ledger` | `bytecode` | ok (1) | verified (1) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.6100 | 0.0312 | 19.55x | 0.0492 | 12.40x |
| `mutex_await_journal` | `bytecode` | ok (1) | verified (1) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.2400 | 0.0193 | 12.44x | 0.0402 | 5.97x |
