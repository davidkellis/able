# Concurrent Graph Visitors application gate — 2026-07-23

## Decision

Retain the source-equivalent `concurrent_graph_visitors` application, exact
verifier, catalog/coverage/selection metadata, two complete measurement
cohorts, and bounded profiles. Also retain one generic bytecode VM change:
the map-backed member-method cache now retains at most 4,096 entries while
the existing hot and instruction-indexed direct caches continue to serve
current call sites.

Retain no compiler, generated-runtime, tree-walker, canonical-stdlib,
language, dependency, named-container, nominal-type, or WASM change. A
64-entry direct-cache experiment was removed because its larger per-VM
footprint produced mixed policy and stateful-pipeline guards.

## Application contract

Four Futures independently build and breadth-first traverse immutable directed
graphs. Each graph node and visit state is a nominal value with inherent
methods. Every visited node crosses a user-defined `GraphVisitor` interface
and invokes a captured scoring callback. The Queue and seen set are local to
each traversal, and Futures are joined only after every traversal completes,
so the exact output is schedule-independent:

```text
8192:8192:5942:2250:204689,740309,923614,61231:464714,115365,968398,760636:274,274,275,277:929840:116698
```

Tree-walker, bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0 emit
that output. Its SHA-256 is
`399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee`.
Canonical and external Able sources are byte-identical with SHA-256
`b445e78cd4c706fc09a3ab008df58fb4a5c90237b2caee4cd52420d04298db56`.
The catalog assigns four CPUs to compiled/Go, one to interpreter lanes, uses
the goroutine executor, and isolates the explicit source root.

## Coverage result

The application genuinely covers lexical matching, nominal and generic
values, Arrays, captured functions, loop/branch/match control flow, inherent
methods, user-defined interfaces and implementations, Option flow through
Queue dequeue, spawn/Future lifecycles, package imports, stdlib protocols,
and program arguments.

The corpus now contains 57 portable applications. Both modes are selected, so
the strict manifest contains 57 compiled and 50 bytecode rows. The weighted
three-feature interaction frontier has no zero-coverage triple and its minimum
depth rises from six to seven.

## Repeated measurements

Every lane received two independent five-process cohorts. All 50 timed
processes passed the exact verifier with no timeout or failure.

| Lane | Processes | Pooled mean | Cohort A | Cohort B | Limiting ratio |
| --- | ---: | ---: | ---: | ---: | ---: |
| Able compiled | 10 | 0.280000 s | 0.2580 s | 0.3020 s | 69.295× Go |
| Go 1.26 | 10 | 0.004041 s | 0.004031 s | 0.004050 s | — |
| Able bytecode | 10 | 0.957000 s | 0.9540 s | 0.9600 s | 16.919× Python / 17.662× Ruby |
| Python 3.14 | 10 | 0.056565 s | 0.056407 s | 0.056723 s | — |
| Ruby 4.0 | 10 | 0.054183 s | 0.053933 s | 0.054433 s | — |

The compiled cohort means differ by 15.7%, so the conclusion uses all ten
workstation samples. The other lanes differ by less than 1%. The final
bytecode cohorts improve from the pre-fix pooled 1.110 seconds to 0.957
seconds, a 13.8% wall-time reduction with identical output.

## Ownership and admission

Three compiled profiles merge to 600 ms of CPU samples.
`bridge.currentGID` and `runtime.Stack` own 93.33% cumulatively beneath all
four traversal Futures, both visitor implementations, inherent state methods,
and captured scorers. This is the established compiled-concurrency owner; its
fixed-context replacement already failed broad concurrent and serial guards.

The pre-fix bytecode census recorded 269,049 inline-call hits with zero
misses, but 191,088 member-method cache misses and only 180,729 hits in one
main call. The cache key safely includes its activation environment. In this
ordinary multi-method workload, many such environments are one-shot, so the
unbounded map retained them and repeatedly rehashed. Three pre-fix profiles
averaged 1.025 seconds/op, 144.96 MB/op, and 1.419 million allocations/op;
`storeCachedMemberMethod` and `lookupCachedMemberMethodEntry` owned 16.04%
and 15.84% cumulatively.

The retained 4,096-entry bound preserves cache validation and existing-entry
refresh, and it continues to populate the hot/direct caches even after the
map reaches its bound. Three final profiles average 755.63 ms/op, 94.09 MB/op,
and 1.205 million allocations/op. The residual profile places lookup at
17.45%, store at 10.72%, and RWMutex reader bookkeeping at 11.69%.

The broader A/B guard used repeated processes. Graph traversal improved
materially, iterator collect was neutral, and the six-sample candidate/control
means were within -0.8% for state machines, +0.8% for policy callbacks, and
+1.5% for the stateful pipeline. The widened direct-cache alternative was
removed; the retained map bound adds no benchmark, container, or nominal-type
branch.

## Verification

- six-lane exact output parity;
- ten verifier-backed processes per compiled, bytecode, and reference lane;
- three compiled, three pre-fix bytecode, and three final bytecode profiles;
- focused member-cache, call-member, reset, and return tests;
- catalog, selection, coverage, operation-depth, pair, and triple checks;
- every added source file and touched implementation file remains below 1,000
  lines;
- JSON, source identity, formatting, and whitespace checks.

## Next recommendation

Add one materially different eighth portable application that raises both
minimum-depth interactions with independently spawned fork/join tree folds.

Why: scorecard and architecture reconciliation is complete, and both
shallowest triples are now at depth seven. A nominal tree/fold topology covers
the same high-value language interactions without repeating graph
breadth-first search or the earlier concurrent topologies.

What it entails: build source-equivalent deterministic Able, Go, Python, and
Ruby implementations where each Future folds an independent nominal tree,
calls inherent node/fold-state methods, chooses a fold algebra through a
user-defined interface, and receives captured weight or pruning callbacks as
data. Use two five-process cohorts and profiles after six-lane parity. Admit
another implementation change only for an exact generic owner repeated across
unlike programs. Update canonical `able-stdlib` only for a reusable API or
correctness defect, and do not begin WASM work.
