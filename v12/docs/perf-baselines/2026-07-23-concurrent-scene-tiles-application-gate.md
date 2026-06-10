# Concurrent Scene Tiles application gate — 2026-07-23

## Decision

Retain the source-equivalent `concurrent_scene_tiles` application, exact
verifier, catalog and coverage memberships, two complete timing cohorts, and
bounded profiles. Retain no compiler, VM, tree-walker, canonical-stdlib,
language, dependency, named-container, non-primitive nominal, or WASM change.

The application raises both minimum-depth interactions from eight to nine:

- concurrency × functions/closures × interface dispatch;
- functions/closures × inherent methods × interface dispatch.

It does so with four independent fixed-point tile renderers, not trees,
graphs, queues, pipelines, state machines, signals, callback batches, or
floating-point geometry.

## Correctness and source equivalence

Able, Go, Python, and Ruby each render four 64×48 integer-grid tiles. Each lane
selects one of two signed-distance fields through an interface, passes a
captured shader through that interface method on every pixel, constructs
nominal point/sample values, and updates a nominal accumulator through
inherent methods. The canonical and external Able sources are byte-identical.

All six execution lanes produce:

```text
12288:12288:4560:1205:6056:940071,96092,396632,704676:805917,90218,566795,592484:-54,-65,-56,-78:80,67,85,76:137465:353691
```

The verifier-backed stdout SHA-256 is
`2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe`.

## Repeated timing

Every measured lane has ten successful samples across two independent
five-process cohorts.

| Lane | Cohort A | Cohort B | Pooled arithmetic mean |
| --- | ---: | ---: | ---: |
| Able compiled | 0.436 s | 0.376 s | 0.406 s |
| Go | 0.004450 s | 0.003867 s | 0.004168 s |
| Able bytecode | 0.672 s | 0.622 s | 0.647 s |
| Python | 0.076455 s | 0.073047 s | 0.074751 s |
| Ruby | 0.077474 s | 0.075759 s | 0.076616 s |

The pooled ratios are 97.398× Go for compiled Able, 8.655× Python for
bytecode Able, and 8.445× Ruby for bytecode Able. Compiled cohort means differ
by 14.8%, so conclusions use all ten workstation samples. The result remains
far outside the target under either cohort.

## Profile gate

Three compiled profiles put `bridge.currentGID` at 96.57% cumulative and
`runtime.Stack` at 95.71%. This repeats the established
compiled-concurrency owner whose fixed-context candidate already failed broad
serial and concurrent guards.

Three bytecode runtime profiles average `378861306 ns/op`, `64543432 B/op`,
and `744133` allocations/op. Their merged profile is split across call
dispatch, call-name dispatch, arithmetic, callee environments, member
dispatch, return completion, allocation, and nominal-field loads. The
215064-call trace shows every dominant application helper, inherent method,
and interface call already inline after four cold interface resolutions.

No exact new child is both dominant and separable across unlike applications.
A new call-name, frame, return, nominal-field, cache, or named-type shortcut
would therefore retry closed evidence or optimize this benchmark shape rather
than a general program boundary.

## Evidence

- `2026-07-23-concurrent-scene-tiles-go-{a,b}.json`
- `2026-07-23-concurrent-scene-tiles-interpreter-{a,b}.json`
- `2026-07-23-concurrent-scene-tiles-cohort-{a,b}.json`
- `2026-07-23-concurrent-scene-tiles-compiled-profile-top.txt`
- `2026-07-23-concurrent-scene-tiles-bytecode-profile-top.txt`
- `2026-07-23-concurrent-scene-tiles-bytecode-trace.json`

## Next recommendation

Add a tenth portable application using independently synthesized fixed-point
audio voices.

Why: the same two high-value interactions remain the shallowest frontier rows
at depth nine. Independent waveform voices exercise callbacks and interface
dispatch over time-series accumulation without repeating geometry, trees,
graphs, queues, pipelines, state machines, signals, or policy batches.

What it entails: four Futures each synthesize an integer sample stream;
nominal phase and mix-accumulator types expose inherent methods; a user-defined
oscillator interface selects different waveforms; and captured envelope
callbacks pass through that interface boundary. Implement exact Able, Go,
Python, and Ruby versions plus a schedule-independent verifier, establish
six-lane parity, run two five-process cohorts, and profile only after
correctness. Admit a production change only if one generic owner repeats
across unlike applications.
