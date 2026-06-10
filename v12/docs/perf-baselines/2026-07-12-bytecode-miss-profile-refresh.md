# Bytecode Miss Profile Refresh (2026-07-12)

## Decision

Keep no bytecode, compiler, runtime, benchmark, fixture, or canonical-stdlib
performance change from this tranche. The six fresh applications confirm that
the material VM work is still split among Array/boxing, plain call/return,
cached member dispatch, raw-float transport, host codec calls, and BigInt host
work. `runResumable` and `execCallOpcode` are shared parents only; they do not
identify an identical implementation leaf.

Do not reopen the already rejected generic call-name/raw-cell, inline-return,
or raw-float experiments from a single row. In particular, do not introduce a
text, codec, MD5, BigInt, Mandelbrot-shaped float, or byte-scanning rule.

The in-use heap captures identify Array-family allocation while the benchmark
interpreter remains live; they are not final-retention measurements. The
subsequent dedicated post-scope probe releases all Base64, I-Before-E, and
Reverse Complement ArrayStore state after three GCs, so there is no remaining
generic Array/extern-return release candidate in this data. See
`2026-07-12-bytecode-array-retention-revalidation.md`.

## Method

All runs used the canonical external stdlib at
`/home/david/sync/projects/able-stdlib/src`, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, and CPU 2 affinity. Each unchanged application first completed
one normal bytecode CLI run under a 60-second bound and passed its existing
Ruby verifier. CPU, allocation, and in-use heap summaries are retained in
`2026-07-12-bytecode-miss-profile-refresh/` beside this record. The heap
profiles are intentionally in-scope snapshots; final retention requires the
separate probe above.

The runtime benchmark harness loads and warms a program before timed calls.
I-Before-E used 30 calls to obtain a 7.66-second CPU sample. Reverse
Complement, Mandelbrot, Base64, and PiDigits used one whole-program warmed
call each; their samples are respectively 6.89, 7.70, 3.39, and 2.63 seconds.
K-Nucleotide requires longer than the bounded warm-plus-sample budget, so its
verifier-backed normal process supplied the 43.36-second CPU and allocation
profiles instead. Its allocation totals therefore include normal program
setup and are not compared as a warmed per-operation metric.

| Application | Normal bytecode run | Verification | Warmed reading |
| --- | ---: | --- | --- |
| Reverse Complement | 7.23 s | passed | 6,946,474,650 ns/op; 705,448,536 B/op; 10,894,379 allocs/op |
| K-Nucleotide | 44.47 s | passed | normal-process profile only |
| I-Before-E | 0.60 s | passed | 257,590,728 ns/op; 9,061,905 B/op; 1,923 allocs/op |
| Mandelbrot | 7.03 s | passed | 6,554,238,713 ns/op; 618,885,824 B/op; 76,303,083 allocs/op |
| Base64 | 3.85 s | passed | 3,442,237,972 ns/op; 2,201,666,200 B/op; 651 allocs/op |
| PiDigits | 2.77 s | passed | 2,660,835,977 ns/op; 335,058,352 B/op; 940,451 allocs/op |

The normal CLI timing and warmed harness are deliberately separate readings:
the former validates public program behavior, while the latter locates
steady-state runtime work. The raw summary table is `rows.tsv` in the artifact
directory.

## Attribution

| Application | Material current-source work | Why it does not validate a CPU candidate |
| --- | --- | --- |
| Reverse Complement | `bytecodeMonoPrimitiveArrayValue` is 40.14% of allocation, integer boxing 28.85%, and Array capacity growth 24.18%; CPU reaches Array-member push and stack materialization. | Primitive array construction/boxing is not the member-cache, plain-call, float, or extern path. |
| K-Nucleotide | `execCallOpcode` is 29.45% cumulative, `execCallName` 18.66%, `finishInlineReturn` 17.09%, and `execBinary` 15.22%. | The previously tested generic call-name and return variants did not clear broad guards; I-Before-E's material call work is a different member-cache path. |
| I-Before-E | `execCallMember` is 29.77%, `lookupCachedMemberMethodEntry` 12.53%, and the cached member fast path 8.62%. | This is receiver/cache validation around text and file processing, absent from the other material workers. |
| Mandelbrot | Normalized raw-float value creation is 92.23% of allocation; `execBinary` is 26.10% CPU cumulative and the fused float jump 20.65%. | This is the already rejected generic raw-float/store lane, not evidence for a program-shaped float rule. |
| Base64 | Exact native calls spend 80.53% cumulative in the host path; encode/decode allocate 55.96%/41.98%. | The dominant work is host codec allocation and computation, not generic VM dispatch. |
| PiDigits | Exact native calls are 59.32% cumulative; `math/big.nat.mul` is 20.53% CPU cumulative and `math/big.nat.make` is 78.48% of allocation. | The BigInt kernel is independent of Base64's codec bridge and of the VM-only lanes. |

The in-scope heap captures are still useful attribution: Base64's decoded-byte
result is 95.59% of its 705.32 MB in-use heap, while I-Before-E's external
String-array clone is 36.10% of its 51.47 MB heap. They must not be called
post-GC leaks, however. The dedicated scope-teardown probe reaches zero direct
ArrayStore state for both workers and for Reverse Complement. Any future
memory change must therefore identify a different shared retained root rather
than reuse a Base64 buffer or special-case a text result.

## Next recommendation

Do not reopen Array/extern release or integer-box cache policy work. The
completed tagged reuse evidence finds 95.77% K-Nucleotide cache hits and 4.59M
Reverse Complement hits despite saturation, while both controls bypass the
tier. The canonical runtime-value architecture design was already completed
and rejected a prototype; the normal-build opcode census also finds no
remaining shared VM leaf. See `2026-07-12-bytecode-opcode-census.md`. Resume
feature completion rather than another unchanged-source micro-tranche,
profiling a newly shared boundary only after it is covered by the full
cross-language guard matrix.
