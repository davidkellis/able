# External Benchmark Comparison

- Generated: `2026-07-29T20:53:57.544970Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-generic-slot-buffer-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-generic-slot-buffer-go-reference.json`
- Suite: `custom`
- Able benchmarks: `generic_slot_buffer`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generic_slot_buffer` | `compiled` | ok (5) | verified (5) | 149cd95dcb57f9309c82ccd148336280f98baa95ea3d91ba34be7989fdab06fe | 0.0420 | 0.0051 | 8.24x | n/a | n/a | n/a | n/a |
| `generic_slot_buffer` | `bytecode` | ok (5) | verified (5) | 149cd95dcb57f9309c82ccd148336280f98baa95ea3d91ba34be7989fdab06fe | 2.2580 | n/a | n/a | 0.1847 | 12.23x | 0.1037 | 21.77x |
