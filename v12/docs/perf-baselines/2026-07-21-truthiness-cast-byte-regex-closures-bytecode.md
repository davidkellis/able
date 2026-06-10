# External Benchmark Comparison

- Generated: `2026-07-22T00:40:22.546768Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-byte-regex-closures-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `config_validation_extraction, log_routing_redaction, policy_record_dispatch, regex_set_audit, regex_stream_audit, regex_suffix_audit`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `config_validation_extraction` | `bytecode` | ok (5) | verified (5) | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 | 1.2820 | 0.0234 | 54.79x | 0.0479 | 26.76x |
| `log_routing_redaction` | `bytecode` | ok (5) | verified (5) | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 | 2.8920 | 0.0176 | 164.32x | 0.0449 | 64.41x |
| `policy_record_dispatch` | `bytecode` | ok (5) | verified (5) | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 | 7.2800 | 0.0250 | 291.20x | 0.0459 | 158.61x |
| `regex_set_audit` | `bytecode` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 4.1640 | 0.0199 | 209.25x | 0.0460 | 90.52x |
| `regex_stream_audit` | `bytecode` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 3.7500 | 0.0188 | 199.47x | 0.0452 | 82.96x |
| `regex_suffix_audit` | `bytecode` | ok (5) | verified (5) | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f | 3.2560 | 0.0181 | 179.89x | 0.0389 | 83.70x |
