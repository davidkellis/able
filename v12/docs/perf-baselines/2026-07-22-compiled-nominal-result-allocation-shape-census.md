# Compiled nominal/Result allocation-shape census — 2026-07-22

## Decision

Admit one general compiled candidate for a separate implementation tranche:
remove the escaping `runtime.NativeBoundMethodValue` box from the already
statically resolved generic-named-union fast path. Keep no compiler, generated
runtime, VM, canonical-stdlib, benchmark, language, or WASM change from this
report-only census.

The exact mechanism is now material in three unlike applications by allocation
objects and bytes, and in four unlike applications by cumulative CPU. This
crosses the project's admission bar without naming `Option`, `Result`, a user
record, or any stdlib container in the proposed lowering rule.

## Cohort and protocol

The cohort contains four source- and verifier-backed applications:

- Binary Event Log: binary decoding, checksums, nominal records, maps, and file
  input;
- Option/Result Config: numeric configuration resolution and chained generic
  union methods;
- Manifest Normalization: text splitting, nominal input/output records, arrays,
  captured normalization, and file input;
- Policy Record Dispatch: regex/NFA matching, nominal records and decisions,
  interfaces, maps, and file input.

Each application was compiled once, then its exact binary was preserved for
every diagnostic launch. All 74 launches passed the public verifier: 50 CPU
profile processes, four exact allocation-profile processes, and twenty
lightweight allocation-stat processes. Every process used `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and a 55-second timeout. No process approached
the timeout.

CPU profiles cover only registered `main`. Allocation shapes subtract each
exact `main-start` allocation profile from its `main-end` profile with
`runtime.MemProfileRate=1`. Five additional independent `main` MemStats deltas
per application provide ordinary exact totals without profile serialization.
All workstation observations are retained.

## CPU breadth

| Application | Processes | CPU samples | Static generic-union cumulative | `call_value_fast` cumulative | bound-value conversion | `mallocgc` cumulative |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 5 | 2.92 s | 34.93% | 28.08% | 3.08% | 53.42% |
| Option/Result Config | 15 | 1.95 s | 58.97% | 43.59% | 9.74% | 60.51% |
| Manifest Normalization | 15 | 1.91 s | 20.94% | 18.32% | 2.09% | 53.93% |
| Policy Record Dispatch | 15 | 1.86 s | 11.83% | 10.75% | 0.54% | 39.78% |

Caller/callee reports show the same path in every application:
`__able_static_generic_union_method_call` finds a compiled method entry, boxes
`runtime.NativeBoundMethodValue`, converts it to `runtime.Value`, and sends it
through `__able_call_value_fast`. No `bridge.CallStaticGenericUnionMember`
fallback receives a CPU or allocation sample.

## Exact allocation shape

The generated allocation is the same line in all four binaries:

```go
value, err := __able_call_value_fast(
    runtime.NativeBoundMethodValue{Receiver: obj, Method: *entry.fn}, args,
)
```

Go escape analysis proves that the composite is interface-converted and
escapes through the `runtime.Value` parameter. Line-level exact profiles show
one 64-byte allocation at that expression for every fast call.

| Application | Mean main allocations | Mean main bytes | escaping boxes | box bytes | objects | bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 4,850,462.8 | 290,399,603.2 | 225,280 | 14,417,920 | 4.64% | 4.96% |
| Option/Result Config | 1,630,658.4 | 57,447,832.0 | 147,456 | 9,437,184 | 9.04% | 16.43% |
| Manifest Normalization | 1,034,216.4 | 47,288,388.8 | 20,480 | 1,310,720 | 1.98% | 2.77% |
| Policy Record Dispatch | 970,530.0 | 49,166,376.0 | 5,632 | 360,448 | 0.58% | 0.73% |

The five allocation-count repetitions vary by at most two allocations in
Binary Event Log, one in Option/Result Config, six in Manifest Normalization,
and three in Policy Record Dispatch. Allocated-byte ranges are respectively
312, 144, 54,848, and 8,240 bytes. The observed box counts and 64-byte size are
therefore far above measurement noise in the first three applications.

Nominal record conversion remains a separate owner: EventRecord conversion is
55.82% cumulative CPU and 41.33% flat allocation bytes, while Manifest and
Policy nominal conversions have different record layouts and smaller shares.
This census does not admit a record-specific or general nominal-layout change.
It admits only the exact generic-union bound-method box shared above those
different consumers.

## Candidate contract

The next candidate should directly invoke the known generated native method
entry with the receiver and arguments instead of first materializing a
`NativeBoundMethodValue` as `runtime.Value`. It must remain a general rule for
all statically resolved methods on generic named unions.

The candidate must preserve:

- receiver injection as argument zero;
- the current runtime environment and `RuntimeData` in `NativeCallContext`;
- native error conversion and Able control propagation;
- nil-result normalization;
- the existing dynamic fallback when no compiled entry exists.

It must not add branches for `Option`, `Result`, EventRecord, ManifestRecord,
PolicyRecord, HashMap, or any other nominal/stdlib type. The current path does
not perform arity validation beyond the generated wrapper, so the direct path
must not silently introduce or remove validation either.

## Required implementation gate

Build the candidate in the generated runtime helper and add focused generated-
source plus executable semantic tests for successful generic-union methods,
error results, nil normalization, receiver order, and a forced fallback. Then
measure preserved current/candidate binaries with repeated averaged cohorts for
Binary Event Log, Option/Result Config, Manifest Normalization, and Policy
Record Dispatch. Guard with Binary Trees, N-Body, K-Nucleotide, Matrix
Multiply, and at least one concurrency application.

Retain the change only if allocation deltas prove that the exact 64-byte box is
removed, the three primary applications improve or remain within noise, and no
unrelated guard regresses. Do not begin WASM work.
