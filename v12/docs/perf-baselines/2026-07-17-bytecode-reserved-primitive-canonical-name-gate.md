# Bytecode reserved-primitive canonical-name gate — 2026-07-17

## Decision

Keep the generic reserved-scalar fast path in `canonicalTypeName(...)`.
Lowercase scalar primitive type names now return immediately instead of walking
the runtime environment to discover whether the name denotes a nominal type.
The v12 language contract reserves those lowercase names for scalar
primitives, so a runtime value binding cannot change their type identity.

Nominal names, imported definitions, interfaces, unions, aliases, `String`,
and every PascalCase built-in retain ordinary environment resolution. The
change does not recognize an application, benchmark, stdlib container, or
user-defined nominal type. No compiler, stdlib, fixture, or language source
changed.

## Profile admission

Fresh retained-baseline profiles used one test binary, canonical external
`able-stdlib`, CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, skipped
benchmark typechecking, and a separate bounded process per program.

| Workload | Iterations | ns/op | CPU samples | String-map cumulative |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 3 | 675,524,497 | 2.81 s | 8.54% |
| Unicode Scalar Pipeline | 1 | 4,659,914,774 | 4.80 s | 7.29% |
| Run-length encode | 1 | 1,423,674,882 | 1.59 s | 6.29% |
| String Split/Join | 1 | 1,201,987,891 | 1.49 s | 7.38% |
| Iterator Collect | 3 | 531,349,840 | 2.35 s | 7.23% |
| Numeric Array Map | 20 | 103,597,778 | 2.38 s | 7.56% |

Direct-parent attribution confirmed that `runtime.mapaccess2_faststr` remained
an aggregate. `lookupIntegerInfo(...)` repeated in all six but its fixed
full-metadata and membership-switch replacements have already failed broad
application gates and were not retried. Environment lookup beneath
`canonicalTypeName(...)` was the new material repeated boundary: 4.63%
cumulative in Boolean, 5.83% in Unicode, 3.14% in Run-length, 1.28% in
Iterator, 1.68% in Array Map, and 0.67% in Split/Join.

## Candidate

`isReservedScalarPrimitiveTypeName(...)` recognizes the closed scalar set:
bool, char, nil, void, signed and unsigned integer widths (including the
runtime-supported pointer widths), and f32/f64. `canonicalTypeName(...)`
returns these names before `Environment.Lookup(...)`.

The guard is deliberately local to canonical type-name resolution. It does
not replace integer metadata lookup, general primitive membership checks, or
nominal type matching, avoiding the application regressions from the earlier
global switch experiments. Tests prove that every reserved scalar stays
canonical even if a runtime environment is constructed manually with a
conflicting value binding, while an ordinary nominal binding still resolves
to its definition identity.

## Repeated workstation gate

Every timing is an independent process. Pairs alternate baseline-first and
candidate-first order under the same profile environment, and every valid
workstation outlier remains in the arithmetic mean. All programs were expanded
to five pairs; Numeric Array Map was expanded to ten because its first three
pairs appeared 4.05% slower.

| Workload | Pairs | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 5 | 596.387 ms | 530.821 ms | 10.99% faster |
| Unicode Scalar Pipeline | 5 | 4.51508 s | 4.51524 s | 0.004% slower; neutral |
| Run-length encode | 5 | 1.52069 s | 1.42910 s | 6.02% faster |
| String Split/Join | 5 | 1.20580 s | 1.16796 s | 3.14% faster |
| Iterator Collect | 5 | 532.207 ms | 528.612 ms | 0.68% faster |
| Numeric Array Map | 10 | 108.071 ms | 106.789 ms | 1.19% faster |

The Boolean mean retains its 732.409 ms baseline outlier. Unicode retains the
4.713 s candidate and 4.807 s baseline processes. Array Map retains the full
82.8–132.7 ms spread on both sides; the expanded mean reverses its misleading
initial three-pair result. No workload crosses the broad regression guard.

Allocation spot checks do not support an allocation claim, as expected for a
lookup-only guard. Iterator and Array Map allocation counts match exactly at
214,686 and 14,646 per operation. Boolean differs by two setup-sensitive
allocations in one one-iteration process.

## Mechanism and correctness

Candidate profiles reduce cumulative `canonicalTypeName(...)` work from 4.63%
to 0.97% in Boolean and from 5.83% to 3.23% in Unicode; the coarse Split/Join
and Array Map profiles no longer sample the helper. Run-length is unchanged at
about 3.2% and Iterator sampling is mixed, so the wall-time result—not the
coarse profile delta—is the retention gate.

Focused alias, canonicalization, integer, numeric, cast, type-match,
type-coercion, named-struct, primitive-kernel, bytecode VM, runtime, and
fixture-parity tests pass. Both changed files remain below 1,000 lines and
`git diff --check` passes. Temporary test binaries, Unicode symlink, and CPU
profiles were removed after this record was written.

## Next recommendation

Reconcile the residual typed-pattern execution family across Unicode,
Run-length, Split/Join, Iterator Collect, and Numeric Array Map, centered on
`execJumpIfNotTypedPattern(...)` and `matchesTypeWithoutRuntimeValue(...)`.

Why: the reserved primitive environment lookup is now removed, the shared
integer metadata representation is already closed by two failed broad gates,
and the remaining string-map owners split among type matching and unrelated
caches. Typed-pattern execution is the remaining concrete semantic owner in
at least five unlike programs.

What it entails: collect post-change call trees and temporary category counts
for simple primitive, exact nominal, interface, union, generic, and miss paths;
separate compile-time simple-type metadata hits from dynamic fallback; and
admit a candidate only if the same category is material in at least three
programs. Any candidate must preserve aliases, coercion, interface/union
semantics, dynamic definitions, both Go interpreters, and the same repeated
workstation controls. Continue to defer WASM.
