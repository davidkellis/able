# External Benchmark Comparison

- Generated: `2026-07-23T10:32:20.354913Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `custom`
- Able benchmarks: `future_pipeline, await_channel_mux, concurrent_stencil_reduction`
- Able modes: `bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `future_pipeline` | `bytecode` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.5200 | n/a | n/a | n/a | n/a | n/a | n/a |
| `await_channel_mux` | `bytecode` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.2500 | n/a | n/a | n/a | n/a | n/a | n/a |
| `concurrent_stencil_reduction` | `bytecode` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 1.8700 | n/a | n/a | n/a | n/a | n/a | n/a |
