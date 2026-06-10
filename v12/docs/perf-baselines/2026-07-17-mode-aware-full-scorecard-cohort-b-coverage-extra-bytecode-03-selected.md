# External Benchmark Comparison

- Generated: `2026-07-17T16:32:44.279961Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `regex_set_audit, regex_stream_audit, array_slice_window`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `regex_set_audit` | `bytecode` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 5.1860 | 0.0273 | 189.96x | 0.0555 | 93.44x |
| `regex_stream_audit` | `bytecode` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 4.5220 | 0.0236 | 191.61x | 0.0483 | 93.62x |
| `array_slice_window` | `bytecode` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.7040 | 0.0337 | 20.89x | 0.0806 | 8.73x |
