# Monomorphic local-lambda constraint retained

Date: 2026-07-31

Decision: **retain the general typechecker rule**

## Outcome

An unannotated lambda stored in a local binding now acquires one exact,
monomorphic callable signature from complete static contexts. Compatible call
arguments, direct invocations, typed bindings, and function returns all feed
the same constraint. A later incompatible complete signature produces the
coded `callable-signature-mismatch` diagnostic, and strict compilation fails
before generation.

Explicitly dynamic uses do not create a static constraint. Explicit generic
lambdas retain their declared polymorphism. Lexically shadowed bindings solve
independently, and reassignment clears the original lambda provenance.

This closes a semantic hole left after invariant callable signatures became
canonical. The compiler may no longer silently accept incompatible static
uses by preserving an erased `runtime.Value` callable and synthesizing
adapters.

## Retained implementation

The typechecker environment records scoped provenance for fresh, unannotated
local lambdas introduced by `:=` or by an implicit declaration through `=`.
The first complete monomorphic constraint contextually checks the lambda and
records its exact signature. Later constraints must be exactly equivalent
after the existing normalization rules. A true `=` reassignment clears the
old provenance.

Constraint propagation covers:

- a local lambda passed to a complete callable parameter;
- direct invocation when argument and result information completes the
  signature;
- a lambda expression checked by a typed binding;
- a local lambda identifier checked by an explicit return type; and
- the same expected-type flow through blocks and conditional branches.

Typed assignment also exposed a local resolver omission:
`FunctionTypeExpression` was not available as an expected assignment type and
therefore became `Unknown`. Local expression annotations now resolve function
parameters and results structurally against the lexical environment, so
renamed imports retain their canonical types.

No compiler generator, runtime, interpreter, bytecode VM, parser, AST,
external stdlib, dependency, benchmark, frozen workspace, or WASM production
change was needed.

## Guards and fixtures

Focused typechecker guards cover compatible constraints, implicit-`=`
declaration, conflicting constraints with a first-use note, direct invocation,
typed binding, return context, explicit dynamic use, lexical shadowing, and
reassignment. Compiler guards prove compatible native lowering remains
available and both direct and nested conflicting uses fail closed before
generation.

Two fixtures make the contract public:

- `07_14_local_lambda_monomorphic_constraint` accepts compatible uses and
  prints `42`;
- `07_15_local_lambda_conflicting_constraints_diag` requires the exact coded
  conflict diagnostic.

The pre-existing `07_02_lambdas_closures_capture` return-context fixture and
both new fixtures pass in tree-walker, bytecode, and strict compiled modes.

## Strict static census

The final census generated all 66 applications under `--no-fallbacks`, with
zero failures. Every application exactly matches the prior compiled-body
boundary-category map, and all 15 aggregate totals are unchanged:

| Category | Sites |
| --- | ---: |
| `array_runtime_conversion` | 43 |
| `bridge_decode` | 1,230 |
| `bridge_encode` | 3,777 |
| `bridge_error` | 10,213 |
| `bridge_other` | 1,538 |
| `callable_runtime_conversion` | 276 |
| `control_error_conversion` | 24,665 |
| `erased_or_dynamic_call` | 1,859 |
| `interface_runtime_conversion` | 3,316 |
| `native_interface_adapter` | 1,192 |
| `native_union_wrap_or_projection` | 27,485 |
| `primitive_or_any_runtime_conversion` | 7,187 |
| `runtime_value_constructor` | 30 |
| `struct_runtime_conversion` | 3,862 |
| `union_runtime_conversion` | 288 |

All 66 final module hashes also equal the pre-completeness candidate hashes.
The raw 77 MB census was intentionally not retained; the compact JSON record
contains the comparison result and aggregate evidence.

## Repeated A/B and Go context

The baseline disabled the new local-lambda constraint registration while
leaving the preceding retained compiler state intact. Baseline and candidate
generated Go were byte-identical for Binary Event Log, Concurrent Document
Pipeline, and Versioned Telemetry Pipeline. Rebuilding with `-trimpath
-buildvcs=false` produced byte-identical binaries for every pair.

Five balanced, alternating baseline/candidate runs per application all passed
their public verifiers and produced identical output hashes:

| Application | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | --- |
| Binary Event Log | 0.046 s | 0.042 s | byte-identical; timer noise |
| Concurrent Document Pipeline | 0.000 s | 0.000 s | below timer resolution |
| Versioned Telemetry Pipeline | 2.152 s | 2.144 s | byte-identical; timer noise |

This is a correctness and fail-closed boundary tranche, not a valid-program
speedup. The retained five-sample external scorecard remains far from the Go
goal for these applications: Able/Go ratios are approximately 9.28, 10.23,
and 9.95 respectively. The unchanged generated code proves this rule is not
the source of those remaining gaps.

## Evidence state

The performance ledger has 23 current closures, zero invalidations, and no
selected owner. The five-node architecture evidence chain is current.

The complete suite passed in 718.50 seconds at 1,846,056 KiB peak RSS. All 85
bounded compiler batches passed; the slowest took 38.659 seconds, and the
canonical outlier took 15.074 seconds.

Broad verification rejected an initially over-broad resolver change. Resolving
function types in the shared global pattern path degraded a renamed imported
generic callable from its native carrier to `runtime.Value`. Scoping callable
annotation resolution to the lexical local-expression path restored the native
carrier. The next run exposed premature arity checking of generic constructors
such as `Array` and `Map`; the local resolver now uses the existing
constructor-aware mode before applying arguments. The full typechecker matrix,
the imported-generic native-carrier guard, the 66-application census, and the
complete suite all pass after both corrections.

Removed the exact 14,363,772 KiB disk-backed task workspace after retaining
only the compact evidence. No `/tmp/able-*`, `/var/tmp/able-*`, or repository
Python cache remains.

## Recommended next tranche

Refresh interpreter-free CPU and allocation profiles for Binary Event Log,
Concurrent Document Pipeline, and Versioned Telemetry Pipeline, using their
public verifiers and equivalent Go applications. Separate samples in generated
compiled bodies from runtime bridges and compare the generated Go operations
with the reference implementations.

Advance only the largest owner that materially repeats in all three and maps
to a general compiler/runtime lowering rule. Verify any candidate with at
least five balanced baseline/candidate/Go pairs; retain no code if no shared
owner clears the broad bar.

This is next because callable admission is now closed and the three programs'
valid outputs are unchanged, yet their Go gaps remain roughly 9-10x. Fresh
profiles are needed to locate the real shared post-lowering cost rather than
guessing at another boundary or adding a benchmark-shaped rule.
