# Concurrent Tree Folds application gate — 2026-07-23

## Decision

Retain the source-equivalent `concurrent_tree_folds` application, exact
verifier, catalog and coverage memberships, two complete timing cohorts, and
bounded profiles. Retain no compiler, VM, tree-walker, canonical-stdlib,
language, dependency, named-container, non-primitive nominal, or WASM change.

The new application raises both minimum-depth interactions from seven to
eight:

- concurrency × functions/closures × interface dispatch;
- functions/closures × inherent methods × interface dispatch.

It does so with four independent complete-tree builds and bottom-up folds,
not graph traversal, channels, a worker queue, an ordered pipeline, or a state
machine.

## Correctness and source equivalence

Able, Go, Python, and Ruby each build four 2,047-node flat nominal trees. Each
lane selects one of two fold algebras through an interface, passes captured
weight and pruning callbacks, and returns a deterministic fold report. The
canonical and external Able sources are byte-identical.

All six execution lanes produce:

```text
8188:8188:7466:722:922643,289372,588929,95674:177548,703803,682349,373118:859333,470066,876991,363530:10,10,10,10:936815:66051
```

The verifier-backed stdout SHA-256 is
`ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7`.

## Repeated timing

Every row has ten successful reference samples and ten successful Able
samples across two independent cohorts. The second compiled cohort reused the
same freshly built binary so timed samples remained outside build work.

| Lane | Cohort A | Cohort B | Pooled arithmetic mean |
| --- | ---: | ---: | ---: |
| Able compiled | 0.384 s | 0.368 s | 0.376 s |
| Go | 0.003681 s | 0.003661 s | 0.003671 s |
| Able bytecode | 0.380 s | 0.390 s | 0.385 s |
| Python | 0.056199 s | 0.056799 s | 0.056499 s |
| Ruby | 0.055536 s | 0.054402 s | 0.054969 s |

The pooled ratios are 102.427× Go for compiled Able, 6.814× Python for
bytecode Able, and 7.004× Ruby for bytecode Able. The cohort means differ by
4.3% compiled and 2.6% bytecode, so the miss decision is not sensitive to
workstation noise.

## Profile gate

Three compiled profiles put `bridge.currentGID` and `runtime.Stack` at 97.92%
cumulative. That is the established compiled-concurrency owner whose prior
fixed-context candidate failed broad serial and concurrent guards; this
application supplies no new general design that makes it safe to reopen.

Three bytecode profiles average `188354435 ns/op`, `23205968 B/op`, and
`381976` allocations/op. Their merged profile is diffuse: call dispatch is
27.59% cumulative, member dispatch 22.41%, binary work 15.52%, allocation
12.07%, RWMutex reader bookkeeping 12.07%, return completion 8.62%, and
member-cache lookup 8.62%. The dispatch trace confirms that the hot inherent
and interface calls are already inline after four cold interface resolutions.
No deeper child is both dominant and separable across unlike programs, so a
new cache, lock, call-frame, return, or named-type shortcut would be
speculative.

## Evidence

- `2026-07-23-concurrent-tree-folds-go-{a,b}.json`
- `2026-07-23-concurrent-tree-folds-interpreter-{a,b}.json`
- `2026-07-23-concurrent-tree-folds-cohort-a.json`
- `2026-07-23-concurrent-tree-folds-compiled-cohort-b.json`
- `2026-07-23-concurrent-tree-folds-bytecode-cohort-b.json`
- `2026-07-23-concurrent-tree-folds-compiled-profile-top.txt`
- `2026-07-23-concurrent-tree-folds-bytecode-profile-top.txt`
- `2026-07-23-concurrent-tree-folds-bytecode-trace.json`
- `2026-07-23-concurrent-tree-folds-interaction-triple-frontier.json`

## Next recommendation

Add a ninth portable application using independently rendered fixed-point
scene tiles.

Why: the same two high-value interactions remain the shallowest frontier rows
at depth eight. A tiled geometry workload can raise both without repeating
trees, graphs, queues, pipelines, state machines, signals, or callback
batches.

What it entails: four Futures each rasterize an independent integer grid
tile; nominal point and tile-accumulator types expose inherent methods; a
user-defined interface selects between signed-distance fields; and captured
shading callbacks cross that interface boundary. Implement exact Able, Go,
Python, and Ruby versions plus a schedule-independent verifier, establish
six-lane parity, run two five-process cohorts, and profile only after
correctness. Admit a runtime change only if one generic owner repeats across
unlike applications.
