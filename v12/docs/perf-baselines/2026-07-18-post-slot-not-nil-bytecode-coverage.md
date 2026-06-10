# External Benchmark Comparison

- Generated: `2026-07-18T19:36:40.500856Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `fixed_width_128, rational_series, word_frequency, document_audit, lexical_rollup, regex_set_audit, regex_stream_audit, array_slice_window, dependency_plan, option_result_config, unicode_scalar_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fixed_width_128` | `bytecode` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 7.6260 | 0.3823 | 19.95x | 0.7323 | 10.41x |
| `rational_series` | `bytecode` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 3.9520 | 0.1162 | 34.01x | 0.1669 | 23.68x |
| `word_frequency` | `bytecode` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.5680 | 0.0228 | 68.77x | 0.0636 | 24.65x |
| `document_audit` | `bytecode` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3420 | 0.0153 | 22.35x | 0.0529 | 6.47x |
| `lexical_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.4320 | 0.0282 | 15.32x | 0.0608 | 7.11x |
| `regex_set_audit` | `bytecode` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 4.0740 | 0.0224 | 181.88x | 0.0497 | 81.97x |
| `regex_stream_audit` | `bytecode` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 3.6940 | 0.0229 | 161.31x | 0.0563 | 65.61x |
| `array_slice_window` | `bytecode` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.6580 | 0.0322 | 20.43x | 0.0710 | 9.27x |
| `dependency_plan` | `bytecode` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.4940 | 0.0193 | 25.60x | 0.0558 | 8.85x |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.8880 | 0.0190 | 46.74x | 0.0552 | 16.09x |
| `unicode_scalar_pipeline` | `bytecode` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 3.5420 | 0.2903 | 12.20x | 0.3431 | 10.32x |
