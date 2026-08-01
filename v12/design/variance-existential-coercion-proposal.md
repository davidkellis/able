# Variance and existential nominal-coercion proposal

Status: **option A selected and retained**

Date: 2026-07-31

The implementation and verification record is
`../docs/perf-baselines/2026-07-31-variance-invariant-callable-contract-retained.md`.

## Retained decision

The selected v12 boundary distinguishes top-level conversion from conversion
beneath a type constructor.

The recommended rule is:

> V12 type parameters are invariant. A conversion that is valid for `A -> B`
> does not imply `F A -> F B`. Function types also have no implicit subtyping:
> their complete signatures must be equivalent. Top-level numeric widening,
> union injection, and concrete-to-interface upcasts remain valid, but they do
> not recursively lift through nominal, container, future, interface-argument,
> or callable construction. Explicit variance declarations are deferred until
> a later language version defines syntax, declaration-site soundness checks,
> and a representation-preserving runtime contract.

This is option A below.

## Canonical language authority

Section 4.1.6 says Able has no general structural-subtyping lattice.
Section 4.1.7 already establishes the minimal v12 rules:

- all type parameters are invariant;
- v12 provides no variance-declaration syntax;
- stdlib containers are invariant by default;
- no implicit subtyping-based coercion exists; and
- conversion occurs only through an explicit constructor/conversion, union
  injection, or another specifically stated language rule.

No current AST node or parser production represents declared variance.
`GenericParameter` contains only its name and constraints. Therefore no v12
type can presently opt out of invariance.

Sections 6.3.5 and 10.3.4 add two top-level operations:

- `as Interface` constructs an existential interface view and captures its
  implementation dictionary; and
- a concrete value may be placed into an interface-typed position when its
  implementation is proven and unambiguous.

Neither rule recursively converts a containing nominal value. The recursive
`Self` contract makes that explicit: `Pair Box` does not become
`Pair Interface`.

## Reproduced implementation gap

The shared checker currently applies one-way assignability recursively:

- `AppliedType` compares each argument with `typeAssignable`;
- `Array`, `Map`, `Range`, `Iterator`, `Future`, and nullable types recurse
  through their element/result arguments;
- some `StructInstanceType` comparisons accept the nominal name without
  comparing all bound arguments; and
- every `FunctionType` has the same `Name()` value (`"Function"`), so the
  fallback equality path admits incompatible callable signatures.

Top-level integer widening is valid, but recursive use makes `Box i8`
assignable to `Box i32` and `Array i8` assignable to `Array i32`. That is
covariance that neither type declared.

The mutation case is observably unsound:

```able
fn overwrite_wide(values: Array i32) -> void {
  values[0] = 300_i32
}

fn main() -> void {
  narrow: Array i8 = [7_i8]
  overwrite_wide(narrow)
  print(narrow[0])
}
```

The checker accepts it. Tree-walker and bytecode share the original array and
print `300`, placing an `i32` outside the declared `i8` element type. Strict
compiled execution converts the `Array i8` through `runtime.Value`, constructs
an `Array i32`, mutates that converted array, and prints `7`. The build uses no
interpreter fallback, but it changes aliasing and crosses a runtime carrier.

Callable signatures have a parallel gap:

```able
fn narrow(value: i8) -> i8 { value }
fn apply_wide(value: (i32 -> i32)) -> i32 { value(300_i32) }
apply_wide(narrow)
```

The checker accepts the call. Tree-walker reports an `i8` parameter mismatch,
bytecode reports an `i8` return mismatch, and strict compiled execution reports
an 8-bit overflow after constructing a `runtime.Value` callable adapter.

These are correctness failures, not permissible backend implementation
choices.

## Boundary that already behaves invariantly

The checker rejects all of these:

```able
show(box_cat)                         ## Box Cat -> Box Display
box_cat as (Box Display)              ## explicit cast of the outer nominal
consume(producer_cat)                 ## Producer Cat -> Producer Display
```

Generic interface arguments are therefore already invariant, and `as` does
not synthesize a recursive nominal conversion.

An element-wise reconstruction remains valid:

```able
rebuilt: Box Display = Box { value: box_cat.value }
```

That expression creates a new `Box Display` deliberately and performs one
top-level `Cat -> Display` upcast for its field. All three execution modes
produce the same result. It does not reinterpret or recursively lift
`Box Cat`.

## Option A — invariant-only v12

Under option A:

1. Every nominal and built-in type argument is invariant after alias
   canonicalization and generic substitution.
2. Top-level literal fitting and numeric widening remain available where the
   numeric rules permit them.
3. Numeric widening does not recurse through `Array`, `Map`, `Range`,
   `Iterator`, `Future`, nullable, a user nominal, an interface application, or
   another type constructor.
4. Function values require equivalent arity, parameter types, return type,
   generic parameters, constraints, and obligations. V12 defines neither
   parameter contravariance nor result covariance.
