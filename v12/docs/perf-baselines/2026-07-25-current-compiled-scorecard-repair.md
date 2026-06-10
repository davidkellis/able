# External Benchmark Comparison

- Generated: `2026-07-26T01:06:35.775320Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-25-current-compiled-go-reference.json`
- Suite: `custom`
- Able benchmarks: `binary_event_log, mutex_ledger, regex_suffix_audit, regex_set_audit, regex_stream_audit, log_routing_redaction, config_validation_extraction, dependency_wave_validation, concurrent_state_machines, policy_record_dispatch`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `5,10,15,11` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `binary_event_log` | `compiled` | ok (5) | verified (5) | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 | 0.1940 | 0.0085 | 22.82x |
| `mutex_ledger` | `compiled` | error (2) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.0800 | 0.0048 | n/a |
| `regex_suffix_audit` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0046 | n/a |
| `regex_set_audit` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0051 | n/a |
| `regex_stream_audit` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0051 | n/a |
| `log_routing_redaction` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0045 | n/a |
| `config_validation_extraction` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0041 | n/a |
| `dependency_wave_validation` | `compiled` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 0.0580 | 0.0045 | 12.89x |
| `concurrent_state_machines` | `compiled` | error (1) | verified (5) | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 | 0.0275 | 0.0043 | n/a |
| `policy_record_dispatch` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0067 | n/a |
