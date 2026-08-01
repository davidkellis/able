# Interface-dictionary capture contract retained

Date: 2026-07-30

## Decision

**Retain the semantic correction and native lowering capability; admit no
performance optimization from this tranche.**

Static interface upcasts now select and retain one visible implementation
dictionary. Distinct package-private implementations for the same concrete
carrier remain distinct, an exact top-level `Self` return preserves the
originating interface view, and statically known incomparable implementations
are rejected at the upcast.

The compiler can represent exact-`Self` interface methods with native Go
interface carriers. Their adapters call the selected compiled implementation
on its native receiver and wrap the concrete result back into the same native
carrier. The strict fixtures do not require `runtime.Value`, an interpreter
method call, or consumer-scope implementation selection.

This is a general interface rule, not a named-container, nominal-type,
application, or benchmark special case. The rule is required for language
correctness. Repeated measurements show it is effectively wall-time neutral
across the selected unlike applications, so this record does not claim it as
a speedup.

## Contract and implementation

- The v12 spec defines dictionary selection at every statically known upcast,
  compile-time and runtime ambiguity handling, fixed default/override slots,
  generic slot instantiation, and preservation of the captured view for exact
  top-level `Self` returns.
- The checker diagnoses known incomparable implementations on typed
  assignment and explicit casts.
- The tree-walker and bytecode VM rewrap an exact-`Self` result with the
  originating interface definition, arguments, shared dictionary, and an
  isolated per-value method overlay.
- Runtime implementation lookup and caches include the originating package,
  so two private implementations for one interface/concrete pair cannot
  replace one another.
- The compiler synthesizes implementation-definition-scoped native adapters,
  selects only adapters visible from the upcast package, and emits exact
  `Self` returns on the native interface carrier.

No stdlib, dependency, language primitive, named non-primitive nominal rule,
or WASM change was required.

## Correctness guards

The following pass under individual 60-second bounds:

- focused typed-assignment and explicit-cast ambiguity tests;
- focused interface-`Self` dictionary preservation;
- tree-walker, bytecode, and parity execution of
  `13_08_static_interface_dictionary_capture` and
  `13_10_static_interface_distinct_dictionary_capture`;
- exact native interface `Self`-return carrier generation;
- strict fallback-free compiled execution of both fixtures;
- the complete interpreter unit group, exec-fixture group, parity group, and
  `go test ./cmd/ablec` captured during implementation.

The monolithic compiler and interpreter package aggregates are not used as
single tests because their independent constituent groups exceed the
one-minute project limit when combined. One accidentally broad compiler test
selection also reached the one-minute cap while the workstation was fully
saturated; its exact intended test passed in 0.888 seconds and its two strict
fixtures passed together in 6.233 seconds.

## Static reach

The checked strict census is
`2026-07-30-interface-dictionary-capture-strict-static-census.json`.
All 66 selected applications generated with `--no-fallbacks`.

Compared with the preceding composite-interface census:

- 58 of 66 generated modules changed and eight were byte-identical;
- aggregate generated source fell from 288,624,837 to 285,352,687 bytes
  (-3,272,150 bytes, -1.13%);
- aggregate generated Go lines fell from 7,145,082 to 7,060,431
  (-84,651 lines, -1.18%);
- generated file count remained 3,316.

The unchanged applications were `base64`, `distance_field`,
`future_await_race`, `future_pipeline`, `monte_carlo_pi`,
`mutex_await_journal`, `option_result_config`, and `rms_norm`.

The broad source contraction establishes static reach, not runtime speed. The
isolated exact-`Self` admission adds one native adapter family to each of the
three timed modules, so their individual source and binary sizes increased
slightly while eliminating a previously unsupported native interface shape.

## Isolated repeated A/B

The current compiler binary had SHA-256
`a907be3cf7568e380494868eef114ab804a94139ee745f16c703531c8875d5bb`.
The isolated baseline had SHA-256
`c77de521babae1967424fb8baf2e796139e780e515544300a78a39f5c33c9ba0`
and differed only by restoring the prior rejection of interface methods whose
return type used `Self`. Both compilers otherwise used the same production
tree.

Every application/variant has two opposite-order five-process cohorts. All 90
outputs verified. The raw samples are
`2026-07-30-interface-dictionary-capture-ab-timings.tsv`.

