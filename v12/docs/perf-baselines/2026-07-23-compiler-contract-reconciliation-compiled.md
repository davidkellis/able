# External Benchmark Comparison

- Generated: `2026-07-23T10:25:52.233859Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `custom`
- Able benchmarks: `sudoku, sudoku_masks, nbody, dependency_plan, future_pipeline, await_channel_mux, concurrent_stencil_reduction`
- Able modes: `compiled`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `sudoku` | `compiled` | timeout (5) | not run | n/a | n/a | 0.1300 | n/a | n/a | n/a | n/a | n/a |
| `sudoku_masks` | `compiled` | ok (5) | verified (5) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 2.0800 | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `compiled` | ok (5) | verified (5) | 40799ff8af9b84a416e8bf940921658787c57be38f638fb4d98c735c8d39e820 | 0.1880 | n/a | n/a | n/a | n/a | n/a | n/a |
| `dependency_plan` | `compiled` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.1360 | n/a | n/a | n/a | n/a | n/a | n/a |
| `future_pipeline` | `compiled` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.3540 | n/a | n/a | n/a | n/a | n/a | n/a |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.4020 | n/a | n/a | n/a | n/a | n/a | n/a |
| `concurrent_stencil_reduction` | `compiled` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 0.2760 | n/a | n/a | n/a | n/a | n/a | n/a |
