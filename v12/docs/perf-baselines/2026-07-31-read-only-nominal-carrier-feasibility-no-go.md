# Read-only nominal-carrier feasibility gate

Date: 2026-07-31

## Decision

Retain the fail-closed census tooling, but retain no compiler or generated
runtime carrier change.

One conservative whole-program rule does not admit all three profiled
applications. `EventRecord` and `Sample` have no reachable field mutation,
identity comparison, or runtime conversion, but each flows through a generated
callable whose non-mutation and non-retention effects are not represented in
the compiler. `DocumentTask` and `DocumentScore` have no reachable field
mutation either, but both cross the runtime-backed Channel ABI through
`runtime.Value`.

Native unions and specialized generic storage are not themselves blockers.
The missing facts are callable effects and an identity-preserving runtime
service adapter. An unconditional Go value carrier would still violate Able's
mutable, alias-visible struct semantics.

No generated-only ceiling is retained. Each ceiling required a different
application lifetime assumption, so their material gains do not constitute
one general lowering rule.

## Conservative proof

The generated boundary census now records a nominal feasibility proof for each
generated whole program. It starts at `__able_compiled_fn_main`, follows local
generated callees, and fails closed unless a constructed nominal has:

- no resolved or unresolved reachable field mutation;
- no pointer identity comparison;
- no runtime, host, native-interface, or opaque field exposure;
- only flat primitive fields; and
- no unresolved mutation-capable call.

Native union wrap/projection and specialized native pointer storage are
recorded as allowed contexts. This diagnostic is deliberately not a compiler
authorization fact: any production rule must derive equivalent facts from
typed Able IR and preserve them through callable and runtime-service adapters.

The analyzer's focused test contains one admitted flat record and rejects
separate mutation, runtime conversion, pointer comparison, unresolved
callable, opaque-handle, and non-primitive-field cases.

## Full strict and stdlib census

All 66 selected applications generated under `--no-fallbacks`; none failed.
The census covered 368 unique declared generated nominal names and 1,103
reachable constructed instantiations spanning 261 unique names.

| Result | Instantiations | Applications |
| --- | ---: | ---: |
| Conservative eligibility | 56 | 30 |
| Runtime/host identity exposure | 900 | 52 |
| Opaque field boundary | 417 | 60 |
| Non-primitive field carrier | 320 | 45 |
| Native-interface identity exposure | 58 | 28 |
| Unresolved field mutation | 58 | 12 |
| Unknown mutation-capable call | 48 | 23 |
| Reachable field mutation | 25 | 16 |

Blocker counts overlap. Passing this feasibility census is not authorization
to change a representation; it identifies cases worth a typed semantic proof.

The separate canonical stdlib inventory covers 70 Able files, 19,090 lines,
258 struct declarations, and 240 syntactic member-assignment sites. Its source
tree identity is
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
The external stdlib was already dirty and remained unchanged by this tranche.
Instantiated stdlib code is also included in every generated whole-program
proof.

## Profiled records

| Application / nominal | Mutation | Allowed native context | Blocking fact |
| --- | ---: | --- | --- |
| Binary Event Log / `EventRecord` | 0 | 3 union sites | one unresolved mutation-capable callable |
| Concurrent Document / `DocumentTask` | 0 | none | 3 runtime conversions and 3 opaque boundary sites |
| Concurrent Document / `DocumentScore` | 0 | none | 3 runtime conversions and 3 opaque boundary sites |
| Versioned Telemetry / `Sample` | 0 | 4 union sites, 3 specialized-storage sites | one unresolved mutation-capable callable |

This confirms the source-level observation that the records are immutable by
use, while also identifying why the compiler cannot yet treat that as a
representation fact.

## Generated-only ceilings

Each temporary candidate changed only emitted Go, passed the public verifier,
and was removed after recording compact evidence.

| Application | Generated-only mechanism | Bytes | Objects | Time | Candidate / Go time |
| --- | --- | ---: | ---: | ---: | ---: |
| Binary Event Log | one audited reusable parse-result slot and immutable default | 9,321,784 -> 6,765,915 (-27.42%) | 171,069.6 -> 117,822.6 (-31.13%) | 49.592 ms -> 47.563 ms (-4.09%) | 7.64x |
| Concurrent Document | application-only native typed Go channels | 648,238 -> 150,776 (-76.74%) | 10,479.4 -> 2,227.4 (-78.74%) | 5.518 ms -> 4.262 ms (-22.77%) | 1.92x |
| Versioned Telemetry | local `Sample`, then audited in-place ring-slot replacement | 430,788,163 -> 8,102,358 (-98.12%) | 13,325,303 -> 116,424.8 (-99.13%) | 1.996 s -> 1.502 s (-24.75%) | 6.89x |

Binary and Concurrent use 25 high-resolution launches per baseline, candidate,
and Go reference, arranged as five alternating cohorts. Telemetry uses five
balanced baseline, candidate, and Go processes because each default run is
already sustained. Allocation rows use five lightweight main-phase samples
per configuration. Every process remained below one minute and all public
verifiers passed.

The ceilings establish materiality but not generality:

- Binary reuses one identity only because the audited callbacks do not retain
  it across iterations.
- Concurrent bypasses the canonical Channel runtime service and therefore does
  not establish its cancellation, close, and dynamic interoperability
  semantics.
- Telemetry mutates an existing window identity only after snapshot-dependent
  scoring, which is safe for this exact lifetime but invalid when an earlier
  alias remains live.

The candidates remain 7.64x, 1.92x, and 6.89x slower than their Go references.
Nominal allocation is therefore important but not the only remaining compiled
cost.

## Verification

- Focused generated-boundary census tests pass.
- Existing caller-owned alias, escaped-array, retained-old-result,
  callee-capture, and conditional-candidate compiler guards pass.
- The static boundary census content-address and contract tests pass.
- All 66 strict applications generated successfully for the new census.
- No external stdlib, runtime, interpreter, VM, parser, AST, dependency,
  benchmark source, frozen workspace, or WASM change was made.
- Removed the exact 1,443,920 KiB disk-backed task workspace; no RAM-backed
  `/tmp/able-*` directory was used or left behind.

## Next recommendation

Add typed nominal mutation, capture, and callable-effect summaries before
attempting another carrier rule.

Why: unresolved mutation-capable callable flow blocks both `EventRecord` and
`Sample`, recurs across 48 constructed instantiations in 23 strict
applications, and is the same missing fact that closed earlier loop-carried
caller-owned reuse work.

What it entails: compute a conservative fixed point over functions, methods,
and statically resolved lambdas. Summaries must record nominal parameter field
writes, capture or storage, return aliasing, unknown or dynamic calls, and
propagate through direct and monomorphic callable edges. Reuse the existing
alias and conditional-retention guards, add closure and interface negatives,
and rerun all 66 strict applications before considering a carrier change.

Why it matters: this turns immutable-by-use from a benchmark observation into
a typed compiler fact. It is prerequisite infrastructure for sound native Go
value carriers and other allocation removal while preserving Able reference
identity.
