# Bytecode six-application clean CPU owner refresh

Date: 2026-07-22

## Decision

Retain no VM, compiler, tree-walker, canonical-stdlib, benchmark, fixture,
language, or WASM change from this tranche. Six fresh bounded runtime-only CPU
profiles expose no new exact non-dispatch VM leaf that is both CPU-material in
at least three unlike applications and still open for a general candidate.

The complete exact-symbol intersection contains only four interpreter symbols
at or above 1% flat CPU in at least three applications. `runResumable` is the
aggregate dispatch loop. The other three are the already-closed slot-to-stack,
raw-integer extraction, and snapshot/carrier families. Their current callers
also reproduce the earlier semantic split rather than exposing a new common
operation.

## Workload and profiling contract

The representatives deliberately span the six current architecture-budget
families:

| Family | Application | Reason selected |
| --- | --- | --- |
| Collection/Array | Array Slice Window | Slice construction, index work, and Array transport |
| Concurrency | Concurrent Event Routing | Futures, channels, executor work, records, and maps |
| Float numeric | Distance Field | Scalar float arithmetic, static math calls, and float slots |
| Wide numeric | Fixed Width 128 | Checked `UInt128`, casts, calls, and integer metadata |
| Text/byte Array | Reverse Complement | Mono-`u8` Array reads/pushes and byte output |
| Text/map | Word Frequency | String-key HashMap, named calls, returns, and type matching |

Every process used the same interpreter build ID
`3fc961f5f7986c29154bc4349c8bef659d60b78d`, normal typechecking, canonical
external `able-stdlib`, source-root-only loading, one warmup, one measured
`main()` call, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 55-second
cap. All six processes passed.

| Application | ns/op | Bytes/op | Objects/op | CPU samples |
| --- | ---: | ---: | ---: | ---: |
| Array Slice Window | 550,384,947 | 14,199,952 | 422,312 | 0.54 s |
| Concurrent Event Routing | 2,824,151,561 | 285,049,280 | 2,808,344 | 2.81 s |
| Distance Field | 5,873,810,729 | 368,086,688 | 26,000,184 | 5.86 s |
| Fixed Width 128 | 7,883,800,811 | 1,242,277,912 | 30,858,407 | 7.86 s |
| Reverse Complement | 3,195,202,773 | 213,618,088 | 3,543,007 | 3.18 s |
| Word Frequency | 1,205,092,608 | 48,520,960 | 637,342 | 1.20 s |

These are attribution runs, not promoted timing comparisons. A candidate
would still require repeated complete workstation samples and arithmetic
means on every side. No candidate reached that gate, so no A/B timing claim is
made from the single profiled calls.

## Complete exact flat intersection

Percentages are flat CPU shares in Array, Event Routing, Distance, Fixed
Width, Reverse Complement, and Word Frequency order. The admission rule is
the same exact non-dispatch interpreter leaf at or above 1% flat CPU in at
least three unlike applications, after excluding completed/rejected families.

| Exact symbol | Array | Event | Distance | Fixed | Reverse | Word | Rows >=1% | Disposition |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `(*bytecodeVM).runResumable` | 11.11% | 5.34% | 5.63% | 4.71% | 9.43% | 3.33% | 6 | Aggregate opcode dispatcher, not a semantic operation |
| `appendSlotStackValueChecked` | 1.85% | 0.36% | 1.19% | 1.27% | 2.83% | 2.50% | 5 | Closed stack-carrier family; two broad ordering/carrier trials failed guards |
| `bytecodeRawIntegerValueInfo` | 1.85% | 0.36% | 0% | 3.05% | 1.26% | 0.83% | 3 | Closed raw-integer family; callers and carriers split |
| `bytecodeStackSnapshotValue` | 1.85% | 0.71% | 2.05% | 0.13% | 0.63% | 1.67% | 3 | Closed snapshot family; raw integer alias severing and float/value copying differ |

No other exact interpreter-owned flat symbol clears 1% in three profiles.
For example, `popCallFrameFields` clears the line only in Distance and Word.
Its cumulative work is broader, but the already-rejected frame ABI and return
families cannot be reopened by cumulative-parent samples.

## Closed-family reconciliation

`appendSlotStackValueChecked` is cumulatively visible in five applications,
but its children are not one cost. Array and Distance spend their sampled work
in value snapshots; Fixed Width and Reverse Complement additionally enter
integer boxing and the tracked raw-value map; Word Frequency includes snapshot
and write-barrier work. The earlier immutable ordering and explicit raw-value
carrier experiments already produced mixed or regressive broad guards and
were removed.

`bytecodeRawIntegerValueInfo` reaches exactly three rows at 1%, but Array's
sample is entirely cast work, Fixed Width divides across checked wide-integer
members, arithmetic, casts, and calls, and Reverse Complement includes Array
push/index work. The full-switch, membership-switch, private-carrier, raw-store,
and producer alternatives have already failed broad gates. This profile does
not identify a new extractor contract or redundant step.

`bytecodeStackSnapshotValue` similarly crosses the numerical threshold in
three rows without sharing one carrier rule. Distance is dominated by float
slot loads and coercion/identity boundaries; Array and Word use ordinary
value/slot snapshots. Previous direct float/value and raw-carrier candidates
were correctness-incomplete or neutral-to-regressive, and the allocation-owner
census already closed the apparently shared symbol by source-line carrier.

The broad Go leaves are symptoms, not Able candidates. Map `matchH2` is
1.37%-6.92% flat in all six profiles, but its owners include integer metadata,
environment/type/member caches, and tracked Array/raw-value state. GC span
scanning is 2.50%-7.00% flat in five rows, while the semantic allocation owners
are not present in a CPU profile. Neither represents one removable VM
operation.

## Verification

No temporary instrumentation or candidate code was added. The unchanged
bytecode suite passes:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter
```

Raw profiles, logs, and complete pprof text are cleanup-eligible under
`/tmp/able-six-owner-refresh-20260722`.

## Next recommendation

Run a sampled allocation-owner matrix for these same six applications, then
return to compiler selection if it also finds no shared owner.

Why: this CPU refresh closes every currently visible shared VM leaf, but five
profiles spend material CPU in Go GC scanning and the six measured calls range
from 14 MB to 1.24 GB allocated. CPU ancestry cannot tell whether that pressure
comes from one removable Able allocation or from six different semantic
objects. Allocation-site evidence is the last responsible local VM gate before
concluding that another helper/carrier experiment cannot materially move the
current architecture budget.

What it entails: capture separate clean `alloc_objects` and `alloc_space`
profiles plus repeated exact measured-main allocation counters; remove package
initialization; intersect fully qualified allocation source lines and their
interpreter callers; and admit a candidate only when the same removable
semantic allocation is material in at least three unlike families. Required
user values, identity-bearing nominal objects, Array backing growth, and
bootstrap caches remain non-candidates. Any admitted rule must receive
repeated alternating A/B processes, retain every workstation sample in its
arithmetic means, and preserve zero-reach controls. Continue to defer WASM.
