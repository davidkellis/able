# External Benchmark Comparison

- Generated: `2026-07-20T19:00:49.379834Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-log-routing-redaction-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-log-routing-redaction-go-reference.json`
- Suite: `custom`
- Able benchmarks: `log_routing_redaction`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `log_routing_redaction` | `compiled` | ok (5) | verified (5) | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 | 0.1160 | 0.0049 | 23.67x | n/a | n/a | n/a | n/a |
| `log_routing_redaction` | `bytecode` | ok (5) | verified (5) | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 | 2.9480 | n/a | n/a | 0.0431 | 68.40x | 0.0187 | 157.65x |
