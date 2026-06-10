# Awaitable cross-application bytecode profile gate — 2026-07-14

## Decision

Keep no bytecode VM, compiler, runtime, or canonical-stdlib change. The three
unlike Awaitable applications repeat the `await` protocol, but they do not
repeat a material concrete descendant beneath that protocol. Optimizing the
shared dispatcher parent would select mutex/channel behavior without evidence
from Future Await Race, violating the application-generality policy.

## Method

The warm bytecode-runtime benchmark loaded and warmed each program before its
CPU profile started. It used the canonical external stdlib, an otherwise empty
temporary run directory (so a sibling external package could not become a
second user root), goroutine execution, and:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 ABLE_EXECUTOR=goroutine
```

Mutex Await Journal and Await Channel Mux used five steady-state `main()`
calls. Future Await Race is much shorter, so it used sixty calls to obtain a
useful 650 ms sample rather than treating a 50 ms capture as evidence. All
three benchmark invocations passed.

| Application | Calls | ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: | ---: |
| Mutex Await Journal | 5 | 102,725,963 | 25,565,217 | 246,118 | 520 ms |
| Await Channel Mux | 5 | 100,587,526 | 28,758,736 | 254,095 | 500 ms |
| Future Await Race | 60 | 10,902,995 | 1,093,499 | 33,846 | 650 ms |

The temporary executable and profiles were removed after recording this
attribution; they are reproducible from the checked-in application sources.

## Attribution

Mutex Await Journal reaches `runAwaitExpression` for 100 ms (19.2% cumulative)
and its `completeAwait` call for 90 ms (17.3%). Await Channel Mux reaches both
for 60 ms (12.0%). In both cases the samples sit on the `commit` call, but the
concrete paths differ: `mutexAwaitable` in the journal and `channelAwaitable`
in the mux.

Future Await Race has only 20 ms (3.1%) under `runAwaitExpression`, at
`selectReadyAwaitArm`; it does not yield enough `completeAwait` samples to
make that parent material. Its leading work instead remains ordinary VM
execution, call dispatch, raw-integer handling, and goroutine stack/runtime
activity.

Thus `completeAwait` is a shared *dispatcher parent* for two applications, not
a demonstrated shared leaf across three. Registration, wakeup, and cancellation
are likewise not a common material descendant. No timing prototype was run,
and no source behavior changed.

## Verification

- All three warm bytecode-runtime benchmark invocations completed successfully.
- The existing verifier-backed source applications remain the semantic checks;
  this tranche changed no runtime code or application output.
- `git diff --check` passes after the documentation updates.

## Next recommendation

The CPU-pinned external scorecard refresh is complete. It confirms a broad
compiled-concurrency miss but does not establish a shared leaf. Next profile
the generated post-bootstrap runtime across Channel Rollup, Future Pipeline,
and Future Await Race, then guard a candidate on the other async applications
and the stable generality suite. See
`2026-07-14-external-scorecard-async-refresh.md`.
