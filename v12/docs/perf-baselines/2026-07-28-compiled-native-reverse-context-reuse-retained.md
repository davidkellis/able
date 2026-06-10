# Compiled native reverse-context reuse retained

Date: 2026-07-28

## Decision

Retain task-payload ownership of the primary generated execution context under
the existing experimental callable-context gate.

When generated code calls a native callable, that callable receives the
embedded `runtime.NativeCallContext` view of the current compiled execution
context. If the native callable subsequently invokes generated code, the
reverse adapter now returns the same primary task context when both the task
payload and environment identity match. Nil contexts, host-created contexts,
and different captured or package environments retain the established
reconstruction path.

This is a general generated-call rule. It does not change
`runtime.NativeCallContext`, its `State` contract, interpreter behavior, or
the callable ABI. It names no benchmark, container, or non-primitive nominal
type.

## Exact construction and caller attribution

Five public-verifier-backed counter runs per application distinguished
necessary task creation from reverse native-call adaptation:

| Application | Baseline task contexts | Baseline reverse contexts | Retained reused contexts | Retained reverse fallbacks |
| --- | ---: | ---: | ---: | ---: |
| Await Channel Mux | 1,536 each run | 3,073 each run | 3,072 each run | 1 each run |
| Mutex Await Journal | 4 each run | 6,222-6,411 | 6,427-6,510 | 1 each run |
| Mutex Work Queue | 4 each run | 12,629-12,816 | 12,794-12,934 | 1 each run |

The varying Mutex counts are scheduler-dependent registration retries. Every
reverse call except the generated main-entry compatibility wrapper reused the
task context. Caller stacks attribute the repeated baseline construction to
general generated callback adapters and channel/mutex Awaitable
`register`/`commit` paths. The task counts are unchanged.

Main-delta allocation profiles independently confirm the counters:

- Await Channel Mux:
  `__able_context_from_environment` falls from 4,609 objects to 1,537.
- Mutex Await Journal:
  6,576 profiled context objects disappear; four task contexts plus the main
  fallback remain below the profile display threshold.
- Mutex Work Queue:
  12,747 profiled context objects disappear; four task contexts plus the main
  fallback remain below the profile display threshold.

## Allocation A/B

Five rotating, public-verifier-backed main-phase processes per side measured:

| Application | Baseline bytes | Retained bytes | Change | Baseline objects | Retained objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 5,554,467.2 | 5,408,678.4 | -2.62% | 101,905.0 | 98,842.0 | -3.01% |
| Mutex Await Journal | 1,800,664.0 | 1,488,499.2 | -17.34% | 37,619.6 | 31,138.6 | -17.23% |
| Mutex Work Queue | 3,813,132.8 | 3,173,387.2 | -16.78% | 79,713.8 | 66,577.6 | -16.48% |

Both allocation measures improve in all three unlike applications.

## Balanced timing and Go comparison

Fifteen rotating baseline/retained/equivalent-Go cohorts per application
measured:

| Application | Baseline | Retained | Raw change | Paired 95% interval | Go | Retained/Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.067605s | 0.068111s | +0.75% | -1.25% to +2.95% | 0.003034s | 22.45x |
| Mutex Await Journal | 0.007495s | 0.006852s | -8.58% | -13.21% to -2.22% | 0.002335s | 2.93x |
| Mutex Work Queue | 0.015219s | 0.013906s | -8.62% | -16.91% to +0.71% | 0.002979s | 4.67x |

Journal improves significantly. Queue's raw mean improves but its paired
interval crosses zero, and Channel is neutral within workstation noise. No
timing regression is claimed. The rule is retained for its exact elimination
of synthetic reverse contexts, broad allocation reduction, and one clear
unlike-program timing win.

The retained applications still reach only 4.45%, 34.09%, and 21.42% of
equivalent-Go performance, so the 95% product target remains unmet.

## Semantic and scope gates

- The reusable pointer is an `atomic.Pointer` stored once on the task payload
  before user task code runs.
- Reuse requires exact environment identity. Cross-package localized
  environments, captured callbacks, host-created native contexts, nil
  contexts, and the main-entry wrapper use the established fallback.
- Full experimental execution-context fixture parity passes in 46.132
  seconds. Focused nested-spawn, cross-package interface, captured callback,
  dynamic named/value boundary, bound-method, lazy Await service, explicit
  protocol, and repeated mutex-contention guards pass.
- All three race-enabled generated applications pass their public verifiers
  without a race report.
- Default builds of all three applications and an experimental await-free
  N-body build are byte-identical to the pre-change generated source.
- All strict dependency graphs omit `pkg/interpreter`.
- Bridge, runtime, and `cmd/ablec` suites pass.
- `generator_render_runtime_future.go` remains at 996 lines and every touched
  source file remains below 1,000 lines.
- The scope-only ledger update changes only the `compiler-production` tree
  hash from `a79251ac...1c4074` to `af726275...f6b1d`. Its seven tests pass
  with 21 current closures and zero invalidations.

No canonical stdlib, runtime package, interpreter, bytecode VM, language,
dependency, application source, non-primitive nominal rule, or WASM change
was needed.

Machine-readable aggregate:

- `2026-07-28-compiled-native-reverse-context-reuse-retained.json`

## Next

Count `__able_await_state.ensureWaitCh`, wake signals, and rearm cycles across
the same three applications, with Future Await Race as a low-reach control.

Why: after reverse-context removal, the retained main-delta profiles show one
buffered wait channel per Await state as the next repeated generated object:
2,048 in Mutex Await Journal and 4,096 in Mutex Work Queue, with the same
mechanism present in Await Channel Mux.

What it entails: distinguish concurrent active waits from sequential task
reuse, then prototype only a task-owned reusable wake signal if exact counters
prove that a task never has overlapping generated Await states. Preserve
synchronous wake-before-park, cancellation, retry/rearm, serial executor, and
dynamic user-Awaitable semantics. Repeat verifier-backed allocation, race, and
balanced Go timing gates.

Why it is important: Array/interface/protocol-arm removal and goroutine-ID
routes are already closed. A safe task-local wait signal would remove one
remaining native Go allocation per Await without reopening those rejected
boundaries or changing the callable/runtime ABI.
