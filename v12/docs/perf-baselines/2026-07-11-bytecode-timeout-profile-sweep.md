# Bytecode Timeout Profile Sweep

## Method

This sweep profiled the four programs that timed out in the corrected external
generality scorecard:

- `binarytrees` (goroutine executor)
- `quicksort`
- `nbody`
- `tapelang_alphabet`

Each used one ordinary bytecode process, its real external input/working
directory, CPU affinity `2-3`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and a
120-second timeout. CPU, cumulative allocation, and heap profiles were
captured through `ABLE_GO_{CPU,ALLOC,MEM}_PROFILE`. The timeout sends `SIGINT`,
which lets the profiler flush before the process exits. These are intentionally
cold-process measurements: they include normal load/lower/run behavior rather
than hiding it behind a repeated-main microbenchmark.

The profile files are retained under `.profiles/` as
`20260711_external_<workload>_bytecode_process.{cpu,alloc,heap}.pprof`; their
matching executable workdirs are under `v12/tmp/perf-profile-20260711-*`.

## Results

| Workload | Main allocation wall | Retained heap at cap | CPU shape |
| --- | --- | ---: | --- |
| BinaryTrees | `execStructLiteralNamedFast` → `NewStructInstancePositionalSized` | 811 MB | 99.4% of 12.2 GB cumulative allocation and 98.9% of 102.9M objects are named struct instances; concurrent task execution and GC are material. |
| QuickSort | i32 slot/array snapshots and boxing | 365 MB | 17.6 GB cumulative; array-read comparison/swap plus `bytecodeBoxedIntegerValue` (39.8% space) and primitive-array wrappers (25.9%). Its retained heap is also dominated by the input array/file path. |
| NBody | raw-float materialization/normalization | 38 MB | 6.4 GB cumulative; `bytecodeMaterializeRawFloatValue` (32.5%), normalized raw-float slot values (22.8%), and division (13.2%). Generic call/lookup/return work is visible in CPU but not a dominant allocation leaf. |
| Tapelang | evaluation-stack growth while loading slot/struct-field values | 2.44 GB | 5.34 GB cumulative; `appendSlotStackValueChecked` (54.8%) and `execLoadSlotStructField` (31.2%) allocate backing stack storage. CPU is dispatch/member/array-slot heavy. |

All four runs reached the 120-second timeout without a process failure. The Go
memory limit is advisory; Tapelang's final heap snapshot exceeds it because its
live stack-backed data continues growing until the timeout.

## Interpretation

There is no single material concrete VM leaf across two independently shaped
programs:

- BinaryTrees is recursive named-struct construction, not array scalar flow.
- QuickSort's i32 boxing/array snapshots do not recur in NBody's float lane.
- NBody's float materialization does not recur in the two structural programs.
- Tapelang's retained evaluation stack does not appear in the other three;
  its two large allocation lines are literal `vm.stack = append(...)` sites.

The broader category “raw scalar materialization” appears in both numeric
programs but through different primitive paths. The generic raw-cell call-name
preservation experiment was already rejected by broad guards, so this evidence
does not justify reviving an integer- or float-specific variation. Likewise,
the Tapelang stack result alone does not authorize a stack special case or
interpreter-language shortcut.

Decision: keep no runtime, compiler, or stdlib code change. No
`able-stdlib` source changed.

## Next Evidence Needed

Before changing stack/value handling, add optional bytecode diagnostics for
maximum value-stack depth, call-frame depth, and stack-capacity growths, then
collect them on Tapelang plus one independently shaped long-running
field/destructuring workload. That distinguishes a generic VM stack-retention
defect from program-specific recursion/data growth without changing observable
language behavior. A candidate still requires the same concrete leaf to recur
in at least two workloads and a broad scorecard after the change.
