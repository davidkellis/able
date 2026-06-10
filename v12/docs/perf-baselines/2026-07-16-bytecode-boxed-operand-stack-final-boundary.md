# Bytecode Boxed Operand-Stack Final Boundary

Date: 2026-07-16

## Outcome

Retain the fifth and final boxed-only operand-stack migration tranche. Every
production bytecode VM path outside the central implementation now uses the
operand-stack API. The representation remains boxed and this tranche does not
claim a speed improvement.

The bounded inventory contained 131 direct storage sites across 12 production
files. A repository-wide audit now scans every production
`bytecode_vm_*.go` file and permits direct access only in
`bytecode_vm_stack.go`. Its 11 central accesses are the complete direct
production surface.

## Retained implementation

The final migrated paths include:

- reusable raw i32, i64, and general integer stack cells;
- canonical stdlib struct, random, Int128, and UInt128 fast paths;
- stack depth, capacity, call-operand, peak, and loop-balance diagnostics;
- definition, import, implicit-slot, placeholder, and duplicate results; and
- pool reset clearing and truncation.

`stackCapacity` joins the central API. Pool reset still explicitly clears all
references before truncating. Raw integer replacement still derives the
reusable cell from the destination index, writes through `setStackValue`, and
then truncates. No stdlib source was changed; the migrated stdlib-named files
are existing generic VM runtime fast paths.

The representation remains:

```go
type bytecodeOperandStack []runtime.Value
```

No scalar tag, carrier, compiler/lowering branch, fixture change, external
stdlib change, or benchmark-specific opcode was added.

## Correctness gate

The repository-wide storage audit and targeted boundaries pass:

```text
go test ./pkg/interpreter -run 'TestBytecodeOperandStack|TestBytecodeVM_.*(Raw|Integer|I32|I64|UInt128|Int128|Struct|Definition|Import|Implicit|Placeholder|Pool|Stack|Diagnostic)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'Test(BytecodeVM_.*(Raw|Reusable)|BytecodeRawInteger|BytecodeVM(StackDiagnostics|Loop|While)|BytecodeVM_Canonical|ApplyBinaryOperatorFast_UsesRawIntegerCarriers|Placeholder|Pipe|Import)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeVM' -count=1 -timeout 55s
```

The full bounded bytecode VM family passed in 17.1 seconds. All touched files
remain below 1,000 lines; the largest is
`bytecode_vm_stdlib_struct_fast.go` at 694 lines.

## Boxed-only parity gate

Each sample ran in a fresh process with `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, typechecking skipped, one measured iteration, and a
55-second process limit. Distance received three additional samples after the
initial cohort spanned 5.96-6.73 seconds. Every sample remains in its mean.

| workload | process mean | CV | range |
| --- | ---: | ---: | ---: |
| Distance Field (6) | 6,332,085,025 ns | 4.40% | 5.959-6.728 s |
| RMS Norm (3) | 5,903,895,620 ns | 5.42% | 5.535-6.106 s |
| reduced NBody (3) | 1,607,069,107 ns | 2.27% | 1.584-1.649 s |
| `matrixmultiply_f64_small` (3) | 321,769,793 ns | 2.35% | 0.317-0.330 s |

Distance and matrix allocation shapes remain about 512.05 MB / 38.00M and
47.64 MB / 1.188M. RMS remains 592.05 MB / 52.00M. NBody and matrix each have
one fresh-process sample with one additional allocation and 48 additional
bytes; the other samples retain the preceding exact counts. Wall-time movement
against the immediately preceding cohort is mixed, from -4.7% to +2.9%, and
falls within the measured CV and historical workstation shifts. This is
performance parity.

Whole-binary disassembly finds no call to `stackDepth`, `stackCapacity`,
`stackValue`, `stackValues`, `stackValuesFrom`, `appendStackValue`,
`appendStackPair`, `setStackValue`, `truncateStack`, or `clearStackFrom`. Go
inlines the boundary back into the existing slice operations.

## Next recommendation

Prototype the first true operand-stack f64 lane under an alternating A/B gate.

Why: repeated profiles across four structurally different numeric programs
attribute major allocation and CPU shares to interface boxing, raw-float
normalization/materialization, and snapshots. The new invariant removes the
former stale-tag risk: production consumers can no longer bypass the central
stack representation.

What it entails: evolve `bytecodeOperandStack` into an owner of boxed storage
and explicit f64 value/kind/tag state while preserving zero-copy boxed argument
ranges. Add raw f64 append/read/write/snapshot operations; clear tags on every
write, truncate, and reused index; materialize only at dynamic, native,
interface, aggregate, closure, snapshot, and VM-exit boundaries; and activate
the lane only for existing type-proven float paths. Add call/return,
rescue/ensure, recursion, yield, borrowed-argument, and index-reuse guards.

Measure alternating boxed and tagged binaries across Distance, RMS, reduced
NBody, and matrix, plus non-float split/join, iterator collect, and numeric
array/map guards. Retain the lane only if it produces a broad allocation
reduction without material guard regressions. Continue to defer WASM.
