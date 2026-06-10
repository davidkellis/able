package interpreter

import "testing"

var benchmarkValidatedIntegerConstSlotsSink []bool

func BenchmarkBytecodeVMValidatedIntegerConstSlotsAlternating(b *testing.B) {
	vm := &bytecodeVM{}
	first := &bytecodeProgram{instructions: make([]bytecodeInstruction, 8)}
	second := &bytecodeProgram{instructions: make([]bytecodeInstruction, 12)}
	_ = vm.validatedIntegerConstSlots(first)
	_ = vm.validatedIntegerConstSlots(second)

	b.ReportAllocs()
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		benchmarkValidatedIntegerConstSlotsSink = vm.validatedIntegerConstSlots(first)
		benchmarkValidatedIntegerConstSlotsSink = vm.validatedIntegerConstSlots(second)
	}
}

func BenchmarkBytecodeVMValidatedIntegerConstSlotsNestedThree(b *testing.B) {
	vm := &bytecodeVM{}
	first := &bytecodeProgram{instructions: make([]bytecodeInstruction, 8)}
	second := &bytecodeProgram{instructions: make([]bytecodeInstruction, 12)}
	third := &bytecodeProgram{instructions: make([]bytecodeInstruction, 16)}
	_ = vm.validatedIntegerConstSlots(first)
	_ = vm.validatedIntegerConstSlots(second)
	_ = vm.validatedIntegerConstSlots(third)

	b.ReportAllocs()
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		benchmarkValidatedIntegerConstSlotsSink = vm.validatedIntegerConstSlots(first)
		benchmarkValidatedIntegerConstSlotsSink = vm.validatedIntegerConstSlots(second)
		benchmarkValidatedIntegerConstSlotsSink = vm.validatedIntegerConstSlots(third)
		benchmarkValidatedIntegerConstSlotsSink = vm.validatedIntegerConstSlots(second)
		benchmarkValidatedIntegerConstSlotsSink = vm.validatedIntegerConstSlots(first)
	}
}
