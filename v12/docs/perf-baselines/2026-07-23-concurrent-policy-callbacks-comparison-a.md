# External Benchmark Comparison

- Generated: `2026-07-23T15:41:47.349921Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-policy-callbacks-interpreter-reference-a.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-policy-callbacks-go-reference-a.json`
- Suite: `custom`
- Able benchmarks: `concurrent_policy_callbacks`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_policy_callbacks` | `compiled` | ok (5) | verified (5) | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 | 0.5420 | 0.0049 | 110.61x | n/a | n/a | n/a | n/a |
| `concurrent_policy_callbacks` | `bytecode` | ok (5) | verified (5) | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 | 0.3820 | n/a | n/a | 0.0574 | 6.66x | 0.0768 | 4.97x |
