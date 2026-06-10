# External Benchmark Comparison

- Generated: `2026-07-23T03:45:56.424058Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-runtime-contract-reconciliation-90s-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `array_slice_window, binary_event_log, dependency_plan, document_audit, lexical_rollup, option_result_config`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `array_slice_window` | `bytecode` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.9760 | 0.0642 | 15.20x | 0.1602 | 6.09x |
| `binary_event_log` | `bytecode` | ok (5) | verified (5) | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 | 8.3120 | 0.3633 | 22.88x | 0.4953 | 16.78x |
| `dependency_plan` | `bytecode` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.6900 | 0.0360 | 19.17x | 0.1168 | 5.91x |
| `document_audit` | `bytecode` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.5560 | 0.0140 | 39.71x | 0.0466 | 11.93x |
| `lexical_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.7480 | 0.0227 | 32.95x | 0.0602 | 12.43x |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 1.3560 | 0.0378 | 35.87x | 0.1144 | 11.85x |
