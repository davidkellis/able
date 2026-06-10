# Bytecode typed-handler profile reconciliation (2026-07-18)

## Decision

Keep no compiler, VM, runtime, canonical-stdlib, benchmark, fixture, or
language change. Bounded warmed CPU and allocation profiles for Monte Carlo
Pi, RMS Norm, and QuickSort do not expose one new handler-level helper that is
material in all three applications.

Primitive scalar boxing is a shared representation boundary, but it is not a
new candidate. Monte Carlo Pi and RMS Norm pay for float-to-interface
transport; QuickSort pays for integer-to-interface transport around primitive
Array operations. The exact f64 operand lane, pointer/raw-cell substitutions,
raw-I32 transport, and producer fusion have already received broad correctness
and wall-time gates. Those trials either moved the allocation or reduced
allocation while regressing application time. The new profiles do not
invalidate those results, so repeating them would optimize profile counters
rather than program performance.

## Admission rule

Before profiling, a candidate required the same concrete helper or
representation boundary to account for at least 2% cumulative CPU in all
three applications. Aggregate parents such as `runResumable`, unrelated leaves
below a shared wrapper, and previously rejected representations did not
qualify. Any implementation also had to remain a primitive/general VM rule,
not an application, opcode-sequence, Array, or other nominal-type special case.

The intended candidate scope was one in-dispatch helper or representation
boundary. A second dispatcher was excluded by the preceding typed loop-block
gate.

## Profile contract

One benchmark binary was built before collection and reused for all three
applications. Each process loaded and lowered the application once, ran one
warmup `main`, then profiled one measured `main` call with:

```text
CPU 0
GOMAXPROCS=1
55-second process cap
canonical external able-stdlib
normal typechecking during setup
```

The benchmark print hook suppresses output, so the same binary also ran each
source through the normal bytecode CLI before profiling. Monte Carlo Pi and
RMS Norm passed their public Ruby verifiers. QuickSort used the unchanged
external source with the first 50,000 canonical input numbers; Ruby sorted the
same input and independently checked its first and last ten values.

| Application | Measured time | B/op | allocs/op | Verification |
| --- | ---: | ---: | ---: | --- |
| Monte Carlo Pi | 3,210,867,859 ns | 177,857,128 | 22,222,190 | public verifier |
| RMS Norm | 5,138,121,374 ns | 288,084,896 | 20,000,171 | public verifier |
| QuickSort, 50,000 numbers | 1,087,779,714 ns | 53,187,792 | 2,600,421 | independent Ruby sort |

These are attribution runs, not replacements for the repeated external
scorecard means.

## CPU attribution

The concrete handler descendants split:

| Owner | Monte Carlo Pi | RMS Norm | QuickSort |
| --- | ---: | ---: | ---: |
| `runResumable` flat | 10.34% | 9.57% | 4.63% |
| normalized raw-float boxing / `convT64` cumulative | 10.66% | 5.66% | absent |
| validated direct float read flat | 2.51% | 3.32% | absent |
| validated direct I32 read cumulative | 10.03% | below 2% | below 2% |
| raw integer extraction flat | below displayed tier | below displayed tier | 6.48% |
| Array slot-index conversion cumulative | absent | absent | 10.19% |
| stack snapshot cumulative | below displayed tier | 2.73% | 6.48% |
| allocator cumulative | 10.03% | 9.18% | 3.70% |

Monte Carlo Pi is led by its cast/divide float store, integer slot/constant
store, multiply/modulo recurrence, and float branch. RMS Norm is led by its
typed float region plus call/native/unary boundaries. QuickSort is led by
primitive Array read/swap/index work, raw integer extraction, and its affine
integer update. The dispatcher and allocator are common Go parents, but their
removable Able descendants are not the same semantic operation.

## Allocation attribution

Workload-focused sampled allocation profiles make the split clearer:

- Monte Carlo Pi is almost entirely
  `bytecodeNormalizedRawFloatSlotValue`, reached from the discarded
  cast/divide store.
- RMS Norm divides among normalized raw floats, unary results, stable stack
  snapshots, raw-float materialization, and Ratio/native results.
- QuickSort divides among boxed I32 values, mono-primitive Array values, and
  raw-I32 stack snapshots.

The common abstraction is therefore the boxed `runtime.Value` transport, not
one bounded producer or consumer.

## Closed representation reconciliation

The repository already contains the required invalidation evidence:

- the true f64 operand lane cut allocations 5.3%-30.3% across four unlike
  numeric workloads but slowed every one by 12.7%-23.6%;
- pointer and slot-sidecar float carriers moved boxing to stable call/snapshot
  boundaries or regressed wall time;
- the raw-I32 cache returns stable allocation-free interface values inside its
  covered range, while a whole scalar lane repeats the rejected per-operation
  tag/reconciliation cost; and
- the shared raw-integer producer fusion was neutral across its primary rows
  and regressed controls, without removing recurring allocation.

The current profiles reproduce those same descendants. They do not show a new
consumer pattern, cheaper coherence rule, or allocation-free transport design
that would justify another implementation. No candidate advanced to timing,
so no guard cohort was needed.

## Verification and cleanup

- All three ordinary CLI executions verified.
- All three warmed CPU/allocation profile processes completed under the
  one-minute rule.
- The feature-coverage checker still reports 15 families, 16 normative
  sections, 35 portable applications, and three intentional local-only
  families.
- No source instrumentation or production code was added.
- Raw profiles, binaries, generated QuickSort input, and output captures were
  removed after this record was written.
- No WASM work was performed.

## Next recommendation

Run a cross-mode canonical-stdlib hot-function census, intersecting fully
qualified Able function owners from bytecode traces with generated compiled-Go
CPU profiles.

Why: the remaining runtime-helper profiles divide below broad dispatcher,
allocator, and boxing parents, while several successful recent improvements
came from general canonical-stdlib algorithms and benefited both execution
modes. A shared stdlib algorithm can advance the compiler and interpreter
targets together without adding runtime knowledge of HashMap, Array, regex, or
another non-primitive nominal type.

What it entails: collect bounded main-only bytecode call counts and existing
or fresh compiled CPU attribution across the selected applications; normalize
owners to Able package/function identity; and require one exact stdlib function
or algorithm family to be material in at least three unlike applications and
both execution modes. Build at most one source-level algorithm/data-flow
candidate in external `able-stdlib`, preserving public semantics and generic
nominal lowering. Measure repeated verifier-backed compiled and bytecode
cohorts plus unrelated guards. Stop at the census if owners remain fragmented,
and continue to defer WASM.
