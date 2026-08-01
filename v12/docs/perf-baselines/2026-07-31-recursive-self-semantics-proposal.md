# Recursive/nested `Self` semantics proposal

Date: 2026-07-31

## Outcome

**Proposal complete; no language or production behavior changed.**

The recommended v12 rule separates two cases:

- when a receiver is concrete or statically constrained, substitute `Self`
  recursively through the complete type expression; and
- when a receiver is an existential interface value, permit only
  dynamically representable methods.

The dispatch receiver `self: Self`, results containing no `Self`, and the
existing exact top-level `Self` result are representable. Other non-receiver
occurrences—including `?Self`, `Pair Self`, another `Self` parameter,
callable occurrences, and `Self T`—are static-only until Able defines a sound
existential carrier.

The complete proposal is
`v12/design/recursive-self-semantics-proposal.md`. It requires a maintainer
decision before specification or implementation work begins.

## Why this boundary is required

Recursively substituting a known type is a normal static type operation.
Recursively converting a runtime value is not.

If an implementation returns `Pair Box`, exposing it as
`Pair DuplicatePair` would require the language to declare variance and define
how an opaque nominal value is transformed without changing identity,
aliasing, mutation, or allocation behavior. A named `Pair`, Array, Result, or
other container rule would violate the shared nominal-lowering contract and
would not solve user-defined nominals.

Additional `Self` parameters have the dual problem: two values with the same
interface type can hide different concrete types, so the caller cannot prove
the implementation receives its own self type.

The exact top-level `Self` result remains valid because the result itself is
reconstructed as the originating interface view with the captured dictionary;
no containing value is transformed.

## Current implementation audit

The shared typechecker already substitutes `Self` recursively through
functions, nominal applications, nullable types, ranges, unions, futures,
generic constraints, and obligations. Implementation validation binds
`Self` to the concrete target.

Tree-walker and bytecode dictionary preservation is deliberately exact-result
only. The compiler's native interface carrier likewise admits exact
top-level `Self` and refuses shapes that merely contain it.

The seam is call admission: the checker currently accepts nested forms on a
dynamic receiver even though no general cross-runtime representation exists.

## Reproduced cross-mode gap

Temporary cases outside the repository tested:

- a `?Self` result;
- a `Pair Self` result;
- `?Self` in an additional parameter; and
- a direct recursive `where Self: Recursive` obligation.

All four typechecked. Tree-walker and bytecode executed all four cases.
The direct recursive constraint also passed strict compiled execution.

Strict compiled dynamic results:

| Shape | Result |
| --- | --- |
| `Pair Self` result | build accepted; execution failed with `return type mismatch: expected DuplicatePair` |
| `?Self` result | build rejected because two fallback wrappers were required |
| `?Self` additional parameter | build rejected because one fallback wrapper was required |

The accepted nominal case omitted `pkg/interpreter`, but generated distinct
native `Pair_Box` and `Pair_DuplicatePair` carriers and crossed the compiled
bridge for dynamic dispatch. The bridge conversion exposed the mismatch.

This is a correctness/specification gap. It does not authorize an optimization
or recursive boxing rule.

## Compatibility census

The current parser processed all `.able` sources under `v12/fixtures`,
`v12/examples`, and `../able-stdlib/src`:

- files: 760;
- parse failures: zero;
- exact `Self` parameters: 162;
- higher-kinded `Self T` parameters: 3;
- exact top-level `Self` results: 19;
- higher-kinded `Self T` results: 2;
- ordinary nested-nominal `Self` results: 1; and
- interface `where` clauses containing `Self`: zero.

The sole ordinary nested result is canonical stdlib
`Integral.div_mod -> DivModResult Self`. It also accepts `other: Self`.
The recommended rule preserves concrete and constrained calls while making a
dynamic existential `Integral.div_mod` call invalid. No audited source makes
that dynamic call.

## Recursive constraints

The proposal makes “well-founded” operational through a finite obligation
graph keyed by canonical `(subject, interface, arguments)`.

An identical repeated key may close only as a back-edge anchored by a real
conformance proof. Every distinct leaf must succeed. A cycle whose
instantiated arguments keep growing never returns to an identical key and is
rejected. This accepts direct finite recursion without allowing an unanchored
cycle to prove itself.

## Alternatives

1. Per-method object safety with recursive static substitution—recommended.
2. Recursive lifting of nested results—rejected as variance-, identity-, and
   representation-unsafe.
3. Existential associated results—sound in principle but a new language
   feature outside v12's current type/runtime model.
4. Dynamic erasure—rejected because current compiled execution demonstrates
   that the promised and stored nominal instantiations diverge.

## Verification

Focused unchanged-behavior guards passed:

```text
go test ./pkg/typechecker \
  -run 'Test(Implementation.*Self|Interface.*Self|CompositeInterfaceSelf|ConstraintSolverMethodSetSelf|MethodsDefinitionAllowsImplicitSelf)' \
  -count=1
ok .../pkg/typechecker 0.003s

go test ./pkg/interpreter \
  -run 'Test(PreserveInterfaceSelfReturn|InterfaceDynamicDispatch|AnalyzeFrameLayout.*Self)' \
  -count=1
ok .../pkg/interpreter 0.043s

go test ./pkg/compiler \
  -run 'TestCompiler(NativeInterface(SelfReturnStaysOnCarrier|Executes)|InterfaceParamAndReturnStayNative|ConcreteReceiverInterfaceMethodStaysNative|GenericInterfaceTouchpointsStayNative)' \
  -count=1
ok .../pkg/compiler 2.262s
```

The complete `./run_all_tests.sh` gate passed in 12:49.83 at
1,322,092 KiB peak RSS:

- 274 seeded exec fixtures, zero planned fixtures, and 275 fixture
  directories;
- 132 current scorecard rows and zero actionable frontier groups;
- 23 current performance closures and zero invalidations;
- a current five-node architecture evidence chain;
- every tree-walker, bytecode, and parity shard;
- all 844 compiler tests;
- canonical compiler outlier: 14.213 seconds; and
- slowest bounded compiler batch: 46.769 seconds.

The mode-aware selector emitted no compiled or bytecode owner.

All task parsing, building, reproduction, and verification used disk-backed
`/var/tmp`. The exact 2,566,140 KiB task workspace and its separate inventory
file were removed after recording the evidence. No task-owned Able temp path
remains.

## Scope

No specification, parser, AST, typechecker, interpreter, bytecode VM,
compiler, generated runtime, stdlib, dependency, fixture, benchmark, frozen
workspace, or WASM production behavior changed.

The earlier exact-top-level-`Self` benchmark route remains closed: the current
benchmark corpus has no natural exact-`Self` dynamic return reach. This tranche
did not manufacture a benchmark or reopen that performance route.

## Next

Obtain a maintainer decision on option A, per-method object safety.

Why: the current checker/interpreter/compiler disagreement cannot be repaired
soundly until v12 says whether nested `Self` is static-only, recursively
lifted, or existential.

What it entails: review the concrete/generic substitution rule, dynamic-safe
signature classification, exact-result exception, and finite recursive
obligation graph. If selected, canonicalize the spec and add checker,
cross-runtime, compiler fail-closed, and `Integral.div_mod` compatibility
coverage.

Why it matters: the decision keeps ordinary compiled work on native Go
carriers and prevents hidden compiler/interpreter crossings or unsound
recursive boxing at dynamic interface boundaries.
