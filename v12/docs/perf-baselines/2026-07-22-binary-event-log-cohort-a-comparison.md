# External Benchmark Comparison

- Generated: `2026-07-22T05:53:20.952018Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-binary-event-log-interpreter-reference-a.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-binary-event-log-go-reference-a.json`
- Suite: `custom`
- Able benchmarks: `binary_event_log`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `binary_event_log` | `compiled` | ok (5) | verified (5) | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 | 0.7480 | 0.0153 | 48.89x | n/a | n/a | n/a | n/a |
| `binary_event_log` | `bytecode` | ok (5) | verified (5) | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 | 8.8660 | n/a | n/a | 0.3106 | 28.54x | 0.2234 | 39.69x |
