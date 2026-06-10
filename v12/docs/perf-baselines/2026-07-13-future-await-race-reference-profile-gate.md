# Future Await Race Reference and Profile Gate — 2026-07-13

## Decision

Keep no bytecode VM, compiler, bridge, runtime, canonical-stdlib, or benchmark
source change. Future Await Race is a material miss in both product lanes, so
it justified fresh attribution. That attribution adds no new removable,
cross-application concrete leaf.

## Pinned reference screen

Three independent processes per language and Able mode ran on CPU 15 with a
45-second cap, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Every process
was accepted by the canonical Ruby verifier and produced stdout hash
`33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4`.

| Mode | Able (s) | Go (s) | Able/Go | Python (s) | Able/Python | Ruby (s) | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| compiled | 0.1233 | 0.0038 | 32.45x | 0.0309 | 3.99x | 0.0506 | 2.44x |
| bytecode | 0.1700 | 0.0038 | 44.74x | 0.0309 | 5.50x | 0.0506 | 3.36x |

The source and verifier are the same cross-language application introduced in
the coverage tranche; no Docker image, stored external result, or synthetic
runtime loop supplied these ratios. The generated reference and comparison
reports were deliberately kept under `/tmp`, where they are cleanup-eligible;
this record preserves the process counts, guards, output hash, and values.

## Normal-process attribution

Because a one-P cap changes goroutine scheduling, the attribution runs kept
the CPU-15, 1-GiB, and 45-second guards but explicitly omitted
`GOMAXPROCS`. Eight independently launched, output-verified processes were
merged per mode. The bytecode captures are ordinary CLI executions, not the
stateful repeated-`main` helper; generated-binary captures use the CPU-only
main-phase profile hook and exclude compiler/bootstrap work.

The merged bytecode profile contains 650 ms of samples. Its material costs
are loader/parser work (52.3% cumulative loader, 41.5% parser, and 35.4% flat
tree-sitter cgo), followed by `bytecodeVM.runResumable` (16.9%),
`GoroutineExecutor.runTask` (13.8%), and `execBinary` (9.2%). Raw-integer
extraction is only 3.1% flat.

This does not repeat one concrete VM operation across the independent
concurrency applications:

- Channel Rollup is loader/text/member shaped: 40.0% loader, 45.0%
  `runResumable`, and 12.5% `execCallMember` cumulative.
- Future Pipeline is numeric-VM shaped: 83.3% `runResumable`, 75.0%
  executor task work, and 27.8% `execBinary` cumulative.
- Future Await Race shares their executor and resumable **parents**, but its
  workload-specific material child is parser/bootstrap work. Its small
  arithmetic and raw-integer leaves do not recur materially in all three.

The merged compiled generated-main profile has 280 ms of samples. It again
places 85.7% cumulatively in the generic
`bridge.currentGID` -> `runtime.Stack` goroutine-identity path, under 92.9%
`RunFuture` task work. The awaited paths (`__able_await_with_state`,
`__able_await_value`, and dynamic method/call helpers) are callers of that
same bridge wall rather than a new leaf. This independently confirms the
91.7% and 93.8% `currentGID`/`runtime.Stack` ownership already observed in
Channel Rollup and Future Pipeline.

The existing generic fixed-context ABI is the only known remedy for that
compiled bridge wall. It remains opt-in: its broad default gate found a stable
54.7% N-body regression, and the allocation-free package-linkage refinement
was separately rejected by a 16.6% K-Nucleotide regression. Re-running the
same rejected ABI experiment for this one additional concurrent application
would not be general evidence.

## Next

Treat Future Await Race as the 24th completed bytecode application row and
return to v12 semantic feature completion and feature-to-application coverage.
This is necessary because the newly added language boundary has now supplied
both product ratios and profiles, yet still adds no shared concrete performance
candidate. The next implementation tranche should add an independently useful
language boundary from the active roadmap, exercise it in fixtures and a
cross-language application, and profile it only after a material verifier-
backed miss appears. That is how the benchmark suite can grow without turning
one concurrency shape into a scheduler-specific optimization target.
