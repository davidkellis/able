# External Benchmark Comparison

- Generated: `2026-07-13T18:29:18.915811Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-13-bytecode-coverage-extension-interpreter-refresh.json`
- Suite: `core`
- Able benchmarks: `fixed_width_128, rational_series, word_frequency, document_audit, lexical_rollup, channel_rollup, future_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fixed_width_128` | `bytecode` | ok (3) | verified (3) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 10.1600 | 0.3694 | 27.50x | 0.7169 | 14.17x |
| `rational_series` | `bytecode` | ok (3) | verified (3) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 4.9567 | 0.1069 | 46.37x | 0.1413 | 35.08x |
| `word_frequency` | `bytecode` | ok (3) | verified (3) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.7500 | 0.0195 | 89.74x | 0.0593 | 29.51x |
| `document_audit` | `bytecode` | ok (3) | verified (3) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3667 | 0.0146 | 25.12x | 0.0419 | 8.75x |
| `lexical_rollup` | `bytecode` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5000 | 0.0186 | 26.88x | 0.0524 | 9.54x |
| `channel_rollup` | `bytecode` | ok (3) | verified (3) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.5633 | 0.0413 | 13.64x | 0.0567 | 9.93x |
| `future_pipeline` | `bytecode` | ok (3) | verified (3) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.5267 | 0.0619 | 8.51x | 0.0713 | 7.39x |
