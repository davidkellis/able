# Fixed-integer alternative arithmetic retained

Date: 2026-07-31

## Decision

Retain one reserved primitive method family for every fixed-width Able integer:

- `wrapping_add`, `wrapping_sub`, and `wrapping_mul` return the same integer
  type with two's-complement/modulo wrapping;
- `saturating_add`, `saturating_sub`, and `saturating_mul` return the same
  integer type clamped to its minimum or maximum; and
- `checked_add`, `checked_sub`, and `checked_mul` return `?T`, using `nil` for
  overflow.

The receiver and right operand must have exactly the same fixed-width type.
Operands evaluate left-to-right exactly once. These methods are reserved
primitive behavior and cannot be replaced by user method sets. Ordinary
arithmetic operators keep their existing raising overflow semantics.

The contract applies to `i8`, `i16`, `i32`, `i64`, `i128`, `u8`, `u16`,
`u32`, `u64`, and `u128`; it does not apply to arbitrary-precision integers,
floats, or non-primitive nominal types.

## General implementation

The tree-walker and bytecode-visible paths share fixed-integer arithmetic in
`pkg/runtime/fixed_integer_arithmetic.go`. Native `i128` and `u128` carriers
use shared wide-integer methods. The typechecker synthesizes the reserved
signatures before ordinary method lookup, and the interpreter dispatches the
same reserved primitive surface.

The compiler lowers calls directly:

- widths through 64 bits use native Go operators/casts for wrapping and
  generated checked/saturating helpers;
- `i128` and `u128` use the native `runtime.Int128` and `runtime.Uint128`
  carriers; and
- checked success/failure stays in the native `__able_nullable[T]` carrier.

Strict generated fixture controls contain no `runtime.Value`, member-call
bridge, or interpreter path in the arithmetic bodies. No named-container,
non-primitive nominal, benchmark, application, or stdlib rule was added.

## Compiled generic-boundary correction

The explicit compiled CLI lane exposed an independent general correctness
gap while verifying the new primitive surface. A successful native
`BigInt.to_i32()` result reached `EqMatcher<Result<i32>>`, but the generated
runtime representation of that specialized generic struct omitted its
concrete type arguments. Dynamic `Matcher<Result<i32>>` recovery therefore
rejected the otherwise correct native value.

Generated specialized generic structs now carry their concrete
`[]ast.TypeExpression` arguments whenever their native nominal carrier is
encoded as a runtime struct instance. In-place boundary updates copy those
arguments as well. This is one shared nominal boundary rule; the correction
does not special-case `BigInt`, `Result`, `EqMatcher`, or any container.

The primitive remains native through `BigInt`, `i64`, `i32`, and
`Result<i32>`. Type metadata is attached only when the enclosing generic
nominal crosses the required runtime-visible interface boundary.

## Coverage and verification

The new `06_03_fixed_integer_alternative_arithmetic` fixture covers all ten
fixed-width types, every operation in all three modes, and checked success.
It passes in tree-walker, bytecode, normal compiled, and strict
no-bootstrap/no-fallback modes.

Focused runtime, typechecker, interpreter, compiler, cache/registration, and
generated-source guards pass. The specialized-generic boundary regression
pins concrete runtime type arguments and their preservation during apply.

The complete default v12 release gate passes:

- 282 seeded fixture rows and 283 fixture directories;
- all eight tree-walker, parity, and bytecode fixture shards;
- all 86 compiler batches;
- 42.032 seconds for the slowest compiler batch; and
- 16.556 seconds for the canonical compiler outlier.

Canonical stdlib passes in tree-walker (22 seconds) and bytecode (17 seconds).
The five previously failing compiled canonical numeric suites pass:

| Suite | Time |
| --- | ---: |
| BigInt | 37.43 s |
| BigUInt | 36.84 s |
| Int128 | 32.79 s |
| UInt128 | 32.75 s |
| Rational | 39.85 s |

The full generated-Go CLI release lane passes in 990.043 seconds. Each named
test remains below one minute; the aggregate is long because the canonical
packages compile and run sequentially.

No canonical stdlib, dependency, benchmark, frozen workspace, or WASM change
was needed.

The exact 12,487,148 KiB owned task workspace, generated probe, and caches
were removed after verification. No Able task directory remains under `/tmp`
or `/var/tmp`.

## Performance evidence

The final compiler change correctly invalidated 12 compiled/cross-family
closures. The exact pre-reconciliation selector is preserved in
`2026-07-31-fixed-integer-alternative-arithmetic-pre-reconciliation-invalidation.{json,md}`.
After correctness review and source reconciliation, the performance ledger
contains 23 current closures, zero invalidations, and no selected owner. The
five-node architecture evidence chain is current.

No execution speedup is claimed: the alternative API was not expressible
before this tranche, and the compiled canonical failures had no valid
baseline runtime. The retained result is a language/correctness/native-lowering
closure, not a benchmark optimization.

## Next

Resolve the remaining full-inference and higher-kinded-type limits in the v12
spec.

Why: v12 supports interface-level type-constructor patterns, but still leaves
the boundary for first-class or more general higher-kinded inference open.
That ambiguity prevents one authoritative decision about when generic code is
statically monomorphizable and when it must fail rather than introduce an
implicit dynamic carrier.

What it entails: inventory the existing grammar, AST, typechecker, compiler,
fixtures, and canonical stdlib HKT surface; state the supported and rejected
forms in the spec first; then implement only gaps required by that selected
contract, with diagnostic and strict native-lowering guards.

Why it matters: an exact inference/HKT boundary keeps generic nominal values
on deterministic native Go carriers, minimizes compiled/interpreter
crossings, and prevents accidental boxing from becoming language behavior.
