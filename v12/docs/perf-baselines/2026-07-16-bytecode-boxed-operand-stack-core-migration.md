# Bytecode Boxed Operand-Stack Core Migration

Date: 2026-07-16

## Outcome

Retain the first correctness-first operand-stack migration tranche. It changes
no Able semantics and does not enable a raw float lane yet.

The VM stack now has a private named representation,
`bytecodeOperandStack`, whose element is still `runtime.Value`. Central helpers
own depth, indexed read/write, ranges, append, pair append, truncate, and pop.
The main run loop, generic/specialized binary dispatch, control-flow spawn,
inline call setup, call-frame depth, and return paths no longer access the
stack field directly.

This creates an enforceable migration boundary without combining it with the
later representation experiment. A source-audit regression test locks the six
migrated production files against reintroducing direct indexing, append,
length, or assignment.

## Retained implementation

The boxed-only representation is deliberately layout-equivalent to the old
slice:

```go
type bytecodeOperandStack []runtime.Value
```

The helper surface is:

- `stackDepth()`;
- `stackValue(...)`;
- `stackValues(...)` and `stackValuesFrom(...)`;
- `appendStackValue(...)` and `appendStackPair(...)`;
- `setStackValue(...)`;
- `truncateStack(...)`; and
- the existing `pop()` and replacement helpers, now implemented through that
  surface.

This tranche migrates:

- `bytecode_vm_run.go`;
- `bytecode_vm_ops.go`;
- `bytecode_vm_controlflow.go`;
- `bytecode_vm_calls.go`;
- `bytecode_vm_call_frames.go`; and
- `bytecode_vm_return.go`.

The generic struct-literal fallback moved into the existing struct-literal VM
module so `bytecode_vm_run.go` remains below 1,000 lines. It preserves the same
evaluation, nil normalization, Array ownership tracking, append, and instruction
advance sequence.

No marker, float sidecar, pointer carrier, boxed-value conversion, compiler
rule, lowering branch, stdlib branch, fixture change, or benchmark-specific
opcode was added.

## Correctness gate

The new focused invariant test proves that boxed and private raw-float values
retain their exact dynamic type through append, range read, replacement,
truncate, and pop. The source audit rejects direct stack access in the migrated
core files.

The following bounded groups pass:

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_.*(Call|Return|Rescue|Ensure|Yield|ProgramSwitch)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeVM_.*(Array|Iterator|Member|Match)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeVM_.*(Binary|Float|Integer|I32|Slot)' -count=1 -timeout 55s
go test ./pkg/interpreter -run 'TestBytecodeOperandStack|TestBytecodeVM_.*(Call|Return|Rescue|Ensure|Yield|ProgramSwitch|Binary|Float|Slot|StructLiteral)' -count=1 -timeout 55s
```

## Boxed-only parity gate

Each workload ran in three fresh processes after one warmup, with
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, typechecking skipped, and a
55-second process limit.

| workload | three-process mean | CV | representative B/op | representative allocs/op |
| --- | ---: | ---: | ---: | ---: |
| Distance Field | 6,364,372,978 ns | 2.58% | 512,052,144 | 38,000,142 |
| RMS Norm | 5,823,091,686 ns | 1.93% | 592,051,589 | 52,000,163 |
| reduced NBody | 1,703,307,209 ns | 1.98% | 97,572,491 | 6,161,057 |
| `matrixmultiply_f64_small` | 296,863,453 ns | 5.43% | 47,636,112 | 1,187,689 |

Allocation counts retain the clean boxed stack's deterministic shape. Timing
is mixed against the earlier attribution processes: Distance is above its old
single-run band, while RMS, NBody, and matrix are at or below their prior
samples. This is parity evidence, not a performance claim.

The Go compiler reports every new accessor as inlineable. Disassembly of
`runResumable`, `execBinary`, `execCall`, and `finishInlineReturn` contains no
call to an operand-stack accessor. The retained core therefore pays no helper
dispatch in those hot paths; the compiler emits the underlying slice
operations directly.

## Next recommendation

Migrate the specialized numeric and slot producer/consumer group next, while
keeping the representation boxed.

Why: those files contain the direct float loads, raw arithmetic results, fused
stores, casts, conditions, and slot-to-stack snapshots that must eventually
read and write a scalar arm. Migrating them now exposes every numeric boundary
to the same checked stack API without yet changing ownership or allocation.

What it entails: move float/integer specialized opcodes, slot operands/stores,
literals, assignment, and jump helpers off direct indexing/appends/truncation;
extend the source audit; add mixed f32/f64, raw integer, condition, cast,
discarded-result, and stack-index-reuse guards; repeat the focused semantic
groups and the same four-application three-process parity gate. Aggregate,
member, iterator, rescue/ensure, and remaining call-specialization files stay
boxed and become later migration groups. Activate no scalar tag until every
production direct access is gone. Continue to defer WASM.