5. A concrete value may upcast to an interface only at the value position
   being converted. `F Concrete -> F Interface` is not implied.
6. Union injection remains an algebraic operation on the value being inserted;
   it does not lift a containing nominal.
7. An explicit constructor or library conversion may allocate, copy, or map
   elements when its API states that behavior. `as` does not invent that
   operation for arbitrary nominal types.

### Why option A is recommended

- It matches the existing normative text.
- It is sound for mutable and identity-bearing values.
- It preserves native Go carrier identity: each static instantiation keeps its
  own exact generated representation.
- It eliminates hidden `runtime.Value` adapters caused only by a checker
  mismatch.
- It needs no named-container or non-primitive nominal special case.
- It leaves a future declared-variance design possible without committing v12
  to syntax or ABI semantics prematurely.

## Option B — add declared variance now

Option B would add syntax such as `out T` and `in T`, then permit conversions
according to declaration-site variance.

This requires substantially more than parser syntax:

- the AST must record variance;
- every field, method parameter/result, callable, constraint, and nested type
  occurrence must be polarity-checked;
- mutable fields and mutating methods must prevent covariance;
- inherited/composite interfaces must compose polarity;
- separate compilation must preserve the declaration facts;
- the runtime must preserve identity and aliasing across a converted view; and
- the compiler needs a general representation-preserving view, not a copied
  nominal or a `runtime.Value` bridge.

Go instantiated structs and slices are distinct types. Declared covariance
does not by itself make `[]Concrete` a `[]Interface`, and copying elements
would change identity/aliasing. Option B therefore cannot be treated as a
checker-only relaxation.

Option B is not recommended for v12.

## Option C — conversion on use

This option would preserve the current checker and allow backends to adapt
values as needed.

It is rejected. The Array reproduction proves that interpreter aliasing and
compiled copying produce different observable results. Callable adaptation
also defers a statically knowable mismatch to runtime and introduces a generic
carrier boundary.

## Option D — recursive existential lifting

This option would treat `F Concrete` as `F Interface` whenever `Concrete`
implements `Interface`.

It is rejected for the same reasons recorded by the recursive `Self` decision:
arbitrary nominal types do not declare covariance; recursive field boxing can
change identity, aliasing, mutation, layout, and allocation; and no general
existential result representation exists. It would invite prohibited
container- or nominal-specific compiler rules.

## Implementation after selection

If option A is selected:

1. Add one canonical type-equivalence predicate that recursively compares
   fully substituted arguments after alias normalization.
2. Use equivalence—not one-way assignability—beneath every invariant type
   constructor.
3. Compare complete function signatures instead of their shared `Name()`.
4. Preserve top-level numeric/literal, union, nullable-injection, and interface
   rules at their explicit entry points.
5. Add coded invariant-argument and callable-signature diagnostics.
6. Make strict compiler entry fail closed on those diagnostics. The low-level
   compiler currently prints ordinary checker errors as warnings and may
   generate an invalid conversion; the selected diagnostics must not reach
   generation.
7. Add negative cross-runtime fixtures for mutable Array widening, user
   nominal widening, callable mismatch, generic interface arguments, and
   nested existential casts.
8. Retain a positive fixture for element-wise existential reconstruction and
   top-level numeric/interface conversion.
9. Re-run the 66-application strict census. Existing valid generated modules
   should remain byte-for-byte unchanged; any changed module requires
   individual review.

This implementation should reject invalid programs before execution. It
should not add a runtime conversion, recursive box, nominal adapter, or
interpreter fallback.

## Compatibility audit

The audited active roots contain 829 Able files:

- `v12/fixtures`;
- `v12/examples`;
- `../able-stdlib/src`; and
- `../able-stdlib/tests`.

All 829 parse successfully. They contain 74 generic struct declarations and
eight generic union declarations. `Array` appears in 204 files, `Map` or
`HashMap` in 33, and `Future` in six. No source contains a variance declaration
or variance keyword.

The breadth of generic use makes an implicit ABI relaxation risky, but the
absence of variance syntax confirms that no declaration has opted into it.
Selection should be followed by the diagnostic implementation and complete
suite rather than by assuming source compatibility from a textual scan.

## Performance consequence

This proposal authorizes no performance candidate.

Option A is nevertheless aligned with the performance mission: valid static
programs retain exact native carriers, while invalid cross-instantiation calls
stop before they can allocate a converted container or construct a
`runtime.Value` callable adapter. Performance measurements are required only
if the selected checker change alters generated output for an existing valid
application.

## Non-goals

This proposal does not:

- select variance syntax;
- add structural subtyping;
- define recursive existential boxes;
- change numeric promotion;
- add named-container lowering;
- modify the canonical stdlib;
- alter a benchmark; or
- begin WASM work.
