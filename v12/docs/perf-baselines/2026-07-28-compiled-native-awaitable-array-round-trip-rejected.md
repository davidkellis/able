# Compiled native Awaitable Array round trip rejected

Date: 2026-07-28

## Decision

Reject and remove the typed Awaitable Array input candidate.

The prototype was a general `await` syntax/kernel rule: when the compiler
already held a static native `Array<Awaitable<T>>`, generated await helpers
collected arms from that native Array rather than first constructing a
`runtime.ArrayValue`. Arbitrary and dynamic `Iterable` inputs retained the
existing runtime path. The rule named no application, benchmark, stdlib
container, user nominal type, or non-primitive nominal type.

The candidate was verifier-correct and reduced allocations in all three
applications, but it did not improve wall time broadly. Per project policy,
the production compiler/runtime prototype and its source guard were removed.
Only the expanded diagnostic overlay and this evidence are retained.

## Exact reach

Three counter runs per side and application were identical:

| Application | Baseline runtime Array conversions / elements | Candidate native Array collections / elements | Interface conversions each direction | Protocol calls |
| --- | ---: | ---: | ---: | ---: |
| Await Channel Mux | 1,024 / 2,560 | 1,024 / 2,560 | 2,560 | 7,680 |
| Mutex Await Journal | 2,048 / 2,048 | 2,048 / 2,048 | 2,048 | 6,144 |
| Mutex Work Queue | 4,096 / 4,096 | 4,096 / 4,096 | 4,096 | 12,288 |

Candidate runtime Array conversion and dynamic Array collection counts were
zero. Native producer counts, lazy-materialization counts, per-element
interface conversions, protocol calls, and verified output remained
unchanged. Thus the prototype removed exactly the intended Array wrapper and
parallel `[]runtime.Value`, but did not remove the subsequent per-element
Awaitable interface-to-runtime crossing.

## Allocation A/B

Five rotating, goroutine-executor, four-CPU main-phase processes per side:

| Application | Baseline bytes | Candidate bytes | Change | Baseline objects | Candidate objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 6,614,006.4 | 6,473,740.8 | -2.12% | 112,681.0 | 110,628.0 | -1.82% |
| Mutex Await Journal | 2,805,137.6 | 2,587,038.4 | -7.77% | 45,840.4 | 41,886.8 | -8.62% |
| Mutex Work Queue | 5,821,430.4 | 5,357,598.4 | -7.97% | 95,975.0 | 87,705.2 | -8.62% |

Allocation improved consistently, confirming that the removed wrapper had a
real heap cost.

## Balanced timing and Go comparison

Fifteen rotating baseline/candidate/equivalent-Go cohorts per application ran
with `ABLE_EXECUTOR=goroutine`, `GOMAXPROCS=4`, and CPU affinity `0-3`:

| Application | Baseline | Candidate | Raw change | Paired 95% interval | Go | Candidate/Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.089115s | 0.088775s | -0.38% | -3.75% to +3.77% | 0.003253s | 27.29x |
| Mutex Await Journal | 0.008904s | 0.009319s | +4.67% | -3.60% to +17.28% | 0.002439s | 3.82x |
| Mutex Work Queue | 0.015510s | 0.015723s | +1.37% | -5.89% to +10.87% | 0.002866s | 5.49x |

Every interval crosses zero, and two raw means regress. The candidate
therefore fails the broad wall-time retention bar despite its allocation
improvement.

All 135 selected timing processes, 30 allocation processes, 18 counter
processes, and six strict census executions passed their public verifier.
Baseline and candidate dependency graphs omit `pkg/interpreter`.

## Correctness and final state

- Default and experimental typed helpers built and verified before removal.
- The ordinary runtime `Iterable` fallback remained emitted.
- Focused Awaitable, Future, Channel, Mutex, execution-context, and contention
  guards passed with the prototype.
- After removal, the focused retained Awaitable/concurrency suite passes in
  28.733 seconds.
- A final strict regeneration passed all three public verifiers, omitted
  `pkg/interpreter`, and produced byte-identical `compiled.go` files to the
  pre-experiment baseline in every application.
- Touched production files returned to their pre-tranche form and remain
  below 1,000 lines.
- No canonical stdlib, runtime package, interpreter, bytecode VM, language,
  dependency, non-primitive nominal, application, or WASM change was retained.
- Removed the exact 2.1 GiB disk-backed tranche workspace after preserving the
  aggregate evidence; unrelated temporary directories were untouched.

Machine-readable aggregate:

- `2026-07-28-compiled-native-awaitable-array-round-trip-rejected.json`

## Next

Measure the remaining per-element native Awaitable interface-to-runtime
crossing and select a general typed arm representation only if it removes that
crossing without closure or interface escape costs.

Why: the rejected Array wrapper route left exactly 2,560/2,048/4,096
interface conversions in both directions and 7,680/6,144/12,288 protocol
calls. Those operations, rather than the Array wrapper alone, now delimit the
next compiled/runtime boundary.

What it entails: profile and count native-interface adapter shapes; prototype
a language-kernel arm abstraction that invokes `is_default`, `is_ready`,
`register`, and `commit` through generated static methods while preserving the
dynamic `Iterable` and runtime Awaitable fallbacks; then repeat semantic,
allocation, and balanced equivalent-Go gates across the same three unlike
applications.

Why it is important: compiled Able remains 3.82x-27.29x slower than the
equivalent Go programs here. Removing only the wrapper was too small; keeping
the Awaitable protocol itself on native Go carriers is the next plausible way
to eliminate a full compiled/interpreted representation crossing.
