# External Benchmark Comparison

- Generated: `2026-07-15T08:54:55.736896Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-async-go-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, mutex_ledger, mutex_await_journal`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `compiled` | ok (3) | verified (3) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3300 | 0.0044 | 75.00x |
| `mutex_ledger` | `compiled` | ok (3) | verified (3) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.5100 | 0.0041 | 124.39x |
| `mutex_await_journal` | `compiled` | ok (3) | verified (3) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.4200 | 0.0038 | 110.53x |
