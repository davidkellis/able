# External Benchmark Comparison

- Generated: `2026-07-18T16:55:00.557230Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-async-go-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, mutex_ledger, mutex_await_journal`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.5180 | 0.0072 | 71.94x |
| `mutex_ledger` | `compiled` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.9240 | 0.0057 | 162.11x |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.8780 | 0.0052 | 168.85x |
