# Compiled task-owned Await signal retained

Date: 2026-07-28

## Decision

Retain one reusable wake signal per compiled asynchronous task under the
existing experimental execution-context option.

Generated Await calls block inside one Go activation: serial execution
suspends and resumes that activation, while goroutine execution blocks it
directly. The activation therefore already owns the live Await state on its Go
stack and does not re-enter through the payload's AST-expression cache. The
experimental payload now uses that existing one-word cache slot for a buffered
`chan struct{}` shared by its sequential Await states. Default generation
retains the established expression cache and state-owned channel byte for byte.

Wake publication is serialized with wait cleanup by the Await-state mutex. A
waker signals only while its state is waiting, and wake consumption or cleanup
drains a losing notification before another state can reuse the task channel.
This preserves synchronous wake-before-park, cancellation races, registration
rearm, late user wakers, and both executor modes.

This is a general generated concurrency rule. It names no benchmark,
container, non-primitive nominal type, or application structure, and it
changes no interpreter/runtime ABI.

## Correctness repair found before measurement

The retained pre-candidate compiler failed Future Await Race with
`Future failed: register expects AwaitWaker`. The earlier lazy Await service
carrier passed `*__able_native_await_waker` to the generated kernel Future
registration helper, but that helper still accepted only a materialized
`AwaitWaker` struct.

The general correction lets a generated Future register the native waker
directly and returns the native Await registration carrier with the same
execution context. A focused compiled guard now covers this boundary. The
allocation and timing baseline was rebuilt after the repair, so no performance
claim includes the correctness failure.

## Exact lifecycle evidence

Five public-verifier-backed goroutine runs per application found no overlapping
active Await states within a task (`max_active=1`, zero overlap events, and
zero active states at exit). Three serial runs independently verified the same
ownership rule where an application actually parked; Journal and Work Queue
complete their mutex arms immediately under serial scheduling and therefore
do not enter the wait cycle.

| Application | States/run | Payloads/run | Baseline channels | Retained channels |
| --- | ---: | ---: | ---: | ---: |
| Await Channel Mux | 1,024 | 1,024 | 1,024 | 1,024 |
| Future Await Race | 192 | 96 | 192 | 96 |
| Mutex Await Journal | 2,048 | 4 | 2,048 | 4 |
| Mutex Work Queue | 4,096 | 4 | 4,096 | 4 |

Channel Mux is the neutral one-state-per-task control. Future Race proves
sequential reuse across two Future awaits. Journal and Work Queue prove reuse
across thousands of repeated mutex Await expressions while retaining
scheduler-dependent registration rearm.

## Allocation A/B

Five rotating baseline/retained processes per side measured the generated
main phase with lightweight `runtime.MemStats`; every stdout passed its public
verifier.

| Application | Baseline bytes | Retained bytes | Change | Baseline objects | Retained objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 5,417,275.2 | 5,220,596.8 | -3.63% | 98,827.0 | 96,777.0 | -2.07% |
| Future Await Race | 737,936.0 | 711,857.6 | -3.53% | 13,442.8 | 13,193.6 | -1.85% |
| Mutex Await Journal | 1,515,747.2 | 1,283,280.0 | -15.34% | 31,099.8 | 29,030.4 | -6.65% |
| Mutex Work Queue | 3,238,902.4 | 2,772,254.4 | -14.41% | 66,595.2 | 62,404.8 | -6.29% |

Both allocation measures improve in all four unlike applications.

## Balanced timing and equivalent Go

Fifteen rotating baseline/retained/equivalent-Go cohorts per application ran
on CPUs 0-3 with the public verifier:

| Application | Baseline | Retained | Raw change | Paired 95% interval | Go | Retained/Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.084901s | 0.077969s | -8.17% | -12.39% to -0.96% | 0.005116s | 15.24x | 6.56% |
| Future Await Race | 0.014718s | 0.015056s | +2.30% | -6.28% to +14.07% | 0.004386s | 3.43x | 29.13% |
| Mutex Await Journal | 0.008441s | 0.007860s | -6.89% | -10.22% to -3.37% | 0.004117s | 1.91x | 52.38% |
| Mutex Work Queue | 0.014272s | 0.014055s | -1.52% | -7.01% to +3.85% | 0.004872s | 2.89x | 34.66% |

Channel Mux and Journal improve significantly. Future Race and Work Queue are
neutral within workstation noise. The 95%-of-Go product goal remains unmet.

## Rejected intermediate design

Do not retry a buffered tagged `chan *__able_await_state`. It reduced channel
counts, but Go allocates a separate non-zero channel buffer object. Channel
Mux regressed 1.03% in objects and Future Race did not improve objects, so the
prototype failed the broad allocation gate and was removed. The retained
zero-sized channel instead rejects late notifications while holding the
originating state's mutex.

## Verification and scope

- The corrected and retained goroutine/serial binaries pass all four public
  verifiers and omit `pkg/interpreter`.
- All four race-enabled retained binaries pass without a race report.
- A dynamic user-Awaitable guard retains a canceled waker and proves that its
  late wake cannot re-register the next Await.
- Full experimental execution-context fixture parity passes in 44.263
  seconds.
- Focused nested-spawn, dynamic-boundary, bound-method, cross-package
  interface, captured-callback, Future Await, and mutex-contention guards pass.
- `go test ./pkg/compiler/bridge ./pkg/runtime ./cmd/ablec` passes.
- All four default generated modules are byte-identical to the pre-tranche
  compiler. The measured experimental `compiled.go` files are byte-identical
  to the final retained generator output.
- The checked performance ledger has 21 current closures and zero
  invalidations. The `compiler-production` scope is
  `b7e73e39...f88c959`.
- Every touched source remains below 1,000 lines; the largest is
  `generator_render_runtime_future.go` at 994.

No canonical stdlib, runtime package, interpreter, bytecode VM, language,
dependency, non-primitive nominal rule, or WASM change was needed.

Machine-readable aggregate:

- `2026-07-28-compiled-task-owned-await-signal-retained.json`

## Next

Refresh post-retention main CPU and exact allocation profiles for Await Channel
Mux, Future Await Race, Mutex Await Journal, and Mutex Work Queue, then select
only the largest generated Await state/arm/registration owner that is material
in at least three unlike applications.

Why: this tranche removed the task cache and repeated wait-channel allocation,
so all earlier profiles overstate those owners. The remaining Go gaps range
from 1.91x to 15.24x and require fresh attribution.

What it entails: capture repeated public-verifier-backed profiles from the
retained source, distinguish semantic Awaitable/registration lifetimes from
compiler-owned state and arm scratch, and prototype reuse or stack ownership
only if late-waker lifetime and serial/goroutine behavior can be proven.

Why it is important: the immediate goal remains native Go performance without
compiled/interpreter crossings or boxing. Fresh ownership evidence prevents
optimizing an object this tranche already removed and prevents unsafe state
reuse merely because the same Await expression repeats.
