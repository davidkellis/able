# External Benchmark Comparison

- Generated: `2026-07-23T13:13:58.403356Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-signal-dispatch-interpreter-reference-a.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-signal-dispatch-go-reference-a.json`
- Suite: `custom`
- Able benchmarks: `concurrent_signal_dispatch`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_signal_dispatch` | `compiled` | ok (5) | verified (5) | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 | 0.2700 | 0.0052 | 51.92x | n/a | n/a | n/a | n/a |
| `concurrent_signal_dispatch` | `bytecode` | ok (5) | verified (5) | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 | 1.6140 | n/a | n/a | 0.0922 | 17.51x | 0.0612 | 26.37x |
