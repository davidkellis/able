# External Benchmark Comparison

- Generated: `2026-07-20T19:52:20.076500Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-log-routing-redaction-async-go-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, mutex_ledger, mutex_await_journal, mutex_work_queue`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3820 | 0.0048 | 79.58x |
| `mutex_ledger` | `compiled` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.8040 | 0.0046 | 174.78x |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.7780 | 0.0044 | 176.82x |
| `mutex_work_queue` | `compiled` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 1.9160 | 0.0048 | 399.17x |
