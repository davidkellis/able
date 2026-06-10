# Await Channel Mux Reference and Profile Gate — 2026-07-13

## Decision

Keep no bytecode VM, compiler, bridge, runtime, canonical-stdlib, or
benchmark source change. Await Channel Mux is a material miss in both product
lanes, so it justified new attribution. The profiles repeat only known parent
costs or workload-specific Channel-Awaitable leaves; they do not authorize a
new generic optimization.

## Pinned reference screen

Three independent processes per language and Able mode ran on CPU 15 with a
45-second cap, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Every Able
process was accepted by the canonical Ruby verifier and produced stdout hash
`0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693`.

| Mode | Able (s) | Go (s) | Able/Go | Python (s) | Able/Python | Ruby (s) | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| compiled | 0.4367 | 0.0047 | 92.91x | 0.1193 | 3.66x | 0.0987 | 4.42x |
| bytecode | 0.2700 | 0.0047 | 57.45x | 0.1193 | 2.26x | 0.0987 | 2.74x |

The local Go/Python/Ruby reference and Able comparison reports remain under
`/tmp` and are cleanup-eligible. This record retains their process counts,
guards, output hash, and values without preserving disposable benchmark
workspaces.

## Normal-process attribution

For profiling, the CPU-15, 1-GiB, and 45-second guards remained in place, but
`GOMAXPROCS=1` was deliberately omitted because forcing one P changes the
goroutine/channel scheduling being examined. Eight independently launched,
output-verified processes were merged for each Able mode. The bytecode capture
is an ordinary CLI process, not the stateful repeated-main helper. The
generated-binary capture uses the CPU-only main-phase hook, so compiler and
bootstrap work are excluded from compiled attribution.

Retained merged profiles:

- `v12/interpreters/go/.profiles/20260713_await_channel_mux_bytecode_process_merged.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_await_channel_mux_compiled_main_merged.cpu.pprof`

The merged bytecode profile has 1.80 seconds of samples. Its material parents
are `bytecodeVM.runResumable` (42.22% cumulative), loader work (36.67%), and
`GoroutineExecutor.runTask` (36.11%). Parsing remains visible (20.00%
cumulative, with 15.56% flat `runtime.cgocall`). The Await-specific concrete
paths are smaller: `runAwaitExpression` is 8.33% cumulative,
`channelAwaitable.toStruct` is 5.56% flat, and `channelAwaitable.commit` is
5.00% cumulative.

This does not create a reusable bytecode candidate:

- Channel Rollup has loader/text/member costs, with `execCallMember` at 12.5%
  cumulative rather than a material Channel-Awaitable leaf.
- Future Pipeline is numeric-VM shaped, with `execBinary` at 27.8%
  cumulative.
- Future Await Race has loader/parser bootstrap as its material child; its
  smaller arithmetic leaves do not recur in the unlike concurrency programs.

Await Channel Mux shares their VM, executor, loader, and parser parents, not a
large concrete descendant. Its only new visible leaf is specifically the
public Channel Awaitable protocol. A Channel-only shortcut would violate the
generality policy and would not be evidence for ordinary programs.

The merged compiled main profile has 2.64 seconds of samples. It repeats the
known generic bridge wall: `bridge.currentGID` is 87.88% cumulative and
`runtime.Stack` is 87.50%, below the generic goroutine-task parent at 84.09%.
The await and Channel helpers are callers of that same cost, not a new leaf.
This independently agrees with Channel Rollup (91.7% `currentGID`), Future
Pipeline (93.8%), and Future Await Race (85.7%).

The existing fixed-context ABI is the only known general remedy for this
compiled bridge cost. It remains opt-in because its broad default scorecard
regressed N-body by 54.7%, and the allocation-free package-linkage refinement
regressed K-Nucleotide by 16.6%. Re-running either rejected candidate for this
one additional concurrent application would be benchmark-shaped evidence, not
a generality gate.

## Next

Treat Await Channel Mux as the 25th completed bytecode application row and
return to v12 semantic feature completion and feature-to-application coverage.
This is necessary because the new public Channel/Awaitable boundary has now
provided correct execution, pinned product ratios, and normal-scheduler
profiles, but it adds no broadly shared removable cost. The next tranche should
select an independently useful unfinished language boundary, cover it with
fixtures and a cross-language application, then profile only if it produces a
material verifier-backed miss.
