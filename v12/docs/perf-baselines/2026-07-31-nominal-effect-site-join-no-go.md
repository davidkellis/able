# Nominal effect/generated-site join and ownership no-go

Date: 2026-07-31

## Decision

Retain the diagnostic join from generated opaque-call sites to typed nominal
callable effects. Retain no compiler carrier, caller-owned storage, generated
runtime, runtime, interpreter, VM, stdlib, language, dependency, benchmark, or
WASM execution change.

The five newly clear nominal rows now have exact generated caller, callee,
argument-index, and source locations. Graph Visitors `VisitState`, Packet
Codecs `CursorState`, and Tree Folds `FoldState` resolve to complete native
interface implementation candidate sets, and every candidate parameter is
typed read-only/non-escaping. The two `EventRecord` rows resolve to indirect
lambda candidate sets.

The three state rows share one material allocation mechanism: loop-carried
flat nominal state enters native interface dispatch and a fresh successor
returns directly or inside a fresh step record. A generated-only ceiling that
overwrote the consumed receiver removed material allocation and improved all
three applications.

That ceiling is not a legal production rule. It merges the old and new Able
struct identities. Callable non-mutation and noncapture do not prove that the
caller has no retained alias or later read of the old value. The existing
retained-old, callee-capture, and conditional-candidate guards demonstrate the
missing semantic requirement. Production admission therefore remains closed
until an interprocedural caller alias, liveness, and ownership-transfer proof
covers direct calls, native interface candidate sets, and embedded results.

## Retained diagnostic link

The compiler's opt-in callable-effect summary now exposes the stable generated
Go name when one exists. The generated boundary census schema is version 4
and records each formerly count-only unknown nominal call with:

- generated caller and callee;
- zero-based generated argument indexes carrying the nominal;
- generated filename, line, and column; and
- typed callable candidates with package, kind, generated Go name, parameter
  index/name, effect blockers, and read-only/non-escaping status.

Resolution is fail-closed:

- `exact-generated-target` requires the emitted target and parameter index to
  match;
- `interface-method-candidate-set` joins the generated selector and
  receiver-adjusted parameter index to every typed implementation;
- `indirect-callable-candidate-set` reports the conservative typed lambda set;
  and
- unresolved runtime, host, helper, or opaque callable sites remain
  `unresolved`.

The full census driver supplies the effect report to the analyzer. Focused
tests cover source-site recording, interface candidate sets, indirect
candidate sets, generated-name publication, and the existing effect
fixed-point positives and negatives.

## Full strict census

All 66 selected compiled applications generated under `--no-fallbacks` with
zero failures. The joined census contains 125 generated unknown-call sites
across 23 applications:

| Resolution | Sites | All candidates read-only/non-escaping |
| --- | ---: | ---: |
| Interface method candidate set | 15 | 8 |
| Indirect callable candidate set | 7 | 6 |
| Unresolved | 103 | not applicable |

The large unresolved remainder is mainly runtime/error helpers, callable
conversion wrappers, host numeric methods, and other sites for which a typed
Able callable identity is not available. The join does not reinterpret those
sites as safe.

The five targeted rows resolve as follows:

| Application / nominal | Generated call | Typed candidates |
| --- | --- | --- |
| Binary Event Log / `EventRecord` | specialized `Result.map` indirect call | two read-only lambdas |
| Concurrent Event Routing / `EventRecord` | specialized `Result.map` indirect call | one read-only lambda |
| Concurrent Graph Visitors / `VisitState` | `visitor.__able_ctx_inspect`, argument 1 | two read-only `GraphVisitor.inspect` implementations |
| Concurrent Packet Codecs / `CursorState` | `codec.__able_ctx_decode`, argument 1 | two read-only `PacketCodec.decode` implementations |
| Concurrent Tree Folds / `FoldState` | `algebra.__able_ctx_combine`, argument 3 | two read-only `FoldAlgebra.combine` implementations |

The generated feasibility counts remain conservative. The old
`unknown-mutation-capable-call` blocker is retained in the generated proof;
the typed join is a separate evidence layer and does not silently convert a
candidate set into carrier authorization.

## Shared owner and generated-only ceiling

