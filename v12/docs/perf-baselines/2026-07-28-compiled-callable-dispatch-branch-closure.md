# Compiled callable-dispatch branch closure

Date: 2026-07-28

## Goal

Split the remaining context-aware callable dispatch cost into exact branches
across Future Await Race, Await Channel Mux, Mutex Await Journal, and Mutex
Work Queue. Attribute every residual `bridge.currentGID` call to its generated
caller, then advance only a general branch that is material in at least three
applications.

## Method

`v12/bench_compiler_callable_dispatch_instrument.py` overlays generated
`compiled.go`, the copied bridge, and generated `main.go`. It counts native
function and native-bound shapes, compatibility fallback, lookup success and
failure, result/error control, environment swap attempts, and complete
four-frame `currentGID` caller stacks. Five processes per application ran with
the catalog executor, CPU budget, working directory, arguments, canonical
stdlib, and public verifier.

The counters are diagnostic only. Their synchronization and stack inspection
make them unsuitable for timing claims.

## Exact branch result

| Application | Context calls | Native-bound calls | Successful method lookups | Legacy native-bound calls |
| --- | ---: | ---: | ---: | ---: |
| Future Await Race | 902.4 | 902.4 | 902.4 | 207.8 |
| Await Channel Mux | 9,216.0 | 8,704.0 | 8,704.0 | 2,048.0 |
| Mutex Await Journal | 8,520.8 | 6,472.8 | 6,472.8 | 120.6 |
| Mutex Work Queue | 17,206.6 | 13,110.6 | 13,110.6 | 296.8 |

Every context compatibility/default call, context environment swap, method
lookup error, callable error, nil result, and node-aware method call was zero.
The remaining non-bound context calls were direct native functions in Await
Channel Mux, Mutex Await Journal, and Mutex Work Queue. Every reached
native-bound function reported `BorrowArgs=false`.

Native-bound lookup plus receiver injection is therefore the only large exact
branch shared by all four applications.

## Residual environment attribution

Fresh total `currentGID` means were 1,293.2, 8,707.0, 130.6, and 306.8. The
only caller present in all four was the legacy Await waker method path, at
108.8, 512.0, 116.6, and 292.8 calls. It is not a material common owner:
Await Channel Mux additionally has 1,536 Future adapter lookups, 3,072 task
swap/restore lookups, and several 512-call named/default paths, while Future
Await Race distributes its remainder across task swaps, yield, named calls,
and definition lookup. Optimizing residual goroutine identification would
again be a two-application route.

## Rejected borrowed-argument candidate

The largest shared branch allocates a receiver-prefixed `[]runtime.Value`.
The existing `NativeFunctionValue.BorrowArgs` contract made a safe general
scratch experiment possible: only compiler-generated native methods proven
not to retain arguments were marked borrowed; nested or concurrent reuse fell
back to owned arguments.

Two forms completed five rotating, verifier-backed allocation runs per side:

| Scratch form | Application | Bytes | Objects |
| --- | --- | ---: | ---: |
| Four slots | Future Await Race | +2.57% | -0.77% |
| Four slots | Await Channel Mux | +1.90% | -5.43% |
| Four slots | Mutex Await Journal | +7.55% | -6.87% |
| Four slots | Mutex Work Queue | +7.14% | -7.02% |
| One slot | Future Await Race | +1.23% | 0.00% |
| One slot | Await Channel Mux | +0.26% | -4.28% |
| One slot | Mutex Await Journal | +1.89% | -6.82% |
| One slot | Mutex Work Queue | +1.78% | -6.75% |

Escape diagnostics explain the tradeoff: passing borrowed storage to the
indirect native implementation makes the storage owner escape. The four-slot
form also enlarges every task context. The one-slot form limits that cost but
still increases bytes in every application and cannot improve Future Await
Race objects.

The candidate failed admission before wall-time/Go-reference timing and was
removed. The retained compiler again emits byte-identical `compiled.go` for
all four applications relative to the pre-experiment census; final graphs
omit `pkg/interpreter`, outputs match, and all public verifiers pass.

## Decision

Retain the diagnostic instrument and no production optimization. No runtime,
interpreter, bytecode VM, canonical stdlib, language, dependency,
named-container/non-primitive nominal, or WASM change is part of this tranche.
The exact 2.6 GiB disk-backed tranche workspace and generated Python cache
were removed after retaining the aggregate evidence.

Machine-readable aggregates:

- `2026-07-28-compiled-callable-dispatch-branch-closure.json`
- `2026-07-28-compiled-callable-dispatch-borrow-candidate.tsv`

## Next

Prototype a diagnostic-only split-receiver native-bound carrier under the
existing experimental callable gate. It should preserve the already-generated
general `direct(runtime, environment, receiver, explicitArgs)` method shape
through method lookup instead of rebuilding a receiver-prefixed
`[]runtime.Value`.

Why: scratch reuse cannot remove the allocation without transferring escape
cost into the context, while a split-receiver call avoids constructing that
slice at all.

What it entails: first count which successful lookups already have a general
compiled direct entry; then use a compiler-owned bound value in an overlay,
exercise value/pointer, captured, cross-package, interface, error, nested, and
dynamic compatibility guards, and repeat allocation plus balanced
baseline/candidate/Go timing across the four reached applications and unrelated
serial controls.

Why it is important: this is the remaining route that directly lowers a
general nominal method call toward its native Go receiver/argument ABI without
adding a named-type rule or crossing back into the interpreter.
