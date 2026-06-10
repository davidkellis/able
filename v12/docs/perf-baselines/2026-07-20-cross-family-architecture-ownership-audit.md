# Cross-family architecture ownership audit

Date: 2026-07-20

## Decision

Complete the architecture-level ownership audit and retain no compiler,
bytecode VM, runtime, parser, canonical-stdlib, benchmark, fixture, language,
nominal-lowering, or WASM performance change.

The audit reconciles the 67 selected misses and 121.969 seconds of aggregate
target excess with the zero-actionable frontier. Broad architecture costs do
exist, especially in the bytecode VM, but none supplies a new causal candidate:
the same mechanisms have either failed unlike-application wall-time guards or
split into different concrete semantic owners below their aggregate parent.
The machine-readable census is
`2026-07-20-cross-family-architecture-ownership-census.json`.

## Admission contract

This tranche used the latest exact CPU/allocation evidence rather than
collecting another overlapping profile cohort. A candidate needed:

- the same concrete non-nominal architectural boundary to be material in at
  least three unlike verifier-backed applications;
- a mechanism not already rejected by broad causal A/B evidence;
- a shared VM, compiler, runtime, or stdlib implementation, not an application,
  algorithm, source shape, or named-container rule; and
- the six source-exact established target guards defined before implementation.

Aggregate `runResumable`, `runtime.mapaccess2_faststr`, `mallocgc`, GC, call,
return, or generated-runtime ancestry was not treated as a leaf without exact
caller/descendant attribution.

## Cross-family census

| Boundary | Cross-family evidence | Exact ownership result | Disposition |
| --- | --- | --- | --- |
| Compiled generated runtime | all 35 selected applications counted | Most boxed/dynamic helpers execute zero times; integer conversion is broad but below 1% in three unlike profiles | no shared material leaf |
| Compiled escape/nominal ABI | broad static emission plus exact dynamic attribution | one nullable consumer and one material nominal definition; other allocations have different required lifetimes | insufficient dynamic breadth |
| Compiled initialization/context | launchers, concurrency, N-Body, Binary Trees, and text/map guards | package/context cuts improve some rows but regress Binary Trees, N-Body, or K-Nucleotide | rejected candidate |
| VM primary dispatch | material flat cost in at least six unlike applications | direct/layout-stable opcodes, typed blocks, restricted register IR, and Go PGO all failed reach, deployability, or broad wall guards | rejected mechanisms |
| VM scalar encoding | integer CPU recurrence and three-program float allocation | extractor/carrier/sidecar/typed-lane variants remove counters or allocation but regress unlike wall time | rejected candidate |
| VM call/return continuation | broad call/return ancestry across major families | concrete call routes and return coercions differ; shared frame/raw-cell/guard candidates regress or are neutral | rejected or no shared child |
| VM nominal construction | four-program allocator aggregate | two text consumers share one definition; two related regex consumers share another; values escape through different observable lifetimes | insufficient lifetime breadth |
| Cross-mode map/allocation parents | repeated Go map, malloc, GC, and collection symbols | exact callers are environments, caches, Able maps, parsing, required result growth, nominals, or host conversion | no shared leaf |

Six of the eight boundaries recur across unlike applications. Recurrence alone
is therefore not the missing evidence. The missing property is a concrete
removable operation whose removal predicts wall time in every affected family.

## Compiled comparison with emitted Go

The coverage-wide generated-helper census is decisive. Boxed binary/unary
operators, generic interface dispatch, checked generic Array indexing, String
conversion, and generic-union fallback recorded zero executions. The only new
broad helper, `__able_int64_from_value`, reached eleven applications but was
0%-0.68% flat in four unlike normal-binary profile groups. Generated-Go escape
and bounds diagnostics were broad statically, yet exact dynamic attribution
reduced the nullable result opportunity to one application and the union
payload opportunity to one material definition.

Representative current compiled misses consequently spend their time in
direct generated application bodies or different semantic subsystems: Fib
recursion, Matrix arithmetic, TapeLang dispatch, HashMap/String work, regex
NFA work, scheduler identity, or required allocation. That explains why a
single generated-runtime rewrite cannot close the compiled product gap even
though the aggregate ratio remains poor.

## Bytecode comparison with reference execution

The bytecode VM has genuine common overhead that Python/Ruby references often
keep inside mature native operator and collection paths: opcode dispatch,
Value transitions, frame/continuation work, and dynamic semantic validation.
The exact evidence nevertheless rejects the available general mechanisms:

- a one-opcode layout change moved unrelated code enough to change applications
  by 3%-9%; hiding it behind an existing opcode merely changed which guards
  regressed;
- a typed four-instruction block executor regressed all three admitted
  families, while restricted register IR had no translated executions;
- cross-suite Go PGO grew `runResumable`, imposed an exact-profile plugin ABI,
  and still regressed unseen JSON;
- integer and float carrier designs reduced local helper/allocation counters
  but added enough Go-side work to lose broad wall time; and
- call-frame, returned-raw-cell, return-guard, positional-struct, and
  caller-owned-result designs either failed broad A/B gates or could not
  preserve the observed lifetimes generally.

This does not mean dispatch or representation is free. It means the current
evidence does not authorize another partial representation or dispatch layer.
A larger rewrite without a newly demonstrated boundary would be speculative
and could produce faster selected benchmarks alongside slower real programs.

## Frontier reconciliation

No ownership disposition changes. The schema-2 frontier remains:

- 75 selected rows: 41 compiled and 34 bytecode;
- 8 snapshot target meets and 67 misses;
- 6 source-exact established guards and 2 unestablished crossings;
- 121.969 seconds aggregate target excess; and
- zero actionable groups.

No raw profiles or binaries were required. The audit consumes already retained
exact reports, causal ledgers, coverage-wide censuses, and source-exact guard
evidence. The canonical external `able-stdlib` has no demonstrated reusable
API boundary to change in this tranche.

## Next recommendation

Expand the portable benchmark corpus with a small set of realistic,
multi-feature applications selected from a feature-interaction coverage audit,
then rebuild the cross-family frontier.

Why: the current suite has enough isolated feature depth to close every known
local family, but it no longer exposes an untested shared implementation leaf.
Real applications that combine parsing, nominal data, iterators/collections,
errors, and concurrency can reveal costs that isolated families hide and are
the safest way to invalidate a closed architecture decision without training
the runtime to an existing benchmark. More repeated profiles of the same 41
applications would only reproduce already closed parents.

What it entails: derive a feature-interaction matrix from the catalog and
fixture coverage; choose the smallest missing combinations with credible
application behavior; implement source-equivalent Able, Go, Python, and Ruby
programs plus external verifiers and bounded inputs; update canonical
`able-stdlib` only for a generally reusable specified API gap; obtain repeated
arithmetic-mean baselines under the existing guardrails; and profile only a
concrete boundary that repeats materially across at least three unlike old/new
applications. Keep all six established guards in the first causal gate and
continue to defer WASM.
