# External Benchmark Comparison

- Generated: `2026-07-21T02:37:54.393923Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-current-product-scorecard-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `regex_set_audit, regex_stream_audit, log_routing_redaction, array_slice_window`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `regex_set_audit` | `bytecode` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 4.0200 | 0.0181 | 222.10x | 0.0406 | 99.01x |
| `regex_stream_audit` | `bytecode` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 3.5480 | 0.0183 | 193.88x | 0.0473 | 75.01x |
| `log_routing_redaction` | `bytecode` | ok (5) | verified (5) | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 | 3.0160 | 0.0242 | 124.63x | 0.0630 | 47.87x |
| `array_slice_window` | `bytecode` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.6480 | 0.0294 | 22.04x | 0.0638 | 10.16x |
