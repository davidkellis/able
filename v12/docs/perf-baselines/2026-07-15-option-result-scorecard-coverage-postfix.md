# External Benchmark Comparison

- Generated: `2026-07-15T21:32:43.761631Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-option-result-scorecard-coverage-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-option-result-scorecard-coverage-go-reference.json`
- Suite: `custom`
- Able benchmarks: `option_result_config`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `option_result_config` | `compiled` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.1960 | 0.0030 | 65.33x | 0.0409 | 4.79x | 0.0152 | 12.89x |
| `option_result_config` | `bytecode` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 3.3880 | 0.0030 | 1129.33x | 0.0409 | 82.84x | 0.0152 | 222.89x |
