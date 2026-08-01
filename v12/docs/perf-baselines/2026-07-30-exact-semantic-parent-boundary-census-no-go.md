# Exact semantic-parent boundary census closure

Date: 2026-07-30

## Decision

Retain the report-only exact semantic-parent census extension and no compiler,
generated-runtime, runtime, interpreter, bytecode VM, canonical-stdlib,
language, dependency, benchmark, fixture, frozen-workspace, or WASM
performance change.

All 66 current strict compiled applications regenerated successfully under
`--no-fallbacks`. Their module hashes match the retained post-dictionary
full-row census 66 for 66, so the earlier interpreter-free dependency proofs
remain exact. The new analyzer joins each direct-main boundary site by:

1. semantic category;
2. exact callee; and
3. exact generated parent.

It then joins those identities to the retained five-process timing frontier,
workload group, profile freshness, and profile owner. No exact safely
lowerable identity is both material and present in three unlike workload
groups. No new CPU, allocation, baseline/candidate, or Go cohort was run.

The machine-readable companion is
`2026-07-30-exact-semantic-parent-boundary-census-no-go.json`.

## Why regeneration was required

The retained 66-row census contained per-application category totals, exact
callee totals, and parent-by-category totals. Those are marginals: when an
application has multiple callees and multiple parents in one category, they
cannot prove which callee occurs under which parent.

The report-only analyzer now records the exact triple at every site. It also
records `runtime.Value` type sites and heap nominal literals through the same
parent-indexed structure. The ordinary compiler output and execution path are
unchanged.

The bounded runner now aggregates direct-main triples across applications and
attaches the existing timing excess and retained profile authority. This
makes future admission mechanical instead of reconstructing pairs from
ambiguous counts.

## Bounded coverage

The workstation carried sustained unrelated Marketlab load. Generation
duration is therefore only a completion guard, not performance evidence.

| Measure | Result |
| --- | ---: |
| Selected strict applications | 66 |
| Successful / failed | 66 / 0 |
| Generation time | 0.309-11.522 s; 3.520 s mean |
| Generated module size | 1,123,264-8,966,497 bytes |
| Generated Go lines | 30,143-211,856 |
| Module identities matching retained census | 66 / 66 |
| Exact direct-main identities | 1,304 |
| Identities in at least three applications | 418 |
| Identities spanning at least three workload groups | 21 |

Every application stayed below the one-minute per-module rule. Two compiler
workers were used to limit interference with unrelated work.

## Exact breadth reduction

Most apparently broad categories collapse after the exact parent join:

| Category | Exact identities | Maximum applications | Maximum workload groups | Cross-group identities |
| --- | ---: | ---: | ---: | ---: |
| `runtime.Value` type | 58 | 36 | 8 | 4 |
| bridge encode | 40 | 29 | 6 | 5 |
| bridge decode | 16 | 14 | 3 | 1 |
| bridge error | 105 | 10 | 3 | 4 |
| erased/dynamic call | 7 | 36 | 8 | 1 |
| interface runtime conversion | 4 | 3 | 1 | 0 |
| callable runtime conversion | 4 | 2 | 1 | 0 |
| union runtime conversion | 1 | 1 | 1 | 0 |
| struct runtime conversion | 92 | 6 | 2 | 0 |
| native interface adapter | 32 | 2 | 1 | 0 |
| native union wrap/projection | 366 | 20 | 7 | 1 |
| heap nominal literal | 449 | 6 | 2 | 0 |

The boundary classes most relevant to eliminating compiled/runtime crossings
fail breadth directly:

- the only three-application interface recovery is
  `Future<i64>.from_value` in Await Channel Mux, Mutex Work Queue, and Mutex
  Ledger, all in the concurrency group;
- the widest callable conversion is `void -> i64` under
  `Mutex.await_lock`, in two related mutex applications;
- the only union runtime conversion occurs in Option/Result Config;
- exact struct conversions remain within regex or numeric families; and
- no exact native interface adapter reaches three applications.

