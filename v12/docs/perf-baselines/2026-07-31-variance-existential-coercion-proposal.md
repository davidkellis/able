# Variance and existential nominal-coercion proposal

Date: 2026-07-31

## Decision

**Recommend option A: invariant-only v12, with explicit variance declarations
deferred. No production behavior is changed by this tranche.**

Top-level numeric widening, union injection, and concrete-to-interface upcasts
remain valid. None of those conversions recursively implies `F A -> F B`.
Function types require equivalent complete signatures.

The full rationale and implementation contract are in
`v12/design/variance-existential-coercion-proposal.md`.

## Canonical rule

Sections 4.1.6 and 4.1.7 already say:

- Able has no general structural-subtyping lattice;
- type parameters are invariant unless explicitly declared otherwise;
- no variance declaration syntax is currently defined; and
- no implicit subtyping-based coercion exists.

The AST `GenericParameter` has no variance field, and the parser has no
variance production. Therefore every present v12 type parameter is invariant.

## Active-source audit

The audit parsed every Able file under `v12/fixtures`, `v12/examples`,
`../able-stdlib/src`, and `../able-stdlib/tests`.

| Measure | Result |
| --- | ---: |
| Able files parsed | 829 |
| Parse failures | 0 |
| Generic struct declarations | 74 |
| Generic union declarations | 8 |
| Files mentioning `Array` | 204 |
| Files mentioning `Map` or `HashMap` | 33 |
| Files mentioning `Future` | 6 |
| Variance declarations/keywords | 0 |

## Reproduced matrix

Temporary sources outside the repository exercised the current boundary:

| Case | Checker | Tree-walker | Bytecode | strict compiled |
| --- | --- | --- | --- | --- |
| `Box i8 -> Box i32` | accepts | `7` | `7` | `7` |
| mutable `Array i8 -> Array i32` | accepts | prints `300` | prints `300` | prints `7` |
| `Future i8 -> Future i32` parameter | accepts | accepts | accepts | accepts |
| `(i8 -> i8) -> (i32 -> i32)`, input `300` | accepts | parameter mismatch | return mismatch | 8-bit overflow |
| `Box Cat -> Box Display` | rejects | not run | not run | low-level compiler warns, builds, prints `Miso` |
| `Box Cat as Box Display` | rejects | not run | not run | low-level compiler warns, builds, prints `Miso` |
| `Producer Cat -> Producer Display` | rejects | not run | not run | low-level compiler warns, builds, runtime type mismatch |
| rebuild `Box Display` from `Cat` field | accepts | `Miso` | `Miso` | `Miso` |

All strict test builds used `-no-fallbacks`; their dependency graphs omit
`pkg/interpreter` and retain `pkg/runtime`.

The Array case is decisive. Both interpreters mutate the original object,
violating its declared element type. Generated Go converts through
`runtime.Value` into a new native `Array i32`, changing aliasing and therefore
program output. The callable case similarly creates a runtime adapter and
defers a statically knowable incompatibility.

The ordinary `able` CLI stops on the rejected nested-existential diagnostics.
Direct `ablec` currently treats most checker errors as warnings; the future
selected implementation must give invariant/callable mismatches coded fatal
diagnostics so strict generation fails closed.

## Options

### A. Invariant-only v12 — recommended

- exact recursively normalized generic arguments;
- exact callable signatures;
- top-level conversions remain local to the converted value;
- explicit constructors perform any element-wise copying or mapping; and
- no runtime carrier is introduced to satisfy a static mismatch.

### B. Declared variance now — defer

Requires parser/AST syntax, polarity validation, separate-compilation facts,
mutable-field rules, and a representation-preserving runtime view. Go slices
and instantiated nominal structs cannot become covariant merely through a
checker rule.

### C. Conversion on use — reject

Already produces observable interpreter/compiler disagreement and hidden
runtime adapters.

### D. Recursive existential lifting — reject

Requires undeclared covariance or recursive boxing and changes nominal
identity, aliasing, layout, and allocation.

## Verification

Passed:

- canonical spec and implementation audit;
- parser/AST variance-surface audit;
- 829-file active-source parse;
- checker probes for numeric generic, Array, Future, callable, nested
  existential, explicit cast, and generic-interface arguments;
- tree-walker and bytecode executions for every checker-admitted probe;
- strict fallback-free compiled builds and dependency inspection; and
- positive element-wise existential reconstruction in all modes;
- focused recursive-`Self` fixture parity after correcting its omitted checked
  diagnostic in `typecheck-baseline.json`; and
- the complete v12 release suite in 654.74 seconds at 1,374,380 KiB peak RSS.
  Every bounded compiler batch passed; the slowest took 42.778 seconds and the
  isolated canonical-stdlib outlier took 14.717 seconds;
- the performance evidence ledger with 23 current closures, zero
  invalidations, and an empty mode-aware selector; and
- the current five-node architecture evidence chain.

The first complete-suite attempt exposed the missing expected diagnostic for
the already-retained `interface_static_only_recursive_self_call` fixture. Its
source and behavior were present from the preceding recursive-`Self` tranche;
this tranche corrected only the checked baseline metadata, then reran the
focused case and complete suite successfully.

Removed the exact 3,010,692 KiB disk-backed task workspace after recording the
results. No `/tmp/able-*` or `/var/tmp/able-*` directory remains.

No production source, specification semantics, parser, AST, checker,
interpreter, bytecode VM, compiler, runtime, stdlib, dependency, new fixture
source or behavior, benchmark, frozen workspace, or WASM behavior changed.

## Next

Obtain a maintainer decision on option A before implementation.

Why: the canonical text already points to invariance, but enforcing it changes
which currently accepted programs receive diagnostics.

What it entails after selection: add canonical recursive type equivalence and
complete callable comparison, issue coded diagnostics, make strict compilation
fail closed, add cross-runtime fixtures, run the complete suite and 66-program
static census, and reconcile only evidence scopes actually invalidated.

Why it matters: invalid cross-instantiation calls must stop before they can
change aliasing or force `runtime.Value` adaptation. Valid compiled programs
then keep exact native Go carriers from source to generated execution.
