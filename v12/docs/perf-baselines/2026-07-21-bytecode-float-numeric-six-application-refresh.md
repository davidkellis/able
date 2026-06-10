# Bytecode float-numeric six-application refresh

Date: 2026-07-21

## Decision

Keep no VM, compiler, runtime, canonical-stdlib, language, benchmark, reference,
or WASM change. Fresh profiles reproduce float-slot extraction and raw float
arithmetic across three unlike numeric programs, but those exact leaves belong
to the already rejected raw-float carrier, operand-lane, and typed-region
family. The two unlike controls separate the residual costs: Matrix Multiply
is dominated by its generic monomorphic-f64 Array dot loop, Monte Carlo mixes
integer recurrence with cast/divide, and Reverse Complement contains no
material float path.

No new concrete operation cleared the admission rule, so this tranche did not
build a candidate merely to reduce allocation counters. The current evidence
does not invalidate the earlier broad wall-time failures of float sidecars,
operand lanes, native scalar-result carriers, or producer fusion.

## Reproducibility contract

One current interpreter test binary was frozen at SHA-256
`6758a13355a1adeebe0984098679c3ad344b0f8cf1a8642694e873e3dd12d53e`.
Each application ran twice for main-only CPU profiles and twice in independent
unprofiled processes for exact measured-main allocation counters. Every
process used CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, the canonical
external stdlib, skipped repeated setup typechecking, and had a 59-second test
cap. Setup, lowering, output capture, and final forced collections were outside
the measured main window.

All six sources match the promoted scorecard hashes. Distance Field and RMS
ran from their source directories; Mandelbrot, Monte Carlo, Matrix Multiply,
and Reverse Complement used their external verifier/input directories. The
workstation was active, so the two samples are interpreted as arithmetic
means and attribution shares, not as a replacement scorecard.

| Application | Profiled main runs | Mean | Merged samples |
| --- | --- | ---: | ---: |
| Distance Field | 6.136 s, 5.733 s | 5.935 s | 11.82 s |
| RMS Norm | 4.951 s, 4.972 s | 4.961 s | 9.88 s |
| Mandelbrot | 6.636 s, 8.626 s | 7.631 s | 15.14 s |
| Monte Carlo Pi | 2.496 s, 3.158 s | 2.827 s | 5.64 s |
| Matrix Multiply | 4.738 s, 5.609 s | 5.173 s | 10.29 s |
| Reverse Complement | 3.409 s, 3.436 s | 3.423 s | 6.82 s |

## Exact CPU intersection

Admission required the same exact symbol to account for at least 1% flat CPU
in at least three unlike applications. Aggregate dispatch/allocation parents
and already rejected representations did not qualify as new candidates.

| Exact symbol | Breadth | Flat shares | Disposition |
| --- | ---: | --- | --- |
| `(*bytecodeVM).runResumable` | 6 | 6.01%-14.73% | aggregate dispatcher parent, already closed |
| `runtime.tryDeferToSpanScan` | 6 | 1.77%-4.29% | Go GC machinery |
| `(*bytecodeVM).appendSlotStackValueChecked` | 5 | 1.21%-6.21% | closed stack-carrier family |
| `runtime.nextFreeFast` | 5 | 1.27%-2.66% | Go allocator machinery |
| `runtime.mallocgc` | 4 | 1.01%-2.77% | different concrete allocation owners |
| `(*bytecodeVM).appendStackValue` | 3 | 1.32%-1.98% | closed stack-carrier family |
| `(*bytecodeVM).slotDirectFloatValueValidated` | 3 | 1.42%-2.91% | closed float-slot/carrier family |
| `bytecodeDirectFloatArithmeticRawFast` | 3 | 1.18%-1.72% | retained raw arithmetic; no new transport design |
| `runtime.mallocgcTiny` | 3 | 1.86%-6.34% | raw-float boxing below different producers |

The float-read callers diverge even inside their three-program breadth. RMS
spends most of the leaf below a typed float region, Mandelbrot divides it
among fused branch, multiply, and add/multiply stores, and Monte Carlo reaches
it mainly through the fused float branch after an integer recurrence and
cast/divide store. Raw arithmetic appears in Distance, RMS, and Mandelbrot,
but it is already the allocation-free retained arithmetic helper; removing
its surrounding transport means reopening the rejected carrier designs.

Matrix is a useful negative discriminator: `tryExecF64DotLoop` alone accounts
for 24.00% flat CPU, while its other material work is Array storage, map/lock,
and allocation machinery. Reverse is led by map matching, Array reads, raw
integer extraction, and synchronization. Their absence from the float leaves
prevents related scalar programs from establishing misleading breadth.

## Exact measured-main allocation

Allocation volume was stable despite wall-time contention: byte spans ranged
from 16 to 176 bytes and object-count spans from zero to four.

| Application | Mean bytes | Mean allocations | Mean frees | Mean GCs |
| --- | ---: | ---: | ---: | ---: |
| Distance Field | 368,035,900 | 26,000,120 | 25,570,716 | 20.0 |
| RMS Norm | 288,033,920 | 20,000,121 | 19,516,587 | 16.0 |
| Mandelbrot | 615,167,296 | 76,303,236 | 75,489,715 | 33.0 |
| Monte Carlo Pi | 177,824,424 | 22,222,134 | 21,198,490 | 11.5 |
| Matrix Multiply | 308,581,952 | 14,032,642.5 | 13,350,081.5 | 14.0 |
| Reverse Complement | 266,698,992 | 4,069,474 | 1,857,873.5 | 7.0 |

This shape agrees with the prior allocation-owner evidence: Distance, RMS,
Mandelbrot, and Monte Carlo pay at different float materialization boundaries;
Matrix pays for Array storage and dynamic boundaries; Reverse pays for byte
Array conversion and snapshots. The true f64 operand lane previously cut
allocation 5.3%-30.3% across four programs yet slowed every broad program
12.7%-23.6%. The current counters add no cheaper coherence rule or new
consumer pattern that would overturn that causal result.

## Verification

The unchanged production tree passes the complete bytecode family:

```text
go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  24.974s
```

Temporary binaries, profiles, output captures, and retention reports were
removed after evidence/frontier validation.

## Next recommendation

Refresh the `bytecode-regex` ownership group across Regex Suffix Audit, Regex
Set Audit, Regex Stream Audit, Log Routing/Redaction, and Configuration
Validation/Extraction, with unrelated text and numeric controls.

Why: after refreshing the larger text/map, wide-numeric, and float-numeric
groups, regex is the largest remaining bytecode group at 15.092 target-excess
seconds. Its current evidence has breadth, but several profiles predate the
latest retained VM and primitive-boundary changes; a same-binary refresh can
tell whether NFA work still dominates or whether a new generic VM wall has
emerged.

What it entails: collect repeated bounded CPU and exact measured-main
allocation processes, intersect exact leaves across the three distinct regex
API/application families and unlike guards, and admit at most one genuinely
new primitive/runtime operation. Existing NFA arenas, state indexes,
character specialization, thread carriers, Array-slot caches, and call/return
alternatives remain closed unless the new profiles provide invalidating
evidence. No regex-benchmark opcode, named stdlib-type shortcut, or WASM work
is admissible.
