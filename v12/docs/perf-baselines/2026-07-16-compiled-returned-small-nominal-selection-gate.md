# Compiled returned-small-nominal selection gate

Date: 2026-07-16

## Decision

The selection gate passes and admits a generic compiler candidate, but this
tranche makes no production-code change. Across signed accumulation, unsigned
accumulation, Fixed Width 128, and Rational Series, the hot returned nominal
values have the same non-escaping static consumer shape. The cohort covers
four unlike programs and three nominal definitions, clearing the required
three-program/two-definition threshold.

The admitted design is caller-owned result storage for fresh small nominal
returns, with the ordinary pointer/reference ABI retained. A by-value nominal
carrier is explicitly rejected: Able values have reference semantics, and a
copy could change observable aliasing after a caller mutates either reference.
No `Int128`, `UInt128`, `Rational`, stdlib-package, or benchmark-specific rule
is permitted.

## Exact allocation refresh

Fresh strict no-fallback binaries received one allocation-only phase process
each with `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 55-second guard.
All completed in 1.81-6.05 seconds. Their output hashes match the established
verified outputs.

| Workload | Main bytes | Main allocations | Exact nominal owner |
| --- | ---: | ---: | --- |
| signed accumulation | 16,000,888 | 1,000,022 | 1,000,003 `int128_from_i128` results |
| unsigned accumulation | 16,000,360 | 1,000,013 | 800,000 `uint128_from_u128` plus 200,003 `from_u64` results |
| Fixed Width 128 | 35,536,224 | 2,220,986 | 1,220,963 `uint128_from_u128` plus 1,000,003 `new` results |
| Rational Series | 4,800,272 | 300,016 | 300,001 `rational_build` results |

The Fixed Width and Rational hashes remain respectively
`eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a`
and
`127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c`.

## Generated escape analysis

Each generated module was rebuilt with Go escape diagnostics. The successful
two-field literal, as well as the zero result on an error-control path, escapes
because it is returned as a pointer. The important constructor helpers do not
inline:

- `int128_from_i128`: cost 441 versus Go's budget 80;
- `uint128_from_u128`: cost 210 versus budget 80; and
- `rational_build`: cost 2759 versus budget 80.

The allocation is therefore a concrete compiled-call ABI cost, not remaining
wide-primitive boxing or a generic runtime bridge.

## Reachable consumer census

A temporary Go-AST census started at each generated
`__able_compiled_fn_main`, followed only direct compiled call edges, found all
reachable functions returning `*Int128`, `*UInt128`, or `*Rational`, and
classified every use of the returned temporary.

| Workload / nominal | Return call sites | Compiled-call forwarding | Local binding/rebinding | Tail return |
| --- | ---: | ---: | ---: | ---: |
| signed / `Int128` | 14 | 3 | 6 | 5 |
| unsigned / `UInt128` | 12 | 3 | 6 | 3 |
| Fixed Width / `UInt128` | 11 | 2 | 7 | 2 |
| Rational / `Rational` | 13 | 2 | 5 | 6 |

Across all 50 sites the returned temporary had zero field mutations, identity
comparisons, collection or aggregate stores, environment captures, dynamic
boxing operations, or runtime-interface crossings. Values are forwarded
through arithmetic methods, retained in ordinary local variables, or returned
again. The canonical stdlib implementations also define all three values as
two primitive fields; no hot method mutates its receiver.

This census is an admission proof for the cohort, not permission to change
Able's nominal semantics globally. In particular, other methods such as
`abs`, `min`, and `max` can return `self` or another input. Those must return
the original pointer so later mutation through an alias remains observable.

## Admitted generic boundary

The safe implementation boundary is an internal caller-owned result slot
(`sret`) for statically compiled fresh nominal returns:

1. A static caller provides a distinct addressable slot for a fresh result.
2. Fresh struct literals initialize that slot and return its pointer.
3. A tail call forwards the caller's slot through intermediate functions and
   methods, removing the constructor's forced heap return.
4. A return of `self`, a parameter, a field, collection storage, or any other
   existing object returns that original pointer and does not copy into the
   slot.
5. Dynamic/runtime/interface wrappers keep the current pointer ABI and
   allocate when a result really crosses that boundary.
6. Go escape analysis remains the final authority: if the caller-owned pointer
   later escapes, Go moves the slot to the heap; otherwise it remains stack
   storage.

This preserves object identity while making non-escape—not a nominal name—the
condition for avoiding allocation. Error-control paths must keep their current
control value and may use a zeroed result slot because callers discard the
result whenever control is non-nil.

## Verification and cleanup

- All four strict generated binaries completed under the project process
  limit and reproduced their established hashes.
- Exact allocation profiles reproduce the post-wide-constant totals and
  owners rather than a timing-only inference.
- Generated Go escape analysis succeeds for all four modules.
- The consumer counter was temporary and is not part of the compiler.
- No compiler, runtime, VM, stdlib, spec, application, verifier, reference, or
  benchmark source changed in this selection tranche.

## Next recommendation

Implement a bounded caller-owned nominal-result prototype in the generic
compiled-call pipeline, then evaluate it across this four-program cohort.

Why: this is now the largest concrete allocation owner shared across four
unlike compiled workloads, and the selection evidence meets the required
generality threshold. Caller-owned pointer storage is also the only identified
route that can remove the heap return without replacing Able references with
Go value semantics.

What it entails: add a structural small-result eligibility check; generate an
internal result-slot variant for fresh-literal returns; forward slots through
static tail calls; and keep existing pointer wrappers for dynamic boundaries.
Add arbitrary user-defined two-field structs to the tests so the rule cannot
key on the three stdlib names. Guard returning `self`, observable alias
mutation, collection escape, dynamic interface calls, and raised control.
Then run repeated alternating binaries for all four workloads plus
small-integer, Binary Trees, startup, interface, and error controls. Retain the
candidate only if allocation removal produces broad wall-time gains without a
guard regression.
