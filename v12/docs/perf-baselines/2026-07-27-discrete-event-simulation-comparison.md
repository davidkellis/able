# External Benchmark Comparison

- Generated: `2026-07-27T20:00:44.352662Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-27-discrete-event-simulation-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-27-discrete-event-simulation-go-reference.json`
- Suite: `custom`
- Able benchmarks: `discrete_event_simulation`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `discrete_event_simulation` | `compiled` | ok (5) | verified (5) | 6aebca9b31a78441438d2321290a7b66dc831ddbc7671d783e4a725aed6e7405 | 0.0580 | 0.0138 | 4.20x | n/a | n/a | n/a | n/a |
| `discrete_event_simulation` | `bytecode` | ok (5) | verified (5) | 6aebca9b31a78441438d2321290a7b66dc831ddbc7671d783e4a725aed6e7405 | 4.7360 | n/a | n/a | 0.1698 | 27.89x | 0.2095 | 22.61x |
