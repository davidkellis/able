# Rank-1 inference and interface-HKT static boundary retained

Date: 2026-07-31

## Decision

Retain a closed v12 inference and higher-kinded-type boundary:

- ordinary inference is local, invariant, and rank-1;
- ordinary generic parameters bind concrete value types and cannot be applied
  as constructors;
- the sole higher-kinded abstraction is an explicit interface self pattern
  such as `for C _`;
- a matching bare implementation target fixes `C` before a call is lowered;
- type-constructor aliases remain compile-time expressions and cannot annotate
  runtime values while unbound; and
- erased `_` value signatures remain limited to explicit host extern/kernel
  ABI declarations.

V12 has no first-class constructor values, constructor-valued fields,
constructor call-site inference, higher-rank inference, impredicative
inference, or runtime constructor dictionaries.

## Inventory

The parser and AST already represent `_` with
`WildcardTypeExpression`. The checker already validates explicit interface
self patterns and bare constructor implementation targets. The compiler
already specializes concrete `Enumerable` calls without a dynamic member or
interpreter bridge.

Canonical stdlib has one constructor-abstract interface,
`Enumerable A for C _`, and twelve constructor implementations spanning
unlike nominal collection families. Current fixtures cover a passing
constructor implementation, a mismatched concrete target, and static calls
on Array and user-defined nominal constructors.

Empirical probes reproduced two leaks:

- a field annotated `Array _` passed typechecking; and
- `type AnyArray = Array _` could be used directly as a runtime parameter
  annotation.

An ordinary `fn keep<F, T>(value: F T) -> F T` was not usable, but failed only
later as a call mismatch rather than at its unsupported declaration.

## General implementation

The declaration and pattern type resolvers now reject unbound constructors in
runtime annotations. Type aliases, explicit interface self patterns, matching
implementation targets, and extern signatures retain their intentional
compile-time or host-boundary behavior.

Ordinary applied generic parameters now receive a focused declaration
diagnostic. Bare constructors passed as ordinary type arguments are rejected
only after the containing generic application has valid arity, avoiding
cascading diagnostics from an already-invalid outer type.

No parser, AST, interpreter, bytecode VM, compiler lowering, runtime
representation, canonical stdlib, dependency, benchmark, frozen workspace, or
WASM change was required.

## Native-lowering guard

The existing strict `Enumerable` compiler guard still proves that concrete HKT
calls use specialized compiled implementation functions and native nominal
carriers without `__able_method_call_node`, `__able_call_value`, or
`__able_member_get_method`. This tranche adds no runtime type-constructor
carrier and no compiled/interpreted boundary.

## Verification

Focused checker tests cover unbound fields, constructor-alias value use,
constructor-alias declaration, and ordinary applied generic parameters. The
new `04_01_hkt_static_boundary_diag` fixture passes in tree-walker, bytecode,
and parity lanes. Existing HKT implementation and strict native `Enumerable`
guards pass.

Canonical stdlib passes in tree-walker and bytecode modes.

The complete default v12 gate passes in 1,012.90 seconds at 1,836,920 KiB peak
RSS:

- 283 seeded exec fixtures and 284 fixture directories;
- every parser, tree-walker, bytecode, and cross-mode parity lane;
- all 86 compiler batches;
- 56.493 seconds for the slowest compiler batch; and
- 16.373 seconds for the canonical compiler outlier.

One initial gate exposed a secondary diagnostic on an intentionally malformed
outer generic application; narrowing validation to well-formed outer arity
removed the cascade. A deliberately tight one-minute aggregate package cap
then expired while `TestBuildEnvFalseAllowsFallbacks` had run for seven
seconds. Three isolated processes completed in 18.31, 17.22, and 16.65
seconds, a 17.393-second mean. The clean full gate passed with no individual
test over one minute.

## Evidence reconciliation

Before review, all 23 performance closures invalidated for the sole reason
`scope-content-drift:v12-spec`. The reviewed spec scope moved from
`b3d251ee1b90d8ddd7d7410a00f9bdf6df99f50cf361262ca1e3eb0d1b85a56c`
to
`e16ca4db0d71270150a6fea676f3b4df0ee8540f49cf2583b76eeff46bdb0b51`.

No closure definition, row identity, benchmark source, non-spec production
scope, disposition, or measurement changed. The final selector has 23 current
closures, zero invalidations, and no compiled or bytecode selection. The
five-node architecture chain is current without decision drift.

No benchmark speedup is claimed: this tranche closes unsupported static
programs and preserves an already-native accepted path.

## Next

Resolve the remaining shared-data race and ownership guidance for v12
concurrency.

Why: `spawn` already maps to Go goroutines and Able nominal values already use
native carriers, but the language does not yet state whether unsynchronized
shared mutation is invalid, undefined, or dynamically checked.

What it entails: inventory captures, mutable nominal carriers, channels,
mutexes, interpreter scheduling, and compiled goroutine lowering; select the
race-free source contract in the spec first; then add only the diagnostics or
guards required by that decision.

Why it matters: a source-level race/ownership rule can preserve direct native
Go references and synchronization without implicit copying, boxing, or a
runtime ownership layer.
