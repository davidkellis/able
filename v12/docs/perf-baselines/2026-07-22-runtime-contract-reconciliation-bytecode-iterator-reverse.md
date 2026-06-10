# External Benchmark Comparison

- Generated: `2026-07-23T03:47:49.085971Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-22-runtime-contract-reconciliation-90s-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `option_result_config, lexical_rollup, document_audit, dependency_plan, binary_event_log, array_slice_window`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 1.4840 | 0.0378 | 39.26x | 0.1144 | 12.97x |
| `lexical_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.7060 | 0.0227 | 31.10x | 0.0602 | 11.73x |
| `document_audit` | `bytecode` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3480 | 0.0140 | 24.86x | 0.0466 | 7.47x |
| `dependency_plan` | `bytecode` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.5760 | 0.0360 | 16.00x | 0.1168 | 4.93x |
| `binary_event_log` | `bytecode` | ok (5) | verified (5) | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 | 9.0140 | 0.3633 | 24.81x | 0.4953 | 18.20x |
| `array_slice_window` | `bytecode` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 1.7580 | 0.0642 | 27.38x | 0.1602 | 10.97x |
