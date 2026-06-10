# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Interpreter linked | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: |
| await_channel_mux | ok | verified | no | 1538 | 1025 | 1 | 2560 |
| mutex_await_journal | ok | verified | no | 1 | 2049 | 1 | 2052 |
| mutex_work_queue | ok | verified | no | 1 | 4097 | 1 | 4100 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
