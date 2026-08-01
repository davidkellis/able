# Open-set exhaustiveness and pattern refutability

Status: **Option A selected and canonicalized; checker facts retained,
compiler branch-elision candidate rejected by broad A/B**

Date: 2026-07-31

## Decision requested

Option A, the conservative static-proof rule in this note for `match`,
`rescue`, and destructuring patterns, was selected on 2026-07-31.

The rule applies to every interface and other open type. `Error` is the
motivating case, not a compiler special case.

The central decision is:

> An open interface component is covered only by an unguarded pattern whose
> accepted type contains that whole component, or by an unguarded universal
> pattern. Enumerating concrete implementations never exhausts an open
> interface.

Under this rule, these two clauses have identical coverage:

```able
case error: Error => recover(error)
case _: Error     => recover()
```

Both cover the complete open `Error` component. The first also binds its
value. A wildcard or bare binding covers the complete remaining subject
domain:

```able
case _     => recover()
case value => recover(value)
```

## Why a decision is needed

The current specification says compilers should check exhaustiveness and says
that an open interface requires a wildcard or `case _: Interface`. It does not
define:

- whether a typed binding such as `case error: Error` is equivalent;
- whether guarded clauses contribute to coverage;
- how coverage composes through unions and nested patterns;
- whether an unproven ordinary match is a compile-time error;
- whether a partial `rescue` is intentionally valid; or
- when a compiler may remove the dynamic no-match/re-propagation branch.

The implementations consequently preserve dynamic checks everywhere. A sound
proof can let static Go lowering omit checks only where no source value can
reach them, while preserving the existing behavior for partial programs.

## Recommended rule

### 1. Coverage is relative to the static subject domain

Let `D` be the values admitted by the subject's static type. Pattern analysis
computes a conservative subset `coverage(P, D)` that pattern `P` is proven to
accept.

A sequence of clauses is exhaustive when the union of their proven coverage
contains `D`. Equivalently, subtract each clause's proven coverage from an
initial remainder `D`; the sequence is exhaustive when the remainder is
empty.

Failure to prove that the remainder is empty is not itself proof that the
source program is invalid.

### 2. Guards never contribute to an exhaustiveness proof

A clause with a guard contributes no static coverage, even when its pattern is
otherwise universal:

```able
case value if predicate(value) => value
```

The guard may be false. Treating all guarded clauses conservatively avoids
making coverage depend on constant folding or effect analysis. The clause
still participates normally at runtime.

A future, separately specified constant-proof optimization may recognize a
guard that is provably always true. It must not be required for the initial
contract.

### 3. Universal patterns cover the remaining domain

An unguarded wildcard or bare binding is irrefutable relative to every valid
subject domain:

```able
case _     => ...
case value => ...
```

Either covers all values remaining at that source position.

### 4. Typed patterns use semantic subtyping

For a typed binding or typed discard, the proven coverage is the intersection
of the remaining domain and the annotation:

```able
case value: T => ...
case _: T     => ...
```

Binding versus discarding does not affect coverage. The pattern is
irrefutable relative to a domain `D` exactly when every value in `D` conforms
to `T`.

This rule is based on resolved semantic types, not source spelling, aliases,
package names, or nominal-container identities.

### 5. Open interfaces are never closed by implementation enumeration

An interface denotes an open set even when the compiler currently sees every
implementation in the program.

For an `Error` subject:

```able
case _: ParseError      => ...
case _: ValidationError => ...
```

does not exhaust `Error`. A later package or dynamic boundary may supply
another implementation.

Either of these unguarded clauses does cover the complete component:

```able
case error: Error => ...
case _: Error     => ...
```

The same rule applies to any interface `I`. There is no `Error`-specific,
named-implementation, container, or whole-program exception.

### 6. Closed unions are covered component by component

Nested unions are normalized to their semantic member domains for coverage.
An unguarded pattern may cover one member, several members through a
supertype, or the whole union.

For `Value | Error`, this is exhaustive:

```able
subject match {
  case value: Value => use(value),
  case error: Error => recover(error),
}
```

The `Error` arm closes the whole open `Error` member. Concrete error arms do
not.

The proof must respect normal overlap and subtyping. It must not assume that a
union's source-level spellings are disjoint when the semantic types overlap.

### 7. Nested patterns are irrefutable only when every required check is

A pattern is irrefutable relative to `D` when its proven coverage contains all
of `D`.

- wildcard and bare binding patterns are irrefutable;
- typed patterns are irrefutable when `D` is a subtype of the annotation;
- a struct pattern can cover its nominal component when the subject is known
  to have that exact structure and every present nested pattern is
  irrefutable; omitted fields do not make it refutable;
- a singleton variant pattern covers that singleton member;
- literals are refutable except where the type system proves that the domain
  is that singleton;
- ordinary Array patterns are refutable because an Array type does not prove
  a length; and
- any guard makes its entire clause refutable for coverage.

The initial implementation may conservatively decline a valid proof for a
complex nested pattern. It must never claim coverage it cannot prove.

### 8. Ordinary `match` remains dynamically checked when proof fails

Non-exhaustive ordinary matches remain valid v12 source programs. The checker
may issue a diagnostic, but lack of a proof alone is not a compile-time
rejection.

When no clause matches at runtime, tree-walker, bytecode, and compiled
execution retain the canonical `Non-exhaustive match` behavior.

The compiler may remove its dynamic no-match branch only when the shared
analysis proves the match exhaustive. This optimization changes no observable
behavior for a well-typed source value.

This preserves canonical fixtures that intentionally exercise runtime
non-exhaustive diagnostics.

### 9. `rescue` may intentionally be partial

A rescue clause list does not need to be exhaustive. When no handler matches,
the original raised value continues propagating.

