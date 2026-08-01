# Invariant generic and callable contract retained

Date: 2026-07-31

Decision: **retain option A**

## Outcome

V12 generic arguments are invariant after alias normalization and generic
substitution. A valid top-level conversion from `A` to `B` does not imply a
conversion from `F A` to `F B`. Complete callable signatures must be
equivalent; v12 defines neither parameter contravariance nor result covariance.

Top-level numeric widening and literal fitting, union injection, nullable
injection, and a proven unambiguous concrete-to-interface upcast remain valid
at the value position where they occur. They do not recursively lift through
an Array, Map, Range, Iterator, Future, nullable, user nominal, interface
application, union application, or callable signature. Declared variance
remains deferred beyond v12.

## Retained implementation

The checker now uses one recursive invariant-equivalence relation beneath type
constructors and compares complete callable signatures, including generic
parameters, constraints, where requirements, and obligations. Alias and
special-type normalization precede comparison. Unknown or unresolved type
parameters remain provisional only while inference is genuinely incomplete.

Two coded error diagnostics identify the rejected boundaries:

- `invariant-type-argument`
- `callable-signature-mismatch`

The compiler treats both as fatal before generation. It does not construct a
recursive nominal conversion, copy an identity-bearing container, synthesize a
callable adapter, or admit an interpreter fallback.

Contextual checking preserves valid source programs without weakening the
contract. Conditional branches are checked at their expected outer type, and
an inline lambda inherits an expected callable signature before its body is
checked. This permits a top-level conversion inside a branch or lambda result
without treating the containing nullable, union, or callable as covariant.

The parser now canonicalizes keyword callable types such as `fn() -> T` and
`fn(i32) -> T`. The kernel declarations for
`__able_String_{from,to}_builtin` use `Array u8`, matching their native byte
carrier. No external stdlib source changed.

## Cross-runtime guards

The focused matrices cover Array, Map, Range, Iterator, Future, nullable, and
user generic arguments; legal top-level widening; callable parameter, result,
arity, and generic mismatch; and alpha-renamed equivalent generic callables.
Compiler guards prove both coded diagnostics fail closed.

Three execution fixtures retain the public behavior:

- `07_11_invariant_type_arguments_diag` rejects `Array i8 -> Array i32`;
- `07_12_callable_signature_diag` rejects `fn(i8) -> i8` where
  `fn(i32) -> i32` is required; and
- `07_13_explicit_invariant_reconstruction` accepts explicit element-wise
  reconstruction and the element’s top-level numeric widening.

Tree-walker, bytecode, and strict compiled fixture runs pass.

## Strict static census

All 66 current applications generated under `--no-fallbacks`. Every
application preserved its complete compiled-body boundary-category map, and
all 15 aggregate boundary totals exactly match
`2026-07-30-interface-dictionary-capture-strict-static-census.json`.

All 66 module hashes changed. They were reviewed individually rather than
accepted from aggregate totals. Canonical callable parsing and exact resolved
type facts changed generated support/type metadata; this tranche did not
change production generator or runtime source. Every row remained strict and
preserved its complete boundary map, so the changes introduce no new boxed,
dynamic, runtime-service, or interpreter boundary.

The compact row-level record is
`2026-07-31-variance-invariant-callable-strict-static-census.json`.

## Evidence state

Focused parser, checker, compiler, fixture, ledger, and architecture-chain
checks pass. The complete fast suite passed in 9:42.25 at 1,868,384 KiB peak
RSS. All 85 bounded compiler batches passed; the slowest took 46.300 seconds,
and the canonical outlier took 13.939 seconds. The performance ledger has 23
current closures and zero invalidations; the five-node architecture evidence
chain is current.

The first durable run exposed a contextual-lambda regression in the existing
Option/Result fixture. A concrete `ProbeError -> Error` conversion inside a
lambda was incorrectly diagnosed as outer callable covariance. That diagnostic
withheld checked receiver facts and made overlapping `unwrap` methods
ambiguous at runtime. The checker now applies the existing top-level interface
rule at the lambda result position. A focused checker guard plus tree-walker,
bytecode parity, and strict compiled Option/Result runs pass; no interpreter or
VM change was retained.

No production interpreter, bytecode VM, compiler generator, runtime,
dependency, external stdlib, benchmark source, frozen workspace, or WASM
change was retained.

Removed the exact 7,944,748 KiB disk-backed task workspace, 36,968 KiB
RAM-backed extern cache, and 44 KiB generated Python cache. No `/tmp/able-*`,
`/var/tmp/able-*`, or repository Python cache remains.

## Recommended next tranche

Close monomorphic local-lambda callable constraints. A locally bound
unannotated lambda used at two incompatible static signatures can still retain
an erased `runtime.Value` callable carrier.

This entails collecting every static callable constraint for the binding,
inferring one exact native signature when the constraints agree, and emitting
a coded callable mismatch before generation when they conflict. It then needs
focused inference/compiler guards and another strict broad census.

This is important because it is the clearest remaining avoidable
compiled/interpreted representation boundary exposed by this work. A general
constraint solution can keep valid monomorphic lambdas on native Go callable
carriers without introducing nominal, container, or benchmark-specific rules.
