# Compiled task-owned Await arm scratch retained

Date: 2026-07-28

## Decision

Retain task-owned Await arm-scratch reuse under the existing experimental
execution-context option.

Each task's reusable primary `__able_await_state` now retains its maximum arm
slice capacity and its `__able_await_arm_state` records. A later sequential
Await refills those records instead of allocating a new slice and one pointer
object per arm. Cleanup cancels registrations, clears waiting state, and then
sets every retained awaitable and registration reference to `nil` before
making the scratch available again.

The existing waker slot is also the activity token. A pending sentinel marks
the state active before user-defined `is_default` protocol code can run; the
real waker replaces it after collection; cleanup clears it under the same
mutex used by late wakes. An Await re-entered from arm inspection or winner
`commit` therefore receives a transient state and independent scratch.

Static Arrays and arbitrary dynamic Iterables use the same collector and
scratch representation. This is a general generated concurrency ownership
rule, not the rejected typed Awaitable protocol-arm fast path. It names no
application, container, or non-primitive nominal type.

## Exact owner result

Three exact main-allocation profiles per side were public-verifier backed:

| Application | Baseline slice + arm objects | Retained objects | Change |
| --- | ---: | ---: | ---: |
| Await Channel Mux | 3,584 | 3,584 | neutral control |
| Future Await Race | 384 | 192 | -50.00% |
| Mutex Await Journal | 4,096 | 8 | -99.80% |
| Mutex Work Queue | 8,192 | 8 | -99.90% |

Channel Mux performs one Await per task and therefore cannot reuse its initial
scratch. Future Race performs two single-arm Awaits per task. Journal and Work
Queue reuse one slice and one arm record in each of four long-lived tasks.

## Allocation A/B

Five rotating baseline/retained processes per side measured the generated
main phase. Every output passed its public verifier.

| Application | Baseline bytes | Retained bytes | Paired change, 95% CI | Baseline objects | Retained objects | Paired change, 95% CI |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 5,221,379.2 | 5,227,689.6 | +0.121% [-0.039%, +0.280%] | 96,787.8 | 96,841.2 | +0.055% [-0.025%, +0.135%] |
| Future Await Race | 695,716.8 | 685,643.2 | -1.448% [-2.620%, -0.276%] | 12,571.6 | 12,317.0 | -2.013% [-4.519%, +0.493%] |
| Mutex Await Journal | 1,119,059.2 | 1,005,712.0 | -10.128% [-10.685%, -9.571%] | 26,980.8 | 22,916.8 | -15.062% [-15.521%, -14.603%] |
| Mutex Work Queue | 2,447,334.4 | 2,222,009.6 | -9.206% [-9.865%, -8.547%] | 58,376.2 | 50,281.2 | -13.866% [-14.314%, -13.419%] |

The one-Await control is neutral. Future bytes improve significantly and its
object result trends down within scheduler noise. Both allocation measures
improve significantly in the two high-reuse applications.

## Balanced timing and equivalent Go

Fifteen cohorts per application rotated all six
baseline/retained/equivalent-Go orders on CPUs 0-3 with `GOMAXPROCS=4` and the
goroutine executor:

| Application | Baseline | Retained | Raw change | Paired change, 95% CI | Go | Retained/Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.078373s | 0.078575s | +0.26% | +1.40% [-7.26%, +10.06%] | 0.003147s | 24.97x | 4.01% |
| Future Await Race | 0.012566s | 0.011790s | -6.17% | -5.13% [-12.76%, +2.49%] | 0.002460s | 4.79x | 20.86% |
| Mutex Await Journal | 0.006228s | 0.005841s | -6.21% | -5.51% [-12.14%, +1.13%] | 0.002420s | 2.41x | 41.42% |
| Mutex Work Queue | 0.010693s | 0.010963s | +2.52% | +2.66% [-1.41%, +6.74%] | 0.002669s | 4.11x | 24.35% |

Every interval includes zero. The allocation win therefore carries no
statistically distinguishable wall-time regression, but compiled performance
still misses the 95%-of-Go product target.

## Rejected intermediate activity field

The first form added a separate boolean activity field. That moved each
Channel Mux state into the next Go allocation class and significantly
increased control bytes by 0.28%. It was removed before timing.

The retained pending-waker sentinel uses an existing field, returns Channel
Mux to neutral, and preserves the same nested and stale-waker safety.

## Verification and scope

- New execution guards cover Await re-entry from both `is_default` arm
  collection and winner `commit`.
- The stale user-waker guard proves that an old cancellation wake cannot
  notify later reused scratch.
- Those three guards pass with race-enabled child builds in 25.069 seconds.
- All four generated race binaries pass with the goroutine executor.
- Three ordinary serial and three ordinary goroutine executions per
  application pass their public verifiers.
- Full experimental fixture parity passes in 44.027 seconds.
- The complete focused Awaitable/service/state suite passes in 7.080 seconds.
- `go test ./pkg/compiler/bridge ./pkg/runtime ./cmd/ablec` and focused
  `go vet` pass.
- All default generated Go files and `go.mod` files are byte-identical to the
  pre-tranche compiler.
- Final strict graphs omit `pkg/interpreter`; measured experimental
  `compiled.go` files match the final generator.
- The ledger update changes only `compiler-production` to
  `a2351cf9...e4eb391a`; all seven ledger tests pass with 21 current closures
  and zero invalidations.
- Every touched source remains below 1,000 lines.

No canonical stdlib, runtime package, interpreter, bytecode VM, language,
dependency, application source, non-primitive nominal rule, or WASM change was
needed.

Machine-readable aggregate:

- `2026-07-28-compiled-task-owned-await-arm-scratch-retained.json`

## Next

Audit the remaining per-Await native waker allocation and split its semantic
lifetime only if native registrations can carry an immutable generation token.

Why: post-scratch exact profiles allocate one
`__able_native_await_waker` per Await—1,024/192/2,048/4,096 objects across the
four applications—making it the next open generated owner shared by all four.

What it entails: distinguish wakers that remain entirely inside native kernel
registrations from wakers materialized or retained by dynamic user Awaitables;
prove that every exposed waker preserves distinct identity; prototype a
task-owned native wake endpoint plus per-Await generation only if late
references cannot target a subsequent Await; then repeat exact allocation,
race, dual-executor, verifier, timing, and Go gates.

Why it is important: blindly reusing the waker object would recreate a
cross-Await lost-wake/cross-wake bug. A lifetime-safe split is the next chance
to remove repeated generated service allocation while keeping compiled
execution native and interpreter-free.
