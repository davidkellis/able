package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_SlotConstImmediateCacheBuildsAndRefreshes(t *testing.T) {
	vm := &bytecodeVM{}
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{
				op:              bytecodeOpBinaryIntSubSlotConst,
				value:           runtime.StringValue{Val: "ignore-me"},
				intImmediate:    runtime.NewSmallInt(3, runtime.IntegerI32),
				hasIntImmediate: true,
			},
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
		op:              bytecodeOpBinaryIntLessEqualSlotConst,
		value:           runtime.StringValue{Val: "ignore-me-too"},
		intImmediate:    runtime.NewSmallInt(9, runtime.IntegerI32),
		hasIntImmediate: true,
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
	unsignedFirst, ok := bytecodeBoxedSmallIntValue(runtime.IntegerU32, 12)
	if !ok {
		t.Fatalf("expected boxed u32 small int")
	}
	unsignedSecond, ok := bytecodeBoxedSmallIntValue(runtime.IntegerU32, 12)
	if !ok {
		t.Fatalf("expected boxed u32 small int on second lookup")
	}
	if unsignedFirst != unsignedSecond {
		t.Fatalf("expected stable boxed cache identity for u32 small int")
	}
	if _, ok := bytecodeBoxedSmallIntValue(runtime.IntegerU32, -1); ok {
		t.Fatalf("expected negative u32 value to bypass boxed cache")
	}
}

func TestBytecodeVM_BoxedIntegerValueDynamicCache(t *testing.T) {
	value := int64(200000)
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
	firstUnsigned, ok := bytecodeBoxedIntegerValue(runtime.IntegerU32, value)
	if !ok {
		t.Fatalf("expected dynamic boxed u32 value")
	}
	secondUnsigned, ok := bytecodeBoxedIntegerValue(runtime.IntegerU32, value)
	if !ok {
		t.Fatalf("expected cached dynamic boxed u32 value")
	}
	if firstUnsigned != secondUnsigned {
		t.Fatalf("expected stable boxed value for dynamic u32 cache lookup")
	}
	if _, ok := bytecodeBoxedIntegerValue(runtime.IntegerU32, -1); ok {
		t.Fatalf("expected negative u32 value to bypass dynamic boxed cache")
	}
}

func TestBytecodeVM_BoxedIntegerI32ValueDynamicCache(t *testing.T) {
	value := int64(200000)
	first := bytecodeBoxedIntegerI32Value(value)
	second := bytecodeBoxedIntegerI32Value(value)
	if first != second {
		t.Fatalf("expected stable boxed value for direct i32 dynamic cache lookup")
	}
	allocs := testing.AllocsPerRun(1000, func() {
		if got := bytecodeBoxedIntegerI32Value(value); got != first {
			t.Fatalf("expected cached direct i32 boxed value")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected zero allocations for cached direct i32 boxed value, got %.2f", allocs)
	}
}
