# Bytecode multi-argument primitive-byte extern admission

Date: 2026-07-18

## Decision

Close the multi-argument primitive-byte extern-conversion gate and retain no
compiler, bytecode VM, runtime, canonical-stdlib, benchmark, fixture, or
language change.

Mandelbrot does reach the same generic host-conversion and mono-`u8`
deoptimization chain previously found in Reverse Complement and FASTA
Generation. It does not repeat that chain materially: the complete conversion
subtree accounts for 0.85% of sampled allocation space, the array
deoptimization accounts for 0.83%, and neither is a material CPU sample. The
predeclared third-unlike-application admission rule therefore fails.

## Contract and inventory

The v12 Array contract requires mutable Array values to preserve their normal
value and aliasing semantics. The host-interop contract describes ordinary
Array arguments as copied into host values. The existing
`able_borrowed_bytes(...)` marker is an internal, explicit synchronous host
implementation opt-in; it is not new Able syntax and does not authorize a
general change to Array semantics.

The current canonical stdlib contains five marker uses:

- one single-argument MD5 host function;
- three single-argument Base64 host functions; and
- the multi-argument `io_write(IoHandle, Array u8)` host function.

The bytecode fast invoker currently admits only exact one-argument primitive
byte signatures. Generated compiled code already recognizes the explicit
marker independently of argument position. Consequently, broadening the
bytecode invoker today would serve only one current multi-argument host
function. A generic implementation shape alone cannot establish broad runtime
materiality.

## Mandelbrot measurement

Five independent warmed bytecode-runtime processes used canonical
`../able-stdlib`, CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, one
warmup call, one measured call, and a 55-second process cap.

| Run | Runtime | Bytes/op | Allocs/op |
| ---: | ---: | ---: | ---: |
| 1 | 6.480 s | 615,201,480 | 76,303,141 |
| 2 | 6.195 s | 615,201,320 | 76,303,132 |
| 3 | 6.097 s | 615,201,704 | 76,303,150 |
| 4 | 6.212 s | 615,201,560 | 76,303,144 |
| 5 | 6.358 s | 615,201,368 | 76,303,135 |
| **Mean** | **6.268 s** | **615,201,486** | **76,303,140** |

The process elapsed mean was 13.258 seconds because each process also loaded,
typechecked, warmed, and—in the final run—profiled the application. The
runtime-only benchmark discards output, so it is attribution evidence rather
than verifier evidence. The identical promoted normal-process source remains
the correctness authority: 5/5 verified bytecode runs, 6.740 seconds mean,
versus Python 1.2431 seconds and Ruby 1.9676 seconds.

## Profile attribution

The bounded CPU profile contains 6.33 seconds of samples. Its leading Able
descendants are the existing float recurrence and slot/stack machinery:

- `execJumpIfFloatMulAddMulCompareConstFalse`, 21.17% cumulative;
- `storeReusableNormalizedFloatSlotRawDiscard`, 13.43% cumulative;
- `appendSlotStackValueChecked`, 6.00% cumulative; and
- validated direct float extraction, 4.11% flat.

Those are the already-profiled raw-float/dispatch family whose operand-lane,
sidecar, carrier, and handler variants failed broad wall-time gates. The host
conversion/deoptimization chain has no material CPU sample.

The separate allocation-space profile sampled 595.83 MiB:

| Exact owner | Sampled space | Share |
| --- | ---: | ---: |
| `coerceRuntimeToHost` subtree | 5.04 MiB | 0.85% |
| `toArrayValue` / `ArrayStoreEnsure` subtree | 4.96 MiB | 0.83% |
| `deoptTypedArrayToDynamic` subtree | 4.96 MiB | 0.83% |
| `deoptTypedArrayToDynamic` flat | 1.23 MiB | 0.21% |

This is the same exact mechanism as Reverse Complement and FASTA, but not the
same material wall. Their preceding sampled shares were 26.97% and 77.23%.

## Reconciliation

The required cohort is therefore two material bulk-output applications plus
one unlike application with an immaterial occurrence. Static inventory also
finds only `io_write` as a current multi-argument consumer. That does not clear
the project's bar for a broadly useful VM change.

No candidate was built, so a baseline/candidate guard cohort would answer no
admitted question. Base64 remains the existing single-argument borrowed-byte
control; JSON, Word Frequency, and Array Slice Window remain available guards
if a future broader host-ABI candidate is independently admitted.

## Verification and cleanup

- 5/5 bounded warmed Mandelbrot processes completed.
- The promoted 5/5 normal bytecode runs remain verifier-backed on the same
  source and input contract.
- The extern marker inventory covers the canonical stdlib and both Go
  execution modes.
- Focused existing extern and primitive-byte tests pass on the unchanged
  production tree.
- Raw profiles, the temporary benchmark binary, stdout/stderr captures, and
  scratch work directories are removed after this record is written.
- No WASM work was performed.

## Next recommendation

Run a current cross-mode exact-leaf selection pass over diverse sustained
misses, while treating the raw-float carriers, broad dispatcher layouts,
primitive-Array growth, returned nominal transport, compiled startup ballast,
and this multi-argument byte boundary as closed designs.

Why: the latest concrete lead failed breadth, and Mandelbrot's remaining large
parents have already received repeated generic implementation and benchmark
gates. Selecting from the current scorecard without a closed-design filter
would simply recreate rejected experiments. The strongest chance of a broad
win is now an exact removable descendant that recurs in unlike compiler and
bytecode programs rather than another local VM micro-cut.

What it entails: reuse current verified scorecard rows and retained profiles,
then collect only missing bounded generated-main and warmed-bytecode profiles
for a deliberately diverse set such as K-Nucleotide, Fixed Width 128, Regex
Set, Distance Field, and one concurrency application. Normalize owners below
allocator, GC, dispatcher, and scheduler parents; record why closed descendants
are ineligible; and admit at most one candidate only when the same previously
untried language/runtime operation is material in at least three unlike
applications. Any candidate then receives repeated, order-balanced,
verifier-backed averages plus unrelated controls. This avoids spending another
tranche measuring a design whose broad result is already known.
