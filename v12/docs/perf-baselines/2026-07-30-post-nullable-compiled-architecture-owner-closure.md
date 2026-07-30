# Post-nullable compiled architecture owner closure

## Decision

Retain no production change.

Fresh post-nullable CPU and exact allocation profiles across six unlike
compiled application families expose no concrete compiler or generated-runtime
owner that is material in three of them. No candidate reaches the admission
gate, so no A/B implementation cohort is warranted.

The common allocation and GC frames are aggregate Go runtime parents. Their
Able descendants are different semantic operations and do not constitute one
optimization.

## Selection

The cohort refreshes the checked `compiled-architecture-target-budget` scope
and spans six invalidated compiled ownership families:

| Application | Family | Current Able / Go | Ratio |
| --- | --- | ---: | ---: |
| Concurrent Event Routing | concurrency and nominal records | 0.0540 / 0.0055 s | 9.82× |
| Distance Field | primitive float numeric | 0.0380 / 0.0147 s | 2.59× |
| Fixed Width 128 | wide numeric and nominal results | 0.1060 / 0.0059 s | 17.97× |
| K-Nucleotide | text and generic map | 1.5920 / 0.0669 s | 23.80× |
| Manifest Normalization | nominal Result/Option normalization | 0.0400 / 0.0050 s | 8.00× |
| Policy Record Dispatch | regex and record dispatch | 0.0960 / 0.0053 s | 18.11× |

These are the current five-process verifier-backed scorecard means. Profile
processes are ownership evidence only and are not substituted for ordinary
timing.

## Strict boundary gate

Every application was freshly built with `--no-fallbacks`. Each final
dependency graph contains 96 packages and omits
`able/interpreter-go/pkg/interpreter`. One smoke, ten CPU-profile processes,
and three allocation-profile processes per application passed the public
verifier with one stable output hash per application.

This directly excludes accidental compiled/interpreted transitions as the
remaining cause.

## CPU ownership

Ten independent `ABLE_GO_PHASE_CPU_PROFILE_DIR` processes per application
were merged using only the generated `main` phase:

| Application | Merged CPU samples | Material owner |
| --- | ---: | --- |
| Concurrent Event Routing | 300 ms | locking/channel scheduling, checked record arithmetic, and record construction |
| Distance Field | 160 ms | native generated `main`, `hypot`, `sqrt`, and `math.Sqrt` |
| Fixed Width 128 | 1.33 s | loop-carried nominal UInt128 allocation, wide arithmetic, and checked multiply |
| K-Nucleotide | 24.88 s | primitive map equality/hash, map lookup, `ToUint`, `ToInt`, and allocation |
| Manifest Normalization | 140 ms | String contains/conversion, nominal parsing, and allocation |
| Policy Record Dispatch | 710 ms | regex NFA closure, move, capture, and allocation |

No exact CPU owner is material in all six, or even in three open unlike
families. The only named helper sampled across three is checked signed
addition, but it has just 10 ms in each of Event Routing and Manifest and 10 ms
of 1.33 seconds in Fixed Width. Broad checked-arithmetic candidates are
already closed by mixed A/B results and are explicitly outside this tranche.

## Exact main allocations

Three independent `ABLE_GO_PHASE_ALLOC_PROFILE_DIR` processes per application
recorded exact main-phase counters:

| Application | Mean bytes | Mean objects | Mean GC |
| --- | ---: | ---: | ---: |
| Concurrent Event Routing | 4,186,712 | 70,680 | 3.67 |
| Distance Field | 512 | 11 | 0.00 |
| Fixed Width 128 | 35,536,416 | 2,220,986 | 24.00 |
| K-Nucleotide | 598,182,443 | 12,232,484 | 323.67 |
| Manifest Normalization | 3,559,528 | 76,234.67 | 3.00 |
| Policy Record Dispatch | 20,768,720 | 435,019.67 | 18.33 |

Representative exact owners are:

- Event Routing: 13,320 positional semantic structs, 12,301 `ToInt`
  conversions, and 8,192 `route_task` objects.
- Distance Field: eleven total application-phase allocations and no material
  Able allocation leaf.
- Fixed Width: 1,220,964 `modular_add_checksum` and 1,000,001
  `ordered_select_checksum` loop-carried nominal results.
- K-Nucleotide: 7,999,998 `ToUint`, 3,961,373 `ToInt`, and 233,358 builtin
  String conversions.
- Manifest: 36,866 `ToString`, 13,312 positional semantic structs, and
  distinct parse/normalization nominal values.
- Policy: 94,464 regex capture clones, 86,305 codepoint results, and the
  regex NFA thread/capture families.

The allocation-profile writer appears in the end snapshots and is
observer-only. Exact main counters above come from `phase-stats.json`.

## Breadth gate

| Concrete owner | Material unlike families | Result |
| --- | ---: | --- |
| `bridge.ToInt` | 2: Event Routing, K-Nucleotide | insufficient breadth |
| `bridge.ToString` | 2: Event Routing, Manifest | insufficient breadth |
| `NewStructInstancePositionalSized` | 2: Event Routing, Manifest | insufficient breadth |
| checked signed arithmetic | 2 material, 1 sparse sample | previously closed |
| generic map hash/storage | 1 | named-container treatment forbidden |
| loop-carried UInt128 nominal values | 1 | nominal specialization forbidden |
| regex NFA storage/execution | 1 | same-family only |
| allocation/GC | 5 | aggregate parent with distinct Able descendants |

The primitive nullable carrier appears only as low cumulative conversion reach
in K-Nucleotide and Manifest. Its former allocation owner does not reappear.

## Retained artifacts

Readable CPU flat/cumulative reports, allocation object/space reports, and all
18 exact phase-stat records are under:

`2026-07-30-post-nullable-compiled-architecture-owner-profiles/`

The machine summary is:

`2026-07-30-post-nullable-compiled-architecture-owner-closure.json`

Raw generated modules, binaries, and binary pprof files are disposable after
this record is verified.

After verification, the exact 833 MiB disk-backed
`/var/tmp/able-post-nullable-architecture.*` workspace and generated Python
cache were removed. No matching active artifact remains.

## Scope

No compiler, generated runtime, runtime package, tree-walker, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No benchmark, named-container, or non-primitive nominal special case was
introduced.

## Next

Reconcile the remaining invalidated compiled closures by causal reach of the
nullable carrier before scheduling more expensive profiles.

Why: all 12 closures are selected by one broad compiler-production hash drift,
but this six-family refresh shows that the retained carrier does not create a
new cross-family owner. Blindly reprofiling every closed family would repeat
work without establishing that the changed representation reaches it.

What it entails: scan the generated source and current profile evidence for
each invalidated closure, starting with `compiled-current-control` and its
Fib, Matrix Multiply, and TapeLang rows. Reprofile only a group where the new
carrier reaches a material path; otherwise retain a causal no-reach
reconciliation. Keep the established target guards protected.

Why it matters: this restores trustworthy closure coverage efficiently while
keeping optimization selection tied to concrete current code rather than a
coarse directory hash.
