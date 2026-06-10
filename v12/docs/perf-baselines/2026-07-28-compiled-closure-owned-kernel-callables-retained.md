# Compiled closure-owned kernel callables retained

Date: 2026-07-28

## Decision

Retain receiver-free storage for generated closure-owned native kernel
methods behind the existing experimental callable execution-context gate.

Generated Awaitable, AwaitRegistration, and AwaitWaker method implementations
already close over their actual service state and expose caller-facing arity.
They now remain `runtime.NativeFunctionValue` values instead of being wrapped
in a `runtime.NativeBoundMethodValue` whose synthetic receiver was rebuilt
into `[]runtime.Value` on every call.

Ordinary compiled methods, Future built-ins, non-primitive nominal lowering,
and default or await-free generation are unchanged.

## Scope and identity

Each reached experimental application replaces the same 20 generated
closure-owned bindings: four methods across four Awaitable surfaces, plus
legacy/context AwaitWaker and AwaitRegistration methods.

The production generator exactly reproduced the successful overlay source in
all four applications. A default Future Await Race build and an experimental
await-free N-body build remained byte-identical at every top-level generated
Go source file relative to the pre-change compiler.

All strict dependency graphs omit `pkg/interpreter`.

## Exact allocation A/B

Five rotating, public-verifier-backed main-phase allocation processes per
side measured:

| Application | Bytes change | Objects change |
| --- | ---: | ---: |
| Future Await Race | -1.64% | -1.96% |
| Await Channel Mux | -3.62% | -5.71% |
| Mutex Await Journal | -5.32% | -7.54% |
| Mutex Work Queue | -4.75% | -6.88% |

Both allocation measures improve in all four applications. Unlike receiver
scratch reuse, the rule removes the receiver wrapper and prefix slice rather
than transferring an escaping buffer into the execution context.

## Balanced timing and Go comparison

Fifteen rotating baseline/candidate/Go cohorts per application measured:

| Application | Baseline | Candidate | Raw change | Go | Candidate/Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| Future Await Race | 0.017714s | 0.017595s | -0.67% | 0.004046s | 4.35x |
| Await Channel Mux | 0.092295s | 0.090249s | -2.22% | 0.004910s | 18.38x |
| Mutex Await Journal | 0.012650s | 0.012452s | -1.56% | 0.003949s | 3.15x |
| Mutex Work Queue | 0.026141s | 0.024981s | -4.44% | 0.004828s | 5.17x |

Every raw mean improves, but every paired 95% interval crosses zero. Wall time
is therefore recorded as neutral; no speedup claim is made. The applications
remain well short of the 95%-of-Go product target.

## Correctness

The following gates pass:

- a new default/experimental representation guard;
- a new captured Awaitable `is_ready`/`commit` parity execution test;
- default-arm, Future, channel, mutex lifecycle, and public mutex contention
  fixtures;
- nested spawn, native-callable value/pointer, bound method, native interface,
  cross-package captured callback, error/rescue, and explicit dynamic-boundary
  execution-context guards;
- dynamic bound/native-bound callback success and failure guards;
- `go test ./cmd/ablec ./pkg/compiler/bridge`.

No canonical stdlib, interpreter, bytecode VM, language, dependency,
named-container/non-primitive nominal, or WASM change was made.

Machine-readable aggregate:

- `2026-07-28-compiled-closure-owned-kernel-callables-retained.json`

## Next

Refresh exact context-call branch counters plus CPU and allocation attribution
across the four reached applications after this representation change.

Why: the former dominant closure-owned native-bound branch has moved to the
allocation-free native-function lane, so the previous residual ownership
ranking is obsolete.

What it entails: recount remaining native-bound Future/interface calls,
successful field and interface lookups, compatibility calls, and residual
`currentGID` stacks; collect repeated profiles; select only a new owner that
is material in at least three unlike applications; then prototype and compare
it against equivalent Go programs.

Why it is important: the retained rule removes a real compiled/runtime
boundary, but the applications are still 3.15x-18.38x slower than Go. Fresh
attribution is required to continue lowering the next actual boundary rather
than optimizing a branch that this tranche already removed.
