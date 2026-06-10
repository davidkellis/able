# Compiled typed/runtime-boundary reachability audit

Suite: `coverage`. This is a debug-only `typed-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Interpreter linked | Any -> runtime | Runtime -> integer | Runtime -> Array | Array -> runtime | Runtime -> struct | Struct -> runtime | Runtime -> union | Union -> runtime | Runtime -> interface | Interface -> runtime | Runtime -> callable | Callable -> runtime | Error -> control | Control -> error |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| await_channel_mux | ok | verified | no | 0 | 3074 | 0 | 1024 | 0 | 0 | 1536 | 0 | 4096 | 2560 | 0 | 2560 | 10756 | 2560 |
| mutex_await_journal | ok | verified | no | 0 | 6144 | 0 | 2048 | 0 | 0 | 0 | 0 | 2048 | 2048 | 0 | 2048 | 12294 | 5300 |
| mutex_work_queue | ok | verified | no | 0 | 20488 | 0 | 4096 | 0 | 0 | 4 | 0 | 4100 | 4096 | 0 | 4096 | 32782 | 11861 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
