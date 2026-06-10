# Bytecode byte/output current-profile gate

## Outcome

This gate retains one generic primitive-array improvement shared by Base64,
Reverse Complement, and FASTA Generation:

- canonical `Array.push` now recognizes raw and boxed `u8` values and appends
  them directly to monomorphic `u8` storage;
- monomorphic `u8` reads return the VM's existing raw `u8` carrier, matching
  the existing raw `i32` and `u32` read paths instead of boxing each byte.

No benchmark, verifier, reference, language syntax, compiler lowering, named
nominal fast path, or canonical stdlib source changed.

## Reproducibility contract

- Go: `go1.26.4 linux/amd64`
- repository HEAD: `237406eccdfb025a519d898daedadee1c8d13a7b`
- baseline test binary SHA-256:
  `53e137130679ee6feab651a14a482a81c958c18aa4cb413e4043faedb4b0d517`
- retained-candidate test binary SHA-256:
  `e5bc433baa1438b48a4cf83d28c8fdbacbfeb2a17d72df0aca9e9933cc0b0bed`
- canonical stdlib tree SHA-256:
  `64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`
- protocol: one warmed `BenchmarkBytecodeProgramRuntime` measured call per
  fresh process, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, CPU 0,
  five independent processes per application, and a 55-second per-process
  timeout.
- CPU and allocation profiles were collected in separate fresh processes so
  memory sampling did not contaminate CPU attribution.

Source SHA-256 values were:

- Base64: `b4676ab1b4392ed4433d7a2ce57c7388907e4719494e6edce32728b071750108`
- FASTA Generation: `f8c67c9ab16e29d92904db2d58091f2512f83319a3bca5caf8ee90c37a2a96d7`
- Reverse Complement: `4bd16d4b6c65362efc5b2515548f59686144553a91eadcc65a0cde6df537e5f9`

## Current attribution

The clean baseline profiles split at their dominant owners:

- Base64 was already a host-kernel workload: Go base64 encode/decode and MD5
  owned most CPU, while two exact output buffers owned almost all allocated
  bytes.
- Reverse Complement spent material CPU in generic Array push/read handling.
  Its baseline measured run allocated about 464 MB in 7.88 million objects;
  monomorphic primitive reads, byte materialization, and later extern
  deoptimization were the main allocation owners.
- FASTA Generation remained call/return/arithmetic heavy, with its byte output
  crossing the generic push path and later the existing extern boundary.

Line attribution exposed the same exact `Array.push` descendant in all three
applications. The dispatcher already had character, `u32`, and `u64` lanes,
but no `u8` lane despite an existing generic `appendArrayU8ValueFast` runtime
operation.

## Candidate reconciliation

The first trial accepted raw bytes by calling `bytecodeIntegerValue`; that
materialized every raw byte and made Reverse Complement allocate about 40 MB
and 1.70 million objects more. It was not retained.

Direct raw-integer inspection removed that conversion, but promoting the Array
then exposed a second representation mismatch: every monomorphic `u8` read
called `runtime.NewSmallInt`. The allocation profile attributed 276.65 MB
directly to that one branch. Returning `bytecodeRawU8ResultValue`, as the
existing primitive read path already does for `i32`/`u32`, removed the boxing
without changing the public runtime value contract.

## Repeated A/B results

Negative percentages are improvements. Every row is the arithmetic mean of
independent fresh processes; medians and full samples remain in the bounded
`/tmp/able-byte-output-frontier-20260720-a` work directory for this session.

| Workload | Runs | Baseline ns/op | Retained ns/op | Change | Allocation result |
| --- | ---: | ---: | ---: | ---: | --- |
| Base64 | 5 + 5 | 2,683,611,007 | 2,492,922,560 | -7.11% | 2.202 GB unchanged; 493 to 484 objects |
| Reverse Complement | 5 + 5 | 3,793,419,051 | 3,199,770,598 | -15.65% | 463.95 MB / 7,876,267 to 213.57 MB / 3,542,787 |
| FASTA Generation | 5 + 5 | 1,774,265,840 | 1,650,046,302 | -7.00% | 70.44 MB / 1,817,097 essentially unchanged |
| JSON guard | 5 + 5 | 507,060,561 | 505,523,402 | -0.30% | exactly unchanged |
| PiDigits guard | 5 + 5 | 2,119,286,736 | 2,069,775,461 | -2.34% | exactly unchanged |
| numeric Array map guard | 10 + 10 | 85,405,107 | 87,508,060 | +2.46% | 14,646 objects unchanged |
| iterator collect guard | 10 + 10 | 427,894,102 | 436,225,192 | +1.95% | 213,286 objects unchanged |

The two unlike fixture guards were extended from five to ten processes and
the A/B order was reversed for the second half because the first cohorts were
volatile. Both final means remain inside the established 5% broad guard with
unchanged allocation counts.

## Post-change profiles

Reverse Complement's measured allocation profile fell from 459.78 MB sampled
to 216.59 MB sampled. The `bytecodeMonoPrimitiveArrayValue` allocation leaf
disappeared completely. Residual allocation is now the separate, already
identified extern/deoptimization boundary (`u8ToValue`,
`deoptTypedArrayToDynamic`) rather than byte-array read boxing.

Post-change CPU ownership now diverges:

- Base64 remains dominated by Go base64 encode/decode and MD5 kernels.
- Reverse Complement is split across primitive Array reads, VM stack/call
  mechanics, map lookup, and the residual extern boundary.
- FASTA Generation is split across raw `i32` slot handling, arithmetic, and
  call/return mechanics.

The retained exact leaf is therefore exhausted for this group; the residual
profiles do not justify another shared three-application candidate.

## Correctness and public-output verification

The normal external harness completed five fresh executions for each target
and guard, with every output verified against the public application verifier:

| Application | Status | Validation | Mean real time | Reference comparison where available |
| --- | --- | --- | ---: | --- |
| Base64 | 5/5 | verified 5/5 | 2.680 s | 0.81x Python; 1.21x Ruby |
| Reverse Complement | 5/5 | verified 5/5 | 3.410 s | no stored Python/Ruby row |
| FASTA Generation | 5/5 | verified 5/5 | 1.744 s | no stored Python/Ruby row |
| JSON | 5/5 | verified 5/5 | 0.852 s | 0.30x Python; 0.55x Ruby |
| PiDigits | 5/5 | verified 5/5 | 2.200 s | 0.24x Ruby |

Focused primitive Array, raw-byte, StringBuilder-byte, runtime Array, Future,
and cleanup tests pass. The full interpreter/runtime package command reached
its 55-second aggregate cap during `TestExecFixtureParity`; the reported
anchored-regex subtest passes independently on both baseline and candidate
binaries in under one second, so this was cumulative suite wall time rather
than a candidate hang.

## Next gate

Refresh the current-binary bytecode concurrency group next. It is now the
largest remaining stale bytecode attribution group in the checked frontier,
covering six unlike channel/Future/mutex applications. Use preserved binaries,
separate bounded CPU/allocation profiles, and repeated target-meeting plus
non-concurrency guards. Advance code only if one exact scheduler/VM descendant
is material across at least three unlike applications; otherwise close the
group and move to compiled byte/output. Do not begin WASM.
