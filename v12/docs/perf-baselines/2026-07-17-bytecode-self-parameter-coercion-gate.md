# Bytecode `Self` parameter coercion gate — 2026-07-17

## Decision

Keep method-set `Self` resolution in bytecode frame-layout analysis and add
`char` to the cached primitive exact-type checks. Method parameters declared as
`Self` now use the owning method set's target type when the VM builds its
parameter coercion plan. Concrete primitives and non-generic named structs can
therefore use the same exact checks as explicitly named parameter types.
Targets containing method-set generics retain the existing generic safety path.

This is method-wide metadata, not an equality, benchmark, stdlib-container, or
named-application branch. The `char` check also applies to ordinary `char`
parameters and returns. No compiler, stdlib, language, fixture, or retained
benchmark-source change was needed.

## Admission audit

Temporary opt-in counters measured only the post-warmup benchmark call. Before
the change, every parameter reached through the four equality-heavy consumers
fell through to general coercion; none used an already-exact or simple
conversion decision.

| Workload | Calls | Params eligible | Already exact | General coercions |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 393,217 | 786,432 | 0 | 786,432 |
| Run-length encode | 959,953 | 1,919,904 | 0 | 1,919,904 |
| Unicode Scalar Pipeline | 1,769,473 | 3,538,944 | 0 | 3,538,944 |
| Temporary custom nominal `Eq` | 262,145 | 524,288 | 0 | 524,288 |

Iterator Collect and linked-list filter/map controls each entered the binder
once, skipped immediately, and visited no parameters. The repeated general
path was therefore specific to concrete method parameters written as `Self`,
but it recurred across boolean, character/text, Unicode scalar, and custom
nominal values.

Resolving `Self` initially exposed a missing primitive metadata case: integers,
floats, strings, and booleans had cached exact checks, but `char` did not. A
first Unicode pair was consequently neutral-to-negative. Adding the generic
`char` exact check completed the existing primitive table; all reported
candidate timings below use the completed candidate.

## Repeated performance gate

Every timing is a separate process with one warmup and one measured call,
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, CPU 0, skipped benchmark
typechecking, and the canonical external `able-stdlib`. Samples alternate
baseline/candidate order. Slow workstation samples remain in the arithmetic
means, and volatile rows were expanded to ten samples.

| Workload | Samples/side | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 5 | 774.906 ms | 713.318 ms | 7.95% faster |
| Custom nominal `Eq` | 10 | 472.158 ms | 449.862 ms | 4.72% faster |
| Run-length encode | 5 | 2.4983 s | 2.4670 s | 1.25% faster |
| Unicode Scalar Pipeline | 3 | 6.2392 s | 5.9046 s | 5.36% faster |
| Iterator Collect guard | 10 | 420.683 ms | 407.554 ms | 3.12% faster |
| Numeric Array Map guard | 10 | 97.287 ms | 98.907 ms | 1.67% slower |
| Base64 guard | 2 | 2.6051 s | 2.5423 s | 2.41% faster; bounded row |
| String Split/Join guard | 3 | 2.0498 s | 2.0021 s | 2.33% faster; bounded row |

Run-length includes a 3.215-second candidate process while its other four
candidate samples are 2.254-2.312 seconds. Custom nominal `Eq` includes a
723-millisecond candidate process while the other nine are 382-476
milliseconds. Iterator includes a 500-millisecond baseline process, and Array
Map includes paired 164-173 millisecond system-load samples among its usual
62-80 millisecond processes. None were discarded.

Base64 allocated about 2.2 GB per measured call. A third pair failed to produce
a complete benchmark result, so the row was stopped under the OOM guardrails.
Split/Join was likewise capped after three complete pairs when the next process
failed to return a result. Both bounded rows are supporting evidence, not the
primary admission basis.

Allocations are unchanged exactly for the nominal, iterator, and numeric-map
rows. Boolean adds five measured allocations and 416 bytes against 350,040
baseline allocations; the text rows vary only by their existing tiny
process-to-process setup tails. The retained decision is driven by CPU work,
not allocation movement.

## Post-change profile

A fresh Unicode CPU profile no longer shows
`invokeFunctionBindArgsForSlotLayout`, `inlineParamCoercionUnnecessary`, or
`coerceValueToType` in the top 80 cumulative nodes. Cached equality remains
30.23% cumulative and the complete equality interface path remains 31.78%, so
the selected subtree was removed without hiding the remaining invocation and
method-body cost.

The next flat owners include `bytecodeRawIntegerValueInfo` (2.79%), cached
call-name execution, binary dispatch, stack loads/stores, small-integer boxing,
and workload-specific method bodies. No one of those is admitted by this
tranche.

## Correctness and cleanup

- New layout tests cover concrete primitive `Self`, exact named-struct `Self`,
  generic method-set targets, and boxed/value `char` checks.
- Existing generic bound methods, overloaded methods, method shorthand,
  method-set generic returns, cached equality coercion, caller-argument
  ownership, custom `Hash`/`Eq`, operator-interface, and mixed-numeric parity
  tests pass.
- The complete `pkg/interpreter` suite and `pkg/runtime` suite pass within their
  55-second bounds.
- Changed Go files remain below 1,000 lines and `git diff --check` passes.
- Temporary counters, custom benchmark source, profiles, and test binaries are
  removed after recording the result.

## Next recommendation

Reconcile why canonical primitive `Eq` implementations still resolve to Able
function bodies during ordinary operator dispatch even though the interpreter
already provides semantics-preserving primitive native `Eq` callables. Profile
Boolean, Run-length, Unicode, mixed numeric equality, and custom nominal `Eq`,
then admit a resolver-precedence change only when it recognizes canonical
primitive implementations and preserves user-defined/nominal dispatch.

Why: parameter coercion has disappeared from the hot profile, but cached
equality still consumes about 30% cumulatively in Unicode. The existing native
primitive boundary could remove generic function invocation and detached VM
frame work for all primitive equality without adding a benchmark-specific
operation. The work entails dispatch-identity tracing, canonical-vs-overridden
implementation tests, mixed numeric/float/NaN semantics checks, and the same
repeated text/nominal/iterator/byte/numeric performance gate. WASM remains
deferred.
