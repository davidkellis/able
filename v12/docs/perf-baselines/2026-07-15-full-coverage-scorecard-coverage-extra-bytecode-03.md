# External Benchmark Comparison

- Generated: `2026-07-15T09:11:34.805359Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `regex_set_audit, regex_stream_audit, array_slice_window`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `regex_set_audit` | `bytecode` | ok (3) | verified (3) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 4.7000 | 0.0187 | 251.34x | 0.0412 | 114.08x |
| `regex_stream_audit` | `bytecode` | ok (3) | verified (3) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 4.0733 | 0.0191 | 213.26x | 0.0414 | 98.39x |
| `array_slice_window` | `bytecode` | ok (3) | verified (3) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.5600 | 0.0270 | 20.74x | 0.0606 | 9.24x |