| Application | Baseline mean | Current mean | Current vs baseline | Current RSS mean | Cohort spread |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Trees | 15.7830s | 16.0870s | +1.93% | 229,275 KiB | 1.41% |
| Quicksort | 1.7270s | 1.7540s | +1.56% | 146,867 KiB | 0.23% |
| k-Nucleotide | 1.3610s | 1.3400s | -1.54% | 27,016 KiB | 0.30% |

The three-application geometric current/baseline wall-time ratio is 1.00637,
or +0.64%. Mean RSS changed by +0.01%, -0.09%, and +0.63% respectively. These
small mixed movements do not establish a causal performance win or material
regression.

Equivalent Go implementations were run and verified in the same cohorts.
Their ratios are descriptive rather than a replacement for the published
scorecard because this cohort used CPU pool `0,1,2,3`, while the preceding
scorecard used `12,13,14,15`. Within this cohort Quicksort exceeded Go, while
Binary Trees and k-Nucleotide remained below the 95% target.

## CPU and allocation profiles

One output-verified CPU and `alloc_space` profile was captured for each
baseline/current application pair:

- Binary Trees remained dominated by `make_tree`, `check_tree`, allocation,
  and GC. Sampled allocation moved from 9,485.94 to 9,407.21 MiB and peak RSS
  from 322,216 to 310,108 KiB.
- Quicksort remained dominated by native quicksort/swap, checked multiply,
  input parsing, and file-buffer allocation. Sampled allocation moved from
  285.47 to 287.11 MiB and peak RSS was unchanged at 148,784 KiB.
- k-Nucleotide remained dominated by primitive HashMap hashing/equality and
  `bridge.ToUint`/`bridge.ToInt`. Sampled allocation moved from 564.41 to
  575.84 MiB and peak RSS from 26,448 to 26,960 KiB.

No interface adapter, dictionary lookup, `runtime.Value` conversion, or other
new owner was hot in all three applications. The sampled allocation movements
are mixed and small enough to be profiling noise. No shared production
optimization clears the three-unlike-application gate.

The k-Nucleotide bridge cost is a real remaining gap, but this tranche does
not reopen it: prior work found no cross-family shared leaf, and a HashMap- or
text-map-specific compiler rule would violate the nominal-lowering guardrail.

## Evidence state

The production and spec changes invalidate all 23 performance closures for
the explicit reasons `scope-content-drift:compiler-production`,
`scope-content-drift:bytecode-production`,
`scope-content-drift:shared-interpreter-semantics`, and
`scope-content-drift:v12-spec`.

A canonical rerun was attempted only after CPU 12 passed the quiet preflight,
but an unrelated Marketlab pipeline subsequently saturated the workstation.
The reverse cohort slowed Quicksort and Binary Trees by roughly 1.7–2.6x.
That task-owned loop was stopped and its samples were rejected rather than
promoted.

The established ten-compiled/nine-bytecode set-cover reconciliation was then
completed conservatively under the sustained load. All 470 opposite-order
Able/Go/Python/Ruby processes verified with zero failures or timeouts. Five
volatile rows remain descriptive. Ten strict boundary guards omit the
interpreter and record zero runtime-service calls. The reviewed ledger now has
23 current closures, zero invalidations, and an empty selector. See
`2026-07-30-interface-dictionary-capture-performance-evidence-reconciliation.md`.

## Next

Profile native dictionary runtime reach in `concurrent_graph_visitors`,
`concurrent_event_routing`, and `validated_job_pipeline`.

Why: those interface-heavy applications have three of the largest
generated-source contractions, while the current algorithmic A/B trio did not
execute an interface adapter as a hot owner.

What it entails: collect CPU, `alloc_space`, and typed-boundary profiles;
prove whether exact-`Self` native adapters execute materially; select one
exact repeated boundary owner; then use a pre-change compiler and equivalent
Go applications for balanced five-or-more-pair A/B.

Why it matters: this is the shortest general path from sound dictionary
semantics to fewer `runtime.Value` conversions and compiled/interpreted
crossings in real interface-heavy programs. Retain no code if the owner is
absent, one-family, or launch-dominated.
