# External Benchmark Comparison

- Generated: `2026-07-16T06:42:57.744615Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-go-reference.json`
- Suite: `custom`
- Able benchmarks: `regex_set_audit, regex_stream_audit, array_slice_window`
- Able modes: `compiled`
- Reference languages: `go`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `regex_set_audit` | `compiled` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 0.2080 | 0.0041 | 50.73x |
| `regex_stream_audit` | `compiled` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 0.1980 | 0.0041 | 48.29x |
| `array_slice_window` | `compiled` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.0920 | 0.0037 | 24.86x |
