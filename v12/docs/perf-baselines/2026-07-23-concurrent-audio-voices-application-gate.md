# Concurrent Audio Voices application gate — 2026-07-23

## Decision

Retain the source-equivalent `concurrent_audio_voices` application, exact
verifier, catalog and coverage memberships, two complete timing cohorts, and
bounded profiles. Retain the generic bytecode concurrency correctness fix:
multi-thread VMs use immutable raw integer value carriers while single-thread
VMs retain mutable carrier reuse.

Retain no compiler, tree-walker, canonical-stdlib, language, dependency,
named-container, non-primitive nominal, or WASM change. Admit no additional
performance candidate from this application.

The application raises both minimum-depth interactions from nine to ten:

- concurrency × functions/closures × interface dispatch;
- functions/closures × inherent methods × interface dispatch.

It does so with four independent fixed-point waveform voices, not geometry,
trees, graphs, queues, pipelines, state machines, signals, policy batches, or
callback transforms.

## Correctness and source equivalence

Able, Go, Python, and Ruby each synthesize four 8,192-sample integer voices.
Each lane selects one of two oscillator implementations through an interface,
passes a captured envelope callback through that interface for every sample,
and evolves nominal phase and mix-accumulator values through inherent methods.
The canonical and external Able sources are byte-identical.

All six execution lanes produce:

```text
32768:32768:16293:3165:1400,1360,1580,1962:341289,608515,436381,456054:449883,9195,50251,344485:927166,319343,609816,158142:64274,64161,64172,60078:166,256,336,412:842236:242361
```

The verifier-backed stdout SHA-256 is
`6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28`.

## Bytecode concurrency defect and fix

The first bytecode cohort failed all five output verifications with different
hashes. Twelve direct repetitions reproduced schedule-dependent corruption.
A race-detector build identified mutable raw integer slot/stack cells being
read and rewritten by different future VMs.

Raw scalar carriers are VM scratch storage and cannot safely cross concurrent
ownership boundaries. Multi-thread bytecode execution now uses immutable raw
integer value carriers for i32, i64, other fixed-width integer slots/stacks,
inline argument copies, and return scratch. The existing mutable cells remain
available in single-thread execution, preserving that fast path.

A source-level regression repeatedly calls the same integer-hot function from
four goroutine-backed Futures and checks independent results. Focused unit
tests, twelve full application repetitions, and a race-detector application
run pass. The invalid first bytecode measurements were discarded and cohort A
was rerun from scratch.

## Repeated timing

Every retained lane has ten successful samples across two independent
five-process cohorts.

| Lane | Cohort A | Cohort B | Pooled arithmetic mean |
| --- | ---: | ---: | ---: |
| Able compiled | 0.926 s | 0.942 s | 0.934 s |
| Go | 0.004900 s | 0.004500 s | 0.004722 s |
| Able bytecode | 1.290 s | 1.260 s | 1.275 s |
| Python | 0.133500 s | 0.128500 s | 0.130977 s |
| Ruby | 0.124000 s | 0.123900 s | 0.123948 s |

The pooled ratios are 197.818× Go for compiled Able, 9.735× Python for
bytecode Able, and 10.287× Ruby for bytecode Able. Both independent cohorts
and their pooled means remain clear target misses.

## Profile and candidate gate

Three compiled profiles put `bridge.currentGID` at 96.44% and `runtime.Stack`
at 96.07% cumulative. This repeats the established compiled-concurrency owner
whose fixed-context candidate already failed broad serial and concurrent
guards.

Three bytecode runtime profiles average `999272925 ns/op`, `181197235 B/op`,
and `4103104` allocations/op. Their merged profile is distributed across
call/name/member dispatch, arithmetic, callee environments, return completion,
allocation, runtime-data synchronization, lookup, and nominal field loads.
The 507,928-call trace shows all dominant helpers and inherent methods inline;
32,764 of 32,768 oscillator interface calls are inline, with only four cold
generic resolutions.

The new correctness rule makes raw integer transport visible, but the profile
does not establish a safe cross-VM mutable-owner design. Reintroducing mutable
cells without an explicit ownership/lifetime proof would recreate the race.
No other exact child is both dominant and separable across unlike
applications. The generality gate therefore admits no additional production
or stdlib optimization.

## Evidence

- `2026-07-23-concurrent-audio-voices-go-{a,b}.json`
- `2026-07-23-concurrent-audio-voices-interpreter-{a,b}.json`
- `2026-07-23-concurrent-audio-voices-cohort-{a,b}.json`
- `2026-07-23-concurrent-audio-voices-compiled-profile-top.txt`
- `2026-07-23-concurrent-audio-voices-bytecode-profile-top.txt`
- `2026-07-23-concurrent-audio-voices-bytecode-trace.json`

## Next recommendation

Add an eleventh materially different portable application using independent
fixed-width packet codec streams.

Why: the same two high-value interactions remain the shallowest frontier rows,
now at depth ten. Packet decoding and encoding can raise both without
repeating audio, geometry, trees, graphs, queues, pipelines, state machines,
signals, or policy/callback batches. It also puts array traversal and
fixed-width integer conversion under the same callable/interface topology,
which can confirm whether immutable concurrent scalar transport is a shared
cost rather than an audio-specific profile effect.

What it entails: four Futures process deterministic integer packet streams;
nominal cursor and packet-stat types expose inherent methods; a user-defined
codec interface selects delta and run-oriented formats; and captured
validation/checksum callbacks cross the interface per decoded field. Implement
exact Able, Go, Python, and Ruby versions plus a schedule-independent verifier,
establish six-lane parity, run two five-process cohorts, and profile only after
correctness. Admit a production change only if one generic owner repeats
across unlike applications. Update canonical `able-stdlib` only for a reusable
API or correctness defect, and do not begin WASM work.
