# Compiled Structural-Control Profiles

## Decision

Keep no compiler, runtime, canonical-stdlib, or benchmark-source performance
change. Fresh collector-free `main` profiles split into three distinct costs:
generated recursive Fib, Sudoku's string-byte materialization and formatting,
and QuickSort's partition loop plus checked integer arithmetic. No concrete
lowering or bridge leaf is material in two independent workloads.

This tranche does retain one profiling-only facility. Generated binaries now
support `ABLE_GO_PHASE_CPU_PROFILE_DIR`, which writes separate bootstrap and
`main` CPU profiles without exact allocation sampling, allocation snapshots,
or forced GCs. The existing `ABLE_GO_PHASE_PROFILE_DIR` remains the exact
allocation mode and is unchanged. The two phase modes, and either mode with
`ABLE_GO_CPU_PROFILE`, are mutually exclusive.

## Why CPU-only phase profiling was necessary

`ABLE_GO_PHASE_PROFILE_DIR` deliberately sets `runtime.MemProfileRate` to one
and writes allocation snapshots at phase boundaries. That is useful for exact
allocation attribution, but its `runtime.profilealloc` and stack-walk work
dominated Sudoku's short allocation-heavy `main` samples. Ordinary
`ABLE_GO_CPU_PROFILE` also cannot isolate a sufficiently short main phase. The
new CPU-only mode keeps the useful launcher/main boundary while leaving normal
allocation behavior intact.

## Method

- Rebuilt current compiled Fib, Sudoku, and QuickSort binaries with the
  canonical external stdlib and the CPU-only phase hook.
- Ran the binaries from their exact sibling `../benchmarks/{fib,sudoku,quicksort}`
  directories. This is required: the public Sudoku and QuickSort corpora are
  suite-local inputs, not the separate Able example-fixture files.
- Used `taskset -c 2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a
  45-second process guard. All outputs were deterministic: Fib 3/3,
  Sudoku 60/60, and QuickSort 3/3 launches had one stdout hash each.
- Merged only each process's `main.cpu.pprof`. Retained inputs, generated
  binaries, run hashes, and merged profiles are under
  `v12/tmp/compiled-structural-profiles/phase-cpu-external/`.

The initially collected exact-allocation Sudoku CPU profile is not used here:
it measures the allocation collector rather than normal application execution.

## Results

| Workload | Recorded main CPU samples | Material descendants | Interpretation |
| --- | ---: | --- | --- |
| Fib | 9.56 s | generated `fib` 99.5% flat | Direct recursive numeric control; no shared bridge/helper cost is sampled. |
| Sudoku | 2.52 s | `parse_board` 68.3% cumulative; `String.bytes`/`validated_bytes` 59.1%; string conversion 23.8%; `board_to_string` 18.3% | Text-byte materialization, per-byte runtime values, and repeated formatting/allocation. |
| QuickSort | 5.22 s | generated `quicksort` 73.6% cumulative; `swap` 13.4% flat; checked multiply 13.2%; parse 24.7% | Primitive mutable partitioning and numeric-input parsing. |

Sudoku's generated `parse_board` spends most of its time obtaining a
`String.bytes` iterator. The current generic implementation creates a
`runtime.Value` array and a small integer value for every byte. Its separate
`board_to_string` path repeatedly formats integers and concatenates strings.
QuickSort instead spends its measurable non-partition time in checked decimal
accumulation; Fib is the recursive function body itself. None of those leaves
repeat across the three profiles, so changing any one would be a
source-shape-specific optimization.

## Verification

```text
cd v12/interpreters/go
go test ./pkg/profilehook -count=1 -timeout 90s
go test ./pkg/compiler -run '^TestCompilerMainUsesInstalledStdlibDiscoveryBeforeSiblingLookup$' -count=1 -timeout 90s
```

Both passed. The rebuilt generated binaries also completed under the stated
guards with deterministic output hashes.

## Next recommendation

Profile the string/byte overlap next: use the CPU-only phase mode on
I-Before-E and Base64, with JSON as a guard, and compare them to Sudoku's
`String.bytes`/runtime-value conversion descendants. Why: Sudoku identifies a
large generic string-to-byte boundary, while I-Before-E and Base64 are already
independent compiled-Go misses in the external scorecard. The work entails
bounded output-checked profiles from their exact external suite directories,
then a candidate only if the same concrete helper is material in at least two
of those applications and preserves the JSON control. Do not special-case
Sudoku, a file name, a codec, or a nominal container.
