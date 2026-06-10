# Compiled channel scheduler-payload fast path retained

Date: 2026-07-25

## Decision

Retain the general generated-channel boundary localization. A channel send or
receive now attempts its nonblocking operation before recovering the current
asynchronous scheduler payload. If it must block, that one recovered payload
supplies both the cancellation context and the executor's blocked/unblocked
bookkeeping.

This removes a compiled/runtime boundary from ordinary successful channel
operations without changing the compiled callable ABI, weakening cancellation,
or selecting a benchmark, container, nominal type, or source method. The
experimental execution-context ABI remains optional and unchanged.

No canonical `able-stdlib`, runtime package, interpreter, tree-walker, bytecode
VM, language, dependency, or WASM change was needed.

Machine-readable samples are in
`2026-07-25-compiled-channel-payload-fast-path-retained.json`.

## Baseline attribution

Fresh strict binaries were built for Concurrent Document Pipeline, Concurrent
Event Routing, and Concurrent Policy Callbacks. All three public verifiers
passed and `go version -m` confirmed that none of the binaries linked
`able/interpreter-go/pkg/interpreter`.

Five clean main-phase CPU profiles and three exact allocation-phase samples
were captured per application with:

- `GOMAXPROCS=1`;
- `GOMEMLIMIT=1GiB`;
- `GOGC=50`; and
- `ABLE_EXECUTOR=goroutine`.

The exact
`__able_current_payload -> bridge.(*Runtime).Env -> bridge.currentGID ->
runtime.Stack` path repeated in all three profiles:

| Application | Merged CPU | `__able_current_payload` | `bridge.currentGID` |
| --- | ---: | ---: | ---: |
| Concurrent Document Pipeline | 170 ms | 160 ms (94.12%) | 160 ms (94.12%) |
| Concurrent Event Routing | 1.48 s | 620 ms (41.89%) | 900 ms (60.81%) |
| Concurrent Policy Callbacks | 330 ms | 280 ms (84.85%) | 280 ms (84.85%) |

Caller attribution placed the shared payload lookup beneath generated channel
send and receive helpers in all three programs. Those helpers recovered the
payload before their first nonblocking attempt. A successful buffered operation
therefore paid for `runtime.Stack` even though it never needed cancellation or
blocked-task state.

## General correction

Generated channel send now:

1. validates the handle and channel state;
2. attempts the nonblocking send;
3. recovers the scheduler payload only after that attempt misses; and
4. reuses the payload for task context and one matched block/unblock pair.

Generated channel receive similarly checks the closed/empty and buffered-value
fast paths before payload recovery, then reuses the recovered payload around
the blocking select.

The previous helpers independently recovered the current payload for context,
`MarkBlocked`, and `MarkUnblocked`. The retained path performs at most one
recovery for an actually blocking channel operation and none for a successful
fast operation.

A generated-source regression test checks both ordering and single recovery.
The mechanically extracted mutex-awaitable renderer keeps the touched
concurrency generator at 991 lines; the new renderer is 68 lines. No mutex
semantics changed.

## Repeated wall-time gate

Each binary received two warmups. Twenty baseline/candidate pairs then ran on
CPU 9 with alternating order, followed by twenty source-equivalent Go runs per
application. All 180 timed processes passed the public verifier. Values are
arithmetic process means:

| Application | Baseline Able | Candidate Able | Change | Go | Candidate / Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Document Pipeline | 41.850 ms | 7.773 ms | -81.43% | 2.182 ms | 3.562x | 28.07% |
| Concurrent Event Routing | 311.582 ms | 177.099 ms | -43.16% | 2.746 ms | 64.489x | 1.55% |
| Concurrent Policy Callbacks | 80.974 ms | 14.930 ms | -81.56% | 2.288 ms | 6.526x | 15.32% |

The unweighted geometric-mean improvement across the three unlike applications
is 73.10%. Every baseline/candidate comparison has the same sign.

The large remaining Go ratios are explicit failures against the eventual
1.052632x target, not evidence against retaining this tranche. In particular,
Event Routing's residual profile exposes a different boundary described below.

## Allocation and profile confirmation

Three exact main-phase allocation samples per side also improve every row:

| Application | Baseline bytes | Candidate bytes | Change | Baseline allocations | Candidate allocations | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Document Pipeline | 1,876,600 | 1,612,877 | -14.05% | 22,912 | 18,794 | -17.97% |
| Concurrent Event Routing | 45,707,680 | 44,658,237 | -2.30% | 972,270 | 955,874 | -1.69% |
| Concurrent Policy Callbacks | 5,277,181 | 4,751,083 | -9.97% | 64,753 | 56,532 | -12.70% |

Twenty short candidate CPU profiles per application were merged to compensate
for the now-short Document and Policy executions. Document and Policy contain
no sampled `__able_current_payload`, `bridge.(*Runtime).Env`,
`bridge.currentGID`, or `runtime.Stack`. Event also contains no sampled
`__able_current_payload` or `bridge.(*Runtime).Env`; its remaining 1.22 seconds
of merged `bridge.currentGID` CPU is called by
`__able_compiled_entry_method_String_split`, not a channel helper.

One exact allocation-profile sample reinforces the separation:

| Application | Baseline `currentGID` allocations | Baseline payload allocations | Candidate `currentGID` allocations | Candidate payload allocations |
| --- | ---: | ---: | ---: | ---: |
| Concurrent Document Pipeline | 4,137 | 4,122 | 0 | 0 |
| Concurrent Event Routing | 24,603 | 16,396 | 8,207 | 0 |
| Concurrent Policy Callbacks | 8,259 | 8,240 | 0 | 0 |

Event's remaining 8,207 identity-recovery allocations belong to the
`String.split` package-environment entry wrapper. The 16,396 allocations under
the channel payload path disappear.

## Correctness gate

Passing gates include:

- generated channel fast-path ordering and single payload recovery;
- the compiled concurrency parity fixture matrix;
- goroutine await/future, blocked flush, mutex contention, and public mutex
  await-lock execution;
- experimental execution-context static kernel calls, fixed helper surface,
  and nested spawn execution;
- all three strict application verifiers;
- all three strict interpreter-dependency checks; and
- `go test ./cmd/ablec`.

One combined focused command reached its cumulative 60-second package timeout
while starting the public mutex-await test. This was not an individual-test
failure: the concurrency parity matrix passed alone in 15.875 seconds, the
other goroutine/future/mutex group in 5.063 seconds, and the named mutex-await
test in 0.746 seconds.

## Next

Refresh strict CPU and allocation profiles across Concurrent Event Routing,
Word Frequency, and Sensor Calibration, which are unlike applications with
known material `String.split` work. Confirm whether the exact generated
`String.split` package-entry wrapper and its
`SwapEnvIfNeeded -> bridge.currentGID -> runtime.Stack` path remain material in
all three current binaries.

If the same exact entry owner clears that gate, find the general
package-environment-effect or static-call lowering fact that keeps a
semantically independent imported inherent method behind its entry wrapper.
Advance one general correction only; do not add a method-name fast path,
reopen the broad execution-context ABI, or select String-heavy benchmarks in
production code. Measure repeated verifier-backed A/B cohorts against all
three source-equivalent Go applications.

This is next because the retained channel correction exposes that entry wrapper
as 36.64% of Event Routing's candidate CPU while Event still runs 64.489x
slower than Go. It entails proving whether the wrapper is conservative or
semantically required through all of `split`'s generated callees, then
localizing the environment boundary only if the proof is general. This matters
because it directly advances the goal of keeping compiled Able execution on
native Go carriers and crossing runtime context boundaries only where language
semantics require them.