The three interface applications use flat primitive-only functional state:

- Graph returns the successor `VisitState` directly from `inspect`;
- Packet embeds the successor `CursorState` in `DecodeStep`; and
- Tree embeds the successor `FoldState` in `FoldStep`.

Go escape diagnostics confirm that each successful inherent state-update
literal escapes to the heap. Three baseline and three candidate lightweight
main-phase allocation runs per application all passed their public verifiers:

| Application | Baseline bytes / objects | Ceiling bytes / objects | Change |
| --- | ---: | ---: | ---: |
| Graph Visitors | 1,202,101 / 46,907 | 803,864 / 38,698 | -33.13% bytes; -17.50% objects |
| Packet Codecs | 2,407,093 / 49,405 | 1,883,125 / 33,023 | -21.77% bytes; -33.16% objects |
| Tree Folds | 929,960 / 24,779 | 537,133 / 16,592 | -42.24% bytes; -33.04% objects |

The temporary ceiling changed only the generated successful state-update
return from a fresh pointer to an overwrite of `self`, then returned `self`.
It was never applied to repository compiler generation.

Ten balanced baseline/candidate/Go cohorts per application produced 90
verified processes:

| Application | Baseline | Ceiling | Go | Ceiling change | Ceiling / Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| Graph Visitors | 0.006963 s | 0.006504 s | 0.002203 s | -6.60% | 2.953x |
| Packet Codecs | 0.008537 s | 0.007930 s | 0.002275 s | -7.11% | 3.485x |
| Tree Folds | 0.006901 s | 0.005612 s | 0.002017 s | -18.68% | 2.783x |

These very short launch measurements are naturally noisy, but all three
directions agree with the exact allocation reductions. They establish
materiality, not semantic legality.

## Semantic rejection

Able structs retain mutable reference semantics. Replacing:

```text
new = update(old)
```

with an in-place overwrite of `old` is observable when any alias to `old`
remains live, even if the update callable neither stores nor returns the input
and even if no pointer equality is performed. A later field read through the
old alias is sufficient.

The current typed effect analysis proves callee behavior. It does not prove:

- that the caller has no alias to the argument;
- that no old-value read occurs after the call;
- that an interface call transfers ownership for every implementation;
- that a successor embedded in another result is the only surviving identity;
  or
- that conditional candidate selection cannot retain multiple generations.

The generated ceiling was audited only for these three exact applications.
Retaining it would therefore be benchmark/lifetime-specific and would violate
the project's general-rule requirement.

## Verification and evidence reconciliation

- `go test ./cmd/able-generated-boundary-census`
- focused nominal-effect and caller-owned alias/capture compiler guards
- `go test ./cmd/ablec`
- 66/66 strict census generation
- 18/18 allocation-ceiling processes verified
- 90/90 balanced timing processes verified
- the three strict modules each contain 96 dependencies and omit
  `able/interpreter-go/pkg/interpreter`
- performance evidence ledger: 23 current closures, zero invalidations
- architecture evidence chain: five current nodes

No observed individual test exceeded one minute. All large generated modules,
binaries, caches, profiles, and disposable reports lived under disk-backed
`/var/tmp`.

The compact machine-readable companion is
`2026-07-31-nominal-effect-site-join-no-go.json`.

## Next recommendation

Add an opt-in interprocedural ownership-transfer analysis; do not select a
carrier or reuse storage yet.

Why: the shared allocation owner is now proven material in three unlike
applications, and the only remaining semantic blocker is caller-side
alias/liveness rather than another opaque callee effect.

What it entails: track nominal binding provenance and aliases; model uses
before and after calls; require unconditional consumption and replacement;
propagate ownership through every native interface implementation and through
fresh result fields; and fail closed for capture, storage, return aliases,
dynamic calls, conditional candidates, and retained old generations. Add
positive direct/interface/embedded-result cases plus the existing retained-old
and capture negatives. Run the 66-program census before enabling any
production path.

Why it matters: this is the missing proof needed either to reuse caller-owned
storage safely or to select a value carrier without violating Able reference
semantics. It directly targets the measured Go gap while keeping the compiler
free of nominal-name, container-name, and benchmark rules.
