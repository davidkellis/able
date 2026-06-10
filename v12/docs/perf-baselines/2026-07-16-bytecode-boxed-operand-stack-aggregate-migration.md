# Bytecode Boxed Operand-Stack Aggregate Migration

Date: 2026-07-16

## Outcome

Retain the third boxed-only operand-stack migration tranche. It closes the
aggregate/member/index/iterator producer-consumer boundary without changing the
stack representation or claiming a runtime improvement.

The bounded inventory contained 167 direct stack-access sites across 20
production files. All are now behind the central operand-stack API. A
filename-family source audit covers array, f64 aggregate, index, iterator,
match, member, propagation, string, and fast struct-literal production files.

## Retained implementation

The migrated paths include:

- Array lookup, mutation, swapping, and member fast paths;
- struct literal construction and match snapshots;
- cached and slot-based index reads;
- string, string-builder, byte/character iterator, and generic iterator paths;
- propagation results; and
- specialized f64 affine, dot, matrix-row, nested-get, and transpose kernels.

Depth checks, indexed reads/writes, truncations, and result pushes use
`stackDepth`, `stackValue`, `setStackValue`, `truncateStack`, and
`appendStackValue`. The representation remains:

```go
type bytecodeOperandStack []runtime.Value
```

No scalar tag, raw aggregate arm, carrier, materialization rule, compiler or
lowering branch, stdlib change, fixture change, or benchmark-specific opcode
was added. Direct stack access now remains in 26 production files, including
the central stack implementation, versus 46 before this tranche.

## Correctness gate

The aggregate source audit joins the existing core and numeric/slot audits.
The focused aggregate and call/error groups pass, followed by the entire
bounded bytecode VM test family:

```text
go test ./pkg/interpreter -run 'TestBytecodeOperandStack|TestBytecodeVM_.*(Array|Iterator|Member|Match|Struct|String|Index|F64|Float)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeOperandStack|TestBytecodeVM_.*(Call|Return|Error|Rescue|Propagation)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeVM' -count=1 -timeout 55s
```

The final command passed in 17.4 seconds. The largest touched production file,
`bytecode_vm_array_member_fast.go`, remains below the project limit at 971
lines.

## Boxed-only parity gate

Each measurement ran in a fresh process with `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, typechecking skipped, one measured iteration, and a
55-second process limit. All samples remain in the means.

| workload | process mean | CV | range |
| --- | ---: | ---: | ---: |
| Distance Field (3) | 6,077,049,314 ns | 1.12% | 6.017-6.151 s |
| RMS Norm (3) | 5,957,935,543 ns | 1.71% | 5.893-6.075 s |
| reduced NBody (3) | 1,670,873,791 ns | 3.06% | 1.627-1.727 s |
| `matrixmultiply_f64_small` (3) | 311,261,281 ns | 3.39% | 0.302-0.323 s |

Allocation shapes remain fixed at about 512.05 MB / 38.00M allocations for
Distance, 592.05 MB / 52.00M for RMS, 97.57 MB / 6.16M for reduced NBody, and
47.64 MB / 1.188M for matrix. Wall times are lower than the preceding retained
cohort, but that cohort already demonstrated large shifts between adjacent,
unchanged NBody and matrix batches. Because this tranche only routes boxed
slice operations through inlineable helpers, the result is performance parity,
not evidence of a speedup.

Whole-binary disassembly finds no call to `stackDepth`, `stackValue`,
`stackValues`, `stackValuesFrom`, `appendStackValue`, `appendStackPair`,
`setStackValue`, `truncateStack`, or `clearStackFrom`. Go inlines the abstraction
back into the existing slice operations.

## Next recommendation

Migrate the remaining call/member-dispatch and control-restoration group next,
still without activating a scalar representation.

Why: captured call arguments, receiver replacement, saved stack depths, and
rescue/ensure/or-else restoration are the remaining ownership-sensitive paths.
A future scalar lane cannot be sound while those paths can bypass the central
representation boundary.

What it entails: migrate call dispatch/member/name/native/static-member files
and rescue, ensure, or-else, loop, and run-prepare restoration; preserve the
current argument backing-slice lifetime and non-clearing truncation semantics;
extend source audits; run call/return/error/control tests; and repeat the four
application averaged parity gate. Stdlib-struct helpers, raw integer carriers,
definitions/imports/placeholders, pool/diagnostics, and the central stack files
remain the final boxed groups. Continue to defer WASM.
