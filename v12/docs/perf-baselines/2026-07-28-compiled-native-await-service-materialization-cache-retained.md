# Compiled native Await service materialization cache retained

Date: 2026-07-28

## Decision

Retain compact atomic materialization caches for generated native Await wakers
and registrations under the existing experimental execution-context option.

Each `__able_native_await_waker` and `__able_native_await_registration` still
has distinct semantic identity. Its lazily created
`runtime.StructInstanceValue` is now published through an
`atomic.Pointer[runtime.StructInstanceValue]` instead of storing a pointer plus
an otherwise dormant `sync.Once`. Concurrent consumers may construct a losing
candidate, but every consumer returns the single compare-and-swap winner, so
the exposed runtime struct identity remains stable.

The rule is a general generated service-carrier layout change. It does not
name an application, container, or non-primitive nominal type, and it does not
change the runtime ABI or create an interpreter boundary.

## Endpoint-reuse audit

Reusing a waker object or mutable task endpoint across sequential Awaits is
unsafe and was not implemented:

- channel and mutex registrations check cancellation and then release their
  lock before invoking the waker;
- Future completion loads cancellation separately from invoking the waker;
- timer completion releases its lock before invoking the waker; and
- dynamic user Awaitables may retain an exposed waker indefinitely.

Cancellation therefore cannot prove that every old invocation has finished.
Mutating a reused endpoint or generation field would permit an old invocation
to target later Await state. An immutable per-Await generation token would
still require one distinct per-Await object and would not remove the remaining
waker allocation. The unique waker lifetime is now a semantic requirement,
not an open object-reuse route.

## Exact owner result

Three exact main-allocation profiles per side were public-verifier backed.
The waker object count is intentionally unchanged, while its generated
allocation class falls from 64 to 48 bytes:

| Application | Wakers | Baseline waker bytes | Retained waker bytes | Change |
| --- | ---: | ---: | ---: | ---: |
| Await Channel Mux | 1,024 | 64 KiB | 48 KiB | -25.00% |
| Future Await Race | 192 | 12 KiB | 9 KiB | -25.00% |
| Mutex Await Journal | 2,048 | 128 KiB | 96 KiB | -25.00% |
| Mutex Work Queue | 4,096 | 256 KiB | 192 KiB | -25.00% |

The same compact cache applies to native registrations, whose count varies
with cancellation and contention.

## Main allocation A/B

Five rotating baseline/retained processes per side measured the generated
main phase. Every output passed its public verifier.

| Application | Baseline bytes | Retained bytes | Paired change, 95% CI | Baseline objects | Retained objects | Paired change, 95% CI |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 5,221,556.8 | 5,178,286.4 | -0.829% [-0.933%, -0.725%] | 96,790.4 | 96,769.8 | -0.021% [-0.061%, +0.019%] |
| Future Await Race | 692,560.0 | 688,792.0 | -0.542% [-1.679%, +0.595%] | 12,544.6 | 12,622.6 | +0.631% [-1.607%, +2.868%] |
| Mutex Await Journal | 1,007,688.0 | 971,708.8 | -3.570% [-3.898%, -3.242%] | 22,931.2 | 22,922.6 | -0.037% [-0.235%, +0.160%] |
| Mutex Work Queue | 2,218,603.2 | 2,152,238.4 | -2.989% [-3.656%, -2.322%] | 50,223.0 | 50,320.0 | +0.196% [-0.551%, +0.942%] |

Byte allocation improves significantly in three unlike applications and is
neutral in Future Await Race. Object counts remain neutral, as expected from
a size-class reduction rather than object reuse.

## Balanced timing and equivalent Go

The first 15 rotating baseline/retained/equivalent-Go cohorts produced a
borderline +1.68% Channel Mux interval excluding zero. An independent
15-cohort confirmation reversed application order and did not reproduce the
result. The combined 30 cohorts rotated all six orders on CPUs 0-3 with
`GOMAXPROCS=4` and the goroutine executor:

| Application | Baseline | Retained | Raw change | Paired change, 95% CI | Go | Retained/Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.074470s | 0.072221s | -3.02% | -0.93% [-6.81%, +4.94%] | 0.002972s | 24.30x | 4.12% |
| Future Await Race | 0.011594s | 0.011413s | -1.56% | +0.30% [-6.50%, +7.11%] | 0.002191s | 5.21x | 19.20% |
| Mutex Await Journal | 0.005418s | 0.005497s | +1.46% | +1.73% [-2.08%, +5.54%] | 0.002139s | 2.57x | 38.92% |
| Mutex Work Queue | 0.010544s | 0.010364s | -1.71% | -0.98% [-5.03%, +3.08%] | 0.002533s | 4.09x | 24.45% |

Every combined interval includes zero. The allocation reduction therefore
has no statistically distinguishable wall-time regression, but compiled
performance remains far below the 95%-of-Go product target.

## Verification and scope

- All four applications pass three serial and three goroutine public-verifier
  runs plus a race-enabled goroutine binary.
- Race-enabled nested collection, nested commitment, stale-waker, and dynamic
  materialization guards pass in 30.762 seconds.
- Full experimental execution-context fixture parity passes in 25.381
  seconds.
- The complete focused native Awaitable/service/state suite passes in 5.629
  seconds.
- `go test ./pkg/compiler/bridge ./pkg/runtime ./cmd/ablec` and focused
  `go vet` pass.
- All default generated Go files and `go.mod` files remain byte-identical to
  the pre-tranche compiler.
- All four strict experimental dependency graphs omit `pkg/interpreter`.
- The performance ledger changes only `compiler-production`, from
  `a2351cf9...e4eb391a` to `8d69533a...e1f7433`; all seven tests pass with 21
  current closures and zero invalidations.
- Every touched source remains below 1,000 lines.

No canonical stdlib, runtime package, interpreter, bytecode VM, language,
dependency, application source, non-primitive nominal rule, or WASM change was
needed.

Machine-readable aggregate:

- `2026-07-28-compiled-native-await-service-materialization-cache-retained.json`

## Next

Refresh post-retention CPU and exact allocation ownership across the four
Await applications, excluding the now-required unique waker object and the
already-closed typed Awaitable Array/protocol conversion routes.

Why: the waker allocation cannot be safely reused, and the current exact
profiles otherwise emphasize conversion families already rejected on broad
timing evidence.

What it entails: collect fresh repeated CPU and allocation profiles from the
retained generator, rank only open generated owners shared by at least three
unlike programs, and prototype nothing if no owner clears that filter.

Why it is important: compiled performance remains 2.57x-24.30x slower than
equivalent Go in these applications. The next work must remove a real,
general compiled/runtime boundary rather than revisit a semantically required
identity or a closed benchmark family.
