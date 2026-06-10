# External Benchmark Comparison

- Generated: `2026-07-20T20:14:26.720817Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `regex_set_audit, regex_stream_audit, log_routing_redaction, array_slice_window`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `regex_set_audit` | `bytecode` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 4.3480 | 0.0222 | 195.86x | 0.0458 | 94.93x |
| `regex_stream_audit` | `bytecode` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 3.7880 | 0.0239 | 158.49x | 0.0500 | 75.76x |
| `log_routing_redaction` | `bytecode` | ok (5) | verified (5) | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 | 2.9580 | 0.0181 | 163.43x | 0.0422 | 70.09x |
| `array_slice_window` | `bytecode` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.6640 | 0.0299 | 22.21x | 0.0655 | 10.14x |
