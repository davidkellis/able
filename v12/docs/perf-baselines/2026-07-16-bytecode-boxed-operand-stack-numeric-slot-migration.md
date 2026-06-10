# Bytecode Boxed Operand-Stack Numeric/Slot Migration

Date: 2026-07-16

## Outcome

Retain the second boxed-only operand-stack migration tranche. It migrates the
specialized numeric and slot producer/consumer family without activating a
scalar tag, changing Able behavior, or claiming a speed improvement.

The bounded inventory contained 107 direct stack-access sites across 13
production files. All are now behind the checked operand-stack operations. A
family-wide source audit covers assignment, cast, float, i32, int/integer,
jump, literal, and slot production files so later specialized opcodes cannot
silently bypass the migration boundary.

## Retained implementation

The migrated paths include:

- raw integer cast results and reusable i32/i64 stack values;
- fused float add/sub/multiply, array-read, affine, and store results;
- slot loads, typed/untyped stores, implicit/slotless call arguments, and
  discarded-result truncation;
- integer affine updates and slot-constant results;
- assignment and compound-assignment results; and
- interpolation, string concatenation, and Array literal operands.

`clearStackFrom(...)` joins the boxed stack API for literal paths that must
release backing-slice references before truncation. Calls deliberately retain
their existing non-clearing truncation semantics because a captured argument
slice may still use the old backing region until the call completes.

The representation remains:

```go
type bytecodeOperandStack []runtime.Value
```

No raw float arm, marker, pointer carrier, materialization rule, compiler or
lowering branch, stdlib change, fixture change, or benchmark-specific opcode
was added.

After this tranche, direct stack access remains in 46 production files. Those
files are aggregate/member/iterator paths, control restoration, remaining call
specializations, pool/diagnostic support, and the central stack/raw-carrier
implementations. A scalar arm remains forbidden until those groups are also
migrated.

## Correctness gate

The operand-stack invariant test now covers suffix reads and explicit clearing
in addition to boxed/raw dynamic-type preservation, indexed reads/writes,
append, truncate, and pop. The source audit expands from the core files to all
numeric/slot filename families.

These bounded groups pass:

```text
go test ./pkg/interpreter -run 'TestBytecodeOperandStack|TestBytecodeVM_.*(Binary|Float|Integer|I32|Slot|Cast|Literal|Assign|Condition|Jump|Call|Return)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeVM_.*(Array|Iterator|Member|Match|Struct|String|Interpolation)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeVM_.*(Rescue|Ensure|Raise|Yield|ProgramSwitch|Inline|CallFrame|CallName)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestInlineCoercion|TestBytecode.*Float|TestBytecode.*Integer|TestBytecode.*Literal' -count=1 -timeout 55s
```

## Boxed-only parity gate

Each process used one warmup and one measured iteration with `GOMAXPROCS=1`,
`GOGC=50`, `GOMEMLIMIT=1GiB`, typechecking skipped, and a 55-second process
limit. Distance completed three processes. RMS, reduced NBody, and matrix each
received six because their first cohorts became volatile.

| workload | process mean | CV | range |
| --- | ---: | ---: | ---: |
| Distance Field (3) | 6,237,366,121 ns | 3.23% | 6.072-6.462 s |
| RMS Norm (6) | 6,507,070,414 ns | 10.07% | 5.618-7.195 s |
| reduced NBody (6) | 2,027,389,276 ns | 22.27% | 1.709-2.709 s |
| `matrixmultiply_f64_small` (6) | 355,685,130 ns | 15.80% | 0.303-0.440 s |

The extra repetitions show a workstation-load shift, not a candidate-specific
allocation change. NBody's first three averaged 2.315 seconds and its next
three averaged 1.740 seconds without a source or binary change. Matrix likewise
moved from 0.402 to 0.309 seconds between adjacent batches. RMS moved in the
opposite time direction while retaining the same allocation counts. All samples
remain in the reported combined means; none were discarded.

Allocation shapes are unchanged: Distance remains about 512.05 MB / 38.00M
allocations, RMS 592.05 MB / 52.00M, reduced NBody 97.57 MB / 6.16M, and matrix
47.64 MB / 1.188M. Whole-binary disassembly finds no call to `stackDepth`,
`stackValue`, `stackValues`, `stackValuesFrom`, `appendStackValue`,
`appendStackPair`, `setStackValue`, `truncateStack`, or `clearStackFrom`.
The compiler inlines the boxed helpers back into slice operations.

This evidence supports semantic/layout parity and retention of the migration;
it does not establish a runtime improvement. The eventual scalar arm still
requires a clean alternating A/B gate.

## Next recommendation

Migrate the aggregate/member/index/iterator producer-consumer group next while
keeping all elements boxed.

Why: arrays, structs, matches, indexes, members, strings, iterators, and the
specialized f64 aggregate kernels are where a future scalar must either remain
raw in a proven primitive container path or materialize at an aggregate or
dynamic boundary. Leaving these paths on direct slice access would make the
eventual tag incomplete and unsafe.

What it entails: move array/member/index swap and access paths, struct literal
and match paths, iterator and string fast paths, propagation helpers, and f64
row/dot/transpose kernels behind the same API; extend the filename-family
source audit; add aggregate materialization, alias, mutation, match, error,
iterator-end, and stack-index-reuse guards; repeat bounded semantic groups and
the averaged four-application parity gate. Rescue/ensure, pool/diagnostics,
imports/definitions, remaining call specializations, and central raw carriers
remain later boxed migration groups. Continue to defer WASM.
