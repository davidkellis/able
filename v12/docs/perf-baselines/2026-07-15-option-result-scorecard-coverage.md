# External Benchmark Comparison

- Generated: `2026-07-15T21:15:02.295159Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-option-result-scorecard-coverage-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-option-result-scorecard-coverage-go-reference.json`
- Suite: `custom`
- Able benchmarks: `option_result_config`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `option_result_config` | `compiled` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.2180 | 0.0030 | 72.67x | 0.0409 | 5.33x | 0.0152 | 14.34x |
| `option_result_config` | `bytecode` | error (5) | not run | n/a | n/a | 0.0030 | n/a | 0.0409 | n/a | 0.0152 | n/a |
