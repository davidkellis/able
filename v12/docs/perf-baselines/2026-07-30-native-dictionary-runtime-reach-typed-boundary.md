# Compiled typed/runtime-boundary reachability audit

Suite: `coverage`. This is a debug-only `typed-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Interpreter linked | Any -> runtime | Runtime -> integer | Runtime -> Array | Array -> runtime | Runtime -> struct | Struct -> runtime | Runtime -> union | Union -> runtime | Runtime -> interface | Interface -> runtime | Runtime -> callable | Callable -> runtime | Error -> control | Control -> error |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| concurrent_graph_visitors | ok | verified | no | 0 | 0 | 1 | 0 | 4 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | 9 | 0 |
| concurrent_event_routing | ok | verified | no | 0 | 16399 | 1 | 0 | 12300 | 13320 | 4101 | 4100 | 0 | 0 | 0 | 0 | 43048 | 0 |
| validated_job_pipeline | ok | verified | no | 0 | 8207 | 1 | 0 | 4104 | 4608 | 1 | 0 | 0 | 0 | 0 | 0 | 16927 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
