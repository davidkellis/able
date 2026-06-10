# External Benchmark Comparison

- Generated: `2026-07-20T20:08:39.997087Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `regex_set_audit, regex_stream_audit, log_routing_redaction, array_slice_window`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `regex_set_audit` | `compiled` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 0.1120 | 0.0050 | 22.40x |
| `regex_stream_audit` | `compiled` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 0.1160 | 0.0045 | 25.78x |
| `log_routing_redaction` | `compiled` | ok (5) | verified (5) | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 | 0.1000 | 0.0043 | 23.26x |
| `array_slice_window` | `compiled` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.0860 | 0.0041 | 20.98x |