An unguarded universal or whole-`Error` typed handler proves handler-selection
exhaustiveness. That proof may remove only the unmatched-handler
re-propagation branch. It does not prove that a selected handler body cannot
raise or propagate another error.

Partial concrete handlers remain useful and valid:

```able
work() rescue {
  case error: ValidationError => repair(error)
}
```

### 10. `or {}` and `!` do not define pattern coverage

`or {}` structurally handles the failure alternative supplied to its handler.
It has no list of patterns to prove exhaustive. The handler body may still
raise, return, or otherwise transfer control.

`!` is a propagation operator, not a pattern match. Its static and generated
control behavior is governed by the operand's success/error type, not by
exhaustiveness analysis.

An ordinary match over a `Result`-like union still uses the ordinary match
rules above.

### 11. Refutable assignment and declaration patterns remain valid

Pattern assignment and declaration may use refutable patterns. A runtime
mismatch retains the specified `Error` result.

Static irrefutability is an optimization fact that may remove impossible
mismatch plumbing. It is not an acceptance requirement for these forms.

## Diagnostics

The first implementation should distinguish:

- **proven exhaustive**: safe to annotate for lowering;
- **not proven exhaustive**: retain runtime fallback; optionally warn where
  useful; and
- **unreachable clause**: optionally warn when an earlier unguarded clause
  provably covers the complete remainder.

An open interface with only concrete implementation clauses is never
classified as proven exhaustive.

Diagnostics should identify the uncovered semantic component where practical,
for example `Error (open interface)`, rather than listing the implementations
currently visible.

## Compiler and AST contract

The proof belongs in one semantic analysis shared by the checker and compiler.
Store the result in an auxiliary fact map keyed by the canonical AST node. Do
not add checker/compiler annotations to the shared AST.

The compiler may consume only positive, sound proofs. If facts are unavailable
or stale, it must retain the existing dynamic branch.

No proof may depend on:

- application or benchmark identity;
- the name of a non-primitive nominal type or stdlib container;
- a closed-world list of interface implementations;
- generated-code size or measured runtime frequency; or
- the presence or absence of the interpreter package.

## Compatibility evidence

A syntactic AST census parsed all 759 Able files under:

- `v12/fixtures`;
- `v12/examples`; and
- `../able-stdlib/src`.

All files parsed successfully. The corpus contains 915 match expressions and
71 rescue expressions.

For matches:

- 351 contain an unguarded universal pattern;
- 102 contain an unguarded typed `Error` pattern;
- 84 contain typed `Error` coverage without a universal;
- 480 contain neither syntactic form, although many cover closed unions or
  other domains through more specific patterns; and
- four expressions contain guarded universal patterns.

For rescues:

- 26 contain an unguarded universal pattern;
- 42 contain typed `Error` coverage without a universal; and
- three are intentionally partial at the syntactic level.

This is a conservative syntactic census, not a semantic exhaustiveness result.
It demonstrates why a hard wildcard-only rule or immediate compile-time
rejection would be disruptive.

The current checker type-checks clause bodies and branch joins but performs no
exhaustiveness proof. Tree-walker and bytecode ordinary matches report
`Non-exhaustive match`; compiled matches emit an equivalent dynamic guard.
All three execution paths re-propagate the original raised value after an
unmatched rescue.

Canonical fixtures explicitly require runtime failure for guarded and
incomplete matches. Canonical stdlib code commonly uses a binding typed as
`Error`, so `case error: Error` must be recognized as equivalent to
`case _: Error` for coverage.

## Alternatives considered

### A. Conservative proof with runtime fallback — recommended

Adopt the rules above. This preserves source compatibility, creates useful
positive facts for lowering, and remains sound for open-world interfaces.

### B. Require compile-time exhaustiveness

Reject every ordinary match that cannot be proven exhaustive.

This provides a stronger safety policy, but it changes current language
behavior and breaks canonical fixtures that deliberately test runtime
non-exhaustiveness. It would also require complete semantic coverage analysis
before adoption. Do not select this option for v12 without a separate
compatibility migration.

### C. Keep runtime-only behavior

Define no static proof and preserve every dynamic check.

This maximizes implementation simplicity but leaves the specification gap open
and prevents safe removal of redundant control-flow checks.

### D. Close `Error` over visible implementations

Treat all concrete `Error` implementations visible during compilation as a
closed set.

Reject this option. It is unsound under separate compilation, imports, and
dynamic boundaries, and would create package-order-dependent semantics.

## Selection and implementation result

The rule is canonical in `spec/full_spec_v12.md`. The Go checker records
positive-only coverage facts outside the AST, with focused coverage for
typed bindings/discards, open interfaces, guards, unions, `Result`, rescue,
structs, and Arrays. The cross-runtime
`08_01_error_exhaustiveness_open_set` fixture preserves partial behavior and
whole-`Error` handling in tree-walker, bytecode, and compiled execution.

The compiler candidate that omitted a proven-impossible no-match or
re-propagation branch was evaluated and removed. Although it reduced generated
guards in three unlike applications by 9.36% to 41.14%, repeated balanced A/B
showed no broad runtime improvement: Policy wall time improved but user time
worsened, Sensor was neutral to slightly worse, and sustained Telemetry was
1.62% slower. The retained compiler therefore continues to emit every dynamic
fallback branch.

The complete decision and measurements are recorded in
`v12/docs/perf-baselines/2026-07-31-error-exhaustiveness-contract-retained-compiler-elision-no-go.md`.

## Out of scope

This decision does not settle:

- which values are legal operands of `raise`;
- the exact catchability/carrier of a runtime non-exhaustive-match diagnostic;
- effect inference for whether a handler body can itself raise;
- constant evaluation of guards; or
- WASM lowering.

Those questions must not be silently bundled into the exhaustiveness rule.
