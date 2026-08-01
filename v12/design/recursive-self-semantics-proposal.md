# Recursive and nested `Self` semantics

Status: **option A selected, implemented, and retained**

Date: 2026-07-31

## Selected decision

The selected boundary between recursive static `Self` substitution and
dynamic interface method calls is:

> Substitute `Self` structurally and recursively when the receiver's concrete
> type is known, including implementation validation and generic constraint
> dispatch. A method called through an existential interface value must have a
> representable object-safe signature. The existing exact top-level `Self`
> result is object-safe because it reconstructs the originating interface
> view. Any other non-receiver occurrence of `Self` is static-only until the
> language defines a sound existential carrier for it.

This preserves native Go lowering for concrete and constrained calls without
requiring a compiler to reinterpret, copy, or recursively box arbitrary
nominal values.

## Why a decision is needed

The current specification establishes two useful facts:

- §10.1.3 permits recursive occurrences such as `?Self`; and
- §10.3.4 gives an exact top-level `Self` result the originating fully bound
  interface view and dictionary.

It then says a result that merely contains `Self` retains its “ordinary
instantiated result type.” That phrase does not say which of these incompatible
dynamic meanings is intended:

1. preserve the hidden concrete result, such as `Pair Box`;
2. expose `Pair Interface` and recursively upcast its contents;
3. expose an existential result such as `exists T: Interface. Pair T`; or
4. permit the call but erase the distinction at runtime.

The same ambiguity exists contravariantly. Given
`fn combine(self: Self, other: Self)`, two values typed as the same interface
may hide different concrete types. The caller cannot prove that the second
value has the receiver's hidden type.

Current implementations disagree observably. A checked dynamic
`Pair Self` call passes in both interpreters, while a strict compiled program
can build and then fail with `return type mismatch: expected DuplicatePair`.
Nullable and parameter-nested forms instead require compiler fallbacks. This
is a language-contract gap, not an admissible performance shortcut.

## Recommended contract

### 1. `Self` substitution is structural for known receivers

For implementation validation, inherent method resolution, and calls where
the receiver is concrete or represented by a statically constrained type
parameter, substitute the fully instantiated receiver for every occurrence of
`Self`.

Substitution descends through:

- function parameter and return types;
- nominal applications such as `Pair Self`;
- nullable types such as `?Self`;
- unions and result types;
- Arrays, ranges, futures, and other type constructors;
- callable parameter and return types; and
- generic constraints and `where` obligations.

For a higher-kinded self pattern such as `for F _`, `Self` denotes the
constructor and `Self T` substitutes the implementing constructor applied to
`T`, as already specified.

This is one type-substitution rule. It is not permission for a compiler to add
rules for `Pair`, `Array`, `Result`, or any other non-primitive nominal type.

### 2. Object safety is checked per interface method

An interface value may still be formed when some methods are static-only.
Calling a method through that existential value is valid only when the
method's signature is dynamically representable.

The following forms are dynamically representable in v12:

- the dispatch receiver as the first parameter, exactly `self: Self`;
- all other parameter and result types that contain no `Self`; and
- an exact top-level `Self` result, using the interface-view reconstruction
  already defined by §10.3.4.

The following forms are static-only:

- an additional parameter of type `Self`;
- `Self` nested in an additional parameter, including `?Self`;
- `Self` nested in a result, including `?Self`, `Pair Self`,
  `Result Self E`, a union member, or a callable type;
- a higher-kinded application such as `Self T`; and
- a method-level obligation whose satisfaction requires revealing or equating
  the interface value's hidden concrete self type.

The checker should diagnose an attempted dynamic call at the call site. It
should not reject the interface declaration, its implementations, concrete
calls, or generic `T: Interface` calls.

### 3. Exact top-level `Self` remains special

For:

```able
interface Clone for Self {
  fn clone(self: Self) -> Self
}
```

calling `clone` on a `Clone` interface value returns the same fully bound
`Clone` view. The runtime or generated adapter wraps the returned concrete
value with the originating interface definition, arguments, and captured
dictionary. Consumer scope does not select another implementation.

This exception is representation-safe because the result itself is the
interface view. It does not recursively transform a containing value.

### 4. Nested results are not recursively lifted

For:

```able
interface DuplicatePair for Self {
  fn duplicate_pair(self: Self) -> Pair Self
}
```

an implementation for `Box` returns `Pair Box`. A concrete or constrained
call may use that type directly. A call on a `DuplicatePair` interface value
is a static error in v12.

The compiler must not manufacture `Pair DuplicatePair` by:

- reinterpreting `Pair Box`;
- walking and boxing its fields;
- relying on covariance that Able has not declared;
- erasing generic arguments at the compiled boundary; or
- adding a `Pair`-specific or named-container-specific lowering rule.

Those transformations can change identity, aliasing, mutation, variance, and
allocation behavior. They also cannot be generalized to opaque user-defined
nominal types.

### 5. Non-receiver `Self` parameters are not dynamically callable

For:

```able
interface Combine for Self {
  fn combine(self: Self, other: Self) -> Self
}
```

two `Combine` interface values do not prove that their hidden concrete types
are equal. Therefore `left.combine(right)` is invalid when `left` is
interface-typed. It remains valid when the receiver and argument share a
known concrete type or a single statically constrained type parameter.

Nesting `Self` under nullable, union, callable, or nominal constructors does
not change this rule.

### 6. Recursive constraints use a finite proof graph

Constraint solving should key an obligation by its canonical fully
instantiated triple:

```text
(subject type, interface, interface arguments)
```

A recursive edge is well-founded for v12 when resolution revisits the exact
same key and the cycle is anchored by an implementation or other canonical
conformance proof currently being validated. The repeated key is a back-edge,
not a request to recurse again. Every distinct non-cyclic leaf must still be
satisfied.

