# Bytecode Float Operand-Lane Profile Gate

Date: 2026-07-16

## Outcome

Complete the retained-state profile refresh with no VM, compiler, stdlib,
fixture, or benchmark production change.

The same concrete interface-boxing descendants do recur across Distance Field,
RMS Norm, reduced NBody, and the unlike array-heavy
`matrixmultiply_f64_small` guard. This admits a primitive-float operand lane as
an architectural candidate. It does **not** admit another raw-value, pointer,
owned-cell, or slot-sidecar substitution: those representations have already
moved the box or regressed broad wall time.

The next implementation must replace the operand stack's value representation
behind one checked abstraction. A marker or sidecar added to the current
`[]runtime.Value` stack would be unsafe as a bounded patch because production
code currently contains 623 direct stack references across 66 files, including
347 indexed accesses, 118 appends, and 159 truncations. An unconverted consumer
could observe a marker instead of the Able float, or leave scalar metadata
attached to a reused stack index.

## Protocol

Each application ran in a fresh process with:

```text
GOMAXPROCS=1
GOGC=50
GOMEMLIMIT=1GiB
ABLE_BENCH_SKIP_TYPECHECK=1
```

Each scalar workload used one warmup and one measured iteration under a
55-second process guard. The smaller matrix guard used five measured iterations
to collect enough CPU samples. CPU and sampled allocation profiles were written
by the existing benchmark hooks; allocation sample rates were 8192 bytes for
Distance/RMS, 2048 for reduced NBody, and 4096 for the matrix guard. The exact
benchmark counters, rather than sampled profile totals, remain authoritative
for bytes and allocation counts.

## Retained measurements

| workload | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Distance Field | 6,189,164,535 | 512,083,808 | 38,000,201 |
| RMS Norm | 6,505,790,562 | 592,083,424 | 52,000,223 |
| reduced NBody | 1,927,287,792 | 97,604,360 | 6,161,118 |
| `matrixmultiply_f64_small` | 369,243,767 | 47,433,147 | 1,187,566 |

These are attribution runs on a shared workstation, not replacements for the
multi-process scorecard means. Their allocation shapes agree with the earlier
retained baselines; timing moved within the workstation's observed range.

## Repeated allocation descendants

Percentages below are each function's flat share of the sampled profile.

| workload | normalized raw float objects / space | materialized raw float objects / space | stack snapshot objects / space |
| --- | ---: | ---: | ---: |
| Distance Field | 40.94% / 33.43% | 22.17% / 27.16% | below the displayed tier |
| RMS Norm | 57.55% / 49.82% | 6.13% / 7.97% | 6.08% / 7.89% |
| reduced NBody | 32.14% / 20.96% | 26.77% / 26.18% | 30.44% / 29.77% |
| `matrixmultiply_f64_small` | 17.93% / 5.75% | 17.95% / 8.64% | 26.56% / 12.79% |

The exact repeated descendants are:

- `bytecodeNormalizedRawFloatSlotValue(...)`, where converting the raw scalar
  to `runtime.Value` boxes it;
- `bytecodeMaterializeRawFloatValue(...)`, where a VM-private float becomes a
  public runtime float at a boundary; and
- for three of the four workloads, `bytecodeStackSnapshotValue(...)`, where a
  mutable owned float must become a stable operand value.

The unlike matrix control remains independently useful: casts and Array growth
are larger owners there, but the same three float transport descendants still
account for 62.48% of sampled allocation objects. This is therefore not a
single scalar benchmark's source shape.

CPU profiles agree with the allocation attribution. RMS spends 13.89%
cumulative in `bytecodeNormalizedRawFloatSlotValue(...)`/`runtime.convT64`;
allocation and GC are material in all four workloads. Distance and NBody also
retain large ordinary-call subtrees, while the matrix control retains its
specialized dot-loop, cast, and Array work. A candidate must therefore improve
float transport without making stable call arguments or array-heavy programs
pay for the new representation.

## Representation decision

The profile gate passes, but only for a true non-interface operand
representation. The following narrower approaches remain closed:

- a slot-only float sidecar, because the next operand load boxes the value;
- reusable pointer stack cells, because calls need stable by-value snapshots
  and the pointers either alias reused storage or move boxing to the call;
- typed float frames layered over pointer carriers, because the previous broad
  gate increased Distance bytes 31.25% and regressed RMS time 12.65%, bytes
  67.56%, and allocations 3.85%; and
- owned-float or raw-carrier setter substitutions, which historical broad gates
  repeatedly rejected on wall time.

A parallel metadata slice grafted onto `[]runtime.Value` is also not a safe
incremental keep. Stack indexes are mutated directly throughout calls,
returns, rescue/ensure control flow, arrays, iterators, struct literals, and
specialized numeric opcodes. The scalar lane must instead make boxed versus raw
state explicit in the stack element type so stale or unhandled state cannot be
silently interpreted as a normal Able value.

## Next recommendation

Introduce the operand-stack representation in correctness-first stages before
turning on raw floats.

First define a VM-private tagged stack element and central checked operations
for append, indexed read/write, truncate, pop, snapshot, and materialization;
initially store every value in its boxed arm so behavior and performance remain
unchanged. Migrate the 66 production files to that abstraction in bounded,
testable groups and add debug invariants for call/return, rescue/ensure,
recursion, yielded execution, and stack-index reuse. Only after all direct
access is gone should the float arm carry `f32`/`f64` scalars by value, seed
eligible callee frames by value, and materialize at dynamic, native,
interface, aggregate, closure, and VM-exit boundaries.

Why: the repeated hotspot is now strong enough to justify the architectural
work, but the two previous broad failures show that partially preserving the
old interface stack merely relocates allocation. Staging the representation
change with a boxed-only parity phase separates correctness risk from the
performance experiment and gives the eventual float candidate a clean A/B
gate across all four applications. Continue to defer WASM.
