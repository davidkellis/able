# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service | semantic_truthy_check | semantic_truthy_error_fallback | explicit_cast_check |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| config_validation_extraction | ok | verified | 2 | 2 | 1 | 0 | 0 | 0 | 0 |
| log_routing_redaction | ok | verified | 2 | 2 | 1 | 0 | 0 | 0 | 0 |
| policy_record_dispatch | ok | verified | 2 | 3074 | 1 | 0 | 2048 | 2048 | 0 |
| regex_set_audit | ok | verified | 2 | 2 | 1 | 0 | 0 | 0 | 0 |
| regex_stream_audit | ok | verified | 2 | 2 | 1 | 0 | 0 | 0 | 0 |
| regex_suffix_audit | ok | verified | 2 | 2 | 1 | 0 | 0 | 0 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
