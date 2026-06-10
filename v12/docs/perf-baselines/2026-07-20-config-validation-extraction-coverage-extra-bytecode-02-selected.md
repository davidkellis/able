# External Benchmark Comparison

- Generated: `2026-07-20T21:54:24.619749Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `document_audit, lexical_rollup, regex_suffix_audit`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `bytecode` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.4200 | 0.0135 | 31.11x | 0.0442 | 9.50x |
| `lexical_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.4760 | 0.0180 | 26.44x | 0.0506 | 9.41x |
| `regex_suffix_audit` | `bytecode` | ok (5) | verified (5) | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f | 3.3800 | 0.0188 | 179.79x | 0.0452 | 74.78x |