Examples:

- `interface I for Self where Self: I` may close at the identical `I` proof
  anchored by the implementation being checked;
- mutually recursive `I`/`J` constraints require proof anchors for both
  interfaces before their repeated keys can close; and
- a cycle whose instantiated arguments grow, such as repeatedly introducing
  another constructor layer, never reaches an identical key and is rejected
  as not well-founded.

This makes “no infinite regress” operational and deterministic. It does not
allow an unanchored cycle to prove its own conformance.

### 7. Compiler and runtime consequence

After selection, one shared checker fact should classify each interface method
as dynamically callable or static-only. The compiler and both interpreters
must consume that fact or reproduce the same canonical predicate.

Static-only methods must fail during typechecking when invoked on an
interface-typed receiver. The compiler must also fail closed if checker facts
are unavailable; it must not admit a strict build that will fail only during
generated execution.

Concrete and generic static calls should continue through ordinary native
lowering. No dynamic dictionary, `runtime.Value`, recursive box, or interpreter
boundary is required for those calls.

## Alternatives

### Option A — per-method object safety

This is the recommended rule above.

Advantages:

- sound for invariant and mutable nominal types;
- preserves concrete and generic source compatibility;
- matches the existing exact top-level `Self` dictionary contract;
- requires no recursive value transformation;
- permits native Go carriers and direct static calls; and
- needs no named-container rule.

Cost:

- a previously accepted but ambiguous dynamic call becomes a checker error.

### Option B — recursively lift nested results

Under this option, `Pair Box` would become `Pair Interface`.

This is not recommended. It requires declared variance plus a general
identity/aliasing-preserving transformation contract. Copying fields is not
valid for opaque or mutable nominals, while reinterpreting their Go carriers
is unsound.

### Option C — existential associated results

Under this option, a dynamic call could return a type equivalent to:

```text
exists T: DuplicatePair. Pair T
```

This is type-theoretically sound but not recommended for v12. Able has no
syntax, inference rule, pattern rule, or runtime carrier for opening such an
existential. It would be a separate language feature, not a lowering tweak.

### Option D — erase nested `Self` dynamically

This approximates the permissive behavior of the current interpreters.

This is rejected. It lets the checker promise `Pair Interface` while runtime
storage contains `Pair Concrete`, and it permits mixed hidden types in
non-receiver `Self` parameters. The reproduced compiled failure demonstrates
that erasure is not cross-runtime semantics.

## Active-source compatibility

The current v12 parser successfully parsed all 760 `.able` files under:

- `v12/fixtures`;
- `v12/examples`; and
- `../able-stdlib/src`.

There were zero parse failures.

Interface signature occurrences:

| Occurrence | Count |
| --- | ---: |
| exact `Self` parameters | 162 |
| higher-kinded `Self T` parameters | 3 |
| exact top-level `Self` results | 19 |
| higher-kinded `Self T` results | 2 |
| ordinary nested-nominal `Self` results | 1 |
| interface `where` clauses containing `Self` | 0 |

The only ordinary nested-nominal result is canonical stdlib
`Integral.div_mod`:

```able
fn div_mod(self: Self, other: Self) -> DivModResult Self
```

Option A preserves its concrete and statically constrained uses. It would
make `div_mod` unavailable on an existential `Integral` value because both the
second parameter and result expose the hidden self type. No active source in
the audited roots performs that dynamic call.

The two higher-kinded result occurrences exercise the established static
constructor substitution rules. They do not establish a dynamic existential
ABI.

## Reproduced behavior

Four temporary source cases were checked outside the repository:

1. `?Self` result;
2. `Pair Self` result;
3. nested `Self` in an additional parameter; and
4. a direct recursive `where Self: Recursive` constraint.

The first three typechecked and ran in tree-walker and bytecode modes. The
direct recursive constraint also passed both modes and strict compiled
execution.

Strict compiled results for the dynamic cases were:

| Shape | Strict build | Execution |
| --- | --- | --- |
| `Pair Self` result | accepted | failed: `return type mismatch: expected DuplicatePair` |
| `?Self` result | rejected: two fallback wrappers required | not run |
| `?Self` parameter | rejected: one fallback wrapper required | not run |

The accepted `Pair Self` generated graph omitted `pkg/interpreter` but used the
compiled bridge to dispatch and convert between distinct generated
`Pair_Box` and `Pair_DuplicatePair` carriers. The conversion is where the
semantic mismatch becomes observable.

## Retained implementation

The retained tranche:

1. canonicalized the static-substitution and object-safety rules in §10.1.3
   and §10.3.4;
2. added one recursive method-safety classifier and a coded,
   source-attributed checker diagnostic;
3. added positive concrete/generic/exact-result and negative dynamic
   cross-runtime fixtures;
4. made tree-walker and bytecode reject static-only interface-value calls
   through the shared checker;
5. made strict compilation fail closed before generation;
6. preserved the exact top-level `Self` dictionary path and native carriers;
7. added anchored direct/mutual and growing-cycle proof-graph tests; and
8. verified canonical `Integral.div_mod` directly and through `T: Integral`.

No stdlib or execution-runtime change was required. The reconciled selector
admits no performance mutation because all 66 valid generated Go modules are
byte-for-byte unchanged.

## Non-goals

This proposal does not:

- change exact top-level `Self` returns;
- define variance;
- introduce associated or existential result syntax;
- authorize recursive container boxing;
- add named-container or non-primitive nominal lowering;
- reopen the exact-`Self` performance route with zero natural benchmark
  reach;
- change canonical stdlib APIs; or
- begin WASM work.
