# External Benchmark Comparison

- Generated: `2026-07-20T17:19:42.891815Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-wide-integer-records-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-wide-integer-records-go-reference.json`
- Suite: `custom`
- Able benchmarks: `wide_integer_records`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `wide_integer_records` | `compiled` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 0.1640 | 0.0238 | 6.89x | n/a | n/a | n/a | n/a |
| `wide_integer_records` | `bytecode` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 5.0920 | n/a | n/a | 0.0636 | 80.06x | 0.1579 | 32.25x |
