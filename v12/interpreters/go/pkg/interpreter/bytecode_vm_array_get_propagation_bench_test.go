package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

var benchmarkBytecodeArrayGetPropagationBoolSink bool

func benchmarkBytecodeArrayGetPropagationVM(finalize bool) *bytecodeVM {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayIndexGetSlot},
		{op: bytecodeOpPropagation},
	}}
	if finalize {
		program = finalizeBytecodeProgramMetadata(program)
	}
	vm.currentProgram = program
	vm.ip = 0
	return vm
}

func BenchmarkBytecodeVMArrayGetSuccessPropagation(b *testing.B) {
	result := runtime.CharValue{Val: 'x'}

	b.Run("guarded_token", func(b *testing.B) {
		vm := benchmarkBytecodeArrayGetPropagationVM(true)
		if !vm.canSkipArrayGetSuccessPropagation(result, bytecodeIndexTypeChar, true) {
			b.Fatalf("warm guarded propagation skip = false, want true")
		}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchmarkBytecodeArrayGetPropagationBoolSink = vm.canSkipArrayGetSuccessPropagation(result, bytecodeIndexTypeChar, true)
		}
	})

	b.Run("exact_primitive_token", func(b *testing.B) {
		vm := benchmarkBytecodeArrayGetPropagationVM(true)
		if !vm.canSkipExactPrimitiveArrayGetSuccessPropagation(result, bytecodeIndexTypeChar, true) {
			b.Fatalf("warm exact propagation skip = false, want true")
		}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchmarkBytecodeArrayGetPropagationBoolSink = vm.canSkipExactPrimitiveArrayGetSuccessPropagation(result, bytecodeIndexTypeChar, true)
		}
	})
}

func BenchmarkBytecodeVMFollowingSuccessPropagationOpcode(b *testing.B) {
	b.Run("fallback_instruction_check", func(b *testing.B) {
		vm := benchmarkBytecodeArrayGetPropagationVM(false)
		if !vm.hasFollowingSuccessPropagationOpcode() {
			b.Fatalf("warm following propagation check = false, want true")
		}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchmarkBytecodeArrayGetPropagationBoolSink = vm.hasFollowingSuccessPropagationOpcode()
		}
	})

	b.Run("metadata_flag", func(b *testing.B) {
		vm := benchmarkBytecodeArrayGetPropagationVM(true)
		if !vm.hasFollowingSuccessPropagationOpcode() {
			b.Fatalf("warm following propagation check = false, want true")
		}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchmarkBytecodeArrayGetPropagationBoolSink = vm.hasFollowingSuccessPropagationOpcode()
		}
	})
}
