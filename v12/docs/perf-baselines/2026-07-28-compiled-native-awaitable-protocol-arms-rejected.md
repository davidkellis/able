# Compiled native Awaitable protocol arms rejected

Date: 2026-07-28

## Decision

Do not retain the typed Awaitable-arm prototype. It removed the complete
Array/interface/protocol-dispatch path and consistently reduced allocation,
but it did not produce a broad wall-time win across the three unlike reached
applications. The production generator was restored exactly to the retained
native-Awaitable-carrier baseline.

No canonical stdlib, interpreter, bytecode VM, language, dependency,
non-primitive nominal, application, or WASM change advanced.

## General candidate

The prototype was confined to the `await` syntax/kernel boundary:

- a statically known native `Array<Awaitable<T>>` entered a typed arm
  collector without first becoming a runtime Array;
- the generated collector unwrapped the general native-interface runtime
  adapter and captured `runtime.NativeAwaitableValue` once per arm;
- arm storage changed from a slice of pointers to a slice of values;
- `is_default`, `is_ready`, `register`, and `commit` called the native
  Awaitable protocol directly;
- arbitrary/dynamic Iterable inputs and non-native Awaitable values retained
  the established runtime fallback.

The rule did not inspect an application, named container, or non-primitive
nominal type.

## Exact protocol census

Three verifier-backed runs per side produced these stable structural counts
(contention-dependent readiness and registration counts are shown as ranges):

| Application | Runtime Array conversions | Interface-to-runtime conversions | Dynamic protocol calls | Native Array collections | Direct native protocol |
| --- | ---: | ---: | ---: | ---: | --- |
| Await Channel Mux baseline | 1,024 | 2,560 | 7,680 | 0 | 0 |
| Await Channel Mux candidate | 0 | 0 | 0 | 1,024 | default 2,560; ready 3,072; register 1,024; commit 1,024 |
| Mutex Await Journal baseline | 2,048 | 2,048 | 6,214-6,348 | 0 | 0 |
| Mutex Await Journal candidate | 0 | 0 | 0 | 2,048 | default 2,048; ready 2,070-2,089; register 22-41; commit 2,048 |
| Mutex Work Queue baseline | 4,096 | 4,096 | 12,578-12,732 | 0 | 0 |
| Mutex Work Queue candidate | 0 | 0 | 0 | 4,096 | default 4,096; ready 4,274-4,285; register 178-189; commit 4,096 |

Interface-from-runtime counts remained 2,560/2,048/4,096 because the Await
expression results still cross the required semantic interface boundary.

Escape diagnostics agreed with the mechanism: the candidate's arm slice
escaped once, as expected for persisted await state, but the baseline's
per-arm `&__able_await_arm_state{...}` allocations disappeared.

## Main-phase allocation gate

Five rotating processes per side used `ABLE_EXECUTOR=goroutine`,
`GOMAXPROCS=4`, CPUs 0-3, and `ABLE_GO_PHASE_STATS_DIR`.

| Application | Baseline bytes | Candidate bytes | Change | Baseline objects | Candidate objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 6,616,540.8 | 6,462,302.4 | -2.33% | 112,707.6 | 107,056.6 | -5.01% |
| Mutex Await Journal | 2,810,977.6 | 2,593,929.6 | -7.72% | 45,929.8 | 39,685.0 | -13.60% |
| Mutex Work Queue | 5,889,142.4 | 5,488,580.8 | -6.80% | 97,023.2 | 84,935.4 | -12.46% |

Every individual paired allocation run improved.

## Balanced wall-time gate

Fifteen cohorts per application rotated all six
baseline/candidate/equivalent-Go orders. Every process used the same four-CPU
budget, each Able process used the goroutine executor, and every output passed
the public verifier outside the timed interval.

| Application | Baseline mean | Candidate mean | Go mean | Raw change | Paired change, 95% CI | Candidate/Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.079739 s | 0.079380 s | 0.002955 s | -0.45% | -0.42% [-3.04%, +2.20%] | 26.86x |
| Mutex Await Journal | 0.008170 s | 0.007551 s | 0.002138 s | -7.57% | -5.93% [-14.37%, +2.51%] | 3.53x |
| Mutex Work Queue | 0.013670 s | 0.013845 s | 0.002624 s | +1.28% | +1.63% [-6.19%, +9.45%] | 5.28x |

All paired intervals include zero and one application regressed by raw mean.
The candidate therefore fails the required broad verifier-backed wall-time
bar despite its allocation improvement.

## Verification and restoration

- Focused native-carrier, explicit-protocol, fixture-parity, goroutine Future,
  and public Mutex tests pass under the one-minute test limit.
- Candidate strict binaries for all three applications passed their public
  verifiers and omitted `pkg/interpreter`.
- After removal, regenerated `compiled.go` files were byte-for-byte identical
  to baseline, with SHA-256:
  - Await Channel Mux:
    `cd21636fa5734b408cf8398ef310573ce0e6299a1ac831a8454ecdb1ca236740`
  - Mutex Await Journal:
    `e6abfe1a284fe6daaf7d097c07aaed58fe6077bbeac1d2b686fd4fde7623c095`
  - Mutex Work Queue:
    `39501f188cbe23cac0155a03e5797c013b793fff4661e39af57a133588185a4e`
- Restored binaries again passed every verifier and omitted
  `pkg/interpreter`.

The reusable diagnostic overlay remains. It now handles both pointer-arm and
value-arm generated layouts and can count direct native protocol branches.

## Next

Refresh retained-baseline CPU and allocation profiles across the same three
applications, with special attention to Await waker and registration
construction, and select a production candidate only if one exact owner is
material in all three.

Why: the complete typed protocol route has now shown that removing boundary
dispatch alone does not reliably move wall time. What it entails: re-rank
post-carrier owners, count the shared waker/registration paths before
prototype, and apply the same verifier-backed allocation and balanced-Go
gates to a general syntax/kernel or runtime rule. Why it is important: this
redirects work toward measured retained-baseline cost rather than adding
compiler complexity to an architecturally attractive but wall-neutral route.
