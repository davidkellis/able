# External Benchmark Comparison

- Generated: `2026-07-14T07:14:09.500986Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-bytecode-coverage-interpreter-refs.json`
- Suite: `custom`
- Able benchmarks: `word_frequency, document_audit, lexical_rollup`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `word_frequency` | `bytecode` | ok (3) | verified (3) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.4933 | 0.0302 | 49.45x | 0.0636 | 23.48x |
| `document_audit` | `bytecode` | ok (3) | verified (3) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3100 | 0.0264 | 11.74x | 0.0543 | 5.71x |
| `lexical_rollup` | `bytecode` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.4033 | 0.0296 | 13.62x | 0.0621 | 6.49x |