## The 21 cross-group identities

The surviving identities form four closed classes.

### Explicit main/host ABI

`__able_call_named`, its `runtime.Value` parameter, `bridge.ToString`, and
`bridge.ToFloat64` occur under generated `main` across several groups. They
are one or a few final host argument/output conversions per process. Retained
normal-binary profiles do not identify them as shared material CPU or
allocation owners.

They are explicit host boundaries, not evidence that the application body
crosses into the interpreter.

### Already-native union lowering

`Path | String` wrapping under generated `main` reaches 20 applications and
seven groups. It constructs the compiler's native Go union carrier. Removing
it would remove semantic union tagging, not boxing.

### Canonical HashMap kernel boundary

Twelve identities occur under:

- `HashMap.raw_set`;
- `HashMap.with_capacity`; and
- `HashMap.raw_get`.

The callees are `bridge.ToDynamicI64`, `bridge.AsInt`, `bridge.ToInt`,
`runtime.Value`, and successful control/error checks. Static reach spans map,
iterator, and regex workload groups because those applications use the same
canonical container.

This is not three unlike semantic parents: it is one named non-primitive
kernel parent copied into several applications. The retained nullable
reconciliations already classify its surviving conversions as explicit
map-kernel dynamic boundaries. k-Nucleotide's current profile makes the
boundary material there, but the other retained profiles do not establish
one general lowerable child in three unlike programs.

A compiler rule for `HashMap` would violate the nominal-lowering guardrail.
No canonical-stdlib change is justified because the kernel must preserve
generic key/value identity and dynamic service semantics.

### Cold Result/error guards

`bridge.RaiseRuntimeErrorWithContext` appears under generated `main`,
`Result.is_ok`, `Result.map`, and `Result.unwrap_or` across three groups.
These are semantic failure paths. Retained normal-path profiles do not sample
them as a shared material owner, and successful control conversions remain
nil checks above unrelated work.

## Admission result

No identity clears all required gates:

- exact callee and parent;
- three unlike applications;
- material retained normal-binary profile reach;
- a semantics-preserving general rule; and
- no benchmark, named-container, or non-primitive nominal special case.

Accordingly:

- no compiler/runtime candidate was implemented;
- no new profile cohort was launched;
- no baseline/candidate/Go A/B was manufactured;
- no dynamic or escaping path was narrowed; and
- no deferred WASM work was touched.

## Retained tooling and verification

The retained report-only changes are:

- analyzer schema 2, which records exact category/callee/parent triples,
  including `runtime.Value` and heap nominal sites;
- aggregate exact-parent ranking in
  `v12/bench_compiled_static_boundary_census`; and
- a focused synthetic contract covering every new identity shape.

The focused Go analyzer test passes below one second. Runner shell syntax
passes. All task source files remain below 1,000 lines. The generated module
identity match proves this tool-only change did not alter ordinary compiler
output.

The exact 433,584 KiB task-created disk-backed workspace was removed after
validation. It contained the compiler, analyzer, Go cache, compact row
results, and disposable raw/enriched census reports. No `/tmp/able-*`
artifact or new Python cache remained.

## Next

Run the current mode-aware performance-evidence selector and the ordinary
release/correctness gates, then reopen execution profiling only for a
genuinely invalidated exact owner.

Why: the refined compiled boundary census is complete and finds no admissible
native/erased crossing. The evidence selector is the fail-closed authority
for whether the analyzer/tooling change or any concurrent retained compiler
work has made a compiled or bytecode closure stale.

What it entails: validate this record and its content-addressed inputs; run
the ledger, scoreboard, frontier, strict dependency, fixture, compiler, and
canonical-stdlib checks; repair only a real shared failure; and profile only
an exact selected owner that is material in three unlike applications.

Why it matters: this protects native primitive and Array carriers while
keeping progress toward Go-native compiled performance and competitive
bytecode performance evidence-driven. It prevents the next tranche from
repeating a named-container, cold-boundary, or already-native route.
