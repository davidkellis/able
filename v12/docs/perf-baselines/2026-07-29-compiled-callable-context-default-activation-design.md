# Compiled callable-context default-activation design

Date: 2026-07-29

## Decision

Complete the authorized broad callable-context design tranche without a
production code change.

The source and retained evidence show that the required allocation-free ABI is
already implemented behind `ExperimentalExecutionContext`. It carries task
context through generated static calls, native interfaces, native callable
types, captured lambdas, bound methods, runtime-callable adapters, Awaitable
callbacks, and native-to-generated reverse calls. Dynamic and host
compatibility entries remain available.

Creating another ABI would duplicate retained work. The actual unresolved
work is safe default activation. The new authoritative design is
`v12/design/compiler-callable-context-default-activation.md`.

## Why this route was reopened

The current compiled frontier has no admitted production candidate under the
standing ownership rules. Its concurrency group accounts for 27.90% of
compiled positive target excess, and fresh unlike-program profiles identify
`bridge.currentGID`/`runtime.Stack` as the dominant exact owner. Local
Channel, Mutex, Future, Awaitable, named-container, or nominal shortcuts are
forbidden.

The maintainer repeatedly authorized proceeding after the no-candidate
closure, explicitly restated the goal of keeping compiled values and calls in
native Go, and authorized the broad boundary direction. This tranche therefore
revisited the formerly excluded callable-context family at design/proof scope
only.

## Evidence reconciled

- The July 24 broad experimental context rollout improved the 61-application
  geometric mean by 7.39% but had repeat-confirmed regressions across numeric,
  Unicode, and concurrency guards. A global default remains rejected.
- The July 27 await-gated callable extension improved Await Channel Mux
  56.94%, Mutex Await Journal 89.66%, and Mutex Work Queue 95.09%.
- The subsequent fourth reached guard, Future Await Race, improved 55.00%.
- A complete 63-application default/experimental census found callable-context
  execution in exactly those four applications, matching output in every row,
  126/126 public verifier passes, and 126/126 interpreter-free graphs.
- Embedded native-context and reverse-context reuse removed per-call synthetic
  allocations while preserving exact payload/environment identity checks.
- Later residual work closed the Await-local owner set; it did not invalidate
  the large baseline-versus-context boundary result.

## Source mapping

The existing implementation already supplies the required pieces:

- `generator_execution_context.go` selects context entries and appends the
  pointer to native callable signatures and invocations.
- `generator_render_execution_context.go` separates task environment, package
  environment, payload, and the embedded native-call view.
- `generator_render_callables.go` adapts runtime callables in both directions
  while keeping compatibility entries.
- `generator_exprs_lambda_cast_range.go` gives generated lambdas their own
  environment-effect proof and callable context.
- `generator_render_runtime_context_calls.go` handles native functions,
  native bound methods, compatibility fallback, and Awaitable calls.
- `generator_callable_execution_context.go` currently detects entry-package
  await only after consulting the experimental option.

That last point is the implementation seam. Production selection must compute
an option-independent scheduler requirement across all loaded modules before
other generator passes consult context activation.

## Tranche result

Retain the design and no production mutation.

The proposed next candidate is not a broad global ABI flip. It is an
activation-only change that uses the existing context ABI by default when the
loaded static program graph contains `await`, while requiring zero generated
source change in await-free applications. Compatibility/dynamic paths remain
unchanged.

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, language, dependency, benchmark, fixture, non-primitive
nominal, or WASM production source changed.

## Verification

This tranche is source- and evidence-backed design work:

- all selected records and active lowering authorities were read in full;
- current context/callable/environment source paths were mapped;
- the exact dirty worktree remained the prior 34-file deferred WASM hold
  before documentation edits;
- no RAM-backed `/tmp/able-*` artifact was created; and
- the implementation and measurement gates are specified in the design.

## Next

Implement the default-activation candidate behind an internal A/B seam, then
run its focused semantic/source-shape gates and the full 63-application
verifier-backed baseline/candidate/Go cohort.

Why: the existing ABI has already proved that explicit context transport can
remove the dominant goroutine-recovery boundary, but the product still ships
the slower compatibility path by default.

What it entails: make await detection option-independent and transitive across
loaded modules; activate the retained context ABI only for scheduler-requiring
programs; prove await-free generated source is identical; verify imported,
captured, cross-package, dynamic, host, and nested-task semantics; then repeat
balanced timing, allocation, CPU, and exact `currentGID` measurements.

Why it is important: this is the only evidence-backed general route that can
move the affected compiled applications materially toward Go without boxing
native carriers, re-entering the interpreter, or imposing the rejected
context overhead on ordinary serial programs.
