# External Benchmark Comparison

- Generated: `2026-07-22T01:17:06.216209Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-regex-concurrency-closures-go-reference.json`
- Suite: `custom`
- Able benchmarks: `config_validation_extraction, log_routing_redaction, policy_record_dispatch, regex_set_audit, regex_stream_audit, regex_suffix_audit`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `config_validation_extraction` | `compiled` | ok (5) | verified (5) | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 | 0.1020 | 0.0049 | 20.82x |
| `log_routing_redaction` | `compiled` | ok (5) | verified (5) | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 | 0.1220 | 0.0059 | 20.68x |
| `policy_record_dispatch` | `compiled` | ok (5) | verified (5) | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 | 0.1940 | 0.0048 | 40.42x |
| `regex_set_audit` | `compiled` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 0.1300 | 0.0058 | 22.41x |
| `regex_stream_audit` | `compiled` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 0.1320 | 0.0048 | 27.50x |
| `regex_suffix_audit` | `compiled` | ok (5) | verified (5) | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f | 0.1260 | 0.0053 | 23.77x |
