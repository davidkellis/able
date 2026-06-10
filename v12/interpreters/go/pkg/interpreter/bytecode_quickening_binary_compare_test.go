package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeQuickenBinaryCompareJumpRetainsFallbackInstruction(t *testing.T) {
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpBinary, operator: ">="},
		{op: bytecodeOpJumpIfFalse, target: 4},
		{op: bytecodeOpReturn},
	}})
	if got := program.instructions[0]; got.op != bytecodeOpJumpIfBinaryCompareFalse || got.target != 4 {
		t.Fatalf("quickened instruction = %#v", got)
	}
	if got := program.instructions[1].op; got != bytecodeOpJumpIfFalse {
		t.Fatalf("fallback instruction = %v, want JumpIfFalse", got)
	}
	nonComparison := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpBinary, operator: "+"},
		{op: bytecodeOpJumpIfFalse, target: 3},
	}})
	if got := nonComparison.instructions[0].op; got != bytecodeOpBinary {
		t.Fatalf("non-comparison instruction = %v, want Binary", got)
	}
}

func TestBytecodeVM_QuickenedBinaryCompareJumpDirectAndFallback(t *testing.T) {
	tests := []struct {
		name  string
		left  runtime.Value
		right runtime.Value
		want  string
	}{
		{name: "direct integer true", left: runtime.NewSmallInt(2, runtime.IntegerI32), right: runtime.NewSmallInt(1, runtime.IntegerI32), want: "true"},
		{name: "direct integer false", left: runtime.NewSmallInt(1, runtime.IntegerI32), right: runtime.NewSmallInt(2, runtime.IntegerI32), want: "false"},
		{name: "direct float true", left: runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64}, right: runtime.FloatValue{Val: 1, TypeSuffix: runtime.FloatF64}, want: "true"},
		{name: "generic string fallback", left: runtime.StringValue{Val: "a"}, right: runtime.StringValue{Val: "b"}, want: "false"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			interp := NewBytecode()
			vm := newBytecodeVM(interp, interp.GlobalEnvironment())
			program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
				{op: bytecodeOpConst, value: test.left},
				{op: bytecodeOpConst, value: test.right},
				{op: bytecodeOpBinary, operator: ">"},
				{op: bytecodeOpJumpIfFalse, target: 6},
				{op: bytecodeOpConst, value: runtime.StringValue{Val: "true"}},
				{op: bytecodeOpReturn},
				{op: bytecodeOpConst, value: runtime.StringValue{Val: "false"}},
				{op: bytecodeOpReturn},
			}})
			result, err := vm.run(program)
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			if got, ok := result.(runtime.StringValue); !ok || got.Val != test.want {
				t.Fatalf("result = %#v, want %q", result, test.want)
			}
		})
	}
}
