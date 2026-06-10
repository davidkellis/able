# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service | semantic_truthy_check | semantic_truthy_error_fallback | explicit_cast_check |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base64 | ok | verified | 3 | 3 | 3 | 0 | 0 | 0 | 0 |
| fasta_generation | ok | verified | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| pidigits | ok | verified | 1001 | 1001 | 1000 | 0 | 0 | 0 | 0 |
| reverse_complement | ok | verified | 1 | 1 | 0 | 0 | 0 | 0 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
