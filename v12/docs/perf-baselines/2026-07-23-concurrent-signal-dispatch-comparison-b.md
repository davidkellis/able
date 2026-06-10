# External Benchmark Comparison

- Generated: `2026-07-23T13:15:05.509974Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-signal-dispatch-interpreter-reference-b.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-signal-dispatch-go-reference-b.json`
- Suite: `custom`
- Able benchmarks: `concurrent_signal_dispatch`
- Able modes: `bytecode, compiled`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_signal_dispatch` | `bytecode` | ok (5) | verified (5) | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 | 1.7260 | n/a | n/a | 0.0801 | 21.55x | 0.0653 | 26.43x |
| `concurrent_signal_dispatch` | `compiled` | ok (5) | verified (5) | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 | 0.2960 | 0.0053 | 55.85x | n/a | n/a | n/a | n/a |
