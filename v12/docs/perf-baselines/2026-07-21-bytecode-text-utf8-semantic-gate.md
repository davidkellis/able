# Bytecode text/UTF-8 semantic gate

Date: 2026-07-21

## Decision

Keep no VM, runtime, stdlib, application, language, compiler, or WASM
performance change. Retain only the opt-in `ABLE_BYTECODE_STRING_STATS`
observer and its measured-main report fields.

K-Nucleotide, Word Frequency, and Policy Record Dispatch repeat the same
host-string to Able-byte-array boundary, but the exact conversion is
CPU-material in only K-Nucleotide and Word Frequency. Policy's conversion,
validation, and decoding work is below 0.1% of its merged samples. Distance
Field executes none of the observed String/UTF-8 operations. The required
three-unlike-family materiality gate therefore fails, so the tempting direct
monomorphic-u8 construction candidate was not built.

This distinguishes operation count from optimization leverage. K-Nucleotide
performs enough inbound conversions to make the generic boundary visible;
Policy proves that merely exercising the same boundary does not make it an
important owner of application wall time.

## Protocol

One optimized interpreter test binary was frozen at SHA-256
`a590027bcab72067238d1450c2e00072afbe221cc34aa1644e82322a3c355c2f`.
Every application ran in a fresh process with the canonical external stdlib,
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 59-second process cap.
Loading, typechecking, lowering, and final forced collections were outside
the measured main. Policy retained typechecking during setup because its
overload annotations are required for unambiguous bytecode lowering; that
setup remains outside the CPU and allocation window.

Each application received two clean main-only CPU profiles, two separate
exact main-allocation processes, and two separate observer processes. CPU
and allocation means are arithmetic means. Observer counts were identical
between their two processes. Source hashes match the current verifier-backed
scorecard rows. The canonical stdlib remained 70 Able sources at tree hash
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.

| Application | CPU runs | CPU mean | Merged samples | Mean bytes | Mean allocations |
| --- | --- | ---: | ---: | ---: | ---: |
| K-Nucleotide | 43.331 s, 45.182 s | 44.257 s | 88.03 s | 1,246,009,140 | 23,947,382 |
| Word Frequency | 1.449 s, 1.448 s | 1.448 s | 2.88 s | 54,156,904 | 625,738 |
| Policy Record Dispatch | 7.513 s, 7.183 s | 7.348 s | 14.64 s | 143,472,244 | 1,439,779 |
| Distance Field | 5.661 s, 5.901 s | 5.781 s | 11.53 s | 368,035,800 | 26,000,119 |

The CPU profiles themselves add bookkeeping allocations; the allocation
columns use the two non-profiled processes only. Their object counts were
stable: the two K values differed by 12, Word by 12, Policy by 41, and
Distance was identical.

## Exact semantic counts

| Application | From-builtin calls / bytes | To-builtin calls / bytes | Canonical validations | Raw validations / bytes | Rune decodes / bytes |
| --- | ---: | ---: | ---: | ---: | ---: |
| K-Nucleotide | 166,685 / 6,233,459 | 14 / 24 | 0 | 0 / 0 | 0 / 0 |
| Word Frequency | 3,874 / 131,073 | 25,948 / 105,125 | 0 | 0 / 0 | 0 / 0 |
| Policy Record Dispatch | 2,048 / 76,224 | 9,216 / 114,048 | 0 | 2,049 / 76,315 | 76,315 / 76,315 |
| Distance Field | 0 / 0 | 0 / 0 | 0 | 0 / 0 | 0 / 0 |

All observed inputs were valid UTF-8. Word's outbound conversions split into
17,979 monomorphic and 7,969 fallback calls; Policy's 9,216 outbound calls
were all monomorphic; K's 14 were all fallback calls.

The observer is allocated only when `ABLE_BYTECODE_STRING_STATS` is set.
Compiler inlining diagnostics confirm that disabled record sites inline to a
single nil check; they do not perform atomic operations or allocate.

## CPU leverage reconciliation

The exact inbound-conversion closure
`(*Interpreter).initStringHostBuiltins.func5` accounts for:

- K-Nucleotide: 1.52 s cumulative, 1.73% of 88.03 s;
- Word Frequency: 0.03 s cumulative, 1.04% of 2.88 s;
- Policy Record Dispatch: below sampler resolution;
- Distance Field: no calls and no samples.

K's time divides between constructing one `runtime.Value` per byte and
building the generic Array state. That makes direct monomorphic-u8
construction an attractive general implementation idea, but not a broad
application optimization on this corpus. Word barely clears 1%, while
Policy's complete String host, raw validation, and character traversal leaves
reach only 0.01 s (0.068%) individually in the merged profile. Outbound
conversion and UTF-8 decode therefore do not supply a different three-family
candidate.

No application-specific algorithm, FASTA path, regex path, String-key map,
named container, or stdlib special case is justified by these counts.

## Retained diagnostic

The runtime-retention harness now includes a `string_stats` object containing
inbound/outbound byte conversion, monomorphic/fallback outbound shape,
canonical/raw validation, rune decode, byte volume, and invalid-UTF-8 counts.
It snapshots the concrete interpreter directly rather than through global
state, so independent or concurrent tests cannot exchange observations.

Verification:

```text
go test ./pkg/interpreter -run 'TestBytecodeStringStats|TestBytecodeProgramRuntimeRetention' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.059s

go test ./pkg/interpreter -run 'TestStringFromBuiltin|TestStringToBuiltin|TestCharFromCodepoint|TestCharToCodepoint|TestCharSimpleFoldNext|TestStringHostBuiltinCallMetadata' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.073s

go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  25.495s
```

## Next recommendation

Run the direct Array semantic-boundary gate across Array Slice Window, Reverse
Complement, and Matrix Multiply, with Word Frequency and Distance Field as
text and numeric controls.

Why: the String/UTF-8 boundary is now closed by exact low leverage in the
third text family. The semantic-work audit's only remaining concrete
two-family survivor is canonical Array slot/member storage; Matrix Multiply
is an independently authored, Array-heavy third family that can either admit
or close it without relying on a named container or workload shape.

What it entails: add or reuse opt-in exact counters for canonical Array-slot
reads, writes, length, push, cache hits, and fallback; collect two clean CPU
profiles and two exact allocation processes per target; and reconcile one
exact storage/member operation below the dispatcher parent. Prototype at most
one generic Array runtime/VM mechanism only if it is CPU-material in all three
unlike targets and has credible end-to-end leverage. Guard any candidate with
the two controls and repeated order-balanced arithmetic means. Do not reopen
closed Array-slot cache, operand-lane, call/return, named-stdlib, benchmark, or
WASM designs without new invalidating evidence.
