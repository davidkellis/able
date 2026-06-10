# Bytecode Boxed Operand-Stack Call/Control Migration

Date: 2026-07-16

## Outcome

Retain the fourth boxed-only operand-stack migration tranche. It closes the
remaining call/member-dispatch and control-restoration boundary without
changing the representation or claiming a runtime improvement.

The bounded inventory contained 92 direct stack-access sites across 13
production files. All are now behind the central API. A source audit covers
all call-family and static-member production files plus rescue, ensure,
or-else, loop, and run preparation.

## Retained implementation

The migrated paths include:

- direct, partial, cached, inline, and exact-native call dispatch;
- dynamic and resolved member calls, call-name caches, and static members;
- receiver materialization and safe-member nil completion; and
- rescue, ensure, or-else, loop-break, and fresh-run stack restoration.

Captured argument ranges use `stackValues` or `stackValuesFrom`, which return
the same backing-array views as the former slices. The following
`truncateStack` calls retain the prior non-clearing behavior, so borrowed or
captured arguments remain valid for exactly the same lifetime. Receiver
materialization uses `setStackValue`; result and control-flow pushes use
`appendStackValue`.

The representation remains:

```go
type bytecodeOperandStack []runtime.Value
```

No scalar tag, carrier, materialization rule, compiler/lowering branch, stdlib
change, fixture change, or benchmark-specific opcode was added. Direct access
now remains at 141 sites in 13 production files, including the central stack
implementation, versus 233 sites in 26 files before this tranche.

## Correctness gate

The call/control source audit joins the core, numeric/slot, and aggregate
audits. These groups pass:

```text
go test ./pkg/interpreter -run 'TestBytecodeOperandStack|TestBytecodeVM_.*(Call|Return|Rescue|Ensure|Raise|OrElse|Loop|ProgramSwitch|Inline|Static|Member)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'Test(BytecodeVM_(NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeBoundMethodBorrowArgsRemainStableAcrossNestedExactCalls|InlineCallArgSnapshotDoesNotAliasRawIntegerStackCell|PrepareResolvedFunctionCallArgsPreservesRawI32StackCell)|BytecodeExecExactNativeBoundCall.*|CallDispatch.*)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeVM' -count=1 -timeout 55s
```

The complete bounded bytecode VM family passed in 17.2 seconds. All touched
production files remain below 1,000 lines; the largest is
`bytecode_vm_call_member.go` at 829.

## Boxed-only parity gate

Each sample ran in a fresh process with `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, typechecking skipped, one measured iteration, and a
55-second process limit. RMS received three additional samples after one
process rose to 6.78 seconds. The spike and every other sample remain in the
reported mean.

| workload | process mean | CV | range |
| --- | ---: | ---: | ---: |
| Distance Field (3) | 6,155,339,951 ns | 2.07% | 6.053-6.298 s |
| RMS Norm (6) | 6,062,033,829 ns | 6.77% | 5.718-6.781 s |
| reduced NBody (3) | 1,686,751,462 ns | 4.27% | 1.621-1.764 s |
| `matrixmultiply_f64_small` (3) | 322,761,623 ns | 2.52% | 0.313-0.328 s |

Allocation shapes remain fixed at about 512.05 MB / 38.00M allocations for
Distance, 592.05 MB / 52.00M for RMS, 97.57 MB / 6.16M for reduced NBody, and
47.64 MB / 1.188M for matrix. Relative to the immediately preceding cohort,
wall times move by 0.9-3.7%, within the measured workstation CV and historical
adjacent-batch shifts. This is performance parity, not evidence of either a
speedup or regression.

Whole-binary disassembly finds no call to `stackDepth`, `stackValue`,
`stackValues`, `stackValuesFrom`, `appendStackValue`, `appendStackPair`,
`setStackValue`, `truncateStack`, or `clearStackFrom`. Go inlines the abstraction
back into the existing slice operations.

## Next recommendation

Complete the final boxed-boundary migration outside
`bytecode_vm_stack.go` next.

Why: the remaining stdlib runtime helpers, raw integer carriers, stack
diagnostics, and definition/import/implicit-slot paths can still bypass the
future representation boundary. A tagged scalar arm would be incomplete and
unsafe until all of them either use the API or are the central implementation.

What it entails: migrate the remaining 131 sites in 12 non-central production
files; add raw-cell reuse, stdlib struct/u128, diagnostic, definition/import,
placeholder, implicit-slot, and pool-reset guards; make the source audit allow
direct storage access only in `bytecode_vm_stack.go`; run the full bytecode VM
and targeted carrier/stdlib/error tests; and repeat the four-application
averaged parity gate. If that remains clean, the following tranche can finally
prototype a genuine tagged scalar stack under alternating A/B measurement.
Continue to defer WASM.
