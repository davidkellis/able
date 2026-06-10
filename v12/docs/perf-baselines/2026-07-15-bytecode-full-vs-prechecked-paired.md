# External Benchmark Comparison

- Generated: `2026-07-15T09:47:09.860007Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `custom`
- Able benchmarks: `sudoku, word_frequency, future_pipeline`
- Able modes: `bytecode, bytecode-prechecked`
- `bytecode-prechecked` performs one `able check` outside timing, then uses trusted `run --skip-typecheck`; it is not scorecard input.
- Reference languages: `python, ruby`
- CPU affinity: `8`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `sudoku` | `bytecode` | timeout (3) | not run | n/a | n/a | 3.0200 | n/a | 5.6700 | n/a |
| `sudoku` | `bytecode-prechecked` | timeout (3) | not run | n/a | n/a | 3.0200 | n/a | 5.6700 | n/a |
| `word_frequency` | `bytecode` | ok (3) | verified (3) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.3300 | n/a | n/a | n/a | n/a |
| `word_frequency` | `bytecode-prechecked` | ok (3) | verified (3) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 1.3200 | n/a | n/a | n/a | n/a |
| `future_pipeline` | `bytecode` | ok (3) | verified (3) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.4100 | n/a | n/a | n/a | n/a |
| `future_pipeline` | `bytecode-prechecked` | ok (3) | verified (3) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.4100 | n/a | n/a | n/a | n/a |
