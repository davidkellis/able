# External Benchmark Comparison

- Generated: `2026-07-20T16:53:38.407335Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `document_audit, lexical_rollup, regex_suffix_audit`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `bytecode` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3360 | 0.0135 | 24.89x | 0.0416 | 8.08x |
| `lexical_rollup` | `bytecode` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5280 | 0.0164 | 32.20x | 0.0472 | 11.19x |
| `regex_suffix_audit` | `bytecode` | ok (5) | verified (5) | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f | 3.3260 | 0.0169 | 196.80x | 0.0431 | 77.17x |
