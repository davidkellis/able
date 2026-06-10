# Bytecode VM Cross-Workload Profile Refresh

This is an evidence tranche, not a runtime change. It refreshes the warmed
bytecode VM view before selecting another candidate, using three deliberately
different Able programs: split/join text, lazy iterator collection, and numeric
array mapping. No benchmark-shaped opcode, compiler lowering rule, or stdlib
special case was considered.

## Method

The direct bytecode runtime benchmark loads and warms each program before CPU
profiling. Runs used the canonical stdlib pinned at
`/home/david/sync/projects/able-stdlib/src`, with
`GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1` and CPU affinity `0`. The program
outputs and benchmark exit status were the correctness checks.

| Workload | Profiled iterations | Profiled result | Samples |
| --- | ---: | ---: | ---: |
| `string_split_join_small` | 5 | 932,707,268 ns/op; 49,499,721 B/op; 513,673 allocs/op | 4.65 s |
| `linked_list_iterator_collect_i64_small` | 5 | 238,128,389 ns/op; 3,254,272 B/op; 29,076 allocs/op | 1.18 s |
| `array_map_i32_small` | 20 | 86,498,142 ns/op; 850,834 B/op; 310 allocs/op | 1.71 s |

The longer numeric run is intentional: its five-iteration profile had too few
samples to distinguish a leaf from sampling noise. Retained profiles are
`20260710_string_split_join_bytecode_refresh_5x.cpu.pprof`,
`20260710_linked_list_iterator_collect_bytecode_refresh_5x.cpu.pprof`, and
`20260710_array_map_i32_bytecode_refresh_20x.cpu.pprof` under
`v12/interpreters/go/.profiles/`.

## Evidence

`execCallOpcode` appears in all three programs, but it is only a VM dispatcher
parent. Its material descendants do not match:

| Workload | Material descendants |
| --- | --- |
| Split/join | direct named calls, inline-return coercion/type matching, and string-keyed type-name lookup |
| Iterator collect | generator `next`, member/struct-call dispatch, native raw-result calls, and iterator scheduling |
| Numeric array/map | array-slot/member calls, inline calls, raw-i32 cache access, binary arithmetic, and raw integer transport |

The potential common subpaths are not a new actionable wall:

- `bytecodeRawIntegerValueInfo(...)` is 110 ms (2.4%) in split/join and
  80 ms (4.7%) in array/map, but received no sample in iterator collect. Its
  callers also differ: split/join includes store and bitwise paths, while
  array/map is mostly raw arithmetic/return transport. It is not a three-way
  candidate.
- `finishInlineReturn(...)` is 880 ms (18.9%) in split/join, 110 ms (9.3%) in
  iterator collect, and 130 ms (7.6%) in array/map. The descendants differ:
  split/join is return coercion/type alias work, collect is array coercion and
  generator returns, and array/map is i32-frame restoration/raw returns. The
  recently rejected slotless-return reorder is therefore not justified again.
- `runtime.mapaccess2_faststr` is visible in all three (470 ms/10.1%,
  130 ms/11.0%, and 90 ms/5.3%). It is reached through existing type-name and
  type-matching maps, not one common cache miss or validation descendant.
  `matchesTypeWithoutRuntimeValue(...)` is still present in every profile, but
  the corresponding small duplicate typed-pattern query was already tested and
  rejected; the current profiles reproduce that conclusion rather than expose a
  new semantic shortcut.

## Decision

Keep no runtime or stdlib code. The profile refresh does not meet the required
bar of one material, concrete VM descendant repeating across independent
feature families. In particular, do not target `execCallOpcode` as though it
were a leaf, do not retry the raw-cell/frame or return-guard experiments, and
do not change the known-type/type-match maps merely because their generic map
access is sampled.

## Next recommendation

Refresh the bounded external scorecard from the full verified Able/Go/Python/
Ruby program suite, then select a new bytecode profile pair only where a
material target gap and a concrete descendant recur in two independently
featured applications. This is preferable to another local VM micro-tweak:
the current representative text/iterator/numeric set rules out the candidate
families it was intended to adjudicate. The work entails preserving the pinned
toolchain and OOM/CPU guardrails, recording only reproducible comparison rows,
and profiling the selected pair plus an unrelated guard before considering one
generic interpreter, compiler, or canonical-stdlib change.
