# Compiled native Awaitable carrier retained

Date: 2026-07-28

## Decision

Retain the general native Awaitable carrier.

Compiled channel, mutex, timer, and default Awaitable values now implement the
language-kernel protocol directly as `runtime.NativeAwaitableValue`. Static
matching, native-interface recovery, Array conversion, and compiled await
selection preserve that carrier. The map-backed `runtime.StructInstanceValue`
is constructed only when an interpreter or ordinary dynamic member consumer
actually needs the semantic runtime representation.

This is a language-kernel boundary rule. It does not name an application,
stdlib container, user-defined nominal type, or benchmark.

## Exact reach and removal

Five baseline counter runs per application established an exact one-for-one
carrier round trip:

| Application | Native producers | Eager struct materializations | Interface conversions | Awaited elements |
| --- | ---: | ---: | ---: | ---: |
| Await Channel Mux | 2,048 channel | 2,048 channel | 2,560 each direction | 2,560 |
| Mutex Await Journal | 2,048 mutex | 2,048 mutex | 2,048 each direction | 2,048 |
| Mutex Work Queue | 4,096 mutex | 4,096 mutex | 4,096 each direction | 4,096 |

Five post-change runs per application preserved every producer, interface
conversion, Array conversion, and await protocol call while recording zero
channel, mutex, or timer materializations. All 30 instrumented executions
passed their public verifier.

Dynamic semantics remain available through
`runtime.RuntimeValueMaterializer`. Interpreter-facing bridge operations
materialize before entering the interpreter, and ordinary generated member
access materializes before inspecting struct fields. Static `bridge.MatchType`
and `bridge.Cast` deliberately keep the native carrier.

## Allocation A/B

Five rotating, public-verifier-backed main-phase processes per side measured:

| Application | Baseline bytes | Candidate bytes | Change | Baseline objects | Candidate objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 8,515,628.8 | 6,609,422.4 | -22.38% | 144,386.6 | 112,650.4 | -21.98% |
| Mutex Await Journal | 4,409,560.0 | 2,841,620.8 | -35.56% | 73,560.0 | 46,400.8 | -36.92% |
| Mutex Work Queue | 9,019,347.2 | 5,886,776.0 | -34.73% | 151,401.0 | 96,983.0 | -35.94% |

Every application improves strongly in both allocation measures.

## Balanced timing and Go comparison

Fifteen rotating baseline/candidate/equivalent-Go cohorts per application
measured:

| Application | Baseline | Candidate | Raw change | Paired 95% interval | Go | Candidate/Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.099467s | 0.090227s | -9.29% | -13.61% to -4.51% | 0.004778s | 18.88x |
| Mutex Await Journal | 0.012018s | 0.010473s | -12.86% | -16.66% to -8.07% | 0.004014s | 2.61x |
| Mutex Work Queue | 0.022841s | 0.018698s | -18.14% | -23.05% to -12.29% | 0.004406s | 4.24x |

All three paired intervals exclude zero. The change therefore improves both
allocation and wall time across three unlike applications. The remaining
5.30%, 38.33%, and 23.57% of equivalent-Go performance still miss the 95%
product goal.

All 135 timing processes passed the public verifier. Every candidate
dependency graph omits `pkg/interpreter`.

## Correctness and scope

- A generated-source guard requires the lazy carrier and forbids eager
  `return awaitable.toStruct(), nil`.
- A bridge regression guard proves static Match/Cast and type inference retain
  the carrier while explicit materialization produces the runtime struct.
- Explicit `is_default`, `is_ready`, and `commit` calls pass in default and
  experimental compiled modes.
- Repeated public mutex-await contention passes in both compiled modes.
- Focused execution-context, native-interface, Future, Await, Channel, Mutex,
  bridge, runtime, interpreter, and `cmd/ablec` suites pass.
- The scope-only performance-ledger bootstrap and all seven ledger tests pass
  with 21 current closures and zero invalidations.
- The three strict experimental applications build, verify, and remain
  interpreter-free.
- Every touched source file remains below 1,000 lines.

No canonical stdlib, tree-walker semantics, bytecode VM, language, dependency,
application source, non-primitive nominal rule, or WASM change was needed.

Machine-readable aggregate:

- `2026-07-28-compiled-native-awaitable-carrier-retained.json`

## Next

Measure and, if broadly beneficial, eliminate the remaining static Awaitable
Array/runtime round trip.

Why: after carrier materialization fell to zero, exact counters still show
1,024/2,048/4,096 native-Array-to-runtime conversions and
2,560/2,048/4,096 element interface conversions in the same three
applications.

What it entails: trace the compiler's native `[]Awaitable<T>` value into the
generated await syntax helper, add only a general typed await-input path that
accepts the native slice directly, retain the existing runtime Iterable path
for dynamic inputs, and repeat the verifier-backed allocation and balanced
Go-comparison protocol.

Why it is important: this is now the next exact shared compiled/runtime
crossing. Removing it would preserve native Go carriers farther through await
selection and directly advance the native-Go performance goal.
