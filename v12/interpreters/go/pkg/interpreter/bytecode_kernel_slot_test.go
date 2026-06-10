package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestKernelArrayWithCapacityUsesSlotLocalParameter(t *testing.T) {
	interp := NewBytecode()
	loadKernelModule(t, interp)

	bucket, ok := interp.packageRegistry["able.kernel"]
	if !ok {
		t.Fatalf("able.kernel package registry missing")
	}
	raw, ok := bucket["Array.with_capacity"]
	if !ok {
		t.Fatalf("Array.with_capacity missing from kernel package registry")
	}
	fn := firstFunction(raw)
	if fn == nil {
		t.Fatalf("Array.with_capacity is not a function: %T", raw)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("Array.with_capacity bytecode = %T, want *bytecodeProgram", fn.Bytecode)
	}
	if program.frameLayout == nil {
		t.Fatalf("Array.with_capacity should use slot frame")
	}
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpLoadName && instr.name == "capacity" {
			t.Fatalf("Array.with_capacity emitted LoadName for parameter %q", instr.name)
		}
	}
	if program.frameLayout.paramSlots != 1 {
		t.Fatalf("Array.with_capacity paramSlots = %d, want 1", program.frameLayout.paramSlots)
	}
	if got := program.frameLayout.paramKinds[0]; got != bytecodeCellKindI32 {
		t.Fatalf("Array.with_capacity parameter slot kind = %v, want i32", got)
	}
	if _, ok := raw.(*runtime.FunctionValue); !ok {
		t.Fatalf("Array.with_capacity registry value = %T, want *runtime.FunctionValue", raw)
	}
	value, err := interp.CallFunction(fn, []runtime.Value{runtime.NewSmallInt(2, runtime.IntegerI32)})
	if err != nil {
		t.Fatalf("Array.with_capacity call failed: %v", err)
	}
	if _, ok := value.(*runtime.ArrayValue); !ok {
		t.Fatalf("Array.with_capacity result = %T, want *runtime.ArrayValue", value)
	}
}
