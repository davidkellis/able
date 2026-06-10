# Bytecode K-Nucleotide and Reverse Complement Profile Refresh — 2026-07-13

## Decision

Keep no VM, compiler, runtime, canonical-stdlib, or benchmark-source change.
The two newly selected severe bytecode misses do not repeat a concrete leaf
with each other or with the retained I-Before-E text profile. Their shared
`runResumable` and call-opcode frames are VM parents, not a safe optimization
target.

## Method

Both runs used canonical external stdlib, CPU 15, `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`. The runtime harness warmed each program before
the profiled measured call. Reverse Complement completed in 18.44 seconds
(8,167,318,398 ns/op; 705,448,504 B/op; 10,894,380 allocs/op). K-Nucleotide
required 102.18 seconds for warmup plus one measured call
(55,542,979,361 ns/op; 1,307,549,920 B/op; 21,600,174 allocs/op); this remains
bounded profile evidence, not a normal test lane.

Retained profiles:

- `20260713_external_reverse_complement_bytecode_refresh.cpu.pprof` (8.11 s)
- `20260713_external_k_nucleotide_bytecode_refresh.cpu.pprof` (55.13 s)

Temporary runtime reports are cleanup-eligible under
`v12/tmp/perf/2026-07-13-bytecode-k-nucleotide-reverse-complement-profile-refresh/`.

## Attribution

| Workload | Concrete material descendants | Result |
| --- | --- | --- |
| K-Nucleotide | `finishInlineReturn` 9.00 s cumulative; `execBinary` 7.73 s; exact-native calls 3.71 s; `bytecodeRawIntegerValueInfo` 2.93 s | call/return, type-match, and raw-integer/counting path |
| Reverse Complement | `execCallMemberArraySlot` 2.39 s cumulative; `appendSlotStackValueChecked` 1.75 s; boxed/snapshot integer helpers 1.42/1.33 s | Array slot, value snapshot, and byte transformation path |
| I-Before-E (retained) | member call/cache and inline return | text member-dispatch path |

Do not infer a raw-integer or boxed-snapshot change from K-Nucleotide and
Reverse Complement: their similarly named integer helpers have different
representations and callers, and the earlier raw-cell/frame routes already
failed broad guards.

## Next Recommendation

Do not take another bytecode micro-candidate from this selection set. Refresh
the deterministic external scoreboard from retained current evidence and use
the feature matrix to add a missing independently authored application only
when it can discriminate a suspected broad VM boundary. The current severe
misses divide into map/counting, byte/Array, and text-member paths, so a source
change now would optimize one benchmark family rather than Able programs
generally.
