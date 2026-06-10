# Packed eager boxed-i32 cache feasibility

Date: 2026-07-21

## Decision

Reject and fully remove the pointer-backed extended boxed-`i32` cache
prototype. Packing the `16,385...262,143` segment into one eager
`[]runtime.IntegerValue` backing allocation removes 245,758 package-init
allocations and improves short compiled startup. It also reduces the live heap
by about 1.97 MB, however, and consistently increases collections in the
compiled allocation-light and allocation-heavy guards. TapeLang is 3.48%
slower across ten order-balanced pairs and Binary Trees performs 2.13% more GC
cycles across three pairs.

The gate explicitly required no GC-count regression because earlier lazy-cache
work traded startup for slower real allocation-heavy programs. The candidate
therefore fails even though its Binary Trees wall is neutral and its ordinary
bytecode rows are neutral-to-positive. No production, stdlib, application,
fixture, benchmark, reference, or language change is retained.

## Handoff reconciliation

The incoming recommendation said to begin with a packed raw-`i32` cache. That
exact representation had already been implemented and rejected on 2026-07-17,
followed by static-table, cache-distribution, call-site, and producer/consumer
gates. Repeating it would not produce new evidence.

This tranche instead tested the next distinct segment. The extended boxed-`i32`
cache owns 245,759 `runtime.IntegerValue` boxes, and the VM's central integer
extractors already recognize both `runtime.IntegerValue` and
`*runtime.IntegerValue`. The prototype retained the existing eager
`[]runtime.Value` lookup table, added one stable concrete backing slice, and
made each interface entry point at its corresponding immutable cell. No new
type case was added to the hot extractor.

## Initialization result

Preserved baseline and candidate interpreter test executables each received
five no-test `GODEBUG=inittrace=1` launches:

| State | Mean init clock | Mean init bytes | Init allocations |
| --- | ---: | ---: | ---: |
| Baseline | 28.4 ms | 38,003,403.2 | 707,336 |
| Packed extended i32 | 26.2 ms | 36,037,377.6 | 461,578 |
| Change | -7.75% | -1,966,025.6 (-5.17%) | -245,758 (-34.74%) |

The allocation reduction is exact. The backing slice replaces one allocation
per extended value, while the byte reduction reflects removal of allocator
size-class overhead around the former individually boxed 40-byte values.

Seven fixed one-million-iteration samples also confirmed why this candidate
was worth reaching the application gate: the already-existing pointer branch
in `bytecodeDirectSmallI32Value(...)` averaged 1.315 ns/op in the baseline
binary versus 6.258 ns/op for the value branch. Both are allocation-free. The
candidate did not add a dispatch branch; its risk was changed representation
and heap pacing.

## Repeated application gate

Every sample below passed its external verifier and retained one output hash
per application. Compiled controls used `GOMAXPROCS=1`, `GOGC=50`, a 1-GiB
memory limit, and `gctrace`; variants alternated order. The normal bytecode
controls used `GOMAXPROCS=1` and the default GC percentage.

| Workload | Mode | Pairs | Baseline mean | Candidate mean | Wall change | Baseline GC | Candidate GC |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Array Slice Window | compiled | 10 | 0.062 s | 0.058 s | -6.45% | 3.90 | 4.00 |
| TapeLang Alphabet | compiled | 10 | 3.677 s | 3.805 s | +3.48% | 3.20 | 4.00 |
| Binary Trees | compiled | 3 | 29.843 s | 29.817 s | -0.09% | 172.33 | 176.00 |
| Word Frequency | bytecode | 5 | 1.310 s | 1.278 s | -2.44% | 6.00 | 5.00 |
| Matrix Multiply | bytecode | 5 | 4.336 s | 4.332 s | -0.09% | 11.20 | 11.00 |

All workstation samples remain in the arithmetic means. In particular,
TapeLang's candidate range is 3.39-5.41 seconds; expanding the initial volatile
five-pair cohort to ten pairs did not remove the outlier and left the candidate
3.48% slower. Mean RSS falls from 48,191 KiB to 45,448 KiB in TapeLang and from
316,071 KiB to 291,829 KiB in Binary Trees, confirming that the additional
collections come from the smaller live heap rather than failed output or extra
application allocation.

The fixed-cache representation cannot recover those heap goals without unused
capacity or padding whose purpose is GC ballast. That would violate the gate's
prohibition on fake ballast and runtime-policy manipulation.

## Verification and cleanup

- Focused fixed/dynamic boxing, raw integer, slot, index, cast, and type tests
  passed with the candidate.
- A package-wide candidate run reached the repository's existing 60-second
  `TestExecFixtures/06_12_28_stdlib_fs_lines` ceiling without reporting a
  candidate failure before the timeout.
- After removal, `go test ./pkg/interpreter -run TestBytecode -count=1
  -timeout 60s` passes in 24.336 seconds.
- The packed backing slice, pointer population, and comments are absent from
  the restored production source.
- Temporary executables, generated trees, traces, and raw timing files live
  only under `/tmp` and are not project artifacts.

## Next recommendation

Run a dynamic boxed-integer reuse and ownership expansion gate across Fixed
Width 128, Reverse Complement, K-Nucleotide, and Rational Series, with Matrix
Multiply and Word Frequency as guards.

Why: the six-program bytecode exact-leaf sweep found
`bytecodeBoxedIntegerValue(...)` materially in Fixed Width and Reverse
Complement, one application short of the three-unlike-program admission rule.
K-Nucleotide and Rational Series are independent map/numeric consumers with
known large integer working sets. They can establish whether the remaining
dynamic cache map/lock path is a genuinely shared VM wall. This tranche closes
packed fixed/eager cache representations; the next work should not retry them.

What it entails: use the existing opt-in `able_bytecode_box_reuse` diagnostic
build to count per-kind dynamic lookups, hits, inserts, capacity misses, and
`i64` bypasses without retaining counters. Pair those counts with clean CPU
profiles that distinguish fixed/extended hits from the dynamic map and lock.
Advance only if the same dynamic mechanism is material in at least three
unlike applications and reuse is high enough to support one generic cache
change. Then require repeated normal-bytecode numeric/map/text guards plus
compiled startup and Binary Trees GC controls. Do not change fixed eager
storage, GC policy, named containers, applications, or WASM.
