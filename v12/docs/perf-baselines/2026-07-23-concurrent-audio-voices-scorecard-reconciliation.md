# Concurrent Audio Voices scorecard reconciliation — 2026-07-23

## Decision

Promote `concurrent_audio_voices` in both the compiled and bytecode selections.
Retain its source-equivalent six-lane application, exact verifier, two
five-process cohorts per lane, bounded profiles, coverage evidence, and
closure records.

Retain the generic bytecode concurrency correctness fix: multi-thread VMs use
immutable raw integer value carriers, while single-thread VMs retain mutable
carrier reuse. Retain no compiler, tree-walker, canonical-stdlib, language,
dependency, named-container, non-primitive nominal, or WASM change. No
additional performance candidate is admitted.

## Measurement and scorecard

All 50 retained timed processes verify. Pooled arithmetic means are `0.934s`
compiled Able versus `0.0047215052s` Go (`197.818x`), and `1.275s` bytecode
Able versus `0.1309773797s` Python (`9.735x`) and `0.1239480368s` Ruby
(`10.287x`). Both independent cohorts and their pooled means remain clear
target misses, so the classification is not sensitive to workstation noise.

The promoted scorecard has 60 applications, 120 full-status rows, and 113
selected rows: 60 compiled and 53 bytecode. Every selected row has five
successful Able samples and five successful reference samples. The selection
manifest SHA-256 is
`4f79a5f4f40d55a96b37df641bef6ba409074eca647b1ffc45b5d7bdef41b947`.

The regenerated performance frontier has eight target meets, 105 misses, five
established guards, zero actionable local groups, and
`182.38084210526316` seconds of summed target excess. The weighted feature
interaction frontier has no zero-depth or depth-one triple and minimum depth
ten. `concurrent_audio_voices` raises both former minimum-depth interactions:

- concurrency × functions/closures × interface dispatch;
- functions/closures × inherent methods × interface dispatch.

## Correctness and candidate gate

The first bytecode cohort failed every verifier run with
schedule-dependent output corruption. Repeated direct runs and the race
detector traced this to mutable raw integer carrier cells crossing Future VM
ownership boundaries. Multi-thread bytecode execution now uses immutable
integer carriers for slots, stacks, inline arguments, and return scratch.
A focused four-Future regression, twelve exact application repetitions, and a
race-detector application run pass.

Three compiled profiles put `bridge.currentGID` at 96.44% and `runtime.Stack`
at 96.07% cumulative, repeating the closed compiled-concurrency owner. Three
bytecode runtime profiles average `999272925 ns/op`, `181197235 B/op`, and
`4103104` allocations/op. Bytecode cost remains distributed across
call/name/member dispatch, arithmetic, callee environments, return
completion, allocation, runtime synchronization, lookup, and nominal field
loads. The 507,928-call trace shows the hot application methods already
inline after four cold generic interface resolutions.

The new correctness rule makes concurrent scalar transport visible, but no
safe cross-VM mutable-owner design has been established. No other exact child
is both dominant and separable across unlike applications. The generality
gate therefore admits no additional runtime or stdlib optimization.

## Closure and architecture reconciliation

The new application changed the compiled- and bytecode-concurrency evidence.
The production correctness fix also changed the bytecode-production scope,
selecting all ten bytecode-dependent closures plus compiled concurrency: 11
closures in total. The ledger's pre-reconciliation report preserves that
selection. After the focused concurrency/raw-integer guard passed, the shared
scope baseline was rebuilt as one coordinated unit; the current ledger
contains 21 closures and zero invalidations. The complete project suite is the
final acceptance gate for that refresh.

The deterministic architecture/ABI dependency chain was regenerated in
order. Its decisions remain unchanged: no current local cross-engine
mechanism, semantic-region tier, native-tier prototype, portable foreign
backend, shared-runtime production migration, or closed-region production
cutover is admitted. The bytecode native proxy now spans 53 applications and
still leaves 42 target misses, while the structural strategy retains zero
concrete admitted routes.

## Verification

- exact output parity across tree-walker, bytecode, compiled Able, Go, Python,
  and Ruby;
- two complete verifier-backed five-process cohorts per runtime lane;
- three compiled and three bytecode profiles plus a bytecode call trace;
- focused bytecode concurrency/raw-integer tests and a race-detector run;
- catalog, selection, coverage, operation-depth, matrix, triple, scorecard,
  frontier, closure-ledger, and architecture dependency checks;
- JSON, source-identity, source-line, formatting, and whitespace checks;
- the complete v12 suite through
  `GOMEMLIMIT=1GiB GOGC=50 ./run_all_tests.sh`.

## Next recommendation

Add an eleventh materially different portable application using independent
fixed-width packet codec streams.

Why: the same two high-value interactions remain the shallowest frontier rows,
now at depth ten. Packet decoding and encoding can raise both without
repeating audio, geometry, trees, graphs, queues, pipelines, state machines,
signals, or policy/callback batches. It also puts array traversal and
fixed-width integer conversion under the same callable/interface topology,
which can test whether immutable concurrent scalar transport is a shared cost
rather than an audio-specific profile effect.

What it entails: four Futures process deterministic integer packet streams;
nominal cursor and packet-stat types expose inherent methods; a user-defined
codec interface selects delta and run-oriented formats; and captured
validation/checksum callbacks cross the interface per decoded field. Implement
exact Able, Go, Python, and Ruby versions plus a schedule-independent verifier,
establish six-lane parity, run two five-process cohorts, and profile only after
correctness. Admit a production change only if one generic owner repeats
across unlike applications. Update canonical `able-stdlib` only for a reusable
API or correctness defect, and do not begin WASM work.
