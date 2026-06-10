# External Bytecode-Miss Profiles

## Decision

Keep no bytecode CPU optimization. Current-source I-Before-E, Base64, and
Monte Carlo all pass their unchanged public verifiers, but their material CPU
work is respectively member-call/cache validation, direct host codec/MD5
kernels, and the already-known float store/cast lane. No exact generic helper
and caller repeat across two of those independent families.

Do not add a text, codec, MD5, base64, float-loop, file, or benchmark-shaped
opcode/lowering rule. No compiler, VM, tree-walker, or canonical-stdlib source
changed in this tranche.

The Base64 profile did reveal a separate generic Array-handle lifetime defect.
It is documented here as a memory-correctness/performance follow-up, not
treated as authorization for a narrow codec fast path.

## Method

All processes used the canonical stdlib at
`/home/david/sync/projects/able-stdlib/src` with `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`.

First, unchanged external applications ran once through the normal bytecode
CLI and their public Ruby verifiers:

| Application | Guarded real time | Verification |
| --- | ---: | --- |
| I-Before-E (`wordlist.txt`) | 0.6400 s | passed |
| Base64 | 3.1800 s | passed |
| Monte Carlo Pi | 2.7800 s | passed |

Next, the steady-state bytecode benchmark harness loaded and warmed each
unchanged program once, then used a fixed `-benchtime` count to avoid
short-process sampling noise: 30 I-Before-E calls and three calls each for
Base64 and Monte Carlo. The retained CPU profiles contain 8.40, 8.97, and
7.94 seconds of samples.

| Application | Steady-state reading |
| --- | --- |
| I-Before-E | 278,158,369 ns/op; 9,061,869 B/op; 1,916 allocs/op |
| Base64 | 2,994,145,892 ns/op; 2,201,607,282 B/op; 433 allocs/op |
| Monte Carlo Pi | 2,651,696,180 ns/op; 177,803,610 B/op; 22,222,114 allocs/op |

Artifacts are retained under
`v12/tmp/external-bytecode-miss-profiles-2026-07-11/`.

## CPU evidence

### I-Before-E

I-Before-E is generic member-call work around text/file processing:

- `execCallOpcode` is 4.60 s cumulative and `execCallMember` is 2.41 s.
- `lookupCachedMemberMethodEntry` is 1.22 s cumulative; its lexical-state
  header and cache-identity checks account for 0.21 s and 0.17 s flat.
- `finishInlineReturn` is 0.92 s; direct string contains is only 0.29 s.

This repeats the member-cache area previously profiled in string workloads,
but no other application in this tranche reaches the same caller/callee
shape materially.

### Base64

Base64 is 85.3% direct host work:

- Go `encoding/base64` encode is 3.43 s flat (38.2%).
- decode is 2.93 s cumulative (32.7%).
- MD5 is 1.07 s flat (11.9%).

The VM dispatch only surrounds exact native calls; its own Array call work is
below 7%. This cannot validate a member-cache or float-path change.

### Monte Carlo Pi

Monte Carlo is the existing primitive-float lane:

- `execStoreSlotCastSlotFloatConstDiv` is 2.97 s cumulative.
- `execStoreSlotCastSlotFloatConstDivDiscardFast` is 2.66 s.
- the fused float multiply/add comparison jump is 1.03 s.

This is the same representation/store family whose broader variants have
already failed cross-workload A/B checks. It has no material member-cache or
extern-codec call path.

## Base64 Array-handle lifetime finding

One additional normal bytecode Base64 run wrote a heap profile at process exit
after the profiler forced a GC. It passed the unchanged public verifier in
3.29 seconds but retained 2.13 GiB:

| Retained allocation | In-use space |
| --- | ---: |
| external `encode_bytes` outputs | 1,198.01 MB |
| external `decode_bytes` outputs | 899.78 MB |
| all other observed storage | 33.63 MB |

The in-use attribution is effectively equal to the process allocation
attribution, so this is not ordinary transient codec allocation. Code
inspection identifies the generic ownership gap:

1. The shared primitive-byte extern fast invoker wraps each host `[]byte`
   result through `newOwnedU8ArrayValueFromBytes`.
2. That wrapper registers each fresh handle in the interpreter's
   `arraysByHandle` tracking map.
3. The runtime additionally retains the backing state in process-wide
   handle-to-state maps. Neither registry has a lifetime/release path.

The same pattern is visible on a different external shape: the repeated
I-Before-E run retains 82.28 MB of returned String-array value slices after a
forced GC. The exact allocation functions differ, so this is not yet a
validated one-line optimization; the shared issue is the Array handle
ownership model. The Go memory limit is necessarily a soft limit here because
the data is still live from the runtime's perspective.

## Next recommendation

Design and test a general Array-handle lifetime model before attempting any
CPU tuning. Why: externally verified Base64 demonstrates multi-gigabyte
retention inside one ordinary application run, and I-Before-E independently
shows the same lifetime category for String arrays. The work entails defining
which values own a handle, how aliases retain it, when last-reference release
is safe across bytecode/tree-walker/extern boundaries, and adding semantic
alias/mutation plus bounded-memory regression tests. It must replace
process-lifetime retention generally; a Base64-only buffer reuse or named
container rule would be invalid.
