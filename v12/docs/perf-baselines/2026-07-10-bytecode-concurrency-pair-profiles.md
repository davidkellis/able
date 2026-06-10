# Bytecode Concurrency Profile Pair

This bounded bytecode tranche tested whether Channel-Rollup's verified target
miss shares a concrete optimization target with an independent `spawn`/Future
workload. It keeps no VM, compiler, runtime, or stdlib implementation change.

## Method

- Pinned canonical stdlib; `GOMEMLIMIT=1GiB` and `GOGC=50`. No
  `GOMAXPROCS=1` cap was used: a one-P cap changes the behavior of the
  workload being evaluated.
- One normal bytecode process per workload, profiled through
  `ABLE_GO_CPU_PROFILE`; each run includes ordinary load/lower/run behavior
  and verified its program output. This is intentionally not the in-process
  repeated-`main` runtime benchmark, which is not a valid concurrency measure:
  it retained executor-sensitive state across calls and timed out under the
  otherwise bounded Channel-Rollup run.
- Channel-Rollup used `ABLE_EXECUTOR=goroutine` and produced
  `16384:4828:502100`. Its capture is
  `.profiles/20260710_channel_rollup_bytecode_process.cpu.pprof`: 344.73
  milliseconds elapsed and 700 milliseconds of CPU samples across goroutines.
- The full external BinaryTrees application at its standard `n=21` target did
  not complete inside the 55-second profile cap, so it supplied no usable CPU
  profile. The reduced `binarytrees_small` fixture (`n=12`) was used only as a
  call-tree control, with the goroutine executor, and produced the checked
  fixture output. Its capture is
  `.profiles/20260710_binarytrees_small_bytecode_process.cpu.pprof`: 279.59
  milliseconds elapsed and 1.31 seconds sampled across goroutines.
- Lexical-Rollup was the serial iterator/string guard and produced
  `16384:4828:502100`. Its capture is
  `.profiles/20260710_lexical_rollup_bytecode_process.cpu.pprof`: 310.46
  milliseconds elapsed and 430 milliseconds sampled.

## Result

Channel-Rollup's limited process sample shows goroutine task execution at
24.29% cumulative, generic call dispatch at 21.43%, loader work at 20.00%,
and atomic `Int32.Add` at 5.71% flat. Its remaining VM leaves are diffuse;
no raw-integer, lookup, return, or channel operation is material enough to
select alone.

The reduced BinaryTrees control also reaches the async parents, but its
material descendants are recursive struct construction, typed-pattern work,
inline returns (11.45% cumulative), and garbage collection. Atomic `Int32.Add`
is 8.40% flat there, but the capture is short and the full application remains
unavailable. The serial Lexical-Rollup guard has no async-task samples; its
largest attributable VM work is `runResumable` at 30.23% cumulative and call
dispatch at 13.95%, while load/parse and GC are also material. Its isolated
raw-integer sample is 2.33% flat.

The Channel-Rollup and reduced BinaryTrees samples repeat executor/task and
atomic work, but differ substantially beneath those broad parents; the serial
guard has neither task execution nor the same atomic leaf. The raw-value,
lookup, return, and type-match leaves likewise do not recur materially across
all three workloads. With the full BinaryTrees target timing out, the reduced
fixture cannot elevate a concurrency-only leaf to an application-level
candidate.

## Decision

Reject a runtime optimization from this evidence. In particular, do not tune
the scheduler parents, channel capacity, `runTask`, atomic bookkeeping, or a
single BinaryTrees call shape. The tree-walker goroutine-executor repair is
now complete: race-clean runtime and interpreter regressions plus the
cross-language Channel-Rollup Docker run verify the same concurrent path. The
next work is a bounded broad-suite cost check of that generic synchronization
boundary; only profile another performance candidate after two full
applications again expose the same material concrete leaf.
