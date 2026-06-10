# External Benchmark Comparison

- Generated: `2026-07-24T04:54:19.958574Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-packet-codecs-interpreter-b.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-packet-codecs-go-b.json`
- Suite: `concurrent-packet-codecs`
- Able benchmarks: `concurrent_packet_codecs`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_packet_codecs` | `compiled` | ok (5) | verified (5) | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 | 0.5100 | 0.0037 | 137.84x | n/a | n/a | n/a | n/a |
| `concurrent_packet_codecs` | `bytecode` | ok (5) | verified (5) | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 | 0.7160 | n/a | n/a | 0.0812 | 8.82x | 0.0824 | 8.69x |
