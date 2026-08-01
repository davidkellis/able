# Recursive `Self` contract retained and performance evidence reconciled

Date: 2026-07-31

## Decision

**Retain option A and atomically rebaseline the reviewed v12-spec and
compiler-production scopes.**

Recursive `Self` substitution is now structural for concrete and statically
constrained receivers. Interface-value calls are admitted per method: the
receiver and an exact top-level `Self` result remain dynamically safe, while
every other non-receiver `Self` occurrence is static-only. Recursive
`Self`-dependent constraints use an anchored finite proof graph.

This closes a correctness gap. It is not a performance optimization and it
does not change generated or executed code for any valid benchmark
application.

## Compatibility and execution proof

The proposal audit parsed all 760 active fixture, example, and canonical
stdlib Able files. Its only ordinary nested result was
`Integral.div_mod -> DivModResult Self`, with no existential call site.

The retained positive fixture proves three distinct paths:

- concrete `Pair Self` substitution;
- generic `T: DuplicatePair` substitution; and
- exact top-level `Self` reconstruction through an interface dictionary.

Tree-walker, bytecode, and strict fallback-free compiled execution all produce
the same verified output. The strict graph omits `pkg/interpreter`.

The canonical stdlib guard invokes `Integral.div_mod` both directly on `i32`
and through `T: Integral`. Both calls produce `3:2` in tree-walker, bytecode,
and strict compiled modes; strict generation needs no fallback and its graph
omits `pkg/interpreter`.

The negative fixture rejects a dynamic `Pair Self` result at the member access
with the coded `static-only-interface-method` diagnostic. The compiler treats
that diagnostic as fatal before generation rather than allowing a generated
return conversion to fail at runtime.

## Performance-scope reconciliation

The evidence selector correctly invalidated all 23 closures:

- all 23 named `scope-content-drift:v12-spec`; and
- the 12 compiled/cross-family closures also named
  `scope-content-drift:compiler-production`.

The invalidation report is
`2026-07-31-recursive-self-contract-pre-reconciliation-invalidation.json`.

A fresh bounded census generated all 66 selected compiled applications with
`--no-fallbacks`:

| Measure | Result |
| --- | ---: |
| Successful / failed | 66 / 0 |
| Generation minimum / maximum | 0.312 / 11.366 seconds |
| Generation arithmetic mean | 3.520 seconds |
| Generated module hashes matching the retained authority | 66 / 66 |
| Aggregate boundary categories matching the retained authority | 15 / 15 |

The comparison authority is
`2026-07-30-interface-dictionary-capture-strict-static-census.json`
(SHA-256
`ba115b6965135d45494090d8c7bba51c2cf076aba8f731579c0334bc11daf0b4`).
Exact equality of every generated module hash proves that the new compiler
branch does not alter valid generated output. Exact equality of all aggregate
boundary categories proves that it does not create or remove a typed/runtime,
interface, union, callable, control, or dynamic boundary.

Because the shared scopes changed deliberately and all affected valid
execution was reviewed, the closure ledger is rebaselined atomically. No prior
timing result is relabeled, no closure disposition changes, and no
optimization candidate is admitted.

## Verification

Passed:

- recursive dynamic-safety classification for parameter, nullable, nominal,
  union, callable, higher-kinded, and obligation occurrences;
- exact-result and type-namespace preservation guards;
- anchored direct and mutual recursive constraints plus growing-cycle
  rejection;
- complete typechecker package;
- all tree-walker/bytecode execution fixtures and parity;
- focused compiler fail-closed and interface-dispatch guards;
- the two compiler audit rows that were active at the monolithic package
  timeout;
- `go test ./cmd/ablec`; and
- canonical fixture export consistency; and
- the complete bounded compiler release lane: core batches, outliers,
  fallback audit, full compiled fixture matrix, strict-dispatch audit,
  interface-lookup audit, and boundary audit.

The bounded compiler release lane completed in 5,161.91 seconds at
2,791,020 KiB peak RSS. Every individual test process remained below the
two-minute harness timeout.

A direct monolithic `go test ./pkg/compiler` invocation hit its 15-minute
aggregate timeout while existing parallel audit rows were still building. The
two named rows passed independently in 4.900 seconds. The repository's bounded
release compiler lane is the authoritative complete-suite gate and passed.

After recording the results, the exact 26,284,576 KiB disk-backed task
workspace and 22,900 KiB RAM-backed extern cache were removed. No
`/tmp/able-*` or `/var/tmp/able-*` path remains.

## No production expansion

No parser, AST, interpreter execution, bytecode VM, runtime, canonical stdlib,
dependency, benchmark, frozen workspace, or WASM change was required. No
recursive boxing, variance rule, named-container rule, or non-primitive
nominal lowering branch was introduced.

## Next

Use the reconciled mode-aware selector as the admission gate. If it selects
nothing, choose the next specified correctness/language completeness decision
instead of repeating a closed optimization. If a later checked change
invalidates a closure, refresh only the named closure evidence and select one
exact CPU/allocation owner shared by at least three unlike applications before
prototyping.

Why: this tranche deliberately leaves valid generated execution byte-for-byte
unchanged, so it creates no new performance owner by itself.

What it entails: keep the evidence ledger and architecture chain green; when a
real trigger appears, use repeated verifier-backed measurements and equivalent
Go/Python/Ruby references, preserve native primitive and static Array
carriers, and reject benchmark-, container-, or nominal-specific rules.

Why it matters: performance work remains aimed at eliminating genuine
compiled/interpreted or boxed boundaries rather than manufacturing a candidate
from a correctness-only scope change.
