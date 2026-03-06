package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_SlotConstImmediateCacheBuildsAndRefreshes(t *testing.T) {
	vm := &bytecodeVM{}
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpBinaryIntSubSlotConst, value: runtime.NewSmallInt(3, runtime.IntegerI32)},
			{op: bytecodeOpCallSelfIntSubSlotConst, value: runtime.StringValue{Val: "oops"}},
		},
	}

	table := vm.slotConstImmediateTable(program)
	if table == nil {
		t.Fatalf("expected non-nil slot-const immediate table")
	}
	if table.instructionCount != len(program.instructions) {
		t.Fatalf("unexpected instruction count: got=%d want=%d", table.instructionCount, len(program.instructions))
	}
	if got, ok := bytecodeSlotConstImmediateAt(program.instructions[0], 0, table); !ok {
		t.Fatalf("expected first slot-const immediate to be cached")
	} else if asInt, intOK := got.ToInt64(); !intOK || asInt != 3 {
		t.Fatalf("expected cached immediate 3, got=%v ok=%v", got, ok)
	}
	if _, ok := bytecodeSlotConstImmediateAt(program.instructions[1], 1, table); ok {
		t.Fatalf("expected invalid immediate slot to remain uncached")
	}

	program.instructions = append(program.instructions, bytecodeInstruction{
		op:    bytecodeOpBinaryIntLessEqualSlotConst,
		value: runtime.NewSmallInt(9, runtime.IntegerI32),
	})

	table = vm.slotConstImmediateTable(program)
	if table == nil {
		t.Fatalf("expected non-nil refreshed slot-const immediate table")
	}
	if table.instructionCount != len(program.instructions) {
		t.Fatalf("expected refreshed instruction count=%d, got=%d", len(program.instructions), table.instructionCount)
	}
	if _, ok := bytecodeSlotConstImmediateAt(program.instructions[2], 2, table); !ok {
		t.Fatalf("expected appended slot-const immediate to be cached after refresh")
	}
}

func TestBytecodeVM_BoxedSmallIntValueCache(t *testing.T) {
	first, ok := bytecodeBoxedSmallIntValue(runtime.IntegerI32, 12)
	if !ok {
		t.Fatalf("expected boxed i32 small int")
	}
	second, ok := bytecodeBoxedSmallIntValue(runtime.IntegerI32, 12)
	if !ok {
		t.Fatalf("expected boxed i32 small int on second lookup")
	}
	if first != second {
		t.Fatalf("expected stable boxed cache identity for i32 small int")
	}
	if _, ok := bytecodeBoxedSmallIntValue(runtime.IntegerI32, bytecodeSmallIntBoxMax+1); ok {
		t.Fatalf("expected out-of-range i32 value to bypass boxed cache")
	}
	if _, ok := bytecodeBoxedSmallIntValue(runtime.IntegerU32, 12); ok {
		t.Fatalf("expected unsupported integer kind to bypass boxed cache")
	}
}

func TestBytecodeVM_BoxedIntegerValueDynamicCache(t *testing.T) {
	value := bytecodeSmallIntBoxMax + 512
	if _, ok := bytecodeBoxedSmallIntValue(runtime.IntegerI32, value); ok {
		t.Fatalf("expected value outside fixed small-int cache")
	}
	first, ok := bytecodeBoxedIntegerValue(runtime.IntegerI32, value)
	if !ok {
		t.Fatalf("expected dynamic boxed i32 value")
	}
	second, ok := bytecodeBoxedIntegerValue(runtime.IntegerI32, value)
	if !ok {
		t.Fatalf("expected cached dynamic boxed i32 value")
	}
	if first != second {
		t.Fatalf("expected stable boxed value for dynamic cache lookup")
	}
	allocs := testing.AllocsPerRun(1000, func() {
		if _, ok := bytecodeBoxedIntegerValue(runtime.IntegerI32, value); !ok {
			t.Fatalf("expected cached dynamic boxed i32 value")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected zero allocations for cached dynamic boxed value, got %.2f", allocs)
	}
	if _, ok := bytecodeBoxedIntegerValue(runtime.IntegerU32, 42); ok {
		t.Fatalf("expected unsupported integer kind to bypass dynamic boxed cache")
	}
}
