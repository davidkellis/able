# Native dictionary runtime-reach profile closure

Date: 2026-07-30

## Decision

**Retain no production code. The selected exact-`Self` native-dictionary path
is absent from all three applications.**

`concurrent_graph_visitors`, `concurrent_event_routing`, and
`validated_job_pipeline` all remain strict, publicly verified, and
interpreter-free. Their interface-heavy generated modules contain native Go
adapters, but none of the three application sources declares an interface
method with an exact top-level `Self` return. Consequently, no exact-`Self`
adapter can execute materially in this cohort.

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, language, dependency, benchmark, fixture, or WASM source
changed.

After retaining compact evidence, the exact 866-MiB disk-backed task workspace
and 16-KiB generated Python cache were removed.

## Protocol

Normal `--no-fallbacks` binaries were built before profiling. Each application
received:

- one untimed public-verifier smoke execution;
- ten independent main-only CPU-profile processes;
- three independent exact main-allocation profile processes; and
- three typed-boundary telemetry processes, including the initial audited
  build and two direct repeats.

All 51 processes verified, with zero failures and zero timeouts. Final Go
dependency graphs omit `able/interpreter-go/pkg/interpreter` in all three
applications. Profiles used CPUs `12-15`, `GOMAXPROCS=4`, `GOGC=50`, and a
1-GiB memory limit.

The workstation carried sustained unrelated Marketlab load. These profiles
are attribution and reachability evidence; they do not replace the current
five-process scoreboard or support a new wall-time claim.

## Exact reach result

| Application | Exact top-level `Self` interface methods | Interface runtime conversions | Merged CPU | Sampled native adapter |
| --- | ---: | ---: | ---: | --- |
| Concurrent Graph Visitors | 0 | 0 | 10 ms | `GraphVisitor.inspect`, 10 ms cumulative |
| Concurrent Event Routing | 0 | 0 | 160 ms | `Map.get`, 10 ms cumulative |
| Validated Job Pipeline | 0 | 0 | 340 ms | none |

The two sampled adapters implement different methods with ordinary concrete or
optional results. Neither is an exact-`Self` method, and neither repeats in the
third application. No generated `Extend`, `Clone`, iterator, error, or other
exact-`Self` adapter frame appears in any merged CPU profile.

Typed-boundary counts reproduced exactly in all three processes per
application:

- interface from runtime: zero in every application;
- interface to runtime: zero in every application; and
- interface lift through runtime: zero in every application.

This is positive evidence that the applications do not box interface values
or cross into the interpreter. It is also negative evidence for the proposed
optimization: the newly admitted exact-`Self` adapter capability is not the
owner of these applications' remaining Go gap.

## Allocation attribution

Exact main-phase allocation counters were stable:

| Application | Allocated bytes, three-run mean | Allocations, three-run mean |
| --- | ---: | ---: |
| Concurrent Graph Visitors | 1,197,728 | 46,892.33 |
| Concurrent Event Routing | 4,184,669.33 | 70,663.67 |
| Validated Job Pipeline | 1,832,490.67 | 32,940 |

Graph Visitors allocates under graph construction, state recording, and its
ordinary visitor call. Event Routing allocates nominal channel/union payloads,
text parsing, and map state. Validated Job Pipeline allocates nominal channel
payloads and spends sampled CPU in the already-closed execution-context
identity path. Those are different semantic parents. Grouping them as
“interface cost” or “allocation” would not identify one removable general
rule.

The current scoreboard still reports substantial descriptive gaps:

| Application | Able | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Concurrent Graph Visitors | 0.0300 s | 0.0044 s | 6.82x |
| Concurrent Event Routing | 0.0540 s | 0.0055 s | 9.82x |
| Validated Job Pipeline | 0.0700 s | 0.0044 s | 15.91x |

Those gaps are real, but this tranche does not explain them with a shared
native-dictionary owner. Channel-handle recovery, nominal scheduler payloads,
`currentGID`, broad execution-context ABI changes, named containers, and
non-primitive nominal shortcuts are already closed or prohibited routes.

## Admission result

No exact boundary owner repeats in three unlike applications. Therefore:

- no candidate was implemented;
- no pre-change compiler was rebuilt;
- no baseline/candidate/Go A/B cohort was manufactured; and
- the retained interface-dictionary correctness capability remains unchanged.

The machine-readable summary is
`2026-07-30-native-dictionary-runtime-reach-no-go.json`; the raw typed-boundary
audit is `2026-07-30-native-dictionary-runtime-reach-typed-boundary.json`.

## Next

Audit sustained benchmark coverage for exact top-level `Self` interface
returns before doing more work on this path.

Why: the current three-application selection was based on generated-source
contraction, but none actually exercises the language surface whose lowering
changed. Static generated support code is not runtime evidence.

What it entails: inspect the 66-application catalog and sibling external suite
for natural builder, state-machine, visitor, or transformation workloads that
return exact `Self` through interface values; rank any candidates by sustained
main-phase duration, feature interactions, source equivalence, and public
verification. Admit a new benchmark only if the newly canonical dictionary
surface exposes a real coverage gap. Do not create a synthetic timing fixture
or change compiler code merely to make the path hot.

Why it matters: a real sustained workload can tell us whether native
dictionary preservation removes boxing in application code. If no such
workload exists, the correct result is to close this direction until a natural
application or production identity change supplies evidence, preserving the
project's generality bar and avoiding benchmark-shaped lowering.
