# Compiled task-owned Await state retained

Date: 2026-07-28

## Decision

Retain one reusable primary generated Await state per compiled asynchronous
task under the existing experimental execution-context option.

The preceding signal tranche proved that generated Await calls within one task
are sequential: a suspended Await keeps its Go activation live, and a nested
Await can re-enter only from a winner's `commit`. The task payload therefore
now owns its primary `__able_await_state`, including the already-retained
buffered zero-sized wake channel. Ordinary sequential Await expressions reuse
that state. If `commit` re-enters Await while the primary state is active, the
nested call receives a separate transient state.

Each Await still receives a distinct waker. The state records that waker under
its mutex, and wake publication succeeds only when both the current waker
identity and waiting state match. A late waker from a completed Await therefore
cannot wake a later reuse of the same state. Arm/default setup and release are
task-local; wait publication, wake state, channel draining, and waker identity
remain mutex-protected.

This is a general generated concurrency ownership rule. It names no
application, container, or non-primitive nominal type and changes no
interpreter or runtime-package ABI. Default generation remains byte-identical.

## Fresh owner selection

Five fresh CPU profiles and three exact main-allocation profiles were captured
for Await Channel Mux, Future Await Race, Mutex Await Journal, and Mutex Work
Queue after the task-owned signal retention. Every process passed its public
verifier and every strict graph omitted `pkg/interpreter`.

The largest shared collector remained `__able_collect_await_arms_ctx`, with
3,584/384/4,096/8,192 objects. That route is closed: the complete native
Awaitable protocol-arm prototype already removed the Array conversion,
interface conversion, dynamic protocol dispatch, slice, and per-arm pointer
objects without a broad wall-time win.

The largest open shared owner was construction of `__able_await_state`:

| Application | Baseline states | Retained states | Change |
| --- | ---: | ---: | ---: |
| Await Channel Mux | 1,024 | 1,024 | neutral control |
| Future Await Race | 192 | 96 | -50.00% |
| Mutex Await Journal | 2,048 | 4 | -99.80% |
| Mutex Work Queue | 4,096 | 4 | -99.90% |

Channel Mux has one Await per task and is the expected neutral control. Future
Race has two sequential Awaits per task. Journal and Work Queue repeatedly
Await mutex acquisition from four long-lived tasks.

## Allocation A/B

Five rotating baseline/retained processes per side measured the generated
main phase with lightweight `runtime.MemStats`. Every stdout passed its public
verifier.

| Application | Baseline bytes | Retained bytes | Paired change, 95% CI | Baseline objects | Retained objects | Paired change, 95% CI |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 5,218,790.4 | 5,219,537.6 | +0.014% [-0.021%, +0.050%] | 96,763.8 | 96,769.4 | +0.006% [-0.006%, +0.018%] |
| Future Await Race | 709,636.8 | 702,361.6 | -1.024% [-1.509%, -0.540%] | 12,878.2 | 13,025.2 | +1.153% [-0.800%, +3.106%] |
| Mutex Await Journal | 1,285,835.2 | 1,122,142.4 | -12.730% [-12.952%, -12.508%] | 29,070.2 | 27,037.8 | -6.991% [-7.235%, -6.747%] |
| Mutex Work Queue | 2,783,553.6 | 2,456,083.2 | -11.763% [-12.393%, -11.133%] | 62,643.4 | 58,518.0 | -6.585% [-7.083%, -6.086%] |

Channel Mux is neutral as required. Future's byte reduction is significant;
its small object sample is noisy around scheduler work despite the exact
96-object state reduction. Both allocation measures improve significantly in
the two high-reuse applications.

## Balanced timing and equivalent Go

Fifteen rotating baseline/retained/equivalent-Go cohorts per application ran
on CPUs 0-3 with `GOMAXPROCS=4`, the goroutine executor, and public output
verification:

| Application | Baseline | Retained | Raw change | Paired change, 95% CI | Go | Retained/Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.072897s | 0.071053s | -2.53% | -2.02% [-4.60%, +0.56%] | 0.003063s | 23.20x | 4.31% |
| Future Await Race | 0.010419s | 0.010348s | -0.68% | -0.23% [-4.63%, +4.16%] | 0.002245s | 4.61x | 21.69% |
| Mutex Await Journal | 0.005473s | 0.005528s | +0.99% | +1.42% [-3.22%, +6.05%] | 0.002103s | 2.63x | 38.05% |
| Mutex Work Queue | 0.010682s | 0.010548s | -1.26% | -1.15% [-6.53%, +4.23%] | 0.002492s | 4.23x | 23.63% |

All four timing intervals include zero. The allocation win therefore carries
no measured wall-time regression, but the 95%-of-Go product target remains
unmet.

## Rejected intermediate locking

The first reusable-state form serialized acquisition and release with the
state mutex. Await Channel Mux then regressed significantly in balanced timing.
Those operations are task-local, so the final design removes those unnecessary
locks while retaining synchronization for waker identity and wake state. The
final fifteen-cohort gate is neutral in all four applications.

## Verification and scope

- A focused nested-commit guard proves that an active primary state falls back
  to a transient nested state and returns the correct value.
- The existing dynamic user-Awaitable guard proves that a retained stale waker
  cannot wake the next Await.
- All four race-enabled goroutine-executor applications pass without a race
  report. Three ordinary serial and three ordinary goroutine runs per
  application also pass their public verifiers.
- A diagnostic serial race build parked for both the pre-tranche baseline and
  candidate; it is an existing race-instrumentation scheduler limitation, not
  a candidate-specific difference.
- Full experimental execution-context fixture parity passes in 44.079 seconds.
- Focused native Awaitable, explicit protocol, user materialization, Future
  registration, stale-waker, nested-reentrancy, and generated-source guards
  pass in 9.956 seconds.
- `go test ./pkg/compiler/bridge ./pkg/runtime ./cmd/ablec` passes.
- All four default generated modules match the pre-tranche compiler byte for
  byte. The measured experimental `compiled.go` files match the final
  generator output byte for byte.
- Every final strict application graph omits `pkg/interpreter`.
- The performance ledger changes only the `compiler-production` tree hash to
  `66f3d91d...6fc5fee`; all seven tests pass with 21 current closures and zero
  invalidations.
- Every touched source remains below 1,000 lines.

No canonical stdlib, runtime package, interpreter, bytecode VM, language,
dependency, application source, non-primitive nominal rule, or WASM change was
needed.

Machine-readable aggregate:

- `2026-07-28-compiled-task-owned-await-state-retained.json`

## Next

Measure and prototype task-owned Await arm-scratch reuse across the same four
applications.

Why: post-state profiles now show `__able_collect_await_arms_ctx` allocating
3,584/384/4,096/8,192 slice-and-arm objects. The typed protocol-arm route is
closed, but sequential task ownership offers a distinct general way to reuse
the existing runtime arm representation.

What it entails: retain arm-slice capacity and arm records only after
registration cleanup, clear every awaitable/registration reference before
reuse, fall back for nested active Await, preserve arbitrary Iterable and
dynamic Awaitable behavior, and repeat exact allocation, race, serial,
goroutine, verifier, balanced timing, and equivalent-Go gates.

Why it is important: it attacks the largest remaining exact generated Await
allocation owner without reopening a rejected representation fast path,
crossing into the interpreter, boxing primitive carriers, or weakening
late-waker semantics.
