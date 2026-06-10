# Concurrent Packet Codecs application gate — 2026-07-23

## Decision

Retain the source-equivalent `concurrent_packet_codecs` application, exact
verifier, catalog and coverage memberships, two complete timing cohorts, and
bounded profiles.

Retain no compiler, interpreter, tree-walker, canonical-stdlib, language,
dependency, named-container, non-primitive nominal, or WASM change. The tested
environment-independence candidate was reverted because a typed callable can
originate in an arbitrary captured package environment; bypassing the guarded
entry path would not have a sufficient semantics proof.

The application raises both minimum-depth interactions from ten to eleven:

- concurrency × functions/closures × interface dispatch;
- functions/closures × inherent methods × interface dispatch.

It does so with four independent fixed-width packet streams and two codec
families, not audio, geometry, trees, graphs, queues, pipelines, state
machines, signals, or policy batches.

## Correctness and source equivalence

Able, Go, Python, and Ruby each decode four deterministic 4,096-field streams.
Two lanes select delta codecs and two select run-length codecs through the
`PacketCodec` interface. Every field crosses a captured validator callback,
and nominal cursor/stat values evolve through inherent methods. The canonical
and external Able sources are byte-identical.

All six execution lanes produce:

```text
16384:16384:14631:10237:2,50,2981,71:65529,65533,9377,65525:478531,117243,169364,286880:610256,204520,462962,420944:695740,94433,664404,109032:8192,3072,8192,3075:52196,22170,3486,31658:0,0,0,0:52015:505771
```

The verifier-backed stdout SHA-256 is
`cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5`.

## Repeated timing

Every retained lane has ten successful samples across two independent
five-process cohorts.

| Lane | Cohort A | Cohort B | Pooled arithmetic mean |
| --- | ---: | ---: | ---: |
| Able compiled | 0.506 s | 0.510 s | 0.508 s |
| Go | 0.0038 s | 0.0037 s | 0.003733 s |
| Able bytecode | 0.882 s | 0.716 s | 0.799 s |
| Python | 0.0831 s | 0.0812 s | 0.082143 s |
| Ruby | 0.0785 s | 0.0824 s | 0.080415 s |

The pooled ratios are 136.09× Go for compiled Able, 9.73× Python for bytecode
Able, and 9.94× Ruby for bytecode Able. The bytecode cohort means differ by
23.18%, so the ten-process arithmetic mean—not either individual cohort—is the
retained workstation estimate. Every independent cohort remains a clear target
miss.

## Profile and candidate gate

Three compiled profiles put both `bridge.currentGID` and `runtime.Stack` at
97.12% cumulative. This repeats the same environment-swap wall already seen in
audio voices, scene tiles, and graph visitors. Packet codec dispatch makes the
owner especially clear: both concrete codec adapters and their guarded entry
functions sit directly beneath it.

A generic candidate allowed proven environment-independent native interface
implementations to enter raw compiled bodies. The packet and audio methods
cannot receive that proof because they call typed callable parameters, whose
values may come from arbitrary captured package environments. Weakening the
proof would trade language correctness for benchmark speed, so the candidate
was fully reverted. The next safe compiler design is to propagate the existing
explicit execution context through native interface adapters, allowing the
callee and callback environments to remain explicit without a
`runtime.Stack` goroutine-ID lookup.

Three bytecode runtime profiles average `548944587 ns/op`, `85525547 B/op`,
and `2023219` allocations/op. Their merged profile is distributed across
call/name/member dispatch, binary arithmetic, callee environments, return
completion, allocation/GC, and load-slot work. No single new bytecode leaf is
both dominant and separable from the existing general VM owners, so no
bytecode candidate clears the broad gate.

## Evidence

- `2026-07-23-concurrent-packet-codecs-go-{a,b}.json`
- `2026-07-23-concurrent-packet-codecs-interpreter-{a,b}.json`
- `2026-07-23-concurrent-packet-codecs-cohort-{a,b}.json`
- `2026-07-23-concurrent-packet-codecs-compiled-profile-top.txt`
- `2026-07-23-concurrent-packet-codecs-bytecode-profile-top.txt`

## Next recommendation

Implement explicit execution-context propagation through generated native
interface adapters, then guard it across packet codecs, audio voices, scene
tiles, graph visitors, and serial interface/callback applications.

Why: four unlike concurrent applications now spend more than 93% of compiled
profile time recovering a goroutine ID through `runtime.Stack`. The existing
context prototype already carries task-local environment and payload state
through ordinary generated calls; adapters are the missing boundary. This is
the largest repeated compiled wall and is far more promising than another
small call-frame or coercion tweak.

What it entails: add context-aware native interface entry methods or adapter
siblings; pass the caller context through compiled interface dispatch; derive
a context from the native call context only at dynamic/runtime boundaries;
retain guarded compatibility entry points; and test captured callbacks,
cross-package globals, nested Futures, cancellation, and concurrent
environment isolation. Benchmark two five-process cohorts for the four
compiled concurrency applications plus serial interface/callback guards.
Admit the change only if correctness stays green and the averaged improvement
is broad. Update canonical `able-stdlib` only for a reusable API or correctness
defect, and do not begin WASM work.
