# Cross-runtime residual census — 2026-07-15

## Scope

This read-only reconciliation follows the repaired `option_result_config`
scorecard row. It determines whether that semantic repair, together with the
now-complete 33-application portable scorecard, admits a new compiler or
bytecode implementation experiment.

It is deliberately not another timing run. The fast typed-pattern repair is a
correctness alignment and the newly verified portable consumer is one
application; profiling that application alone would not satisfy the
three-unlike-application admission rule. Existing repeated workstation means
and CPU profiles remain selection evidence: the repair has not created a
natural three-application cohort that could support a new performance
conclusion.

## Current coverage checks

The current promoted scorecard contains 33 portable applications, 66
compiled/bytecode rows, and 25 explicit source reports. It has 32 rankable
compiled rows, of which 4 meet the 95%-of-Go goal, and 24 rankable bytecode
rows, of which 3 meet both interpreter targets. Those ratios identify product
gaps; they do not identify a removable inner cost by themselves.

The current lowering audit also passed across all 33 application entry
modules:

```sh
env GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  ./v12/bench_bytecode_audit --suite coverage
```

It lowered 148 functions and 7,827 instructions. `LoadSlot` appears in 32
applications and `JumpIfNotTypedPattern` in 16. Those are static reachability
counts, rather than invocation or CPU counts, so neither authorizes a slot or
typed-pattern shortcut. In particular, the audit must not turn the
imported-interface matching repair into an `Option`, `Result`, or `Error`
special case.

## Residual reconciliation

| Runtime | Repeated parent or change | Concrete evidence | Decision |
| --- | --- | --- | --- |
| Bytecode VM | `finishInlineReturn` / call dispatch | The text, iterator, numeric, Option/Result, dependency-plan, and document-profile gates divide into return coercion/type matching, generator/member dispatch, Array/index work, or lambda lowering. The only three-way return leaf has already been neutral-to-mixed under broad guards. | Closed; do not retry guard order or frame work. |
| Bytecode VM | raw integer / map / member cache | Existing unlike workloads use different carriers and callers; each generic candidate already lost its broad guard. | Closed; do not add a raw-value, fixed-map, Array, or iterator fast path. |
| Generated compiler | `__able_call_value_fast` | The call-path census found four material unlike consumers, but normal-binary profiles split below the helper: Option/Result is generic-union allocation/dispatch, while the three concurrent applications reach the rejected `bridge.currentGID` → `runtime.Stack` identity boundary. | Closed; do not tune the helper parent or retry the fixed execution-context ABI. |
| Recent typed-pattern semantic repair | imported lexical type canonicalization | The repair now verifies the complete Option/Result application in both interpreters, but it establishes no natural three-application performance cohort and no timed leaf. | Correctness only; no profile or performance experiment. |
| Recent generic interface-default repair | `Iterator<T>` default return metadata | Document Audit and Lexical Rollup are the two natural public consumers; their profile descendants diverge. | Correctness only; do not create a synthetic third application. |

The source evidence is the current full scorecard, the Option/Result
reconciliation, the bytecode candidate-admission inventory, the three-shape
profile refresh, and the compiled call-path profile gate. Their conclusions
remain compatible after the scorecard promotion.

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib, benchmark,
or fixture performance change. No bounded profile was taken because there is
no newly admitted, concrete, non-nominal leaf shared by at least three unlike
verifier-backed applications. Repeating an unchanged profile or optimizing a
static opcode count would measure noise and risk making a benchmark fast while
leaving real programs slower.

## Next recommendation

Wait for a material cross-cutting semantic/compiler/runtime change, or for a
genuinely needed portable application that naturally creates a third unlike
consumer of an unprofiled boundary. Then run repeated verifier-backed
workstation averages and bounded CPU/allocation profiles for the three
applications; implement only a concrete shared descendant that remains neutral
on broad controls. This preserves the target of broadly useful compiler and VM
performance rather than optimizing a named type, container, protocol, or
benchmark shape.
