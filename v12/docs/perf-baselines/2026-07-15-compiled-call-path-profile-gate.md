# Compiled call-path profile gate — 2026-07-15

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib, or benchmark-source performance change. The call-path census identified `__able_call_value_fast` as a shared parent, but the four fresh normal-binary profiles divide below it. Option/Result Configuration is generic-union dispatch/allocation work. The three unlike concurrent applications repeat the previously rejected goroutine-identity boundary `bridge.currentGID` -> `runtime.Stack`.

Do not add an Option/Result, generic-union, Future, channel, Mutex, or fast-method helper special case. Do not retry the fixed execution-context ABI: its prior broad guards regressed independent non-concurrent applications.

## Method and results

Every binary was built once from the current tree with the external canonical stdlib and **without** call-path telemetry. The generated CPU-only phase hook captured bootstrap and registered-`main` CPU profiles, merged only within the same application. Every launch passed its canonical Ruby verifier and produced the shown stable stdout hash.

The first three sets used their own immediate quiet-host preflight and one-CPU affinity, with `GOMEMLIMIT=1GiB`, `GOGC=50`, and normal (unset) `GOMAXPROCS` for concurrency semantics. Later host variability was handled at maintainer direction with 120 independent unpinned Future Await Race launches. Ambient load does not enter a Go CPU profile; its timing spread is recorded, but is not a Go-performance comparison.

| Application | Launches | Main samples | Key descendant | Stable stdout SHA-256 |
| --- | ---: | ---: | --- | --- |
| Option/Result Configuration | 50 | 5.34 s | generic-union dispatch / allocation | `28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112` |
| await-channel-mux | 8 | 2.07 s | `bridge.currentGID` 90.82% cumulative | `0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693` |
| mutex-await-journal | 30 | 9.36 s | `bridge.currentGID` 95.73% cumulative | `e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e` |
| Future Await Race | 120 | 4.18 s | `bridge.currentGID` 82.30% cumulative | `33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4` |

Future Await Race's wall-clock samples had a 59.8 ms mean, 60.0 ms median, 3.2 ms population standard deviation, and 50–70 ms range.

## Attribution

Option/Result Configuration has `__able_static_generic_union_method_call` at 57.30% cumulative, `__able_call_value_fast` at 41.95%, and `runtime.mallocgc` at 51.87%. That path is material but does not appear in the three unlike concurrent controls.

await-channel-mux, mutex-await-journal, and Future Await Race reach `__able_call_value_fast` at 71.01%, 82.37%, and 44.26% cumulative. In each it is a caller of await/method dispatch whose repeated concrete descendant is `bridge.currentGID` / `runtime.Stack` (89.86%, 95.41%, and 80.14% `runtime.Stack` cumulative). Mutex also reaches `SwapEnvIfNeeded`, another member of the already rejected fixed-context family. The helper parent is therefore not an implementation candidate.

Retained profiles: `v12/interpreters/go/.profiles/20260715_compiled_call_path_*_{main,bootstrap}_merged.cpu.pprof`.

## Next recommendation

Wait for a material cross-cutting semantic/compiler change or a genuinely needed portable application to expose a different repeated concrete leaf. Record repeated verifier-backed workstation mean, median, and spread, and require the same descendant in at least three unlike applications before admitting an implementation experiment.
